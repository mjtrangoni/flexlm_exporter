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

type lmstatInformation struct {
	arch    string
	build   string
	version string
}

// Take as reference the structures design from
// http://search.cpan.org/~odenbach/Flexnet-lmutil/lib/Flexnet/lmutil.pm

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

//feature points to a structure like
// 'feature' => {
//     'MATLAB' => {
//         'reservations' => [
//             {
//                 'reservations' => '1',
//                 'group' => 'etechnik-labor',
//                 'type' => 'HOST_GROUP'
//             }
//         ],
//         'issued' => '115',
//         'used' => '36',
//         'users' => [
//             {
//                 'serverhost' => 'dabu.uni-paderborn.de',
//                 'startdate' => 'Wed 8/12 17:18',
//                 'port' => '27000',
//                 'licenses' => 1,
//                 'display' => 'bessel',
//                 'host' => 'bessel',
//                 'handle' => '4401',
//                 'user' => 'hangmann'
//             },
//         ]
//     },
// },
// ...

type feature struct {
	issued float64
	used   float64
}
