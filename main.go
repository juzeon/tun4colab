package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/go-resty/resty/v2"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type TunnelUpdateInfo struct {
	URL  string
	Port int
}
type TunnelCmd struct {
	Cmd  *exec.Cmd
	Port int
}

func main() {
	if len(os.Args) <= 1 {
		PrintUsage()
		return
	}
	var commandList []string
	var portList []int
	for i, arg := range os.Args {
		if arg == "-c" || arg == "-p" {
			if i+1 >= len(os.Args) {
				slog.Error("Malformed flags")
				PrintUsage()
				return
			}
			switch arg {
			case "-p":
				port, err := strconv.Atoi(os.Args[i+1])
				if err != nil {
					slog.Error(err.Error())
					PrintUsage()
					return
				}
				portList = append(portList, port)
			case "-c":
				commandList = append(commandList, os.Args[i+1])
			}
		}
	}
	slog.Info("Ports to open:", "ports", portList)
	slog.Info("Commands to execute:", "commands", strings.Join(commandList, ", "))
	cloudflaredFilename := "cloudflared-linux-amd64"
	if runtime.GOOS == "windows" {
		cloudflaredFilename = "cloudflared-windows-amd64.exe"
	}
	client := resty.New().SetRetryCount(3).
		AddRetryCondition(func(response *resty.Response, err error) bool {
			return response.StatusCode() != 200
		})
	if _, err := os.Stat(cloudflaredFilename); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			slog.Info("Downloading cloudflared from github...")
			resp, err := client.R().Get("https://api.github.com/repos/cloudflare/cloudflared/releases")
			if err != nil {
				panic(err)
			}
			var releaseResp []GithubReleaseResponse
			err = json.Unmarshal(resp.Body(), &releaseResp)
			if err != nil {
				panic(err)
			}
			url := ""
			for _, asset := range releaseResp[0].Assets {
				if asset.Name == cloudflaredFilename {
					url = asset.BrowserDownloadUrl
					break
				}
			}
			if url == "" {
				slog.Error("Cannot find release download url")
				return
			}
			slog.Info("Downloading from:", "url", url)
			resp, err = client.R().Get(url)
			if err != nil {
				panic(err)
			}
			err = os.WriteFile(cloudflaredFilename, resp.Body(), 0755)
			if err != nil {
				panic(err)
			}
			slog.Info("Downloading cloudflared completed")
			if runtime.GOOS == "linux" {
				slog.Info("Setting permission...")
				err = SetExecutable(cloudflaredFilename)
				if err != nil {
					panic(err)
				}
				slog.Info("Permission is set")
			}
		} else {
			panic(err)
		}
	}
	var cmdList []*exec.Cmd
	for _, command := range commandList {
		cmd := GetCmd(command)
		cmd.Stderr = os.Stdout
		cmd.Stdout = os.Stdout
		cmdList = append(cmdList, cmd)
		err := cmd.Start()
		if err != nil {
			slog.Error("Command cannot be started", "command", command, "error", err)
		}
		time.Sleep(200 * time.Millisecond)
	}
	cfLog, err := os.OpenFile("cloudflared.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		panic(err)
	}
	tunnelOutCh := make(chan TunnelUpdateInfo)
	defer close(tunnelOutCh)
	var tunnelCmdList []TunnelCmd
	for _, port := range portList {
		prefix := ""
		if runtime.GOOS == "linux" {
			prefix = "./"
		}
		cmd := GetCmd(prefix + cloudflaredFilename + " tunnel --url http://127.0.0.1:" + strconv.Itoa(port))
		out := &bytes.Buffer{}
		cmd.Stdout = out
		cmd.Stderr = out
		tunnelCmdList = append(tunnelCmdList, TunnelCmd{
			Cmd:  cmd,
			Port: port,
		})
		err := cmd.Start()
		if err != nil {
			slog.Error("cloudflared cannot be started", "port", port, "error", err)
		}
		go func(out *bytes.Buffer, port int) {
			for {
				if !strings.Contains(out.String(), "INF") {
					continue
				}
				line, err := out.ReadString('\n')
				_, err0 := cfLog.WriteString(line + "\n")
				if err0 != nil {
					slog.Warn("failed to write to cloudflared.log", "line", line)
				}
				if err != nil {
					if !errors.Is(err, io.EOF) {
						slog.Error("read from tunnel output", "error", err)
					}
					break
				}
				arr := regexp.MustCompile("https://(.*?)\\.trycloudflare\\.com").FindStringSubmatch(line)
				if len(arr) < 2 {
					continue
				}
				tunnelOutCh <- TunnelUpdateInfo{
					URL:  arr[0],
					Port: port,
				}
			}
		}(out, port)
	}
	go func() {
		for info := range tunnelOutCh {
			slog.Info("Tunnel created", "port", info.Port, "url", info.URL)
		}
	}()
	for i, cmd := range cmdList {
		err := cmd.Wait()
		if err != nil {
			slog.Error("Command exited", "command", commandList[i], "error", err)
		}
	}
	for _, cmd := range tunnelCmdList {
		err := cmd.Cmd.Wait()
		if err != nil {
			slog.Error("cloudflared exited", "port", cmd.Port, "error", err)
		}
	}
}
func GetCmd(command string) *exec.Cmd {
	if runtime.GOOS == "windows" {
		return exec.Command("cmd.exe", "/c", command)
	} else {
		return exec.Command("/bin/sh", "-c", command)
	}
}
func SetExecutable(filename string) error {
	_, err := GetCmd("chmod +x " + filename).CombinedOutput()
	if err != nil {
		return err
	}
	return nil
}
