# Talos Linux Support

Due to specifics of architecture and enforced security policies of Talos OS additional actions have to be taken 
to to run JovianDSS CSI plugin.

## iSCSI

JovianDSS CSI plugin relies on iscsi service provided by host, but Talos OS do not provide out of the box, at least not in 1.6 version.

To add iSCSI to your Talos distro user have to use Talos extensions.
Below are links for detailed information:
1. [Boot assets](https://www.talos.dev/latest/talos-guides/install/boot-assets/)
2. [System Extensions](https://www.talos.dev/latest/talos-guides/configuration/system-extensions/)
3. [Talos Image Factory](https://factory.talos.dev/)

If user is using [Image Factory](https://factory.talos.dev/) to generate OS image, result can be installed by running:

```bash
talosctl upgrade --image factory.talos.dev/installer/<new_schematic_id>:<version>
```

For instance
```bash
talosctl upgrade --image  factory.talos.dev/installer/c9078f9419961640c712a8bf2bb9174933dfcf1da383fd8ea2b7dc21493f8bac:v1.6.7 
```

If installation completed successfully user will be able to see `tgtd` and `iscsid` services.
Please notice that `iscsi` extension have to be present on all working nodes.

```bash
talosctl get service -n node2.my-talos-cluster.lan
```

```
NODE                         NAMESPACE   TYPE      ID           VERSION   RUNNING   HEALTHY   HEALTH UNKNOWN
cntr1.my-talos-cluster.lan   runtime     Service   ext-iscsid   1         true      false     true
cntr1.my-talos-cluster.lan   runtime     Service   ext-tgtd     1         true      false     true
```

and extensions

```bash
talosctl get extensions -n node2.my-talos-cluster.lan
```

```
NODE                         NAMESPACE   TYPE              ID   VERSION   NAME          VERSION
node2.my-talos-cluster.lan   runtime     ExtensionStatus   0    1         iscsi-tools   v0.1.4
```

Listings above emphasize only services and extensions that is required for `plugin` to operate and does not include all possible services that might be present on your cluster.


## Patching security

Talos OS security policies preventing direct interaction of containerised services with extensions and systems of Talos out of the box.
To give JovianDSS CSI plugin access to iSCSI extension and parts of Talos OS file system(required to attach volumes hosted by JovianDSS to user containers), user have to change Talos OS security policies.

Change can be done through patching [Patching Talos OS](https://www.talos.dev/v1.6/talos-guides/configuration/patching/).

Here is patch.

```yaml
cluster:
  apiServer:
    admissionControl:
      - name: PodSecurity
        configuration:
          exemptions:
            namespaces:
              - joviandss-csi
```

Assuming that user have current Talos node config file `node.yaml`, patch file with context provided above in the same directory and named `securitypatch.yaml` and node with FQDN or ip with name `node.myclaster.lan`
, patching can be done by:

Additionally user might want to ensure that `iscsi_tcp` will be loaded on machine start, that can be done by adding following patch
```yaml
machine:
    install:
        extraKernelArgs:
            - iscsi_tcp=1
    kernel:
        # Kernel modules to load.
        modules:
            - name: iscsi_tcp # Module name.
```


1. Applying patch to present configuration files of worker node and controller node of your running Talos cluster locally.
```bash
talosctl machineconfig patch worker.yaml --patch @securitypatch.yaml -o worker_cfg_v2.yaml
talosctl machineconfig patch controller.yaml --patch @securitypatch.yaml -o controller_cfg_v2.yaml
```
Ensuring module load on start 
```bash
talosctl machineconfig patch worker.yaml --patch @modulepatch.yaml -o worker_cfg_v2.yaml
```

2. Uploading config to Talos

```bash
talosctl apply-config -n node1.my-talos-cluster.lan,node2.my-talos-cluster.lan,...<and all other worker nodes you have in your cluster>... --file node_cfg_v2.yaml
talosctl apply-config -n cntr1.my-talos-cluster.lan,cntr2.my-talos-cluster.lan,...<and all other controller nodes you have in your cluster>... --file cntr_cfg_v2.yaml
```
```bash
kubectl label ns joviandss-csi pod-security.kubernetes.io/audit=privileged pod-security.kubernetes.io/enforce=privileged pod-security.kubernetes.io/warn=privileged
```
