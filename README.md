# tun4colab

A helper program that allows for parallel execution of API services and utilizes Cloudflare Quick Tunnel for port forwarding on Google Colab.

## Build

Requires Go 1.21+

```bash
# If using Windows
set GOOS=linux
set GOARCH=amd64

# Build for colab
go build
```

## Usage

```bash
-p <Port 1> -p <Port 2> ... -p <Port N>
-c <Command 1> -c <Command 2> ... -c <Command 3>
```

e.g.

```bash
./tun4colab -p 8080 -p 8081 -c "http-server -p 8080" -c "python api.py -p 8081"
```

The output of the API services will be redirected to stdout, while the  endpoints of the tunnels will be printed once they are created.

## Example

```bash
> ./tun4colab -p 8081 -c "cd D:\Go && http-server -p 8081"
2023/11/08 22:24:00 INFO Ports to open: ports=[8081]
2023/11/08 22:24:00 INFO Commands to execute: commands="cd D:\\Go && http-server -p 8081"
Starting up http-server, serving ./

http-server version: 14.1.1

http-server settings:
CORS: disabled
Cache: 3600 seconds
Connection Timeout: 120 seconds
Directory Listings: visible
AutoIndex: visible
Serve GZIP Files: false
Serve Brotli Files: false
Default File Extension: none

Available on:
  http://2.0.0.1:8081
  http://172.20.160.1:8081
  http://192.168.137.1:8081
  http://192.168.58.1:8081
  http://192.168.168.1:8081
  http://10.27.214.248:8081
  http://127.0.0.1:8081
Hit CTRL-C to stop the server

2023/11/08 22:24:03 INFO Tunnel created port=8081 url=https://mas-production-complicated-tag.trycloudflare.com
```

![](https://public.ptree.top/ShareX/2023/11/08/1699453519/AEjbKXeXOd.png)