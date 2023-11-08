# tun4colab

A helper program that allows for parallel execution of API services and utilizes Cloudflare Quick Tunnel for port forwarding on Google Colab.

## Build

Requires Go 1.21+

```bash
go build
```

## Usage

```bash
-p <Port 1> -p <Port 2> ... -p <Port N>
-c <Command 1> -c <Command 2> ... -c <Command 3>
```

Example:

```bash
./tun4colab -p 8080 -p 8081 -c "http-server -p 8080" -c "python api.py -p 8081"
```

The output of the API services will be redirected to stdout, while the  endpoints of the tunnels will be printed once they are created.

