![proxyproxy](./logo.svg "proxyproxy")

## Why would I need such a thing?

Operating systems (or at least some common Linux distributions) are rather terrible
in handling HTTP proxies which are often used in cooperate IT infrastructures. References
to these proxy addresses are scattered all over your system, for example in Ubuntu you have:

- `apt` package manager has its `/etc/apt/apt.conf`.
- for some terrible reason the is a [setting hidden in `dconf`](https://askubuntu.com/questions/1133286/set-ubuntu-desktop-network-proxy-settings-programmatically-on-18-04).
- a lot (but not all) of shell utilities use the `http(s)_proxy` environment variables. 
- Docker needs some [config](https://docs.docker.com/network/proxy/) to work behind a proxy.
- Some browsers such as `firefox` are required to be configured.
- etc.

If you need to do this configuration once that is annoying but manageable. But 
if you have to change these settings every once in a while because for example you 
need to use your notebook in a network without proxy things get quite unhandy.

## So what is `proxyproxy`?

`proxyproxy` is a HTTP/HTTPS proxy that runs on your local machine and is aware of the 
network the machine is attached to. Based on than it redirects your requests
accordingly.

## Requirements

`proxyproxy` is tested on Ubuntu 18.04 and should work on any current Linux distribution. 

## Installing `proxyproxy`

There are a couple of ways to install `proxyproxy` on your system:

### Via `snapcraft`

The easiest way to install `proxyproxy` is to use the `snapcraft` package. This also 
sets up a system service:

```
# sudo snap install proxyproxy
# sudo snap start --enable proxyproxy.proxyproxy
```

### Binary Download

Navigate to [Releases](https://github.com/unprofession-al/proxyproxy/releases), grab
the package that matches your operating system and architecture. Unpack the archive
and put the binary file somewhere in your `$PATH`.

### From Source

Make sure you have [go](https://golang.org/doc/install) installed, then run: 

```
# go get -u https://github.com/unprofession-al/proxyproxy
```

## Preparing the service

_This step can be skipped in `proxyproxy` was installed via `snapcraft`._

Since `proxyproxy` should be started as daemon/service for comfort a systemd service 
file must be created. For example create the file `/etc/systemd/system/proxyproxy.service`
with the following content (customize as needed):

```
[Unit]
Description=Service for proxyproxy
Wants=network.target

[Service]
ExecStart=/usr/bin/proxyproxy -c /etc/proxyproxy/config.yaml
SyslogIdentifier=proxyproxy
Restart=on-failure
TimeoutStopSec=30
Type=simple

[Install]
WantedBy=multi-user.target
```

Then, prepare the configuration (see below), reload systemd and start/enable the service:

```
# sudo systemctl daemon-reload
# sudo systemctl start proxyproxy.service
# sudo systemctl enable proxyproxy.service
```

## Configuring `proxyproxy`

The configuration is made via a simple configuration file. It's default location is 
`/etc/proxyproxy/config.yaml` (`/var/snap/proxyproxy/common/config.yaml` if 
installed via `snapcraft`). The following example contains comments to guide you:

``` yaml
---
# Write a ton of logs?
verbose: true

# On which host:port should proxyproxy listen to. This is where you need to point your 
# applications to.
proxy_address: 127.0.0.1:8080

# Where should the management interface listen to. 
admin_address: 127.0.0.1:8081

# Which interface names should be watched for changes. This is matched via 'contains'.
# For example 'enp0' would match 'enp0s25'.
interfaces:
  - eth0
  - enp0
  - wlan0
  - wlo1

# The 'proxyproxy' configs:
proxy_proxy_configs:
  # Profile name
  company_a:
    # Do a man in the middle... Currently this must be 'false'
    # Also if you want to set this to 'true' at one point, make sure you have read
    # the code! MITM can be evil...
    mitm: false
    # The 'company_a' profile will be used if an IP af the interfaces in in the
    # following subnet:
    in_net: 10.131.49.192/26
    # 'proxyproxy' will then redirect your traffic to the following proxy server 
    remote_proxy: "http://proxy.company.a:8080"

# If no profile is matched, no proxy server is used. 'proxyproxy' sends your 
# requests to the internet directly.
```
