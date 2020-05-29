# Open-E JovianDSS Kubernetes CSI plugin

[![Build Status](https://travis-ci.org/open-e/JovianDSS-KubernetesCSI.svg?branch=master)](https://travis-ci.org/open-e/JovianDSS-KubernetesCSI)
[![Go Report Card](https://goreportcard.com/badge/github.com/open-e/JovianDSS-KubernetesCSI)](https://goreportcard.com/report/github.com/open-e/JovianDSS-KubernetesCSI)

## Deployment

### Configuring

Plugin has 2 config files. Controller and node configs. Controller is responsible for a management of particular volumes on JovianDSS storage. When nodes responsibility is limited to connecting particular volume to particular host. Configuration file examples can be found in 'deploy/cfg/' folder.

 - **llevel** the logging level of the plugin

 - **plugins** specify the services that should be run in the plugin.
    Possible values: *IDENTITY_SERVICE*, *CONTROLLER_SERVICE*, *NODE_SERVICE*
    + **IDENTITY_SERVICE** - starts identity service, expected to run on each physical node with plugin
    + **CONTROLLER_SERVICE** - starts controller service, cluster should have only one instance of this service in running at a time
    + **NODE_SERVICE** - starts node service, this service is responsible for attaching physical volumes stored on JovianDSS
        This service should be running on every physical host that is expected to have containers with such feature.
 - **controller** - describes properties of controller service
    + **name** - name of JovianDSS storage, not used at the moment
    + **addr** - ip address of JovianDSS storage
    + **port** - port of JovianDSS storage, the port that is asigned to REST interface
    + **user** - user to execute REST requests
    + **pass** - password for the user specified above
    + **prot** - protocol that is gona be used for sending REST
    + **pool** - name of the pool created on JovianDSS
    + **tries** - number of attempts to send REST request if network related failure occured
    + **iddletimeout** - time maintain iddle session
 - **node** - describes properties of node service
    + **id** - prefix for a node name
    + **addr** - ip address of JovianDSS storage
    + **port** - port of JovianDSS storage, the port that is asigned to iSCSI volume sharing    


Add config files as secrets:

``` bash
kubectl create secret generic jdss-controller-cfg --from-file=./deploy/cfg/controller.yaml

kubectl create secret generic jdss-node-cfg --from-file=./deploy/cfg/node.yaml
```
Node config do not provides nothing but storage address and request to create proper services.

### Deploy plugin

Make sure that you have iscsi\_tcp module installed on the machines running node plugin.

If you change confing names from the previous step. Dont forget to modify  *joviandss-csi-controller.yaml* and *joviandss-csi-node.yaml* accordingly.
To deploy plugins to a cluster:

``` bash
kubectl apply -f ./deploy/joviandss/joviandss-csi-controller.yaml

kubectl apply -f ./deploy/joviandss/joviandss-csi-node.yaml 

kubectl apply -f ./deploy/joviandss/joviandss-csi-sc.yaml
```

If everything is OK, you should see something like:

```bash
[kub@kub-master /]$ kubectl get csidrivers

NAME                       CREATED AT
com.open-e.joviandss.csi   2019-06-07T22:52:01Z
```
and 

```bash
[kub@kub-master /]$ kubectl get pods

NAME                         READY   STATUS    RESTARTS   AGE
joviandss-csi-controller-0   3/3     Running   0          10d
joviandss-csi-node-q55k5     2/2     Running   0          10d
joviandss-csi-node-w2cp8     2/2     Running   0          10d
```


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



