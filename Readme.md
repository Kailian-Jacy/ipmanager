# ipmanager

This is a proxy layer used for probe and control qsc-idc ip-pool.

In our case, our ip-pool is deployed on a nginx server with hundreds of IPs, who plainly takes proxy and sends out requests through one of its IP.

But managing IP availabiliy from nginx is restricted, this is what this program used for.

## Usage

This is a monitor and load-balancer for ip-pool with huge amount of ip.

`ipmanager` is parsing `accesslog` of `nginx` with regex to fetch history of each ip, and load balance requests to available IPs dynamically and apparently.

First, with `config.UpstreamConfPath` given, the program would read and parse ips to build a full ip list during initialization. And it scans `config.AccessLogPath` for requests behavior.

Scanned request records are reserved in `IP.IP.History []*Entry` and can be access with `GET HOST.com/probe`. 
Every `config.ScanInterval` time the program will scan the log for newly incoming requests logs and update the history of each ip as well as available IP list to tune its load-balancing strategy.

With request comes in, the program select an IP and send it to the port

At will, you'll possibly be able to take a look and analyse the behavior of upstream ips with promethuse and your favorate visualization plane.

## Deployment

Build from source and run with:
```bash
cd ipmanager
go version
env GO111MODULE=on go build -o ipmanager
./ipmanager --config [absolute/full/path/to/config.json]
```
Note: 
1. This package by default requires go 1.18 or higher
2. For nginx log is protected by default, you may want to run it with sudo.

### Configuration
Example configuration file built in with default value:
```json
{
  "mode": "serve",
  "debug": true,
  "probe_port": "9095",
  "proxy_port": "9096",
  "next": "127.0.0.1",
  "scan_interval_min": 5,
  "access_log_path": "/var/log/nginx/host.access.log",
  "upstream_conf_path": "/etc/nginx/conf.d/001-upstream.conf",
  "max_history_log_each_ip": 1000,
  "max_cool_down_log_each_ip": 1000,
  "cool_down_min": 10,
  "dial_timeout": 15,
  "max_connection_timeout": 100,
  "strategy": "consecutive",
  "consecutive_failure": 3
}
```

- `mode`: Enum of `serve` and `parse`.
    - `parse`: The program parses the log after last rotate and continue to serve.
    - `serve`: The program do not take a look at the last rotate, but cares only about the later requests. It may take some failure to make the Available IPs converge.
- `debug`: Causes verbose logging.
- `probe_port`: The port for probe. Router `/probe` `/ping` and `/config` is available.
- `next`: The nginx entry of upstreams.

## Parsing

All files it fetched is parsed with regex, and the regex is defined in `ip/scanner.go`. With logging style changes and other possibly change, you may
want to modify the regex by your own.

## Proxying

For now, load-balancing is naive: random. You can extend it on your own.

The proxy is meant to be apparent and supporting only tcp.

## IP Control

IP are classified as:
```azure
Available
Unavailable
```

You may want to control the switching condition of IPs on your own. The control `config.strategy` is defined in `ip/strategy.go` and selected by configuration file.

Fow now, only `consecutive` mode is supported. Which means:
- An ip would be marked as `Unavailable` if it **consecutively** fails `config.consecutive_failure` times in 5 minutes.
- An ip would be marked as `Available` if it has been banned for `config.Cooldown` minutes.

## Extend

You may want to extend:
- Parsing regex in: `ip/scanner.go`
- Load-Balancing strategy: `Web/server.go`
- IP banning and cool down strategy: `ip/strategy.go`

For more extensibility, you can raise an issue to let me know.

## TODO
TODO: Add probe to prometheus for visualization.
TODO: Add debug tracing to upstream. Which one has the very last request been sent to?
TODO: Make performance test and tune.