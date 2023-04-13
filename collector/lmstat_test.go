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
	"os"
	"testing"

	"github.com/go-kit/log"
)

const (
	v117String                  = "v11.7"
	v11130String                = "v11.13.0"
	testParseLmstatVersionNew   = "fixtures/lmstat_new.txt"
	testParseLmstatVersionOld   = "fixtures/lmstat_old.txt"
	testParseLmstatLicenseInfo1 = "fixtures/lmstat_app1.txt"
	testParseLmstatServerDown   = "fixtures/lmstat_server_down.txt"
	testParseLmstatServerUp     = "fixtures/lmstat_server_up_win.txt"
)

func TestContains(t *testing.T) {
	t.Parallel()

	containsOut := contains([]string{"a", "b"}, "b")
	if !containsOut {
		t.Fatalf("contains = %t - expected true", containsOut)
	}

	containsOut = contains([]string{"a", "b"}, "c")
	if containsOut {
		t.Fatalf("contains = %t - expected false", containsOut)
	}
}

func TestParseLmstatVersion(t *testing.T) {
	t.Parallel()

	dataByte, err := os.ReadFile(testParseLmstatVersionNew)
	if err != nil {
		t.Fatal(err)
	}

	dataStr, err := splitOutput(dataByte)
	if err != nil {
		t.Fatal(err)
	}

	lmstatInfo := parseLmstatVersion(dataStr)
	if lmstatInfo.arch != "x64_lsb" || lmstatInfo.build != "188735" || lmstatInfo.version != "v11.14.0.1" {
		t.Fatalf("Unexpected values %s, %s, %s != x64_lsb, 188735, v11.14.0.1", lmstatInfo.arch, lmstatInfo.build, lmstatInfo.version)
	}

	dataByte, err = os.ReadFile(testParseLmstatVersionOld)
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

	t.Parallel()

	dataByte, err = os.ReadFile(testParseLmstatLicenseInfo1)
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
			if info.version != v117String || info.master ||
				!info.status {
				t.Fatalf("Unexpected values for %s: %s, %t, %t",
					info.fqdn, info.version, info.master, info.status)
			}
		} else if info.fqdn == "host2.domain.net" {
			if info.version != v117String || !info.master ||
				!info.status {
				t.Fatalf("Unexpected values for %s: %s, %t, %t",
					info.fqdn, info.version, info.master, info.status)
			}
		}
	}

	dataByte, err = os.ReadFile(testParseLmstatServerDown)
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
			if info.version != v11130String || info.master ||
				!info.status {
				t.Fatalf("Unexpected values for %s: %s, %t, %t",
					info.fqdn, info.version, info.master, info.status)
			}
		} else if info.fqdn == "host2" {
			if info.version != v11130String || !info.master ||
				!info.status {
				t.Fatalf("Unexpected values for %s: %s, %t, %t",
					info.fqdn, info.version, info.master, info.status)
			}
		} else if info.fqdn == "host3" {
			if info.version != "" || info.master || info.status {
				t.Fatalf("Unexpected values for %s: %s, %t, %t",
					info.fqdn, info.version, info.master, info.status)
			}
		}
	}

	dataByte, err = os.ReadFile(testParseLmstatServerUp)
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
			!info.master || !info.status {
			t.Fatalf("Unexpected values for %s: %s, %t, %t",
				info.fqdn, info.version, info.master, info.status)
		}
	}
}

func TestParseLmstatLicenseInfoVendor(t *testing.T) {
	t.Parallel()

	dataByte, err := os.ReadFile(testParseLmstatLicenseInfo1)
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
			if !info.status || info.version != "v11.6" {
				t.Fatalf("Unexpected values for %s: %t, %s", name, info.status,
					info.version)
			}
		} else {
			t.Fatalf("Unexpected feature: %s", name)
		}
	}
}

func TestParseLmstatLicenseInfoFeature(t *testing.T) {
	t.Parallel()

	dataByte, err := os.ReadFile(testParseLmstatLicenseInfo1)
	if err != nil {
		t.Fatal(err)
	}

	dataStr, err := splitOutput(dataByte)
	if err != nil {
		t.Fatal(err)
	}

	features, licUsersByFeature, reservGroupByFeature := parseLmstatLicenseInfoFeature(dataStr, log.NewNopLogger())
	for name, info := range features {
		if name == "feature11" {
			if info.issued != 16384 || info.used != 80 {
				t.Fatalf("Unexpected values for %s: %v!=16384 %v!=80", name,
					info.issued, info.used)
			}
		}
	}

	foundUser11 := false

	const (
		licUsed1  = 1
		licUsed8  = 8
		licUsed12 = 12
		licUsed16 = 16
		licUsed26 = 26
	)

	for username, licused := range licUsersByFeature["feature34"] {
		for i := range licused {
			if username == "user1" {
				if licused[i].num != licUsed16 {
					t.Fatalf("Unexpected values for feature34[%s]: %v!=16",
						username, licused[i].num)
				}
			} else if username == "user11" {
				foundUser11 = true
				if licused[i].num != licUsed26 {
					t.Fatalf("Unexpected values for feature34[%s]: %v!=26",
						username, licused[i].num)
				}
			} else if username == "user17" {
				if licused[i].num != licUsed12 {
					t.Fatalf("Unexpected values for feature34[%s]: %v!=12",
						username, licused[i].num)
				}
			}
		}
	}

	if !foundUser11 {
		t.Fatalf("Couldn't parse user \"user11\" from feature34")
	}

	foundCmfy211 := false

	for username, licused := range licUsersByFeature["feature31"] {
		for i := range licused {
			if username == "user33" {
				if licused[i].num != licUsed16 {
					t.Fatalf("Unexpected values for feature31[%s]: %v!=16",
						username, licused[i].num)
				}
			} else if username == "cmfy211" {
				foundCmfy211 = true
				if licused[i].num != licUsed1 {
					t.Fatalf("Unexpected values for feature31[%s]: %v!=1",
						username, licused[i].num)
				}
			} else if username == "cmfy212" {
				if licused[i].num != licUsed16 {
					t.Fatalf("Unexpected values for feature31[%s]: %v!=16",
						username, licused[i].num)
				}
			}
		}
	}

	if !foundCmfy211 {
		t.Fatalf("Couldn't parse user \"cmfy211\" from feature31")
	}

	var (
		found        = false
		foundJohnDoe = false
		foundJaneDoe = false
	)

	for username, licused := range licUsersByFeature["feature100"] {
		for i := range licused {
			if username == "user13" {
				if licused[i].num != licUsed1 {
					t.Fatalf("Unexpected values for feature1[%s]: %v!=1",
						username, licused[i].num)
				}
			} else if username == "Administrator" {
				// There is 2 users, and this should always enter here.
				found = true
			} else if username == "John Doe" {
				foundJohnDoe = true
			} else if username == "Jane Doe Jr." {
				foundJaneDoe = true
			}
		}
	}
	if !found {
		t.Fatalf("Couldn't parse user Administrator from feature100")
	}
	if !foundJohnDoe {
		t.Fatalf("Couldn't parse user \"John Doe\" from feature100")
	}
	if !foundJaneDoe {
		t.Fatalf("Couldn't parse user \"Jane Doe Jr.\" from feature100")
	}

	if licUsersByFeature["feature12"] != nil {
		t.Fatalf("Unexpected value for feature12: shouldn't match any user")
	}

	for group, licreserv := range reservGroupByFeature["feature38"] {
		if group == "GROUP10" {
			if licreserv != licUsed8 {
				t.Fatalf("Unexpected values for feature38[%s]: %v!=8", group,
					licreserv)
			}
		}
	}
	if reservGroupByFeature["feature11"] != nil {
		t.Fatalf("Unexpected value for feature11: shouldn't match any reservation")
	}
}

func TestParseLmstatLicenseInfoUserSince(t *testing.T) {
	t.Parallel()

	dataByte, err := os.ReadFile(testParseLmstatLicenseInfo1)
	if err != nil {
		t.Fatal(err)
	}

	dataStr, err := splitOutput(dataByte)
	if err != nil {
		t.Fatal(err)
	}

	_, licUsersByFeature, _ := parseLmstatLicenseInfoFeature(dataStr, log.NewNopLogger())

	const (
		sinceUser1  = "Fri 10/20 14:12"
		sinceUser17 = "Fri 10/20 12:36"
	)

	for username, licused := range licUsersByFeature["feature34"] {
		for i := range licused {
			if username == "user1" {
				if licused[i].since != sinceUser1 {
					t.Fatalf("Unexpected values for feature34[%s]: %s!=%s",
						username, licused[i].since, sinceUser1)
				}
			} else if username == "user17" {
				if licused[i].since != sinceUser17 {
					t.Fatalf("Unexpected values for feature34[%s]: %s!=%s",
						username, licused[i].since, sinceUser17)
				}
			}
		}
	}
}
