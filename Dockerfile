FROM docker.io/rockylinux/rockylinux:8
LABEL maintainer="Mario Trangoni <mjtrangoni@gmail.com>"

# Install dependencies and clean cache
RUN rpm --import /etc/pki/rpm-gpg/RPM-GPG-KEY-rockyofficial && \
    dnf -y install bash-completion redhat-lsb-core strace && \
    dnf -y clean all

COPY flexlm_exporter /bin/flexlm_exporter

EXPOSE      9319
USER        nobody
ENTRYPOINT  [ "/bin/flexlm_exporter" ]
