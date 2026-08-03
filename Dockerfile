FROM quay.io/rockylinux/rockylinux:9.8-minimal
LABEL maintainer="Mario Trangoni <mjtrangoni@gmail.com>"
LABEL org.opencontainers.image.source="https://github.com/mjtrangoni/flexlm_exporter"

# Install dependencies and clean cache
RUN microdnf -y update && \
    microdnf -y install bash-completion strace && \
    microdnf -y clean all && \
    ln -s /lib64/ld-linux-x86-64.so.2 /lib64/ld-lsb-x86-64.so.3

COPY flexlm_exporter /bin/flexlm_exporter

# Add exporter user and group
RUN groupadd -g 30001 exporter && \
  useradd --no-log-init -m -d /home/exporter -u 30001 -g 30001 exporter

EXPOSE      9319
USER        exporter
WORKDIR     /home/exporter

RUN mkdir -p /home/exporter/config &&\
  chown -R 30001:30001 /home/exporter/config

# Default home dir
ENV HOME=/home/exporter

ENTRYPOINT  [ "/bin/flexlm_exporter" ]
