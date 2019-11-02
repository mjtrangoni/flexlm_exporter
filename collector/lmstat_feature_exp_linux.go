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

package collector

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

func parseLmstatLicenseFeatureExpDate(outStr [][]string) map[int]*featureExp {
	var (
		expires float64
		index   int
	)

	featuresExp := make(map[int]*featureExp)
	// iterate over output lines
	for _, line := range outStr {
		lineJoined := strings.Join(line, "")
		if !lmutilLicenseFeatureExpRegex.MatchString(lineJoined) {
			continue
		}
		matches := lmutilLicenseFeatureExpRegex.FindStringSubmatch(lineJoined)
		// Parse date, month has to be capitalized.
		slice := strings.Split(matches[4], "-")
		day, month, year := slice[0], slice[1], slice[2]
		if len(day) == 1 {
			day = "0" + day
		}
		if len(year) == 1 {
			year = "000" + year
		}
		expireDate, err := time.Parse("02-Jan-2006",
			fmt.Sprintf("%s-%s-%s", day,
				strings.Title(month), year))
		if err != nil {
			log.Errorf("could not convert to date: %v", err)
		}

		if expireDate.Unix() <= 0 {
			expires = math.Inf(1)
		} else {
			expires = float64(expireDate.Unix())
		}

		index++
		featuresExp[index] = &featureExp{
			name:     matches[1],
			expires:  expires,
			licenses: matches[3],
			vendor:   matches[5],
			version:  matches[2],
		}
	}
	return featuresExp
}

// getLmstatFeatureExpDate returns lmstat active and inactive licenses expiration date
func (c *lmstatFeatureExpCollector) getLmstatFeatureExpDate(ch chan<- prometheus.Metric) error {
	var (
		outBytes []byte
		err      error
	)

	for _, licenses := range LicenseConfig.Licenses {
		// Call lmstat with -i (lmstat -i does not give information from the server,
		// but only reads the license file)
		if licenses.LicenseFile != "" {
			outBytes, err = lmutilOutput("lmstat", "-c", licenses.LicenseFile, "-i")
			if err != nil {
				continue
			}
		} else if licenses.LicenseServer != "" {
			outBytes, err = lmutilOutput("lmstat", "-c", licenses.LicenseServer, "-i")
			if err != nil {
				continue
			}
		} else {
			log.Fatalf("couldn`t find `license_file` or `license_server` for %v",
				licenses.Name)
			return nil
		}

		outStr, err := splitOutput(outBytes)
		if err != nil {
			log.Errorln(err)
			return err
		}

		// features
		var (
			featuresToExclude = []string{}
			featuresToInclude = []string{}
		)

		if licenses.FeaturesToExclude != "" && licenses.FeaturesToInclude != "" {
			log.Fatalln("%v: can not define `features_to_include` and "+
				"`features_to_exclude` at the same time", licenses.Name)
			return nil
		} else if licenses.FeaturesToExclude != "" {
			featuresToExclude = strings.Split(licenses.FeaturesToExclude, ",")
		} else if licenses.FeaturesToInclude != "" {
			featuresToInclude = strings.Split(licenses.FeaturesToInclude, ",")
		}

		featuresExp := parseLmstatLicenseFeatureExpDate(outStr)

		for idx, feature := range featuresExp {
			if contains(featuresToExclude, feature.name) {
				continue
			} else if licenses.FeaturesToInclude != "" &&
				!contains(featuresToInclude, feature.name) {
				continue
			}

			ch <- prometheus.MustNewConstMetric(c.lmstatFeatureExp,
				prometheus.GaugeValue, feature.expires,
				licenses.Name, feature.name, strconv.Itoa(idx),
				feature.licenses, feature.vendor,
				feature.version)
		}
	}
	return nil
}
