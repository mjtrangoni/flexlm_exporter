// Copyright 2018 Mario Trangoni
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
)

type lmstatFeatureExpCollector struct {
	lmstatFeatureExp *prometheus.Desc
}

func init() {
	registerCollector("lmstat_feature_exp", defaultEnabled,
		NewLmstatFeatureExpCollector)
}

// NewLmstatFeatureExpCollector returns a new Collector exposing lmstat license
// feature expiration date.
func NewLmstatFeatureExpCollector() (Collector, error) {
	return &lmstatFeatureExpCollector{
		lmstatFeatureExp: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "feature",
				"expiration_seconds"),
			"License feature expiration date in seconds labeled by app, name, index, licenses, vendor, version.",
			[]string{"app", "name", "index", "licenses", "vendor",
				"version"}, nil,
		),
	}, nil
}

// Update calls (*lmstatFeatureExpCollector).getLmstatFeatureExpDate to get the
// platform specific memory metrics.
func (c *lmstatFeatureExpCollector) Update(ch chan<- prometheus.Metric) error {
	err := c.getLmstatFeatureExpDate(ch)
	if err != nil {
		return fmt.Errorf("couldn't get licenses feature expiration date: %s", err)
	}
	return nil
}
