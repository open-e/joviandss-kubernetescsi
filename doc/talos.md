# Talos Linux Support

Talos Linux has no package manager and enforces strict security defaults, so additional
actions have to be taken before installing the JovianDSS CSI plugin, further `plugin`.

Instructions below are written for Talos 1.14. The `plugin` itself has been verified on
Talos 1.6 only.

All steps apply to every node that runs the `plugin` node service. By default these are the
worker nodes: the node DaemonSet has no tolerations, so it does not run on control plane
nodes unless they accept workloads (`cluster.allowSchedulingOnControlPlanes`).

## 1. iSCSI tools extension

The `plugin` relies on the iSCSI daemon `iscsid` and the `iscsiadm` tool provided by the host.
Talos does not ship them; they come from the `siderolabs/iscsi-tools` system extension, which
has to be part of the installed image.

Further information:

1. [Boot assets](https://www.talos.dev/latest/talos-guides/install/boot-assets/)
2. [System Extensions](https://www.talos.dev/latest/talos-guides/configuration/system-extensions/)
3. [Talos Image Factory](https://factory.talos.dev/)

Create a schematic in the [Image Factory](https://factory.talos.dev/) with the extension:

```yaml
customization:
  systemExtensions:
    officialExtensions:
      - siderolabs/iscsi-tools
```

For a new node, boot from the factory ISO and install with the factory installer image, both
built for that schematic:

```
https://factory.talos.dev/image/<schematic_id>/<talos_version>/metal-amd64.iso
factory.talos.dev/metal-installer/<schematic_id>:<talos_version>
```

For a running node, switch it to the installer image with the extension:

```bash
talosctl upgrade -n node1.my-talos-cluster.lan \
    --image factory.talos.dev/metal-installer/<schematic_id>:<talos_version>
```

For instance

```bash
talosctl upgrade -n node1.my-talos-cluster.lan \
    --image factory.talos.dev/metal-installer/c9078f9419961640c712a8bf2bb9174933dfcf1da383fd8ea2b7dc21493f8bac:v1.14.2
```

Once the extension is installed, the node reports it and runs the `ext-iscsid` service:

```bash
talosctl get extensions -n node2.my-talos-cluster.lan
talosctl get services -n node2.my-talos-cluster.lan
```

```
NODE                         NAMESPACE   TYPE              ID   VERSION   NAME          VERSION
node2.my-talos-cluster.lan   runtime     ExtensionStatus   0    1         iscsi-tools   <extension version>

NODE                         NAMESPACE   TYPE      ID           VERSION   RUNNING   HEALTHY   HEALTH UNKNOWN
node2.my-talos-cluster.lan   runtime     Service   ext-iscsid   1         true      false     true
```

Listings above show only the extension and service required by the `plugin`; the node lists
more.

## 2. Machine configuration patches

Two patches are needed, both in `deploy/talos/`:

- `modulepatch.yaml`: loads the `iscsi_tcp` kernel module on every boot.

  ```yaml
  apiVersion: v1alpha1
  kind: KernelModuleConfig
  name: iscsi_tcp
  ```

- `workload-isolation-patch.yaml`: Talos 1.14 runs kubelet and all pods in an isolated
  mount/PID namespace by default. The `plugin` node service runs `iscsiadm` of the host and
  mounts host paths; it has not been verified with workload isolation enabled, so disable it.

  ```yaml
  apiVersion: v1alpha1
  kind: SecurityProfileConfig
  workloadIsolation: false
  ```

Apply the patches to running nodes:

```bash
talosctl patch machineconfig \
    -n node1.my-talos-cluster.lan,node2.my-talos-cluster.lan \
    --patch @deploy/talos/modulepatch.yaml \
    --patch @deploy/talos/workload-isolation-patch.yaml
```

Or, for a new cluster, add them when generating the configuration:

```bash
talosctl gen config my-talos-cluster https://<endpoint>:6443 \
    --config-patch-worker @deploy/talos/modulepatch.yaml \
    --config-patch-worker @deploy/talos/workload-isolation-patch.yaml
```

Use `--config-patch` instead of `--config-patch-worker` if control plane nodes run the
`plugin` too.

## 3. Pod Security

Talos enforces the Pod Security level `baseline` cluster-wide, which rejects the privileged
pods of the `plugin`. No Talos patch is needed for that: the namespace manifest
`deploy/joviandss/namespace.yaml` labels `joviandss-csi` with the `privileged` level, see
[installation guide](install.md).
