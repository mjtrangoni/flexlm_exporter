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
	"bytes"
	"encoding/csv"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"

	"github.com/mjtrangoni/flexlm_exporter/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

// contains check if an array contains a string.
func contains(slice []string, item string) bool {
	set := make(map[string]struct{}, len(slice))
	for _, s := range slice {
		set[s] = struct{}{}
	}

	_, ok := set[item]

	return ok
}

// execute lmutil utility.
func lmutilOutput(args ...string) ([]byte, error) {
	cmd := exec.Command(*lmutilPath, args...)
	// Disable localization for parsing.
	cmd.Env = append(os.Environ(), "LANG=C")

	out, err := cmd.Output()
	if err != nil {
		// convert error to strings
		errorToString := errorDescriptionString[err.Error()]
		if errorToString != "" {
			log.Errorf("error while calling '%s %s': %v:'%s'", *lmutilPath,
				strings.Join(args, " "), err, errorToString)
		} else {
			log.Errorf("error while calling '%s %s': %v:'unknown error'",
				*lmutilPath, strings.Join(args, " "), err)
		}
	}

	return out, err
}

func splitOutput(lmutilOutput []byte) ([][]string, error) {
	r := csv.NewReader(bytes.NewReader(lmutilOutput))
	// It seems that some vendors used to encrypt the display, and contains
	// pipes. That is why we have to use other special characters.
	// r.Comma = '|'
	r.Comma = 'Ž'
	r.LazyQuotes = true
	r.Comment = '#'

	result, err := r.ReadAll()
	if err != nil {
		log.Errorf("could not parse lmutil output: %v", err)
		return result, err
	}

	keys := make(map[string]int)

	res := make([][]string, len(result))

	for _, v := range result {
		key := v[0]
		if _, ok := keys[key]; ok {
			keys[key]++

			v[0] = strings.TrimSpace(v[0]) + strconv.Itoa(keys[key])
		} else {
			keys[key] = 1
		}
		res = append(res, v)
	}

	return res, err
}

func parseLmstatVersion(outStr [][]string) lmstatInformation {
	for _, line := range outStr {
		lineJoined := strings.Join(line, "")
		if lmutilVersionRegex.MatchString(lineJoined) {
			names := lmutilVersionRegex.SubexpNames()
			matches := lmutilVersionRegex.FindAllStringSubmatch(lineJoined, -1)[0]

			md := map[string]string{}
			for i, n := range matches {
				md[names[i]] = n
			}

			return lmstatInformation{
				arch:    md["arch"],
				build:   md["build"],
				version: md["version"],
			}
		}
	}

	return lmstatInformation{arch: notFound, build: notFound, version: notFound}
}

func parseLmstatLicenseInfoServer(outStr [][]string) map[string]*server {
	servers := make(map[string]*server)

	for _, line := range outStr {
		lineJoined := strings.Join(line, "")
		if lmutilLicenseServersRegex.MatchString(lineJoined) {
			matches := lmutilLicenseServersRegex.FindStringSubmatch(lineJoined)[1]
			for _, portServer := range strings.Split(matches, ",") {
				fqdn := strings.Split(portServer, "@")[1]
				servers[strings.Split(fqdn, ".")[0]] = &server{
					fqdn: fqdn, port: strings.Split(portServer, "@")[0],
				}
			}
		} else if lmutilLicenseServerStatusRegex.MatchString(lineJoined) {
			matches := lmutilLicenseServerStatusRegex.FindStringSubmatch(lineJoined)
			servers[strings.Split(matches[1], ".")[0]].version = matches[4]
			if matches[2] == upString {
				servers[strings.Split(matches[1], ".")[0]].status = true
			}
			if matches[3] == " (MASTER)" {
				servers[strings.Split(matches[1], ".")[0]].master = true
			}
		}
	}

	return servers
}

func parseLmstatLicenseInfoVendor(outStr [][]string) map[string]*vendor {
	vendors := make(map[string]*vendor)

	for _, line := range outStr {
		lineJoined := strings.Join(line, "")
		if lmutilLicenseVendorStatusRegex.MatchString(lineJoined) {
			matches := lmutilLicenseVendorStatusRegex.FindStringSubmatch(lineJoined)
			if matches[2] == upString {
				vendors[matches[1]] = &vendor{
					status: true, version: matches[3],
				}
			} else {
				vendors[matches[1]] = &vendor{
					status: false, version: matches[3],
				}
			}
		}
	}

	return vendors
}

func parseLmstatLicenseInfoFeature(outStr [][]string) (map[string]*feature,
	map[string]map[string]float64, map[string]map[string]float64) {
	features := make(map[string]*feature)
	licUsersByFeature := make(map[string]map[string]float64)
	reservGroupByFeature := make(map[string]map[string]float64)
	// featureName saved here as index for the user and reservation information.
	var featureName string

	for _, line := range outStr {
		lineJoined := strings.Join(line, "")
		if lmutilLicenseFeatureUsageRegex.MatchString(lineJoined) {
			matches := lmutilLicenseFeatureUsageRegex.FindStringSubmatch(lineJoined)

			issued, err := strconv.Atoi(matches[2])
			if err != nil {
				log.Errorf("could not convert %s to integer: %v", matches[2],
					err)
			}

			featureName = matches[1]

			used, err := strconv.Atoi(matches[3])
			if err != nil {
				log.Errorf("could not convert %s to integer: %v", matches[3],
					err)
			}

			features[featureName] = &feature{
				issued: float64(issued),
				used:   float64(used),
			}
		} else if lmutilLicenseFeatureUsageUserRegex.MatchString(lineJoined) {
			if licUsersByFeature[featureName] == nil {
				licUsersByFeature[featureName] = map[string]float64{}
			}
			matches := lmutilLicenseFeatureUsageUserRegex.FindStringSubmatch(lineJoined)
			username := matches[1]
			if strings.TrimSpace(username) == "" {
				log.Debugln("username couldn't be found for '" + lineJoined +
					"', using lmutilLicenseFeatureUsageUser2Regex.")
				matches = lmutilLicenseFeatureUsageUser2Regex.FindStringSubmatch(lineJoined)
				username = matches[1]
			}
			if matches[3] != "" {
				licUsed, err := strconv.Atoi(matches[3])
				if err != nil {
					log.Errorf("could not convert %s to integer: %v",
						matches[3], err)
				}
				licUsersByFeature[featureName][username] += float64(licUsed)
			} else {
				licUsersByFeature[featureName][username] += 1.0
			}
		} else if lmutilLicenseFeatureGroupReservRegex.MatchString(lineJoined) {
			if reservGroupByFeature[featureName] == nil {
				reservGroupByFeature[featureName] = map[string]float64{}
			}
			matches := lmutilLicenseFeatureGroupReservRegex.FindStringSubmatch(lineJoined)
			groupReserv, err := strconv.Atoi(matches[2])
			if err != nil {
				log.Errorf("could not convert %s to integer: %v", matches[1],
					err)
			}
			reservGroupByFeature[featureName][matches[4]] = float64(groupReserv)
		}
	}

	return features, licUsersByFeature, reservGroupByFeature
}

// getLmstatInfo returns lmstat binary information
func (c *lmstatCollector) getLmstatInfo(ch chan<- prometheus.Metric) error {
	outBytes, err := lmutilOutput("lmstat", "-v")
	if err != nil {
		log.Errorln(err)
		return err
	}

	outStr, err := splitOutput(outBytes)
	if err != nil {
		log.Errorln(err)
		return err
	}

	lmstatInfo := parseLmstatVersion(outStr)

	ch <- prometheus.MustNewConstMetric(c.lmstatInfo, prometheus.GaugeValue, 1.0, lmstatInfo.arch, lmstatInfo.build, lmstatInfo.version)

	return nil
}

// getLmstatLicensesInfo returns lmstat active licenses information
func (c *lmstatCollector) getLmstatLicensesInfo(ch chan<- prometheus.Metric) error {
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	for _, licenses := range LicenseConfig.Licenses {
		wg.Add(1)
		go func(licenses config.License) {
			defer wg.Done()
			err := c.collect(licenses, ch)
			if err == nil {
				ch <- prometheus.MustNewConstMetric(scrapeErrorDesc, prometheus.GaugeValue, 0, "lmstat", licenses.Name)
			} else {
				ch <- prometheus.MustNewConstMetric(scrapeErrorDesc, prometheus.GaugeValue, 1, "lmstat", licenses.Name)
			}
		}(licenses)
	}
	return nil
}

func (c *lmstatCollector) collect(licenses config.License, ch chan<- prometheus.Metric) error {
	var (
		outBytes []byte
		err      error
	)

	// Call lmstat with -a (display everything)
	if licenses.LicenseFile != "" {
		outBytes, err = lmutilOutput("lmstat", "-c", licenses.LicenseFile, "-a")
		if err != nil {
			return err
		}
	} else if licenses.LicenseServer != "" {
		outBytes, err = lmutilOutput("lmstat", "-c", licenses.LicenseServer, "-a")
		if err != nil {
			return err
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

	servers := parseLmstatLicenseInfoServer(outStr)
	for _, info := range servers {
		if info.status {
			ch <- prometheus.MustNewConstMetric(c.lmstatServerStatus,
				prometheus.GaugeValue, 1.0, licenses.Name, info.fqdn,
				strconv.FormatBool(info.master), info.port, info.version)
		} else {
			ch <- prometheus.MustNewConstMetric(c.lmstatServerStatus,
				prometheus.GaugeValue, 0, licenses.Name, info.fqdn,
				strconv.FormatBool(info.master), info.port, info.version)
		}
	}

	vendors := parseLmstatLicenseInfoVendor(outStr)
	for name, info := range vendors {
		if info.status {
			ch <- prometheus.MustNewConstMetric(c.lmstatVendorStatus,
				prometheus.GaugeValue, 1.0, licenses.Name, name,
				info.version)
		} else {
			ch <- prometheus.MustNewConstMetric(c.lmstatVendorStatus,
				prometheus.GaugeValue, 0, licenses.Name, name, info.version)
		}
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

	features, licUsersByFeature, reservGroupByFeature := parseLmstatLicenseInfoFeature(outStr)
	for name, info := range features {
		if contains(featuresToExclude, name) {
			continue
		} else if licenses.FeaturesToInclude != "" &&
			!contains(featuresToInclude, name) {
			continue
		}
		ch <- prometheus.MustNewConstMetric(c.lmstatFeatureUsed,
			prometheus.GaugeValue, info.used, licenses.Name, name)
		ch <- prometheus.MustNewConstMetric(c.lmstatFeatureIssued,
			prometheus.GaugeValue, info.issued, licenses.Name, name)
		if licenses.MonitorUsers && (licUsersByFeature[name] != nil) {
			for username, licused := range licUsersByFeature[name] {
				ch <- prometheus.MustNewConstMetric(
					c.lmstatFeatureUsedUsers, prometheus.GaugeValue,
					licused, licenses.Name, name, username)
			}
		}
		if licenses.MonitorReservations && (reservGroupByFeature[name] != nil) {
			for group, licreserv := range reservGroupByFeature[name] {
				ch <- prometheus.MustNewConstMetric(
					c.lmstatFeatureReservGroups, prometheus.GaugeValue,
					licreserv, licenses.Name, name, group)
			}
		}
	}

	return nil
}
