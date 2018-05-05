FROM centos:latest
LABEL maintainer="Mario Trangoni <mjtrangoni@gmail.com>"

# Install dependencies
RUN rpm --import /etc/pki/rpm-gpg/RPM-GPG-KEY-CentOS-7 && \
    yum -y install bash-completion redhat-lsb-core strace &&\
    # Clean cache
    yum -y clean all

COPY flexlm_exporter /bin/flexlm_exporter

EXPOSE      9319
USER        nobody
ENTRYPOINT  [ "/bin/flexlm_exporter" ]
