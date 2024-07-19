# Installation

This document provides detailed guide on installation of JovianDSS CSI NFS Plugin, further `plugin`.


## Preparation

`Plugin` is delf contained and does not expect host to contain eny specific software.


## Installation

Installation process of JovianDSS CSI plugin goes through creation of appropriate Kubernetes resources.


### 1. Create namespace `joviandss-csi` that will host `plugin`. 

```bash
kubectl apply -f ./deploy/joviandss/namespace.yaml
```

Grant additional security privileges to `joviandss-csi` namespace:

```bash
kubectl label ns joviandss-csi pod-security.kubernetes.io/audit=privileged pod-security.kubernetes.io/enforce=privileged pod-security.kubernetes.io/warn=privileged
```

### 2. Install CRDT and classes related snapshot support provided for snapshot support. This CRDT's are inherited from [github.com/kubernetes-csi/external-snapshotter](https://github.com/kubernetes-csi/external-snapshotter/tree/release-5.0/client/config/crd`)

```bash
kubectl apply -f ./deploy/joviandss/crdt/volumesnapshotclasses.yaml
kubectl apply -f ./deploy/joviandss/crdt/volumesnapshotcontents.yaml
kubectl apply -f ./deploy/joviandss/crdt/volumesnapshots.yaml
```

### 3. Create snapshot controller

[Snapshot controller](https://kubernetes-csi.github.io/docs/snapshot-controller.html) is based on [external snapshoter project](https://github.com/kubernetes-csi/external-snapshotter).
Keep in mind that `Cluster roler` for `snapshot-controller` is bound to `joviandss-csi-controller-service-account` in definition of `joviandss-csi-controller`,
therefore possible renaming might require additional changes in `./deploy/joviandss/joviandss-csi-controller.yaml`.

Original RBAC for `snapshot-controller` been extended with `update` for `resources` `volumesnapshotcontents/status`

Controller can be installed by:

```bash
kubectl apply -f ./deploy/joviandss/snapshot-controller/rbac-snapshot-controller.yaml
kubectl apply -f ./deploy/joviandss/snapshot-controller/setup-snapshot-controller.yaml
```

### 4. Setup config file for `plugin`

Main plugin configuration get provided by config file that get attached to controller service of plugin through `secret`.
Check this [guide on configuration](configuration.md) and [nfs specifci guide](configuration-nfs.md) to get more information on `plugin` configuration.

```bash
kubectl create secret -n joviandss-csi generic jdss-controller-cfg --from-file ./deploy/cfg/cfg.yaml 
```

Please keep in mind that config file name and secret name are both referenced in `./deploy/joviandss/joviandss-csi-controller.yaml`.

[TODO]: # Provide detailed description of how config name and secret name affects controller config

And change in config filename and secret name will require appropriate changes in `./deploy/joviandss/joviandss-csi-controller.yaml`.

### 5. Install actual JovianDSS CSI plugin

``` bash
kubectl apply -f ./deploy/joviandss/nfs/joviandss-csi-controller.yaml

kubectl apply -f ./deploy/joviandss/nfs/joviandss-csi-node.yaml 
```

### 6. Checking

If everything is OK, you should be able to find JovianDSS CSI Driver in list of CSI drivers:

```bash
kubectl get csidrivers.storage.k8s.io
```
```
NAME                             ATTACHREQUIRED   PODINFOONMOUNT   STORAGECAPACITY   TOKENREQUESTS   REQUIRESREPUBLISH   MODES        AGE
nfs.csi.joviandss.open-e.com   true             true             false             <unset>         false               Persistent   1d12h
``` 

You should be able to see that `controller` and `node` services in `joviandss-csi` namespace:

```bash
kubectl get pods -n joviandss-csi
```
```
NAME                         READY   STATUS    RESTARTS        AGE
joviandss-csi-controller-0   4/4     Running   12 (3d1h ago)   3d12h
joviandss-csi-node-hltwk     2/2     Running   0               3d23h
joviandss-csi-node-nzqkp     2/2     Running   5 (3d23h ago)   3d23h
joviandss-csi-node-qzdf6     2/2     Running   5 (3d23h ago)   3d23h
```

Also `snapshot-controller` should be running in `kube-system` namespace:
```bash
kubectl get pods -n kube-system
```
```
NAME                                   READY   STATUS    RESTARTS      AGE
snapshot-controller-7c5dccb849-8vvzm   1/1     Running   0             61m
snapshot-controller-7c5dccb849-q4hnx   1/1     Running   0             61m
```
