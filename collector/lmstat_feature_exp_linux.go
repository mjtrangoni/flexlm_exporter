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

func parseLmstatLicenseFeatureExpDate(outStr [][]string) map[string]*featureExp {
	featuresExp := make(map[string]*featureExp)
	var featureName string
	var expires float64
	// iterate over output lines
	for _, line := range outStr {
		lineJoined := strings.Join(line, "")
		if !lmutilLicenseFeatureExpRegex.MatchString(lineJoined) {
			continue
		}
		matches := lmutilLicenseFeatureExpRegex.FindStringSubmatch(lineJoined)
		featureName = matches[1]
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

		if _, ok := featuresExp[featureName]; !ok {
			featuresExp[featureName] = &featureExp{
				expires:  expires,
				licenses: matches[3],
				vendor:   matches[5],
				version:  matches[2],
			}
		} else {
			log.Debugf("Feature %s exists already, sum lic counts and use the earliest exp date", featureName)
			// We take the earliest expiration date
			if featuresExp[featureName].expires > expires {
				featuresExp[featureName].expires = expires
			}
			// We have to convert licenses to int to sum them
			newLicCount, err := strconv.Atoi(matches[3])
			if err != nil {
				log.Errorf("could not convert %s to integer: %v", matches[3], err)
			}
			oldLicCount, err := strconv.Atoi(featuresExp[featureName].licenses)
			if err != nil {
				log.Errorf("could not convert %s to integer: %v", featuresExp[featureName].licenses, err)
			}
			featuresExp[featureName].licenses = strconv.Itoa(newLicCount + oldLicCount)
		}
	}
	return featuresExp
}

// getLmstatFeatureExpDate returns lmstat active and inactive licenses expiration date
func (c *lmstatFeatureExpCollector) getLmstatFeatureExpDate(ch chan<- prometheus.Metric) error {
	var outBytes []byte
	var err error

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
		var featuresToExclude = []string{}
		var featuresToInclude = []string{}
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

		for name, featureExp := range featuresExp {
			if contains(featuresToExclude, name) {
				continue
			} else if licenses.FeaturesToInclude != "" &&
				!contains(featuresToInclude, name) {
				continue
			}

			ch <- prometheus.MustNewConstMetric(c.lmstatFeatureExp,
				prometheus.GaugeValue, featureExp.expires,
				licenses.Name, name, featureExp.licenses,
				featureExp.vendor, featureExp.version)
		}
	}
	return nil
}
