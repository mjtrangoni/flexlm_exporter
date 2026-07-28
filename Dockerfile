FROM docker.io/rockylinux/rockylinux:9.3-minimal

LABEL maintainer="Mario Trangoni mjtrangoni@gmail.com"
LABEL org.opencontainers.image.source="https://github.com/mjtrangoni/flexlm_exporter"

ENV HOME=/home/exporter

RUN microdnf -y update &&
microdnf -y install shadow-utils bash-completion strace glibc &&
microdnf -y clean all &&
ln -s /lib64/ld-linux-x86-64.so.2 /lib64/ld-lsb-x86-64.so.3 &&
groupadd -g 30001 exporter &&
useradd --no-log-init -m -d /home/exporter -u 30001 -g 30001 exporter &&
mkdir -p /home/exporter/config &&
chown -R exporter:exporter /home/exporter/config

COPY flexlm_exporter /bin/flexlm_exporter

WORKDIR /home/exporter
USER exporter

EXPOSE 9319

ENTRYPOINT [ "/bin/flexlm_exporter" ]