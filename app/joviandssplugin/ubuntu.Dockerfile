FROM ubuntu:18.04
LABEL maintainers="Andrei Perepiolkin"
LABEL description="JovianDSS CSI Plugin"

RUN mkdir -p /run/lock/iscsi
RUN apt-get update -y
RUN apt-get install -y util-linux open-iscsi e2fsprogs iproute2
COPY ./_output/jdss-csi-plugin /jdss-csi-plugin
ENTRYPOINT ["/jdss-csi-plugin"]
