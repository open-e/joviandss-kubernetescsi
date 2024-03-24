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

1. Applying patch to present config localy
```bash
talosctl machineconfig patch node.yaml --patch @securitypatch.yaml -o node1.yaml
```

2. Uploading config to Talos

```yaml
talosctl apply-config -n node.mycluster.lan --file node1.yaml
```
