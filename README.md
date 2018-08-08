# jmftrindade - Modified fork:

Go client periodically sends random data duration and data sizes for fictional read / write events.

To run Prometheus server:

```
$ cd /usr/local/prometheus_install_folder
$ ./prometheus -config.file=prometheus.yml
```

Prometheus's config file should have a section under "static config" listing the group of lineage data producing clients , e.g.:

```
# my global config
global:
  scrape_interval:     15s # Set the scrape interval to every 15 seconds. Default is every 1 minute.
#  evaluation_interval: 15s # Evaluate rules every 15 seconds. The default is every 1 minute.
  # scrape_timeout is set to the global default (10s).

  # Attach these labels to any time series or alerts when communicating with
  # external systems (federation, remote storage, Alertmanager).
  external_labels:
      monitor: 'codelab-monitor'

# Load rules once and periodically evaluate them according to the global 'evaluation_interval'.
#rule_files:
  # - "first.rules"
  # - "second.rules"

# A scrape configuration containing exactly one endpoint to scrape:
# Here it's Prometheus itself.
scrape_configs:
  # The job name is added as a label `job=<job_name>` to any timeseries scraped from this config.
  - job_name: 'prometheus'

    # metrics_path defaults to '/metrics'
    # scheme defaults to 'http'.

    # Override the global default and scrape targets from this job every 5 seconds.
    scrape_interval: 5s

    static_configs:
      - targets: ['localhost:9090']
        labels:
          group: 'prometheus_server'

      - targets: ['localhost:8080', 'localhost:8082', 'localhost:8083']
        labels:
          group: 'lineage_servers'
```

To start a few data lineage producing clients:
```
$ cd examples/random
$ ./random_lineage_generator -listen-address=:8082 &
$ ./random_lineage_generator -listen-address=:8080 &
$ ./random_lineage_generator -listen-address=:8082 &
```

To view data lineage timeseries data, go to Prometheus dashboard at http://localhost:9090/ and look for "lineage" on the search bar.

To delete existing timeseries recorded by the data lineage clients:
```
$ curl -XDELETE -g 'http://localhost:9090/api/v1/series?match[]=lineage_op_durations'
$ curl -XDELETE -g 'http://localhost:9090/api/v1/series?match[]=lineage_op_durations_sum'
$ curl -XDELETE -g 'http://localhost:9090/api/v1/series?match[]=lineage_op_durations_count'

etc...
```

# Prometheus Go client library

[![Build Status](https://travis-ci.org/prometheus/client_golang.svg?branch=master)](https://travis-ci.org/prometheus/client_golang)
[![Go Report Card](https://goreportcard.com/badge/github.com/prometheus/client_golang)](https://goreportcard.com/report/github.com/prometheus/client_golang)

This is the [Go](http://golang.org) client library for
[Prometheus](http://prometheus.io). It has two separate parts, one for
instrumenting application code, and one for creating clients that talk to the
Prometheus HTTP API.

## Instrumenting applications

[![code-coverage](http://gocover.io/_badge/github.com/prometheus/client_golang/prometheus)](http://gocover.io/github.com/prometheus/client_golang/prometheus) [![go-doc](https://godoc.org/github.com/prometheus/client_golang/prometheus?status.svg)](https://godoc.org/github.com/prometheus/client_golang/prometheus)

The
[`prometheus` directory](https://github.com/prometheus/client_golang/tree/master/prometheus)
contains the instrumentation library. See the
[best practices section](http://prometheus.io/docs/practices/naming/) of the
Prometheus documentation to learn more about instrumenting applications.

The
[`examples` directory](https://github.com/prometheus/client_golang/tree/master/examples)
contains simple examples of instrumented code.

Example queries:
```
metrics:
- lineage_op_sizes_histogram_bytes_bucket
- lineage_op_durations_histogram_seconds_bucket

the 95th %ile over 5m windows:
histogram_quantile(0.95, sum(rate(lineage_op_sizes_histogram_bytes_bucket[5m])) by (le))

rate(lineage_op_durations_histogram_seconds_sum[5m]) / rate(lineage_op_durations_histogram_seconds_count[5m]) 
```


## Client for the Prometheus HTTP API

[![code-coverage](http://gocover.io/_badge/github.com/prometheus/client_golang/api/prometheus)](http://gocover.io/github.com/prometheus/client_golang/api/prometheus) [![go-doc](https://godoc.org/github.com/prometheus/client_golang/api/prometheus?status.svg)](https://godoc.org/github.com/prometheus/client_golang/api/prometheus)

The
[`api/prometheus` directory](https://github.com/prometheus/client_golang/tree/master/api/prometheus)
contains the client for the
[Prometheus HTTP API](http://prometheus.io/docs/querying/api/). It allows you
to write Go applications that query time series data from a Prometheus
server. It is still in alpha stage.

## Where is `model`, `extraction`, and `text`?

The `model` packages has been moved to
[`prometheus/common/model`](https://github.com/prometheus/common/tree/master/model).

The `extraction` and `text` packages are now contained in
[`prometheus/common/expfmt`](https://github.com/prometheus/common/tree/master/expfmt).

## Contributing and community

See the [contributing guidelines](CONTRIBUTING.md) and the
[Community section](http://prometheus.io/community/) of the homepage.
