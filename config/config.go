// Package config includes all individual types and functions to gather
// the monitored licenses.
// (C) Copyright 2017 Mario Trangoni.
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
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"gopkg.in/yaml.v3"
)

// YAML Type definitions

// License individual configuration type.
type License struct {
	Name                string `yaml:"name"`
	LicenseFile         string `yaml:"license_file,omitempty"`
	LicenseServer       string `yaml:"license_server,omitempty"`
	FeaturesToExclude   string `yaml:"features_to_exclude,omitempty"`
	FeaturesToInclude   string `yaml:"features_to_include,omitempty"`
	MonitorUsers        bool   `yaml:"monitor_users"`
	MonitorReservations bool   `yaml:"monitor_reservations"`
	MonitorVersions     bool   `yaml:"monitor_versions,omitempty"`
}

// Configuration type for all licenses.
type Configuration struct {
	Licenses []License `yaml:"licenses"`
}

// Load parses the YAML file.
func Load(filename string, logger log.Logger) (Configuration, error) {
	level.Info(logger).Log("msg", "Loading license config file:")
	level.Info(logger).Log(" - ", filename)

	bytes, err := os.ReadFile(filepath.Clean(filename))
	if err != nil {
		return Configuration{}, fmt.Errorf("failed to read %s: %w", filename, err)
	}

	var c Configuration

	err = yaml.Unmarshal(bytes, &c)

	if err != nil {
		level.Error(logger).Log("Couldn't load config file: ", err)
		return c, err
	}

	return c, nil
}
