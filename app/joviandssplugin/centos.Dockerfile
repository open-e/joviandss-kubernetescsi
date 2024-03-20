FROM centos:stream9
LABEL maintainers="Andrei Perapiolkin"
LABEL description="JovianDSS CSI Plugin"

RUN yum -y install iscsi-initiator-utils ca-certificates e2fsprogs util-linux iproute
RUN mkdir -p /var/lib/kubelet/plugins_registry/joviandss-csi-driver/
COPY ./_output/jdss-csi-plugin /jdss-csi-plugin
ENTRYPOINT ["/jdss-csi-plugin"]
