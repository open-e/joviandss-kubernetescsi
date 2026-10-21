# Pod Security for the JovianDSS CSI Plugin

The JovianDSS CSI plugin, further `plugin`, needs privileged pods. The node service runs with
`privileged: true`, adds the `SYS_ADMIN` capability, uses `hostNetwork` and mounts host paths.
Clusters that enforce Pod Security Standards reject such pods unless the `plugin` namespace
is allowed to run them.

This page compares the two ways to allow it:

1. **Namespace labels**: the namespace `joviandss-csi` declares its own Pod Security level.
2. **Admission exemption**: the API server configuration excludes the namespace from Pod
   Security checks (`deploy/talos/securitypatch.yaml` for Talos).

Recommendation: use namespace labels. Use an exemption only when the cluster operator does
not allow namespace labels to raise the level (see [When to use which](#4-when-to-use-which)).

## 1. Background: how Pod Security works today

Kubernetes checks pods with the built-in Pod Security Admission controller. It knows three
levels, from the most permissive to the most strict:

- `privileged`: no restrictions.
- `baseline`: blocks known privilege escalations: privileged containers, host namespaces
  (`hostNetwork`, `hostPID`, `hostIPC`), `hostPath` volumes and added capabilities such as
  `SYS_ADMIN`.
- `restricted`: `baseline` plus hardening rules (non-root user, dropped capabilities,
  seccomp profile).

Each level can be applied in three modes:

- `enforce`: pods that violate the level are rejected.
- `audit`: violations are recorded in the API server audit log; pods are allowed.
- `warn`: violations are returned to the client as warnings; pods are allowed.

The level for a namespace comes from, in order of precedence:

1. Labels on the namespace, `pod-security.kubernetes.io/<mode>: <level>`.
2. Cluster defaults in the `PodSecurity` admission configuration of the API server.

The same admission configuration can also list **exemptions**: usernames, runtime classes and
namespaces that are not checked at all.

The `plugin` node pods need `privileged`. Under `baseline` the DaemonSet controller cannot
create them: no pod appears, and the reason is only visible in the DaemonSet events
(`FailedCreate ... violates PodSecurity "baseline:latest"`).

## 2. The two approaches

### 2.1 Namespace labels

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: joviandss-csi
  labels:
    pod-security.kubernetes.io/enforce: privileged
    pod-security.kubernetes.io/audit: privileged
    pod-security.kubernetes.io/warn: privileged
```

Or on an existing namespace:

```bash
kubectl label namespace joviandss-csi \
    pod-security.kubernetes.io/enforce=privileged \
    pod-security.kubernetes.io/audit=privileged \
    pod-security.kubernetes.io/warn=privileged
```

- Part of the `plugin` manifests; installed and removed together with the `plugin`.
- Works on every Kubernetes distribution with Pod Security Admission, including managed
  services (EKS, GKE, AKS), where the API server configuration is not accessible.
- Needs only Kubernetes API access (`patch` on namespaces), no node or control plane access.
- Visible to anyone who inspects the namespace (`kubectl get ns joviandss-csi --show-labels`).

### 2.2 Admission exemption

On Talos, `deploy/talos/securitypatch.yaml` is a machine configuration patch for control plane
nodes:

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

On other distributions the same entry goes into the `AdmissionConfiguration` file passed to
`kube-apiserver --admission-control-config-file`, for example through `kubeadm`
`ClusterConfiguration` `apiServer.extraArgs` and `extraVolumes`.

- Changes the cluster, not the `plugin`: must be applied to every control plane node and
  restarts `kube-apiserver`.
- Needs node or control plane access (`talosctl`, SSH); not possible on managed services.
- The namespace is skipped completely: no `enforce`, but also no `audit` and no `warn`.
- Not visible from the namespace; to find it, read the API server configuration.
- Stays in place after the `plugin` is removed, until someone removes it.

On Talos the patch is merged with the default configuration: the namespace is added to the
exempted namespaces (`kube-system` by default) and the cluster defaults stay as they are.
Check the result in the rendered control plane configuration before applying it.

## 3. Comparison

| | Namespace labels | Admission exemption |
|---|---|---|
| Where it lives | `plugin` manifest (`namespace.yaml`) | API server configuration (Talos machine config) |
| Who can apply it | Kubernetes user allowed to label namespaces | Cluster operator with control plane access |
| Managed Kubernetes (EKS, GKE, AKS) | Yes | No |
| Talos | Yes, no machine configuration change | Yes, control plane patch |
| Needs API server restart | No | Yes |
| Removed with the `plugin` | Yes, with the namespace | No, manual cleanup |
| `audit` / `warn` records | Configurable per mode | None, namespace not checked |
| Can be blocked by policy | Yes, by RBAC or a policy engine | No, only by the operator |
| Discoverable | `kubectl get ns --show-labels` | Only in the API server configuration |

The resulting permission is the same: pods in `joviandss-csi` may be privileged.

## 4. When to use which

**Namespace labels**: the default for installing the `plugin`.

- Any cluster where the installer can create or label namespaces.
- Managed Kubernetes, where it is the only option.
- Talos clusters: no machine configuration change for one application.
- GitOps setups (Argo CD, Flux): the label is part of the versioned manifests.

**Admission exemption**: for clusters where namespace labels are not allowed to raise the level.

- Multi-tenant clusters where tenants may create and label namespaces, and the operator does
  not want labels to be the way to get `privileged`. RBAC on namespaces or a policy engine
  (Kyverno, OPA Gatekeeper) usually denies the `privileged` label; the exemption is then
  managed centrally by the operator.
- Clusters where the platform team owns all security configuration as node or control plane
  configuration (for example Talos machine configs kept in one repository).

In both cases the namespace has to allow `privileged` **before** the `plugin` DaemonSet and
StatefulSet are created, or pod creation fails. The controllers retry, so pods appear once the
label or exemption is in place, but with a delay due to back-off.

## 5. How things changed over time

| Kubernetes | Pod security mechanism |
|---|---|
| up to 1.20 | PodSecurityPolicy (PSP): cluster objects bound to service accounts through RBAC; optional admission plugin |
| 1.21 | PSP deprecated |
| 1.22 | Pod Security Admission introduced as alpha |
| 1.23 | Pod Security Admission beta, enabled by default |
| 1.25 | Pod Security Admission stable; PSP removed |

Consequences for CSI drivers:

- With PSP, a driver shipped a privileged `PodSecurityPolicy` and RBAC that allowed its
  service accounts to use it. These objects no longer exist since Kubernetes 1.25.
- Upstream Kubernetes has no cluster default level: without an admission configuration every
  namespace is `privileged`, and privileged pods are accepted. Many manifests written in that
  period work on plain clusters without any label.
- Distributions with secure defaults set a cluster default. Talos sets `enforce: baseline`,
  `audit: restricted`, `warn: restricted` and exempts `kube-system`. On these clusters a
  driver without labels or an exemption fails.
- Namespace labels became the common way for drivers and operators to declare their needs:
  a driver that ships its own namespace sets `pod-security.kubernetes.io/enforce: privileged`
  on it.

History of the `plugin` itself:

- The `plugin` was verified on Talos 1.6 with `deploy/talos/securitypatch.yaml`, the
  admission exemption.
- Namespace labels are the replacement. `deploy/joviandss/namespace.yaml` does not carry the
  labels yet; add them as in [section 2.1](#21-namespace-labels) or label the namespace after
  applying it.

## 6. Troubleshooting

Node pods missing, DaemonSet shows 0 pods:

```bash
kubectl -n joviandss-csi describe daemonset joviandss-csi-node
```

Look for `FailedCreate` events mentioning `violates PodSecurity`.

Effective labels of the namespace:

```bash
kubectl get namespace joviandss-csi --show-labels
```

Check what a level would reject without changing it (server-side dry run):

```bash
kubectl label --dry-run=server --overwrite namespace joviandss-csi \
    pod-security.kubernetes.io/enforce=baseline
```

The API server lists the existing pods that would violate the level as warnings.

Talos: current admission configuration of the API server:

```bash
talosctl -n <control-plane-node> get machineconfig -o yaml | grep -A 20 admissionControl
```

Further information:

1. [Pod Security Admission](https://kubernetes.io/docs/concepts/security/pod-security-admission/)
2. [Pod Security Standards](https://kubernetes.io/docs/concepts/security/pod-security-standards/)
3. [Enforce Pod Security Standards with namespace labels](https://kubernetes.io/docs/tasks/configure-pod-container/enforce-standards-namespace-labels/)
4. [Talos: Pod Security](https://www.talos.dev/latest/kubernetes-guides/configuration/pod-security/)
