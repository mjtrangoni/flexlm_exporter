// Copyright 2017 Mario Trangoni
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

// +build linux windows

package collector

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/mjtrangoni/flexlm_exporter/config"
)

type lmstatCollector struct {
	lmstatInfo                *prometheus.Desc
	lmstatServerStatus        *prometheus.Desc
	lmstatVendorStatus        *prometheus.Desc
	lmstatFeatureUsed         *prometheus.Desc
	lmstatFeatureUsedUsers    *prometheus.Desc
	lmstatFeatureReservGroups *prometheus.Desc
	lmstatFeatureIssued       *prometheus.Desc
}

// LicenseConfig is going to be read once in main, and then used here.
var LicenseConfig config.Configuration

const (
	notFound = "not found"
)

func init() {
	registerCollector("lmstat", defaultEnabled, NewLmstatCollector)
}

// NewLmstatCollector returns a new Collector exposing lmstat license stats.
func NewLmstatCollector() (Collector, error) {
	return &lmstatCollector{
		lmstatInfo: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "lmstat", "info"),
			"A metric with a constant '1' value labeled by arch, build and version of the lmstat tool.",
			[]string{"arch", "build", "version"}, nil,
		),
		lmstatServerStatus: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "server", "status"),
			"License server status labeled by app, fqdn, master, port and version of the license.",
			[]string{"app", "fqdn", "master", "port", "version"}, nil,
		),
		lmstatVendorStatus: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "vendor", "status"),
			"License vendor status labeled by app, name and version of the license.",
			[]string{"app", "name", "version"}, nil,
		),
		lmstatFeatureUsed: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "feature", "used"),
			"License feature used labeled by app and feature name of the license.",
			[]string{"app", "name"}, nil,
		),
		lmstatFeatureUsedUsers: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "feature", "used_users"),
			"License feature used by user labeled by app, feature name and "+
				"username of the license.", []string{"app", "name", "user"}, nil,
		),
		lmstatFeatureReservGroups: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "feature", "reserved_groups"),
			"License feature reserved by group labeled by app, feature name "+
				"and group name of the license.", []string{"app", "name", "group"},
			nil,
		),
		lmstatFeatureIssued: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "feature", "issued"),
			"License feature issued labeled by app and feature name of the license.",
			[]string{"app", "name"}, nil,
		),
	}, nil
}

// Update calls (*lmstatCollector).getLmStat to get the platform specific
// memory metrics.
func (c *lmstatCollector) Update(ch chan<- prometheus.Metric) error {
	err := c.getLmstatInfo(ch)
	if err != nil {
		return fmt.Errorf("couldn't get lmstat version information: %s", err)
	}

	err = c.getLmstatLicensesInfo(ch)
	if err != nil {
		return fmt.Errorf("couldn't get licenses information: %s", err)
	}

	return nil
}
