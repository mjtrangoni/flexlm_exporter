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
	"os"
	"path/filepath"
	"testing"

	"github.com/mjtrangoni/flexlm_exporter/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/promslog"
)

// stubLmutil writes a script that answers "lmstat -v" with a version banner and
// every other call with the given fixture, then points lmutilPath at it.
func stubLmutil(t *testing.T, fixture string) {
	t.Helper()

	abs, err := filepath.Abs(fixture)
	if err != nil {
		t.Fatal(err)
	}

	script := "#!/bin/sh\n" +
		"for a in \"$@\"; do\n" +
		"  if [ \"$a\" = \"-v\" ]; then\n" +
		"    echo 'lmstat v11.19.7 build 1 x64_lsb'\n" +
		"    exit 0\n" +
		"  fi\n" +
		"done\n" +
		"cat " + abs + "\n"

	path := filepath.Join(t.TempDir(), "lmutil")
	if err := os.WriteFile(path, []byte(script), 0o700); err != nil { //nolint:gosec // test stub must be executable
		t.Fatal(err)
	}

	old := *lmutilPath
	*lmutilPath = path

	t.Cleanup(func() { *lmutilPath = old })
}

// TestGatherDuplicateUserLabels guards the flexlm_feature_used_users metric,
// which carries no version label. A user holding the same feature under two
// client versions since the same start time produced two metrics with identical
// labels, and Gather() then failed the whole scrape with "was collected before
// with the same name and label values".
func TestGatherDuplicateUserLabels(t *testing.T) {
	stubLmutil(t, "fixtures/lmstat_app8.txt")

	oldCfg := LicenseConfig

	LicenseConfig = config.Configuration{Licenses: []config.License{{
		Name:                "Vendor Sim",
		LicenseFile:         "1717@SERVER1",
		FeaturesToInclude:   "msimhdlsim",
		MonitorUsers:        true,
		MonitorReservations: true,
		MonitorVersions:     false,
	}}}

	t.Cleanup(func() { LicenseConfig = oldCfg })

	c, err := NewLmstatCollector(promslog.New(&promslog.Config{}))
	if err != nil {
		t.Fatal(err)
	}

	reg := prometheus.NewPedanticRegistry()
	if err := reg.Register(wrapCollector{c}); err != nil {
		t.Fatal(err)
	}

	families, err := reg.Gather()
	if err != nil {
		t.Fatalf("Gather() failed: %v", err)
	}

	// user1 holds 3 licenses across two versions, user2 holds 1. Without the
	// per-start-time aggregation user1 would appear twice.
	want := map[string]float64{"user1": 3, "user2": 1}
	got := map[string]float64{}

	for _, mf := range families {
		if mf.GetName() != "flexlm_feature_used_users" {
			continue
		}

		for _, m := range mf.GetMetric() {
			for _, l := range m.GetLabel() {
				if l.GetName() == "user" {
					if _, dup := got[l.GetValue()]; dup {
						t.Fatalf("user %s emitted more than once", l.GetValue())
					}

					got[l.GetValue()] = m.GetGauge().GetValue()
				}
			}
		}
	}

	if len(got) != len(want) {
		t.Fatalf("unexpected users: got %v want %v", got, want)
	}

	for user, num := range want {
		if got[user] != num {
			t.Errorf("user %s: got %v licenses, want %v", user, got[user], num)
		}
	}
}

// wrapCollector adapts the internal Collector interface to prometheus.Collector.
type wrapCollector struct{ c Collector }

// Describe sends nothing, making this an unchecked collector. The registry
// still rejects duplicate label sets at Gather() time, which is what we test.
func (w wrapCollector) Describe(_ chan<- *prometheus.Desc) {}

func (w wrapCollector) Collect(ch chan<- prometheus.Metric) {
	if err := w.c.Update(ch); err != nil {
		panic(err)
	}
}
