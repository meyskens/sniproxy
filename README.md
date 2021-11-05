# SNIProxy

SNIProxy is a very simple reverse proxy server that uses TLS SNI to route to hosts (in HTTP mode it just uses the Host header).
It is designed to proxy as much on the TCP level where possible but still allow for the same port to be used on the outside.

## How to use

1. Compile the server
2. Setup endpoints.txt in the `host,ip` format

```
test.stuvm.be,10.1.2.3
student2.stuvm.be,10.1.2.9
```

3. Start the server `./sniproxy serve`

`test.stuvm.be` now proxies port 80 and 443 to 10.1.2.3 on port 80 and 443 respectively.
SNIProxy due to be designed for the specific use case for student hosting also forwards `*.test.stuvm.be` to the same IP allowing for infinite subdomains to be used.

## Where is it used?

I use this as reverse proxy for StuVM (class VM hosting for students) where we share many VMs behind one public IPv4 (they are expensive you know).
