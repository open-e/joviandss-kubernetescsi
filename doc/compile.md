# Compilation

This guide provides overview on how user can compile CSI Kubernetes Plugin.


Compilation is done with `make` utility.

Specification of supported protocol type is done on the compilation stage through `make` option `PROTOCOL_TYPE`.

#### NFS Plugin

```bash
make dev PROTOCOL_TYPE=nfs
```

#### iSCSI Plugin

```bash
make dev PROTOCOL_TYPE=iscsi
```

## Build targets

There is 3 main build targets:

1. dev -- compiles plugin with additional debug symbols and generates container
2. prod -- compiles plugin with no debug symbols and generates container
3. cli -- generates cli that provides command line interface to plugin implementation
