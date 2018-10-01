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
	"io/ioutil"
	"path/filepath"

	"github.com/prometheus/common/log"
	"gopkg.in/yaml.v2"
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
}

// Configuration type for all licenses.
type Configuration struct {
	Licenses []License `yaml:"licenses"`
}

// Load parses the YAML file.
func Load(filename string) (Configuration, error) {

	log.Infoln("Loading license config file:")
	log.Infof(" - %s", filename)

	bytes, err := ioutil.ReadFile(filepath.Clean(filename))
	if err != nil {
		return Configuration{}, err
	}

	var c Configuration
	err = yaml.Unmarshal(bytes, &c)
	if err != nil {
		log.Fatalf("Couldn't load config file: %s", err)
	}

	return c, nil
}
