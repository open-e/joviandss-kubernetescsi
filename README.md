# Open-E JovianDSS Kubernetes CSI plugin

[![Go Report Card](https://goreportcard.com/badge/github.com/open-e/joviandss-kubernetescsi)](https://goreportcard.com/report/github.com/open-e/joviandss-kubernetescsi)

This repo provide plugin source code along side with the resource deffinitions and instructions on how to use [Jovian Data Storage Solution](https://www.open-e.com/products/jovian-data-storage-software/general-information/) as a storage for containers running in Kubernetes cluster.

## Supported platforms

JovianDSS CSI plugin been tested on following platforms: Talos OS 1.6


## Plugin installation

### iSCSI
[Here is a guide](doc/install.md) on how user can install iSCSI plugin using `kubectl`.

### NFS
[Here is a guide](doc/install-nfs.md) on how user can install NFS plugin.

`Helm` charts are comming...

## Plugin configuration

General plugin configuration gets done by passing config file in form of kubernetes `secret`.
You can check for example on how to expose config file to plugin in [installation guide](doc/install.md). 
Check [iSCSI configuration document](doc/configuration.md) to learn about configurational options.
For NFS please check [nfs installation guide](doc/install-nfs.md) and [NFS configuration document](doc/configuration.md)

## Deploy iSCSI example applications

This section describes installing application with iSCSI.

Once `plugin` installation is completed user can deploy applications that use volumes from *JovianDSS* as [persistent volumes](https://kubernetes.io/docs/concepts/storage/persistent-volumes/)

Here is examples of volumes and snapshots:

### NGINX app with PVC

Keep in mind that PVC volumes will get automatically created and deleted with container creation and deletion.
``` bash
kubectl apply -f ./deploy/example/nginx-pvc.yaml
```

### NGINX app with PV

In order to deploy application with pre provisioned volume administrator first have to create volume.
It can be done with the help of [csc](https://github.com/rexray/gocsi/tree/master/csc) tool.
Or it can be done manually with `JovianDSS` user interface or cli tool, keep in mind that for existing volume to be used it name have to start with `vp_` prefix.

For instance if you have existing `zvol` on `JovianDSS` named `pv-test` and you want to attach it to nginx application running inside kubernetes:

1. Rename volume `pv-test` to `vp_pv-test`.
2. Create *persistent volume*
User can find example of persistent volumes at *deploy/example/pv-test.yaml*. In order to set your volume user is expected to change `spec: csi: volumeHandle` to exact name that volume have on JovianDSS pool.
Once it it done *pv* can be created by calling:
```
kubectl apply -f ./deploy/example/pv-test.yaml
```
3. Create *persistent volume claim*
Once you obtain Id of the volume you can create *pvc* based on specific *pv*:

```bash
kubectl apply -f ./deploy/example/pv-test-pvc.yaml
```
4. Deploy application

```bash
kubectl apply -f ./deploy/example/pv-test-pvc-nginx.yaml
```

### Making snapshot

Snapshot of volume associated with *pvc* `pv-test-pvc` provided in previous example can be created by:
```bash
kubectl apply -f ./deploy/example/pv-test-pvc-snapshot.yaml
```

## Deploy NFS example applications

User can use same approach to for NFS based volumes.
Examples for NFS volumes can be found in folder:
```
deploy/examples/nfs
```
For instance installation of NGINX with NFS can be done by 

``` bash
kubectl apply -f ./deploy/example/nfs/nginx-pvc.yaml
```
