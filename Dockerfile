FROM centos:8
LABEL maintainer="Mario Trangoni <mjtrangoni@gmail.com>"

# Install dependencies
RUN rpm --import /etc/pki/rpm-gpg/RPM-GPG-KEY-centosofficial && \
    dnf -y install bash-completion redhat-lsb-core strace && \
    # Clean cache
    dnf -y clean all

COPY flexlm_exporter /bin/flexlm_exporter

EXPOSE      9319
USER        nobody
ENTRYPOINT  [ "/bin/flexlm_exporter" ]
