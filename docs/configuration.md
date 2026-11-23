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
iscsi:
  iqn: iqn.csi.2019-04
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
- `logfile` user can specify file to output log to, by default log would be printed to standard output.
- `pool` Pool name of the JovianDSS storage that would be used to store volumes, pool have to be created manually on the side of JovianDSS by user. `pool` is a key of the top level of the file, not of the `endpoint` section.

- `endpoint` is a section of config file instructing controller on how to connect to JovianDSS endpoint using REST API. REST API have to be enabled on the side of JovianDSS storage to make `plugin` work. The plugin uses REST API v4.
    - `addrs` list of addresses of the storage. Only the first address is used to send REST commands; the addresses after it are not used.
    - `port` port that would be used to connect to storage
    - `user` name of the JovianDSS Web/REST user
    - `pass` password of that user
    - `idletimeout` time to wait for one REST request to complete before considering it as failed, such as `30s`.
    - `tries` how many times a `GET` request that the storage answers without data (status 204) is sent in all: `3` means the first request and two more, `1` means no second request. `3` when the key is not set. No other request is sent again: not a `DELETE` answered 204, which is a success, and not a request that failed.
    - `retrydelay` the pause before such a `GET` is sent again, written like `idletimeout`. `1s` when the key is not set.
- `iscsi` is a section of config file containing information on how to connect to JovianDSS iscsi targets.
    - `iqn` iqn prefix that would be used for target creation. It may hold lower case letters, digits, `.`, `:` and `-`, has to start and end with a letter or a digit, and may have 155 characters at most: the plugin adds `:` and 64 characters to it, and a target name has 220 at most. `iqn.csi.2019-04` when the key is not set.
    - `addrs` list of addresses that would be used to connect targets
    - `port` iscsi port provided by JovianDSS storage

Keys that the plugin does not read, such as `name` in the `endpoint` section, are ignored.
