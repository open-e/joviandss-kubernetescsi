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
iscsi:
  iqn: iqn.csi.2024-04 
  addrs:
    - 192.168.0.100
  port: 3260
```

- `loglevel` the logging level of the plugin. Logging can be done on following levels
    1. Panic
    2. Fatal
    3. Error
    4. Warn
    5. Info
    6. Debug
    7. Trace
- `logpath` user can specify file to output log to, by default log would be printed to standard output.

- `endpoint` is a section of config file instructing controller on how to connect to JovianDSS endpoint using REST API. REST API have to be enabled on the side of JovianDSS storage to make `plugin` work.
    - `name` name of storage, does not affect anything at the moment
    - `addrs` list of addresses that would be used to send REST commands to storage
    - `port` port that would be used to connect to storage, this port would be used for every address user provides for `addrs`
    - `pool` Pool name of the JovianDSS storage that would be used to store volumes, pool have to be created manually on the side of JovianDSS by user
    - `tries` how many attempts should be taken to sent single rest request to JovianDSS network interface before failing CSI request.
    - `iddletimeout` time to wait for REST request to complete before considering it as failed.
- `iscsi` is a section of config file containing information on how to connect to JovianDSS iscsi targets.
    `iqn` iqn prefix that would be used for target creation
    `addrs` list of addresses that would be used to connect targets
    `port` iscsi port provided by JovianDSS storage
