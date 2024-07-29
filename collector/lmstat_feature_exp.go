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

package collector

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/mjtrangoni/flexlm_exporter/config"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const (
	lenghtOne   = 1
	posInfinity = 1
	yearLength  = 4
)

type lmstatFeatureExpCollector struct {
	lmstatFeatureExp     *prometheus.Desc
	lmstatFeatureAggrExp *prometheus.Desc
	logger               log.Logger
}

func init() {
	registerCollector("lmstat_feature_exp", defaultEnabled,
		NewLmstatFeatureExpCollector)
}

// NewLmstatFeatureExpCollector returns a new Collector exposing lmstat license
// feature expiration date.
func NewLmstatFeatureExpCollector(logger log.Logger) (Collector, error) {
	return &lmstatFeatureExpCollector{
		lmstatFeatureExp: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "feature",
				"expiration_seconds"),
			"License feature expiration date in seconds labeled by app, name, index, licenses, vendor, version.",
			[]string{
				"app", "name", "index", "licenses", "vendor",
				"version",
			}, nil,
		),
		lmstatFeatureAggrExp: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "feature", "aggregate_expiration_seconds"),
			"Aggregate by license features expiration day in seconds. Labeled by app, licenses, features.",
			[]string{"app", "index", "licenses", "features"}, nil,
		),
		logger: logger,
	}, nil
}

// Update calls (*lmstatFeatureExpCollector).getLmstatFeatureExpDate to get the
// platform specific memory metrics.
func (c *lmstatFeatureExpCollector) Update(ch chan<- prometheus.Metric) error {
	err := c.getLmstatFeatureExpDate(ch)
	if err != nil {
		return fmt.Errorf("couldn't get licenses feature expiration date: %w", err)
	}

	return nil
}

func parseLmstatLicenseFeatureExpDate(outStr [][]string, logger log.Logger) map[int]*featureExp {
	var (
		expires     float64
		index       int
		expIndex    int
		vendorIndex int
		matches     []string
	)

	featuresExp := make(map[int]*featureExp)
	// iterate over output lines
	for _, line := range outStr {
		lineJoined := strings.Join(line, "")
		if !lmutilLicenseFeatureExpRegex.MatchString(lineJoined) &&
			!lmutilLicenseFeatureExpRegex2.MatchString(lineJoined) {
			continue
		} else if lmutilLicenseFeatureExpRegex.MatchString(lineJoined) {
			matches = lmutilLicenseFeatureExpRegex.FindStringSubmatch(lineJoined)
			expIndex = 4
			vendorIndex = 5
		} else {
			matches = lmutilLicenseFeatureExpRegex2.FindStringSubmatch(lineJoined)
			expIndex = 5
			vendorIndex = 4
		}

		level.Debug(logger).Log(matches)
		// Parse date, month has to be capitalized.
		slice := strings.Split(matches[expIndex], "-")
		if len(slice) > lenghtOne {
			day, month, year := slice[0], slice[1], slice[2]
			if len(year) > yearLength {
				lenToRemove := len(year) - yearLength
				year = year[:len(year)-lenToRemove]
			}
			if len(day) == lenghtOne {
				day = "0" + day
			}
			if len(year) == lenghtOne {
				year = "000" + year
			}

			expireDate, err := time.Parse("02-Jan-2006",
				fmt.Sprintf("%s-%s-%s", day,
					cases.Title(language.English).String(month), year))
			if err != nil {
				level.Error(logger).Log("could not convert to date:", err)
			}

			if expireDate.Unix() <= 0 {
				expires = math.Inf(posInfinity)
			} else {
				expires = float64(expireDate.Unix())
			}
		} else {
			// every string matching the expiration position will be considered
			// as permanent
			expires = math.Inf(posInfinity)
		}

		index++

		featuresExp[index] = &featureExp{
			name:     matches[1],
			expires:  expires,
			licenses: matches[3],
			vendor:   matches[vendorIndex],
			version:  matches[2],
		}
	}

	return featuresExp
}

// getLmstatFeatureExpDate returns lmstat active and inactive licenses expiration date.
func (c *lmstatFeatureExpCollector) getLmstatFeatureExpDate(ch chan<- prometheus.Metric) error {
	wg := &sync.WaitGroup{}
	defer wg.Wait()

	for _, licenses := range LicenseConfig.Licenses {
		wg.Add(lenghtOne)

		go func(licenses config.License) {
			defer wg.Done()

			if err := c.collect(&licenses, ch); err == nil {
				ch <- prometheus.MustNewConstMetric(scrapeErrorDesc, prometheus.GaugeValue, 0, "lmstat_feature_exp", licenses.Name)
			} else {
				ch <- prometheus.MustNewConstMetric(scrapeErrorDesc, prometheus.GaugeValue, 1, "lmstat_feature_exp", licenses.Name)
			}
		}(licenses)
	}

	return nil
}

func (c *lmstatFeatureExpCollector) collect(licenses *config.License, ch chan<- prometheus.Metric) error {
	var (
		outBytes []byte
		err      error
	)
	// Call lmstat with -i (lmstat -i does not give information from the server,
	// but only reads the license file)
	if licenses.LicenseFile != "" {
		outBytes, err = lmutilOutput(c.logger, "lmstat", "-c", licenses.LicenseFile, "-i")
		if err != nil {
			return err
		}
	} else if licenses.LicenseServer != "" {
		outBytes, err = lmutilOutput(c.logger, "lmstat", "-c", licenses.LicenseServer, "-i")
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("couldn't find `license_file` or `license_server` for %v", licenses.Name)
	}

	outStr, err := splitOutput(outBytes)
	if err != nil {
		return err
	}

	// features
	var (
		featuresToExclude = []string{}
		featuresToInclude = []string{}
	)

	if licenses.FeaturesToExclude != "" && licenses.FeaturesToInclude != "" {
		return fmt.Errorf("%v: can not define `features_to_include` and "+
			"`features_to_exclude` at the same time", licenses.Name)
	} else if licenses.FeaturesToExclude != "" {
		featuresToExclude = strings.Split(licenses.FeaturesToExclude, ",")
	} else if licenses.FeaturesToInclude != "" {
		featuresToInclude = strings.Split(licenses.FeaturesToInclude, ",")
	}

	featuresExp := parseLmstatLicenseFeatureExpDate(outStr, c.logger)
	aggrFeaturesExpMap := make(map[float64]*aggrFeaturesExp)

	for idx, feature := range featuresExp {
		if contains(featuresToExclude, feature.name) {
			continue
		} else if licenses.FeaturesToInclude != "" &&
			!contains(featuresToInclude, feature.name) {
			continue
		}

		licenseCount, _ := strconv.Atoi(feature.licenses)
		if val, ok := aggrFeaturesExpMap[feature.expires]; ok {
			val.licenses += licenseCount
			val.features++
		} else {
			aggrFeaturesExpMap[feature.expires] = &aggrFeaturesExp{
				app:      licenses.Name,
				features: lenghtOne,
				licenses: licenseCount,
			}
		}

		ch <- prometheus.MustNewConstMetric(c.lmstatFeatureExp,
			prometheus.GaugeValue, feature.expires,
			licenses.Name, feature.name, strconv.Itoa(idx),
			feature.licenses, feature.vendor,
			feature.version)
	}

	aggrFeaturesKeys := make([]float64, 0, len(aggrFeaturesExpMap))

	for exp := range aggrFeaturesExpMap {
		aggrFeaturesKeys = append(aggrFeaturesKeys, exp)
	}

	sort.Float64s(aggrFeaturesKeys)

	for idx, exp := range aggrFeaturesKeys {
		val := aggrFeaturesExpMap[exp]
		ch <- prometheus.MustNewConstMetric(c.lmstatFeatureAggrExp,
			prometheus.GaugeValue, exp,
			val.app, strconv.Itoa(idx), strconv.Itoa(val.licenses), strconv.Itoa(val.features))
	}

	return nil
}
