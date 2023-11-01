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

package collector

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/mjtrangoni/flexlm_exporter/config"
	"github.com/prometheus/client_golang/prometheus"
)

type lmstatCollector struct {
	lmstatInfo                     *prometheus.Desc
	lmstatServerStatus             *prometheus.Desc
	lmstatVendorStatus             *prometheus.Desc
	lmstatFeatureUsed              *prometheus.Desc
	lmstatFeatureUsedUsers         *prometheus.Desc
	lmstatFeatureUsedUsersVersions *prometheus.Desc
	lmstatFeatureReservGroups      *prometheus.Desc
	lmstatFeatureIssued            *prometheus.Desc
	logger                         log.Logger
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
func NewLmstatCollector(logger log.Logger) (Collector, error) {
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
				"username of the license.", []string{"app", "name", "user", "since"}, nil,
		),
		lmstatFeatureUsedUsersVersions: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "feature", "used_users"),
			"License feature used by user labeled by app, feature name, "+
				"username of the license and version.", []string{"app", "name", "user", "since", "version"}, nil,
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
		logger: logger,
	}, nil
}

// Update calls (*lmstatCollector).getLmStat to get the platform specific
// memory metrics.
func (c *lmstatCollector) Update(ch chan<- prometheus.Metric) error {
	err := c.getLmstatInfo(ch)
	if err != nil {
		return fmt.Errorf("couldn't get lmstat version information: %w", err)
	}

	err = c.getLmstatLicensesInfo(ch)
	if err != nil {
		return fmt.Errorf("couldn't get licenses information: %w", err)
	}

	return nil
}

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
func lmutilOutput(logger log.Logger, args ...string) ([]byte, error) {
	_, err := os.Stat(*lmutilPath)
	if os.IsNotExist(err) {
		level.Error(logger).Log("err", *lmutilPath, "missing")
		os.Exit(1)
	}

	cmd := exec.Command(*lmutilPath, args...)
	// Disable localization for parsing.
	cmd.Env = append(os.Environ(), "LANG=C")

	out, err := cmd.Output()
	if err != nil {
		// convert error to strings
		errorToString := errorDescriptionString[err.Error()]
		if errorToString != "" {
			return nil, fmt.Errorf("error while calling '%s %s': %v:'%s'", *lmutilPath,
				strings.Join(args, " "), err, errorToString)
		}

		return nil, fmt.Errorf("error while calling '%s %s': %v:'unknown error'",
			*lmutilPath, strings.Join(args, " "), err)
	}

	return out, nil
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
		return result, fmt.Errorf("could not parse lmutil output: %w", err)
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

	return res, nil
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

func parseLmstatLicenseInfoFeature(outStr [][]string, logger log.Logger) (map[string]*feature,
	map[string]map[string][]*featureUserUsed, map[string]map[string]float64) {
	features := make(map[string]*feature)
	licUsersByFeature := make(map[string]map[string][]*featureUserUsed)
	reservGroupByFeature := make(map[string]map[string]float64)
	// featureName saved here as index for the user and reservation information.
	var featureName string

	for _, line := range outStr {
		lineJoined := strings.Join(line, "")
		if lmutilLicenseFeatureUsageRegex.MatchString(lineJoined) {
			matches := lmutilLicenseFeatureUsageRegex.FindStringSubmatch(lineJoined)

			issued, err := strconv.Atoi(matches[2])
			if err != nil {
				level.Error(logger).Log("could not convert", matches[2], "to integer:", err)
			}

			featureName = matches[1]

			used, err := strconv.Atoi(matches[3])
			if err != nil {
				level.Error(logger).Log("could not convert", matches[3], "to integer:", err)
			}

			features[featureName] = &feature{
				issued: float64(issued),
				used:   float64(used),
			}
		} else if lmutilLicenseFeatureUsageUserRegex.MatchString(lineJoined) {
			if licUsersByFeature[featureName] == nil {
				licUsersByFeature[featureName] = map[string][]*featureUserUsed{}
			}
			matches := reSubMatchMap(lmutilLicenseFeatureUsageUserRegex, lineJoined)
			username := matches["user"]
			if strings.TrimSpace(username) == "" {
				level.Debug(logger).Log("username couldn't be found for '", lineJoined,
					"', using lmutilLicenseFeatureUsageUser2Regex.")
				matches = reSubMatchMap(lmutilLicenseFeatureUsageUser2Regex, lineJoined)
				username = matches["user"]
			}
			if matches["ver"] != "" {
				var found = -1
				for i := range licUsersByFeature[featureName][username] {
					if licUsersByFeature[featureName][username][i].version == matches["ver"] {
						found = i
					}
				}
				if found < 0 {
					unixSince := convertLmstatTimeToUnixTime(matches["since"], logger).Unix()
					sinceString := strconv.FormatInt(unixSince, 10)
					licUsersByFeature[featureName][username] = append(licUsersByFeature[featureName][username],
						&featureUserUsed{num: 0, version: matches["ver"], since: sinceString})
				}
			}
			if matches["licenses"] != "" {
				licUsed, err := strconv.Atoi(matches["licenses"])
				if err != nil {
					level.Error(logger).Log("could not convert", matches["licenses"], "to integer:", err)
				}
				for i := range licUsersByFeature[featureName][username] {
					if licUsersByFeature[featureName][username][i].version == matches["ver"] {
						licUsersByFeature[featureName][username][i].num += float64(licUsed)
					}
				}
			} else {
				for i := range licUsersByFeature[featureName][username] {
					if licUsersByFeature[featureName][username][i].version == matches["ver"] {
						licUsersByFeature[featureName][username][i].num += 1.0
					}
				}
			}
		} else if lmutilLicenseFeatureGroupReservRegex.MatchString(lineJoined) {
			if reservGroupByFeature[featureName] == nil {
				reservGroupByFeature[featureName] = map[string]float64{}
			}
			matches := lmutilLicenseFeatureGroupReservRegex.FindStringSubmatch(lineJoined)
			groupReserv, err := strconv.Atoi(matches[2])
			if err != nil {
				level.Error(logger).Log("could not convert", matches[1], "to integer:", err)
			}
			reservGroupByFeature[featureName][matches[4]] = float64(groupReserv)
		}
	}

	return features, licUsersByFeature, reservGroupByFeature
}

// getLmstatInfo returns lmstat binary information.
func (c *lmstatCollector) getLmstatInfo(ch chan<- prometheus.Metric) error {
	outBytes, err := lmutilOutput(c.logger, "lmstat", "-v")
	if err != nil {
		return err
	}

	outStr, err := splitOutput(outBytes)
	if err != nil {
		return err
	}

	lmstatInfo := parseLmstatVersion(outStr)

	ch <- prometheus.MustNewConstMetric(c.lmstatInfo, prometheus.GaugeValue, 1.0, lmstatInfo.arch, lmstatInfo.build, lmstatInfo.version)

	return nil
}

// getLmstatLicensesInfo returns lmstat active licenses information.
//
//nolint:unparam
func (c *lmstatCollector) getLmstatLicensesInfo(ch chan<- prometheus.Metric) error {
	wg := &sync.WaitGroup{}
	defer wg.Wait()

	for _, licenses := range LicenseConfig.Licenses {
		wg.Add(lenghtOne)

		go func(licenses config.License) {
			defer wg.Done()

			if err := c.collect(&licenses, ch); err == nil {
				ch <- prometheus.MustNewConstMetric(scrapeErrorDesc, prometheus.GaugeValue, 0, "lmstat", licenses.Name)
			} else {
				ch <- prometheus.MustNewConstMetric(scrapeErrorDesc, prometheus.GaugeValue, 1, "lmstat", licenses.Name)
			}
		}(licenses)
	}

	return nil
}

func (c *lmstatCollector) collect(licenses *config.License, ch chan<- prometheus.Metric) error {
	var (
		outBytes []byte
		err      error
	)

	// Call lmstat with -a (display everything)
	if licenses.LicenseFile != "" {
		outBytes, err = lmutilOutput(c.logger, "lmstat", "-c", licenses.LicenseFile, "-a")
		if err != nil {
			return err
		}
	} else if licenses.LicenseServer != "" {
		outBytes, err = lmutilOutput(c.logger, "lmstat", "-c", licenses.LicenseServer, "-a")
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("couldn't find `license_file` or `license_server` for %v",
			licenses.Name)
	}

	outStr, err := splitOutput(outBytes)
	if err != nil {
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
		return fmt.Errorf("%v: can not define `features_to_include` and "+
			"`features_to_exclude` at the same time", licenses.Name)
	} else if licenses.FeaturesToExclude != "" {
		featuresToExclude = strings.Split(licenses.FeaturesToExclude, ",")
	} else if licenses.FeaturesToInclude != "" {
		featuresToInclude = strings.Split(licenses.FeaturesToInclude, ",")
	}

	features, licUsersByFeature, reservGroupByFeature := parseLmstatLicenseInfoFeature(outStr, c.logger)
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
			if licenses.MonitorVersions {
				for username, licused := range licUsersByFeature[name] {
					for i := range licused {
						ch <- prometheus.MustNewConstMetric(
							c.lmstatFeatureUsedUsersVersions, prometheus.GaugeValue,
							licused[i].num, licenses.Name, name, username, licused[i].since, licused[i].version)
					}
				}
			} else {
				for username, licused := range licUsersByFeature[name] {
					for i := range licused {
						ch <- prometheus.MustNewConstMetric(
							c.lmstatFeatureUsedUsers, prometheus.GaugeValue,
							licused[i].num, licenses.Name, name, username, licused[i].since)
					}
				}
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

// from https://stackoverflow.com/a/46202939
func reSubMatchMap(r *regexp.Regexp, str string) map[string]string {
	match := r.FindStringSubmatch(str)
	subMatchMap := make(map[string]string)

	for i, name := range r.SubexpNames() {
		if i != 0 {
			subMatchMap[name] = match[i]
		}
	}

	return subMatchMap
}

func convertLmstatTimeToUnixTime(lmtime string, logger log.Logger) time.Time {
	matches := reSubMatchMap(lmutilTimeRegex, lmtime)

	// current time and offset (lmstat outputs the time for the current time zone of the server where it is executed)
	ctime := time.Now()
	_, offset := ctime.UTC().Zone()

	closure := func(m map[string]string, year int) time.Time {
		month, _ := strconv.Atoi(m["month"])
		day, _ := strconv.Atoi(m["day"])

		// RFC3339 time string from the lmstat time information
		ltimes := fmt.Sprintf("%v-%02d-%02dT%v:00Z", year, month, day, m["time"])
		ltime, err := time.Parse(time.RFC3339, ltimes)

		// correct the created time (local) to UTC time by subtracting the offset
		ltime = ltime.Add(-time.Duration(offset) * time.Second)

		if err != nil {
			level.Error(logger).Log("could not convert", ltime, "to unix time:", err)

			// fallback, just return the current time in case of errors
			return ctime
		}

		return ltime
	}

	unixtime := closure(matches, ctime.Year())

	// if the created unixtime > current time, then the year is wrong, and we just subtract 1 year here. This is a guess
	// and not necessarily correct, but lmstat does not provide any information about the year
	if unixtime.After(ctime) {
		unixtime = closure(matches, ctime.Year()-1)
	}

	return unixtime
}
