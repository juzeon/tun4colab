package main

import "fmt"

func PrintUsage() {
	fmt.Println(`
Usage:
	-p <Port 1> -p <Port 2> ... -p <Port N>
	-c <Command 1> -c <Command 2> ... -c <Command 3>
	Example: ./tun4colab -p 8080 -p 8081 -c "http-server -p 8080" -c "python api.py -p 8081"`)
}
