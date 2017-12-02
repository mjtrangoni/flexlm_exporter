# FLEXlm Exporter [![Build Status](https://travis-ci.org/mjtrangoni/flexlm_exporter.svg)][travis]

[![Coverage Status](https://coveralls.io/repos/github/mjtrangoni/flexlm_exporter/badge.svg?branch=master)](https://coveralls.io/github/mjtrangoni/flexlm_exporter?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/mjtrangoni/flexlm_exporter)](https://goreportcard.com/report/github.com/mjtrangoni/flexlm_exporter)

[Prometheus](https://prometheus.io/) exporter for FLEXlm License Manager
`lmstat` license information.

## Getting

```
go get github.com/mjtrangoni/flexlm_exporter
```

## Building

```
cd $GOPATH/src/github.com/mjtrangoni/flexlm_exporter
make
```

## Configuration

This is an illustrative example of the configuration file in YAML format.

```
# FlexLM Licenses to be monitored.
---
licenses:
  - name: app1
    license_file: /usr/local/flexlm/licenses/license.dat.app1
    features_to_exclude: feature1,feature2
    monitor_users: True
    monitor_reservations: True
  - name: app2
    license_server: 28000@host1,28000@host2,28000@host3
    features_to_exclude: feature1,feature2
    monitor_users: True
    monitor_reservations: True
```

## Running

```
./flexlm_exporter <flags>
```

## What's exported?

 * `lmutil lmstat -v` information.
 * `lmutil lmstat -c license_file -a` or `lmutil lmstat -c license_server -a`
   license information.

## Dashboards

TODO

## Contributing

Refer to [CONTRIBUTING.md](https://github.com/mjtrangoni/flexlm_exporter/blob/master/CONTRIBUTING.md)

## License

Apache License 2.0, see [LICENSE](https://github.com/mjtrangoni/mjtrangoni/blob/master/LICENSE).

[travis]: https://travis-ci.org/mjtrangoni/flexlm_exporter
