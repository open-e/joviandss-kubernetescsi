# JovianDSS CSI plugin configuration

Plugin configuration goes through config file. Config file itself gets provided to plugin in form of Kubernetes `secret` and has `yaml` format.

Please take notice that name of a file and the name of secret resource is the same as specified in controller config.

```bash
kubectl create secret -n joviandss-csi generic jdss-controller-cfg --from-file ./deploy/cfg/cfg.yaml 
```

If you change config file user have to restart controller service as well.

Here is example config file:

```
loglevel  : Debug
logpath   : /tmp/csi-log
endpoint:
  name: MainStorage
  addrs:
    - 192.168.0.100
  port: 82
  user: admin
  pass: admin
  prot: https
  pool: Pool-0
  tries: 3
  idletimeout: 5s
nfs:
  addrs:
    - 192.168.0.100
```

Read about top level configuration options and `endpoint` section in [configuration guide](configuration.md).

- `nfs` is a section of config file containing information on how to connect to JovianDSS nfs resources.
    - `addrs` list of addresses that would be used to connect shares
