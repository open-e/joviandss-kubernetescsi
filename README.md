# Open-E JovianDSS Kubernetes CSI plugin

[![Build Status](https://travis-ci.org/open-e/JovianDSS-KubernetesCSI.svg?branch=master)](https://travis-ci.org/open-e/JovianDSS-KubernetesCSI)
[![Go Report Card](https://goreportcard.com/badge/github.com/open-e/joviandss-kubernetescsi)](https://goreportcard.com/report/github.com/open-e/joviandss-kubernetescsi)

This repo provide plugin source code along side with the resource deffinitions and instructions on how to use [Jovian Data Storage Solution](https://www.open-e.com/products/jovian-data-storage-software/general-information/) as a storage for containers running in Kubernetes cluster.

## Supported platforms

JovianDSS CSI plugin been tested on following platforms: Talos OS 1.6


## Plugin installation

[Here is a guide](doc/install.md) on how user can install plugin using `kubectl`.

`Helm` charts are comming...

## Plugin configuration

General plugin configuration gets done by passing config file in form of kubernetes `secret`.
You can check for example on how to expose config file to plugin in [installation guide](doc/install.md). 
Check [configuration document](doc/configuration.md) to learn about configurational options.

### Deploy application

In order to deploy application with automatic storage allocation run: 
``` bash
kubectl apply -f ./deploy/examples/nginx-pvc.yaml

kubectl apply -f ./deploy/examples/nginx.yaml
```

In order to deploy application with pre provisioned volume administrator first have to create volume.
It can be done with the help of [csc](https://github.com/rexray/gocsi/tree/master/csc) tool.
Once you obtain Id of the volume you can create persistent volume placing proper name of the volume into the file.
```bash
kubectl apply -f ./deploy/examples/nginx-pv.yaml
```



