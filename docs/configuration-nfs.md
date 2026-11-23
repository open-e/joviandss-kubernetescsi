# JovianDSS CSI plugin configuration

Plugin configuration goes through config file. Config file itself gets provided to plugin in form of Kubernetes `secret` and has `yaml` format.

Please take notice that name of a file and the name of secret resource is the same as specified in controller config.

```bash
kubectl create secret -n joviandss-csi generic jdss-controller-cfg --from-file ./deploy/cfg/cfg.yaml 
```

If you change config file user have to restart controller service as well.

Here is example config file:

```
loglevel: Info
# logfile: /tmp/csi-log
pool: Pool-0
endpoint:
  addrs:
    - 192.168.0.100
  port: 82
  user: admin
  pass: admin
  idletimeout: 30s
  tries: 3
  retrydelay: 1s
nfs:
  addrs:
    - 192.168.0.100
```

Read about top level configuration options, `pool` among them, and `endpoint` section in [configuration guide](configuration.md).

- `nfs` is a section of config file containing information on how to connect to JovianDSS nfs resources.
    - `addrs` list of addresses that would be used to connect shares
