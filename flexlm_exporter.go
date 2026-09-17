// Copyright 2017 Mario Trangoni
// Copyright 2015 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"os/user"
	"runtime"
	"sort"
	"syscall"
	"time"

	//nolint:gosec
	_ "net/http/pprof"

	"github.com/prometheus/common/promslog"
	"github.com/prometheus/common/promslog/flag"

	kingpin "github.com/alecthomas/kingpin/v2"
	"github.com/mjtrangoni/flexlm_exporter/collector"
	"github.com/mjtrangoni/flexlm_exporter/config"
	"github.com/prometheus/client_golang/prometheus"
	promcollectors "github.com/prometheus/client_golang/prometheus/collectors"
	promcollectorsversion "github.com/prometheus/client_golang/prometheus/collectors/version"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/version"
	"github.com/prometheus/exporter-toolkit/web"
	"github.com/prometheus/exporter-toolkit/web/kingpinflag"
)

const (
	serverReadHeaderTimeout = 60
)

// handler wraps an unfiltered http.Handler but uses a filtered handler,
// created on the fly, if filtering is requested. Create instances with
// newHandler.
type handler struct {
	unfilteredHandler       http.Handler
	exporterMetricsRegistry *prometheus.Registry
	includeExporterMetrics  bool
	configPath              string
	maxRequests             int
	logger                  *slog.Logger
	flexlmCollector         *collector.FlexlmCollector
}

func newHandler(includeExporterMetrics bool, configPath string, maxRequests int, logger *slog.Logger) *handler {
	h := &handler{
		exporterMetricsRegistry: prometheus.NewRegistry(),
		includeExporterMetrics:  includeExporterMetrics,
		configPath:              configPath,
		maxRequests:             maxRequests,
		logger:                  logger,
	}
	if h.includeExporterMetrics {
		h.exporterMetricsRegistry.MustRegister(
			promcollectors.NewProcessCollector(promcollectors.ProcessCollectorOpts{}),
			promcollectors.NewGoCollector(),
		)
	}

	innerHandler, err := h.innerHandler()
	if err != nil {
		panic(fmt.Sprintf("Couldn't create metrics handler: %s", err))
	}

	h.unfilteredHandler = innerHandler

	return h
}

// ServeHTTP implements http.Handler.
func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	filters := r.URL.Query()["collect[]"]
	h.logger.Debug("collect query:", "filters", filters)

	if len(filters) == 0 {
		// No filters, use the prepared unfiltered handler.
		h.unfilteredHandler.ServeHTTP(w, r)
		return
	}
	// To serve filtered metrics, we create a filtering handler on the fly.
	filteredHandler, err := h.innerHandler(filters...)
	if err != nil {
		h.logger.Warn("Couldn't create filtered metrics handler:", "err", err)
		w.WriteHeader(http.StatusBadRequest)
		_, _ = fmt.Fprintf(w, "Couldn't create filtered metrics handler: %s", err)

		return
	}

	filteredHandler.ServeHTTP(w, r)
}

// innerHandler is used to create both the one unfiltered http.Handler to be
// wrapped by the outer handler and also the filtered handlers created on the
// fly. The former is accomplished by calling innerHandler without any arguments
// (in which case it will log all the collectors enabled via command-line
// flags).
func (h *handler) innerHandler(filters ...string) (http.Handler, error) {
	nc, err := collector.NewFlexlmCollector(h.logger, filters...)
	if err != nil {
		return nil, fmt.Errorf("couldn't create collector: %w", err)
	}

	// Save the reference to the main collector for a subsequent graceful shutdown.
	if len(filters) == 0 {
		h.flexlmCollector = nc
		h.logger.Info("Enabled collectors")

		collectors := []string{}

		for n := range nc.Collectors {
			collectors = append(collectors, n)
		}

		sort.Strings(collectors)

		for _, c := range collectors {
			h.logger.Info(c)
		}
	}

	// Load LicenseConfig from a YAML file.
	collector.LicenseConfig, err = config.Load(h.configPath, h.logger)
	if err != nil {
		h.logger.Error("couldn't load config file", "path", h.configPath, "err", err)
		if len(filters) == 0 {
			return nil, err
		}
	}

	r := prometheus.NewRegistry()
	r.MustRegister(promcollectorsversion.NewCollector("flexlm_exporter"))

	if err := r.Register(nc); err != nil {
		return nil, fmt.Errorf("couldn't register node collector: %w", err)
	}

	handler := promhttp.HandlerFor(
		prometheus.Gatherers{h.exporterMetricsRegistry, r},
		promhttp.HandlerOpts{
			ErrorLog:            slog.NewLogLogger(h.logger.Handler(), slog.LevelError),
			ErrorHandling:       promhttp.ContinueOnError,
			MaxRequestsInFlight: h.maxRequests,
			Registry:            h.exporterMetricsRegistry,
		},
	)

	if h.includeExporterMetrics {
		// Note that we have to use h.exporterMetricsRegistry here to
		// use the same promhttp metrics for all expositions.
		handler = promhttp.InstrumentMetricHandler(
			h.exporterMetricsRegistry, handler,
		)
	}

	return handler, nil
}

func main() {
	var (
		metricsPath = kingpin.Flag(
			"web.telemetry-path",
			"Path under which to expose metrics.",
		).Default("/metrics").String()
		configPath  = kingpin.Flag("path.config", "Configuration YAML file path.").Default("licenses.yml").String()
		maxRequests = kingpin.Flag(
			"web.max-requests",
			"Maximum number of parallel scrape requests. Use 0 to disable.",
		).Default("40").Int()
		disableExporterMetrics = kingpin.Flag(
			"web.disable-exporter-metrics",
			"Exclude metrics about the exporter itself (promhttp_*, process_*, go_*).",
		).Bool()
		maxProcs = kingpin.Flag(
			"runtime.gomaxprocs", "The target number of CPUs Go will run on (GOMAXPROCS)",
		).Envar("GOMAXPROCS").Default("1").Int()
		toolkitFlags = kingpinflag.AddFlags(kingpin.CommandLine, ":9319")
	)

	promslogConfig := &promslog.Config{}
	flag.AddFlags(kingpin.CommandLine, promslogConfig)
	kingpin.Version(version.Print("flexlm_exporter"))
	kingpin.CommandLine.UsageWriter(os.Stdout)
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	logger := promslog.New(promslogConfig)

	logger.Info("Starting flexlm_exporter", "version", version.Info())
	logger.Info("Build context", "build_context", version.BuildContext())

	if userCurrent, err := user.Current(); err == nil && userCurrent.Uid == "0" {
		logger.Warn(`FLEXlm Exporter is running as root user. ` +
			`This exporter is designed to run as unprivileged user, root is not required.`)
	}

	runtime.GOMAXPROCS(*maxProcs)
	logger.Debug("Go MAXPROCS", "procs", runtime.GOMAXPROCS(0))

	h := newHandler(!*disableExporterMetrics, *configPath, *maxRequests, logger)
	http.Handle(*metricsPath, h)
	http.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`<html>
			<head><title>FLEXlm Exporter</title></head>
			<body>
			<h1>FLEXlm Exporter</h1>
			<p><a href="` + *metricsPath + `">Metrics</a></p>
			</body>
			</html>`))
	})

	server := &http.Server{
		ReadHeaderTimeout: serverReadHeaderTimeout * time.Second,
	}

	// Configuring the context for intercepting OS signals (SIGINT, SIGTERM)
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	serverErrors := make(chan error, 1)
	go func() {
		if err := web.ListenAndServe(server, toolkitFlags, logger); err != nil {
			serverErrors <- err
		}
	}()

	select {
	case err := <-serverErrors:
		if err != nil {
			logger.Error("HTTP server error", "err", err)
			os.Exit(1)
		}
	case <-ctx.Done():
		logger.Info("Shutting down flexlm_exporter...")

		// Gracefully stop the collector's background timer goroutines.
		if h.flexlmCollector != nil {
			h.flexlmCollector.Close()
		}

		// Perform a graceful shutdown of the web server (10-second timeout).
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Error("Graceful shutdown failed", "err", err)
			_ = server.Close()
		}

		logger.Info("Server successfully stopped")
	}
}
