# FLEXlm Exporter ![Build Status](https://github.com/mjtrangoni/flexlm_exporter/workflows/Build/badge.svg)

[![CircleCI](https://circleci.com/gh/mjtrangoni/flexlm_exporter.svg?style=svg)](https://circleci.com/gh/mjtrangoni/flexlm_exporter)
[![Docker Repository on Quay](https://quay.io/repository/mjtrangoni/flexlm_exporter/status)][quay]
[![Docker Pulls](https://badgen.net/docker/pulls/mjtrangoni/flexlm_exporter?icon=docker)][hub]
[![Go Reference](https://pkg.go.dev/badge/github.com/mjtrangoni/flexlm_exporter.svg)](https://pkg.go.dev/github.com/mjtrangoni/flexlm_exporter)
[![Coverage Status](https://coveralls.io/repos/github/mjtrangoni/flexlm_exporter/badge.svg?branch=main)](https://coveralls.io/github/mjtrangoni/flexlm_exporter?branch=main)
[![Go Report Card](https://goreportcard.com/badge/github.com/mjtrangoni/flexlm_exporter)](https://goreportcard.com/report/github.com/mjtrangoni/flexlm_exporter)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://raw.githubusercontent.com/mjtrangoni/flexlm_exporter/main/LICENSE)
[![StyleCI](https://github.styleci.io/repos/107779392/shield?branch=main)](https://github.styleci.io/repos/107779392)

[Prometheus](https://prometheus.io/) exporter for FLEXlm License Manager
`lmstat` license information.

If you are looking for other License Managers metrics (like rlm,lmx, and mathlm),
please check the [License Manager Exporter](https://codeberg.org/Zauberbutter/license_manager_exporter/).

## Install

```console
go install github.com/mjtrangoni/flexlm_exporter
```

## Building

```console
cd $GOPATH/src/github.com/mjtrangoni/flexlm_exporter
make
```

## Configuration

This is an illustrative example of the configuration file in YAML format.

```yaml
# FlexLM Licenses to be monitored.
---
licenses:
  - name: app1
    license_file: /usr/local/flexlm/licenses/license.dat.app1
    features_to_exclude: feature1,feature2
    monitor_users: True
    monitor_reservations: True
    monitor_versions: False
  - name: app2
    license_server: 28000@host1,28000@host2,28000@host3
    features_to_include: feature5,feature30
    monitor_users: True
    monitor_reservations: True
    monitor_versions: False
```

Notes:

 1. It is possible to define a license with a path in `license_file`, that has to
 be readable from the exporter instance, **or** with `license_server` in a
 `port@host` combination format.
 2. You can exclude some features from exporting with `features_to_exclude`,
 **or** export some defined and exclude the rest with `feature_to_include`.

## Running

```console
./flexlm_exporter <flags>
```

### Docker images

Docker images are available on,

 1. [Quay.io](https://quay.io/repository/mjtrangoni/flexlm_exporter).
    `$ docker pull quay.io/mjtrangoni/flexlm_exporter:latest`
 1. [Docker](https://hub.docker.com/r/mjtrangoni/flexlm_exporter/).
    `$ docker pull mjtrangoni/flexlm_exporter:latest`
 1. [GHCR](https://github.com/mjtrangoni/flexlm_exporter/pkgs/container/flexlm_exporter/).
    `$ docker pull ghcr.io/mjtrangoni/flexlm_exporter:latest`

Please make sure that SELinux is not running in your host, or run the container
as root.

You can launch a *flexlm_exporter* container with,

```console
$ export DOCKER_REPOSITORY="quay.io/mjtrangoni/flexlm_exporter:latest"
$ export LMUTIL_LOCAL="PATH where your lmutil binary is located"
$ export CONFIG_PATH_LOCAL="PATH where your exporter config file is located"
$ docker run --name flexlm_exporter -d -p 9319:9319 \
    --volume $LMUTIL_LOCAL:/usr/bin/flexlm/ \
    --volume $CONFIG_PATH_LOCAL:/home/exporter/config/licenses.yml \
    $DOCKER_REPOSITORY --path.lmutil="/usr/bin/flexlm/lmutil" \
    --path.config="/home/exporter/config/licenses.yml"
```

Metrics will now be reachable at <http://localhost:9319/metrics>.

## What's exported?

 1. `lmutil lmstat -v` information.
 1. `lmutil lmstat -c license_file -a` or `lmutil lmstat -c license_server -a`
   license information.
 1. `lmutil lmstat -c license_file -i` or `lmutil lmstat -c license_server -i`
   license features expiration date.

## Dashboards

 1. [Grafana Dashboard](https://grafana.com/grafana/dashboards/3854-flexlm)

## Alerting

### Prometheus rules

Prometheus rules for alerting with [Prometheus Alertmanager](https://prometheus.io/docs/alerting/alertmanager/).

```yaml

groups:
- name: FlexLM
  rules:
  - alert: FlexLmServerDown
    expr: flexlm_server_status == 0
    for: 5m
    labels:
      severity: error
    annotations:
      summary: "Flexlm Error (instance {{ $labels.instance }})"
      description: "FlexLm {{ $labels.collector }} was not successful\n  VALUE = {{ $value }}\n  LABELS: {{ $labels }}"
  - alert: LicenceAvailable
    expr: 100*(flexlm_feature_used / flexlm_feature_issued) > 95
    for: 5m
    labels:
      severity: warning
    annotations:
      summary: "Licence Available Status (instance {{ $labels.instance }})"
      description: "Licence fully used \n  VALUE = {{ $value }}\n  LABELS: {{ $labels }}"
  - alert: LicenseExpiring
    expr: ((flexlm_feature_expiration_seconds - time()) / 86400) < 14
    for: 30m
    labels:
      severity: warning
    annotations:
      summary: License {{ $labels.app }} expiring soon on {{ $labels.instance }}
      description: License {{ $labels.app }} on {{ $labels.instance }} has {{ $labels.features }} features ({{ $labels.licenses }} licenses) expiring in {{ $value }} days
```

## Contributing

Refer to [CONTRIBUTING.md](https://github.com/mjtrangoni/flexlm_exporter/blob/main/CONTRIBUTING.md)

## License

Apache License 2.0, see [LICENSE](https://github.com/mjtrangoni/mjtrangoni/blob/main/LICENSE).

[hub]: https://hub.docker.com/r/mjtrangoni/flexlm_exporter/
[quay]: https://quay.io/repository/mjtrangoni/flexlm_exporter
