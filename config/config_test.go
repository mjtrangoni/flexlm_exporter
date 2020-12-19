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

package config_test

import (
	"regexp"
	"testing"

	"github.com/mjtrangoni/flexlm_exporter/config"
)

const (
	testLoadYml = "fixtures/licenses.yml"
)

func TestLoad(t *testing.T) {
	t.Parallel()

	testLicenseConfig, err := config.Load(testLoadYml)
	if err != nil {
		t.Fatal(err)
	}

	appRegex := regexp.MustCompile(`^app\d`)

	for _, licenses := range testLicenseConfig.Licenses {
		if !appRegex.MatchString(licenses.Name) {
			t.Fatalf("'%s' not matching expected app name.", licenses.Name)
		}
		if licenses.Name == "app1" && licenses.FeaturesToExclude != "feature1,feature2" {
			t.Fatalf("'%s' not matching expected feature1,feature2", licenses.FeaturesToExclude)
		}
		if licenses.Name == "app2" && licenses.FeaturesToInclude != "feature5,feature30" {
			t.Fatalf("'%s' not matching expected feature5,feature30", licenses.FeaturesToInclude)
		}
		if licenses.Name == "app3_domain1" && licenses.FeaturesToInclude != "" && licenses.FeaturesToExclude != "" {
			t.Fatalf("'%s' and '%s' expected to be empty", licenses.FeaturesToInclude, licenses.FeaturesToExclude)
		}
		if licenses.Name == "app3_domain2" && licenses.FeaturesToInclude != "" && licenses.FeaturesToExclude != "" {
			t.Fatalf("'%s' and '%s' expected to be empty", licenses.FeaturesToInclude, licenses.FeaturesToExclude)
		}
	}
}
