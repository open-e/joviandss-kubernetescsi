FROM centos:stream9

LABEL maintainers="Andrei Perapiolkin"
LABEL description="JovianDSS CSI Plugin"

#RUN yum -y install iscsi-initiator-utils ca-certificates e2fsprogs util-linux iproute
RUN yum -y install ca-certificates e2fsprogs util-linux iproute wget zip bzip2 socat e2fsprogs exfatprogs xfsprogs dosfstools cifs-utils gdisk rsync procps util-linux nvme-cli fuse3
#RUN yum -y install netbase btrfs-progs fatresize ntfs-3g nfs-common fdisk cloud-guest-utils
RUN mkdir -p /var/lib/kubelet/plugins_registry/joviandss-csi-driver/

COPY ./_output/jdss-csi-plugin /usr/local/bin/jdss-csi-plugin
COPY ./_output/jdss-csi-cli /usr/local/bin/jdss-csi-cli

COPY ./deploy/container/scripts/iscsiadm /usr/local/bin/
COPY ./deploy/container/scripts/donothing.sh /usr/local/bin/
COPY ./deploy/container/bin/csc /usr/local/bin/

WORKDIR /root

#ENTRYPOINT ["/usr/local/bin/donothing.sh"]
ENTRYPOINT ["/usr/local/bin/jdss-csi-plugin"]
