FROM        quay.io/prometheus/busybox:glibc
MAINTAINER  Mario Trangoni <mjtrangoni@gmail.com>

COPY flexlm_exporter /bin/flexlm_exporter

EXPOSE      9319
USER        nobody
ENTRYPOINT  [ "/bin/flexlm_exporter" ]
