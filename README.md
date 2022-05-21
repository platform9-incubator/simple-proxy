
# simple-proxy

A simple proxy to test airgapped browser access
[![Release](https://github.com/platform9-incubator/simple-proxy/actions/workflows/release.yml/badge.svg)](https://github.com/platform9-incubator/simple-proxy/actions/workflows/release.yml)

# how to run

This is supposed to be simple to run. Here are the arguments

* `--port`: the port to listen on
* `--target`: the host to filter (no other hosts will be allowed)
* `--target-ip`: use the IP address of the target instead of resolving the hostname
* `--verbose`: print more information (by default only prints errors)


```
simple-proxy --port=8080 --target pf9-1.local.net:8080 --target-ip 192.168.210.233
```

# Start browser with proxy

The following works but you will need to exit windows other wise Chrome uses the existing session

```
/Applications/Google\ Chrome.app/Contents/MacOS/Google\ Chrome --user-data-dir=/tmp  --proxy-server=localhost:8080 --incognito https://www.google.com
```
