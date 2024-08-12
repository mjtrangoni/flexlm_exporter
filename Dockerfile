FROM docker.io/rockylinux/rockylinux:8
LABEL maintainer="Mario Trangoni <mjtrangoni@gmail.com>"
LABEL org.opencontainers.image.source="https://github.com/mjtrangoni/flexlm_exporter"

# Install dependencies and clean cache
RUN rpm --import /etc/pki/rpm-gpg/RPM-GPG-KEY-rockyofficial && \
    dnf -y update && \
    dnf -y install bash-completion redhat-lsb-core strace && \
    dnf -y clean all && \
    rm -f /etc/pki/tls/private/postfix.key

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
