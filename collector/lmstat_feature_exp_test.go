// Copyright 2020 Mario Trangoni
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
	"math"
	"os"
	"testing"

	"github.com/prometheus/common/promslog"
)

const (
	feature12String                       = "feature12"
	v201812String                         = "2018.12"
	vendor2String                         = "vendor2"
	testParseLmstatLicenseFeatureExpDate1 = "fixtures/lmstat_i_app1.txt"
	testParseLmstatLicenseFeatureExpDate2 = "fixtures/lmstat_i_app2.txt"
)

func TestParseLmstatLicenseFeatureExpDate1(t *testing.T) {
	t.Parallel()

	dataByte, err := os.ReadFile(testParseLmstatLicenseFeatureExpDate1)
	if err != nil {
		t.Fatal(err)
	}

	dataStr, err := splitOutput(dataByte)
	if err != nil {
		t.Fatal(err)
	}

	logger := promslog.New(&promslog.Config{})
	featuresExp := parseLmstatLicenseFeatureExpDate(dataStr, logger)
	found := false

	for index, feature := range featuresExp {
		if feature.name == "feature_11" {
			if feature.version != v201812String ||
				feature.licenses != "150" ||
				feature.expires != 1546214400 ||
				feature.vendor != vendor2String {
				t.Fatalf("Unexpected values %s, %s, %s, %s, != %f",
					feature.name, feature.version,
					feature.licenses, feature.vendor,
					feature.expires)
			}
		} else if feature.name == feature12String && index == 12 {
			if feature.version != v201812String ||
				feature.licenses != "50" ||
				feature.expires != 1546214400 ||
				feature.vendor != vendor2String {
				t.Fatalf("Unexpected values %s, %d, %s, %s, %s, != %f",
					feature.name, index,
					feature.version, feature.licenses,
					feature.vendor, feature.expires)
			}
		} else if feature.name == feature12String && index == 13 {
			if feature.version != v201812String ||
				feature.licenses != "2" ||
				feature.expires != 1538265600 ||
				feature.vendor != vendor2String {
				t.Fatalf("Unexpected values %s, %d, %s, %s, %s, != %f",
					feature.name, index,
					feature.version, feature.licenses,
					feature.vendor, feature.expires)
			}
		} else if feature.name == "feature15" {
			if feature.version != "2018.09" ||
				feature.licenses != "2" ||
				feature.expires != math.Inf(posInfinity) ||
				feature.vendor != vendor2String {
				t.Fatalf("Unexpected values %s, %s, %s, %s, != %f",
					feature.name, feature.version,
					feature.licenses, feature.vendor,
					feature.expires)
			}
		} else if feature.name == "feature16" {
			if feature.version != "0.1" ||
				feature.licenses != "1" ||
				feature.expires != math.Inf(posInfinity) ||
				feature.vendor != vendor2String {
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

func TestParseLmstatLicenseFeatureExpDate2(t *testing.T) {
	t.Parallel()

	dataByte, err := os.ReadFile(testParseLmstatLicenseFeatureExpDate2)
	if err != nil {
		t.Fatal(err)
	}

	dataStr, err := splitOutput(dataByte)
	if err != nil {
		t.Fatal(err)
	}

	logger := promslog.New(&promslog.Config{})
	featuresExp := parseLmstatLicenseFeatureExpDate(dataStr, logger)
	found := false

	for index, feature := range featuresExp {
		if feature.name == "FEATURE_NAME_2_38_05" && index == 1 {
			if feature.version != "1.1" ||
				feature.licenses != "1" ||
				feature.expires != 1591920000 ||
				feature.vendor != "VENDOR" {
				t.Fatalf("Unexpected values %s, %s, %s, %s, != %f",
					feature.name, feature.version,
					feature.licenses, feature.vendor,
					feature.expires)
			}
		} else if feature.name == "feature_name_9" && index == 8 {
			if feature.version != "15.0" ||
				feature.licenses != "2" ||
				feature.expires != math.Inf(posInfinity) ||
				feature.vendor != "vendor" {
				t.Fatalf("Unexpected values %s, %s, %s, %s, != %f",
					feature.name, feature.version,
					feature.licenses, feature.vendor,
					feature.expires)
			}

			found = true
		}
	}
	if !found {
		t.Fatalf("feature_name_9 not found")
	}
}
