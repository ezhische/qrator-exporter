# Qrator exporter
Inspired by https://github.com/StupidScience/qrator-exporter

Tottaly rewrited for new API

## Configuration

Exporter configurates via environment variables:

|Env var|Description|required|
|---|---|---|
|QRATOR_CLIENT_ID|Your client ID for qrator. Only digits required.|true|
|QRATOR_X_QRATOR_AUTH|X-Qrator-Auth header for access to Qrator Api. It's not required if you use IP auth|true|
|QRATOR_DOMAINS_IDS|Domain IDs to select domains and not export all|false|
|QRATOR_PROXY_URL|Proxy string if need to use proxy http://login:password@host:port|false|
|QRATOR_API_URL|API URL (default https://api.qrator.net/request)|false|
|QRATOR_TIMEOUT|API Call timeous (default 5s)|false|
|QRATOR_EXPORTER_PORT|Metrics port (default 9502)|false|
|QRATOR_EXPORTER_CONCURENT|Number of parralel connections to API (default 10)|false|

Exporter listen on tcp-port **9502**. Metrics available on `/metrics` path.

## Exposed metrics

It returns all statistics that defined in 3 methods [StatisticsCurrentIP](https://api.qrator.net/#types-statisticscurrentip), [StatisticsCurrentHTTP](https://api.qrator.net/#types-statisticscurrenthttp), [Billable](https://api.qrator.net/#domain-methods-statistics).

## Run via Docker

The latest release is automatically published to the [Docker registry](https://hub.docker.com/r/ezhische/qrator-exporter).

You can run it like this:
```
$ docker run -d --name qrator-exporter \
            -e QRATOR_CLIENT_ID=12345 \
            -e QRATOR_X_QRATOR_AUTH=12345abcdef \
            -p 9502:9502 \
            ezhsiche/qrator-exporter
```
