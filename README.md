# sshmon-check-elasticsearch
Nagios/Checkmk-compatible SSHMon-check for Elasticsearch-Clusters

## Installation
* Download [latest Release](https://github.com/indece-official/sshmon-check-elasticsearch/releases/latest)
* Move binary to `/usr/local/bin/sshmon_check_elasticsearch`


## Usage
```
$> sshmon_check_elasticsearch -service Elasticsearch_testcluster.default -dns 10.96.0.10:53 -host testcluster.default.svc.cluster.local
```

```
Usage of sshmon_check_elasticsearch:
  -dns string
        Use alternate dns server
  -host string
        Host
  -port int
        Port (default 9200)
  -service string
        Service name (defaults to Elasticsearch_<host>)
  -v    Print the version info and exit
```

Output:
```
0 Elasticsearch_testcluster.default - OK - Elasticsearch cluster 'testcluster' on testcluster.default.svc.cluster.local has status 'green'
```

### Supported Elasticsearch versions
| Version | Tested |
| --- | --- |
| v7 | Yes |

## Development
### Snapshot build

```
$> make --always-make
```

### Release build

```
$> BUILD_VERSION=1.0.0 make --always-make
```