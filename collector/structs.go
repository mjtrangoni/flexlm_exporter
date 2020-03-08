// Copyright 2017-2018 Mario Trangoni
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

type lmstatInformation struct {
	arch    string
	build   string
	version string
}

type server struct {
	fqdn    string
	port    string
	version string
	status  bool
	master  bool
}

type vendor struct {
	status  bool
	version string
}

type feature struct {
	issued float64
	used   float64
}

type featureExp struct {
	name     string
	expires  float64
	licenses string
	vendor   string
	version  string
}

type aggrFeaturesExp struct {
	app      string
	features int
	licenses int
}
