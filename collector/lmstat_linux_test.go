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
	"io/ioutil"
	"testing"
)

const (
	testParseLmstatVersionNew   = "fixtures/lmstat_new.txt"
	testParseLmstatVersionOld   = "fixtures/lmstat_old.txt"
	testParseLmstatLicenseInfo1 = "fixtures/lmstat_app1.txt"
)

func TestParseLmstatVersion(t *testing.T) {
	dataByte, err := ioutil.ReadFile(testParseLmstatVersionNew)
	if err != nil {
		t.Fatal(err)
	}

	dataStr, err := splitOutput(dataByte)
	if err != nil {
		t.Fatal(err)
	}

	lmstatInfo = parseLmstatVersion(dataStr)
	if lmstatInfo.arch != "x64_lsb" || lmstatInfo.build != "188735" || lmstatInfo.version != "v11.14.0.1" {
		t.Fatalf("Unexpected values %s, %s, %s != x64_lsb, 188735, v11.14.0.1", lmstatInfo.arch, lmstatInfo.build, lmstatInfo.version)
	}

	dataByte, err = ioutil.ReadFile(testParseLmstatVersionOld)
	if err != nil {
		t.Fatal(err)
	}

	dataStr, err = splitOutput(dataByte)
	if err != nil {
		t.Fatal(err)
	}

	lmstatInfo = parseLmstatVersion(dataStr)
	if lmstatInfo.arch != notFound || lmstatInfo.build != notFound ||
		lmstatInfo.version != notFound {
		t.Fatalf("Unexpected values %s, %s, %s != %s", lmstatInfo.arch,
			lmstatInfo.build, lmstatInfo.version, notFound)
	}
}

func TestParseLmstatLicenseInfoServer(t *testing.T) {
	dataByte, err := ioutil.ReadFile(testParseLmstatLicenseInfo1)
	if err != nil {
		t.Fatal(err)
	}

	dataStr, err := splitOutput(dataByte)
	if err != nil {
		t.Fatal(err)
	}

	servers := parseLmstatLicenseInfoServer(dataStr)
	for _, info := range servers {
		if info.fqdn == "host1.domain.net" || info.fqdn == "host3.domain.net" {
			if info.version != "v11.7" || info.master != false ||
				info.status != true {
				t.Fatalf("Unexpected values for %s: %s, %t, %t",
					info.fqdn, info.version, info.master, info.status)
			}
		} else if info.fqdn == "host2.domain.net" {
			if info.version != "v11.7" || info.master != true ||
				info.status != true {
				t.Fatalf("Unexpected values for %s: %s, %t, %t",
					info.fqdn, info.version, info.master, info.status)
			}
		}
	}
}

func TestParseLmstatLicenseInfoVendor(t *testing.T) {
	dataByte, err := ioutil.ReadFile(testParseLmstatLicenseInfo1)
	if err != nil {
		t.Fatal(err)
	}

	dataStr, err := splitOutput(dataByte)
	if err != nil {
		t.Fatal(err)
	}

	vendors := parseLmstatLicenseInfoVendor(dataStr)
	for name, info := range vendors {
		if name == "VENDOR1" {
			if info.status != true || info.version != "v11.6" {
				t.Fatalf("Unexpected values for %s: %t, %s", name, info.status,
					info.version)
			}
		} else {
			t.Fatalf("Unexpected feature: %s", name)
		}
	}
}
