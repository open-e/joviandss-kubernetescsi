# Installation

This document provides detailed guide on installation of JovianDSS CSI Plugin, further `plugin`.


## Preparation

`Plugin` expects host machine to have `iscsid` daemon, `iscsiadm` cli and `iscsi\_tcp` kernel module installed.
That is needed as `plugin` does not contain mentioned iscsi tool yet relies on them heavily.
Installation of this components is pretty straightforward on most of linux distribution and will not be coverer here except for `TalosOS`.
Due to architecture and security policies [additioanal actions have to be take prior to following instructions provided below](talos.md).


## Installation

Installation process of JovianDSS CSI plugin goes through creation of appropriate Kubernetes resources.


### 1. Create user need to create namespace for `plugin` and resources associates with it.

```bash
kubectl apply -f ./deploy/joviandss/namespace.yaml
```

### 2. User have to install CRDT provided for snapshot support. This CRDT's are inherited from [github.com/kubernetes-csi/external-snapshotter](https://github.com/kubernetes-csi/external-snapshotter/tree/release-5.0/client/config/crd`)

```bash
kubectl apply -f ./deploy/joviandss/svolumesnapshotclasses.yaml
kubectl apply -f ./deploy/joviandss/svolumesnapshotcontents.yaml
kubectl apply -f ./deploy/joviandss/svolumesnapshots.yaml
```

### 3. Create snapshoting service

```bash
kubectl apply -f ./deploy/joviandss/snapshot-controller.yaml
```

### 4. Setup config file for `plugin`

Main plugin configuration get provided by config file that get attached to controller service of plugin through `secret`.
Check this [guide on configuration](configuration.md) to get more information on `plugin` configuration.

```bash
kubectl create secret -n joviandss-csi generic jdss-controller-cfg --from-file ./deploy/cfg/cfg.yaml 
```

Please keep in mind that config file name and secret name are both referenced in `./deploy/joviandss/joviandss-csi-controller.yaml`.

[TODO]: # Provide detailed description of how config name and secret name affects controller config

And change in config filename and secret name will require appropriate changes in `./deploy/joviandss/joviandss-csi-controller.yaml`.

### 5. Install actual JovianDSS CSI plugin

``` bash
kubectl apply -f ./deploy/joviandss/joviandss-csi-controller.yaml

kubectl apply -f ./deploy/joviandss/joviandss-csi-node.yaml 
```

### 6. Checking

If everything is OK, you should be able to find JovianDSS CSI Driver in list of CSI drivers:

```bash
kubectl get csidrivers.storage.k8s.io
```
```
NAME                             ATTACHREQUIRED   PODINFOONMOUNT   STORAGECAPACITY   TOKENREQUESTS   REQUIRESREPUBLISH   MODES        AGE
iscsi.csi.joviandss.open-e.com   true             true             false             <unset>         false               Persistent   3d12h
``` 

Also you should be able to see that `controller`, `node` services running along side with `snapshot-controller`
```bash
kubectl get pods -n joviandss-csi
```
```
NAME                         READY   STATUS    RESTARTS        AGE
joviandss-csi-controller-0   4/4     Running   12 (3d1h ago)   3d12h
joviandss-csi-node-hltwk     2/2     Running   0               3d23h
joviandss-csi-node-nzqkp     2/2     Running   5 (3d23h ago)   3d23h
joviandss-csi-node-qzdf6     2/2     Running   5 (3d23h ago)   3d23h
snapshot-controller-0        1/1     Running   0               3d13h
```
