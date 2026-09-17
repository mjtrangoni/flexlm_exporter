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

// Package collector includes all individual collectors to gather and export flexlm metrics.
package collector

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"sync"
	"time"

	kingpin "github.com/alecthomas/kingpin/v2"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	// Namespace defines the common namespace to be used by all metrics.
	namespace       = "flexlm"
	defaultEnabled  = true
	appString       = "app"
	collectorString = "collector"
	nameString      = "name"
	upString        = "UP"
	versionString   = "version"
)

var (
	scrapeDurationDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "scrape", "collector_duration_seconds"),
		"flexlm_exporter: Duration of a collector scrape.",
		[]string{collectorString},
		nil,
	)
	scrapeSuccessDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "scrape", "collector_success"),
		"flexlm_exporter: Whether a collector succeeded.",
		[]string{collectorString},
		nil,
	)
	scrapeErrorDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "scrape", "error"),
		"flexlm_exporter: Whether a license scrape had an error.",
		[]string{collectorString, nameString},
		nil,
	)
)

var (
	factories              = make(map[string]func(logger *slog.Logger) (Collector, error))
	initiatedCollectorsMtx = sync.Mutex{}
	initiatedCollectors    = make(map[string]Collector)
	collectorState         = make(map[string]*bool)
	forcedCollectors       = map[string]bool{} // collectors which have been explicitly enabled or disabled
)

func registerCollector(collector string, isDefaultEnabled bool, factory func(logger *slog.Logger) (Collector, error)) {
	var helpDefaultState string
	if isDefaultEnabled {
		helpDefaultState = "enabled"
	} else {
		helpDefaultState = "disabled"
	}

	flagName := "collector." + collector
	flagHelp := fmt.Sprintf("Enable the %s collector (default: %s).", collector, helpDefaultState)
	defaultValue := strconv.FormatBool(isDefaultEnabled)

	flag := kingpin.Flag(flagName, flagHelp).Default(defaultValue).Action(collectorFlagAction(collector)).Bool()
	collectorState[collector] = flag

	factories[collector] = factory
}

// FlexlmCollector implements the prometheus.Collector interface.
type FlexlmCollector struct {
	Collectors    map[string]Collector
	logger        *slog.Logger
	cacheMu       sync.RWMutex
	cachedMetrics map[string][]prometheus.Metric
	ctx           context.Context
	cancel        context.CancelFunc
	wg            sync.WaitGroup // Tracks active background goroutines for graceful shutdown.
}

// collectorFlagAction generates a new action function for the given collector
// to track whether it has been explicitly enabled or disabled from the command line.
// A new action function is needed for each collector flag because the ParseContext
// does not contain information about which flag called the action.
// See: https://github.com/alecthomas/kingpin/issues/294
//
//revive:disable:unused-parameter
func collectorFlagAction(collector string) func(ctx *kingpin.ParseContext) error {
	return func(ctx *kingpin.ParseContext) error {
		forcedCollectors[collector] = true
		return nil
	}
}

// NewFlexlmCollector creates a new FlexlmCollector and starts background asynchronous scrapers.
//
//revive:enable:unused-parameter
func NewFlexlmCollector(logger *slog.Logger, filters ...string) (*FlexlmCollector, error) {
	f := make(map[string]bool)

	for _, filter := range filters {
		enabled, exist := collectorState[filter]
		if !exist {
			return nil, fmt.Errorf("missing collector: %s", filter)
		}

		if !*enabled {
			return nil, fmt.Errorf("disabled collector: %s", filter)
		}

		f[filter] = true
	}

	collectors := make(map[string]Collector)

	initiatedCollectorsMtx.Lock()
	defer initiatedCollectorsMtx.Unlock()

	for key, enabled := range collectorState {
		if !*enabled || (len(f) > 0 && !f[key]) {
			continue
		}

		if collector, ok := initiatedCollectors[key]; ok {
			collectors[key] = collector
		} else {
			collector, err := factories[key](logger.With("collector", key))
			if err != nil {
				return nil, err
			}

			collectors[key] = collector
			initiatedCollectors[key] = collector
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	fc := &FlexlmCollector{
		Collectors:    collectors,
		logger:        logger,
		cachedMetrics: make(map[string][]prometheus.Metric),
		ctx:           ctx,
		cancel:        cancel,
	}

	// Launching background workers for asynchronous metric collection on a timer
	fc.startBackgroundScrapers()

	return fc, nil
}

// startBackgroundScrapers starts individual timers for each active collector.
func (n *FlexlmCollector) startBackgroundScrapers() {
	for name, c := range n.Collectors {
		n.wg.Add(1)
		go func(name string, c Collector) {
			defer n.wg.Done()

			// Determine the timer interval. If the collector implements an interface with a GetInterval() method, use it;
			// otherwise, apply the default value (30 seconds).
			interval := 30 * time.Second
			if timedCollector, ok := c.(interface{ GetInterval() time.Duration }); ok {
				if d := timedCollector.GetInterval(); d > 0 {
					interval = d
				}
			}

			ticker := time.NewTicker(interval)
			defer ticker.Stop()

			n.logger.Info("Starting background asynchronous scraper", "collector", name, "interval", interval)

			// Perform the initial data collection immediately upon startup.
			n.runAndCache(name, c)

			for {
				select {
				case <-n.ctx.Done():
					n.logger.Info("Stopping background asynchronous scraper", "collector", name)
					return
				case <-ticker.C:
					n.runAndCache(name, c)
				}
			}
		}(name, c)
	}
}

// Close stops all background workers and waits for them to complete fully (graceful shutdown).
func (n *FlexlmCollector) Close() {
	n.cancel()
	n.wg.Wait()
	n.logger.Info("All background scrapers have been stopped cleanly")
}

// runAndCache collects metrics in the background, measures execution time, handles errors, and saves the result to the cache.
func (n *FlexlmCollector) runAndCache(name string, c Collector) {
	var success float64
	begin := time.Now()

	ch := make(chan prometheus.Metric, 100)
	var metrics []prometheus.Metric

	// Asynchronous collection of metrics from the channel returned by the collector
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for m := range ch {
			metrics = append(metrics, m)
		}
	}()

	err := c.Update(ch)
	close(ch)
	wg.Wait()

	duration := time.Since(begin)

	if err != nil {
		if IsNoDataError(err) {
			n.logger.Debug("collector returned no data", nameString, name, "duration_seconds", duration.Seconds(), "err", err)
		} else {
			n.logger.Error("collector failed", nameString, name, "duration_seconds", duration.Seconds(), "err", err)
		}
		success = 0
	} else {
		n.logger.Debug("collector succeeded", nameString, name, "duration_seconds", duration.Seconds())
		success = 1
	}

	// Adding service metrics (duration and collection success status) to the cache
	metrics = append(metrics,
		prometheus.MustNewConstMetric(scrapeDurationDesc, prometheus.GaugeValue, duration.Seconds(), name),
		prometheus.MustNewConstMetric(scrapeSuccessDesc, prometheus.GaugeValue, success, name),
	)

	n.cacheMu.Lock()
	n.cachedMetrics[name] = metrics
	n.cacheMu.Unlock()
}

// Describe implements the prometheus.Collector interface.
func (n *FlexlmCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- scrapeDurationDesc
	ch <- scrapeSuccessDesc
	ch <- scrapeErrorDesc
}

// Collect implements the prometheus.Collector interface.
// It now serves cached data instantly, without blocking or waiting for external utilities.
func (n *FlexlmCollector) Collect(ch chan<- prometheus.Metric) {
	n.cacheMu.RLock()
	defer n.cacheMu.RUnlock()

	for _, metrics := range n.cachedMetrics {
		for _, m := range metrics {
			ch <- m
		}
	}
}

// Collector is the interface a collector has to implement.
type Collector interface {
	// Get new metrics and expose them via prometheus registry.
	Update(ch chan<- prometheus.Metric) error
}

// ErrNoData indicates the collector found no data to collect, but had no other error.
var ErrNoData = errors.New("collector returned no data")

func IsNoDataError(err error) bool {
	return errors.Is(err, ErrNoData)
}
