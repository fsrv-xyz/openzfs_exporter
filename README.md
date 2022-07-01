# openzfs_exporter

This is a Prometheus exporter for OpenZFS metrics on FreeBSD.

## usage

```
-discover-pools
      use autodiscovery for zfs pools
-exported-pools value
      list of pools to export metrics for
-filter string
      filter queried datasets (default "^.*$")
-filter-reverse
      reverse filter functionality; if set, only not matching datasets would be exported
-interval duration
      refresh interval for metrics (default 5s)
-version
      print binary version
-web.listen-address string
      address listening on (default ":8080")
```

## exported metrics

All sysctl properties displayed in
```shell
$ sysctl kstat.zfs.<poolname>.dataset
```
are dynamically exported.
Metrics will look like:
```prometheus
# HELP openzfs_zfs_parameter sysctl openzfs dataset parameters
# TYPE openzfs_zfs_parameter gauge
openzfs_zfs_parameter{dataset="tank/postgres/data14",parameter="nread",pool="tank"} 3.1909632408e+10
openzfs_zfs_parameter{dataset="tank/postgres/data14",parameter="nunlinked",pool="tank"} 633421
openzfs_zfs_parameter{dataset="tank/postgres/data14",parameter="nunlinks",pool="tank"} 633421
openzfs_zfs_parameter{dataset="tank/postgres/data14",parameter="nwritten",pool="tank"} 3.909428281e+10
openzfs_zfs_parameter{dataset="tank/postgres/data14",parameter="reads",pool="tank"} 3.854329e+06
openzfs_zfs_parameter{dataset="tank/postgres/data14",parameter="writes",pool="tank"} 3.8178e+06
```

