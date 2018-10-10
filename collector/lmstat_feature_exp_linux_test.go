// Copyright 2018 Mario Trangoni
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
	"io/ioutil"
	"math"
	"testing"
)

const (
	testParseLmstatLicenseFeatureExpDate1 = "fixtures/lmstat_i_app1.txt"
)

func TestParseLmstatLicenseFeatureExpDate(t *testing.T) {
	dataByte, err := ioutil.ReadFile(testParseLmstatLicenseFeatureExpDate1)
	if err != nil {
		t.Fatal(err)
	}

	dataStr, err := splitOutput(dataByte)
	if err != nil {
		t.Fatal(err)
	}

	featuresExp := parseLmstatLicenseFeatureExpDate(dataStr)
	found := false
	for index, feature := range featuresExp {
		if feature.name == "feature_11" {
			if feature.version != "2018.12" ||
				feature.licenses != "150" ||
				feature.expires != 1546214400 ||
				feature.vendor != "vendor2" {
				t.Fatalf("Unexpected values %s, %s, %s, %s, != %f",
					feature.name, feature.version,
					feature.licenses, feature.vendor,
					feature.expires)
			}
		} else if feature.name == "feature12" && index == 12 {
			if feature.version != "2018.12" ||
				feature.licenses != "50" ||
				feature.expires != 1546214400 ||
				feature.vendor != "vendor2" {
				t.Fatalf("Unexpected values %s, %d, %s, %s, %s, != %f",
					feature.name, index,
					feature.version, feature.licenses,
					feature.vendor, feature.expires)
			}
		} else if feature.name == "feature12" && index == 13 {
			if feature.version != "2018.12" ||
				feature.licenses != "2" ||
				feature.expires != 1538265600 ||
				feature.vendor != "vendor2" {
				t.Fatalf("Unexpected values %s, %d, %s, %s, %s, != %f",
					feature.name, index,
					feature.version, feature.licenses,
					feature.vendor, feature.expires)
			}
		} else if feature.name == "feature15" {
			if feature.version != "2018.09" ||
				feature.licenses != "2" ||
				feature.expires != math.Inf(1) ||
				feature.vendor != "vendor2" {
				t.Fatalf("Unexpected values %s, %s, %s, %s, != %f",
					feature.name, feature.version,
					feature.licenses, feature.vendor,
					feature.expires)
			}
		} else if feature.name == "feature16" {
			if feature.version != "0.1" ||
				feature.licenses != "1" ||
				feature.expires != math.Inf(1) ||
				feature.vendor != "vendor2" {
				t.Fatalf("Unexpected values %s, %s, %s, %s, != %f",
					feature.name, feature.version,
					feature.licenses, feature.vendor,
					feature.expires)
			}
			found = true
		}
	}
	if !found {
		t.Fatalf("feature16 not found")
	}
}
