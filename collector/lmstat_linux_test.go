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
	testParseLmstatServerDown   = "fixtures/lmstat_server_down.txt"
	testParseLmstatServerUpWin  = "fixtures/lmstat_server_up_win.txt"
)

func TestContains(t *testing.T) {
	containsOut := contains([]string{"a", "b"}, "b")
	if containsOut != true {
		t.Fatalf("contains = %t - expected true", containsOut)
	}
	containsOut = contains([]string{"a", "b"}, "c")
	if containsOut != false {
		t.Fatalf("contains = %t - expected false", containsOut)
	}
}

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
	var (
		err      error
		dataByte []byte
		dataStr  [][]string
	)

	dataByte, err = ioutil.ReadFile(testParseLmstatLicenseInfo1)
	if err != nil {
		t.Fatal(err)
	}

	dataStr, err = splitOutput(dataByte)
	if err != nil {
		t.Fatal(err)
	}

	servers := parseLmstatLicenseInfoServer(dataStr)
	for _, info := range servers {
		if info.fqdn == "host-1.domain.net" || info.fqdn == "host3.domain.net" {
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

	dataByte, err = ioutil.ReadFile(testParseLmstatServerDown)
	if err != nil {
		t.Fatal(err)
	}

	dataStr, err = splitOutput(dataByte)
	if err != nil {
		t.Fatal(err)
	}

	servers = parseLmstatLicenseInfoServer(dataStr)
	for _, info := range servers {
		if info.fqdn == "host1" {
			if info.version != "v11.13.0" || info.master != false ||
				info.status != true {
				t.Fatalf("Unexpected values for %s: %s, %t, %t",
					info.fqdn, info.version, info.master, info.status)
			}
		} else if info.fqdn == "host2" {
			if info.version != "v11.13.0" || info.master != true ||
				info.status != true {
				t.Fatalf("Unexpected values for %s: %s, %t, %t",
					info.fqdn, info.version, info.master, info.status)
			}
		} else if info.fqdn == "host3" {
			if info.version != "" || info.master != false ||
				info.status != false {
				t.Fatalf("Unexpected values for %s: %s, %t, %t",
					info.fqdn, info.version, info.master, info.status)
			}
		}
	}

	dataByte, err = ioutil.ReadFile(testParseLmstatServerUpWin)
	if err != nil {
		t.Fatal(err)
	}

	dataStr, err = splitOutput(dataByte)
	if err != nil {
		t.Fatal(err)
	}

	servers = parseLmstatLicenseInfoServer(dataStr)
	for _, info := range servers {
		if info.fqdn != "BVS15004" || info.version != "v11.12" ||
			info.master != true || info.status != true {
			t.Fatalf("Unexpected values for %s: %s, %t, %t",
				info.fqdn, info.version, info.master, info.status)
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

func TestParseLmstatLicenseInfoFeature(t *testing.T) {
	dataByte, err := ioutil.ReadFile(testParseLmstatLicenseInfo1)
	if err != nil {
		t.Fatal(err)
	}

	dataStr, err := splitOutput(dataByte)
	if err != nil {
		t.Fatal(err)
	}
	features, licUsersByFeature, reservGroupByFeature = parseLmstatLicenseInfoFeature(dataStr)
	for name, info := range features {
		if name == "feature11" {
			if info.issued != 16384 || info.used != 80 {
				t.Fatalf("Unexpected values for %s: %v!=16384 %v!=80", name,
					info.issued, info.used)
			}
		}
	}
	for username, licused := range licUsersByFeature["feature34"] {
		if username == "user1" {
			if licused != 16 {
				t.Fatalf("Unexpected values for feature34[%s]: %v!=16",
					username, licused)
			}
		} else if username == "user17" {
			if licused != 12 {
				t.Fatalf("Unexpected values for feature34[%s]: %v!=12",
					username, licused)
			}
		}
	}

	if licUsersByFeature["feature12"] != nil {
		t.Fatalf("Unexpected value for feature12: shouldn't match any user")
	}

	for group, licreserv := range reservGroupByFeature["feature38"] {
		if group == "GROUP10" {
			if licreserv != 8 {
				t.Fatalf("Unexpected values for feature38[%s]: %v!=8", group,
					licreserv)
			}
		}
	}
	if reservGroupByFeature["feature11"] != nil {
		t.Fatalf("Unexpected value for feature11: shouldn't match any reservation")
	}
}
