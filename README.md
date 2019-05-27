# `proxyproxy`

## What would I need such a thing?

Operating systems (or at least some common Linux distributions) are rather terrible
in handling HTTP proxies which are often used in corperate IT infrasturctures. References
to these proxy adresses are scattered all over your system, for example in Ubuntu you have:

- `apt` package manager has its `/etc/apt/apt.conf`.
- for some terrible reason the is a [setting hidden in `dconf`](https://askubuntu.com/questions/1133286/set-ubuntu-desktop-network-proxy-settings-programmatically-on-18-04).
- a lot (but not all) of shell utilities use the `http(s)_proxy` environment variables. 
- Docker needs some [config](https://docs.docker.com/network/proxy/) to work behind a proxy.
- Some browsers such as `firefox` are required to be configured.
- etc.

If you need to do this configuration once that is annoying but manageable. But 
if you have to change these settigs everyonce in a while because for example you 
need to use your notebook in a network without proxy things get quite unhandy.

## So what is `proxyproxy`?

`proxyproxy` is a HTTP/HTTPS proxy that runs on your local machine and is aware of the 
network the machine is attached to. Based on than it redirects your requests
accordingly.
