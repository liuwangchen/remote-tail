package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/BurntSushi/toml"
	"github.com/mylxsw/remote-tail/command"
	"github.com/mylxsw/remote-tail/console"
)

var mossSep = ".--. --- .-- . .-. . -..   -... -.--   -- -.-- .-.. -..- ... .-- \n"

var welcomeMessage = getWelcomeMessage() + console.ColorfulText(console.TextMagenta, mossSep)

var filePath = flag.String("file", "", "-file=\"/var/log/*.log\"")
var hostStr = flag.String("hosts", "", "-hosts=root@192.168.1.101,root@192.168.1.102")
var configFile = flag.String("conf", "", "-conf=example.toml")
var tailFlags = flag.String("tail-flags", "--retry --follow=name", "flags for tail command, you can use -f instead if your server does't support `--retry --follow=name` flags")
var slient = flag.Bool("slient", false, "-slient=false")

var Version = ""
var GitCommit = ""

func usageAndExit(message string) {

	if message != "" {
		fmt.Fprintln(os.Stderr, message)
	}

	flag.Usage()
	fmt.Fprint(os.Stderr, "\n")

	os.Exit(1)
}

func printWelcomeMessage(config command.Config) {
	fmt.Println(welcomeMessage)

	for _, server := range config.Servers {
		// If there is no tail_file for a service configuration, the global configuration is used
		if server.TailFile == "" {
			server.TailFile = config.TailFile
		}

		serverInfo := fmt.Sprintf("%s@%s:%s", server.User, server.Hostname, server.TailFile)
		fmt.Println(console.ColorfulText(console.TextMagenta, serverInfo))
	}
	fmt.Printf("\n%s\n", console.ColorfulText(console.TextCyan, mossSep))
}

func parseConfig(filePath string, hostStr string, configFile string, slient bool, tailFlags string) (config command.Config) {
	if configFile != "" {
		if _, err := toml.DecodeFile(configFile, &config); err != nil {
			log.Fatal(err)
		}

	} else {

		hosts := strings.Split(hostStr, ",")

		config = command.Config{}
		config.TailFile = filePath
		config.Servers = make(map[string]command.Server, len(hosts))
		config.Slient = slient
		config.TailFlags = tailFlags

		for index, hostname := range hosts {
			hostInfo := strings.Split(strings.Replace(hostname, ":", "@", -1), "@")
			var port int
			if len(hostInfo) > 2 {
				port, _ = strconv.Atoi(hostInfo[2])
			}
			config.Servers["server_"+string(index)] = command.Server{
				ServerName: "server_" + string(index),
				Hostname:   hostInfo[1],
				User:       hostInfo[0],
				Port:       port,
			}
		}
	}

	if config.TailFlags == "" {
		config.TailFlags = "--retry --follow=name"
	}

	return
}

func main() {

	flag.Usage = func() {
		fmt.Fprint(os.Stderr, welcomeMessage)
		fmt.Fprint(os.Stderr, "Options:\n\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	if (*filePath == "" || *hostStr == "") && *configFile == "" {
		usageAndExit("")
	}

	config := parseConfig(*filePath, *hostStr, *configFile, *slient, *tailFlags)
	if !config.Slient {
		printWelcomeMessage(config)
	}

	outputs := make(chan command.Message, 255)
	var wg sync.WaitGroup

	for _, server := range config.Servers {
		wg.Add(1)
		go func(server command.Server) {
			defer func() {
				if err := recover(); err != nil {
					fmt.Printf(console.ColorfulText(console.TextRed, "Error: %s\n"), err)
				}
			}()
			defer wg.Done()

			// If there is no tail_file for a service configuration, the global configuration is used
			if server.TailFile == "" {
				server.TailFile = config.TailFile
			}

			if server.TailFlags == "" {
				server.TailFlags = config.TailFlags
			}

			if server.User == "" {
				server.User = config.User
			}

			if server.Password == "" {
				server.Password = config.Password
			}

			if server.PrivateKeyPath == "" {
				server.PrivateKeyPath = config.PrivateKeyPath
			}

			// If the service configuration does not have a port, the default value of 22 is used
			if server.Port == 0 {
				server.Port = 22
			}

			cmd := command.NewCommand(server)
			cmd.Execute(outputs)
		}(server)
	}

	if len(config.Servers) > 0 {
		go func() {
			for output := range outputs {
				content := strings.Trim(output.Content, "\r\n")
				// 去掉文件名称输出
				if content == "" || (strings.HasPrefix(content, "==>") && strings.HasSuffix(content, "<==")) {
					continue
				}

				if config.Slient {
					fmt.Printf("%s -> %s\n", output.Host, content)
				} else {
					fmt.Printf(
						"%s %s %s\n",
						console.ColorfulText(console.TextGreen, output.Host),
						console.ColorfulText(console.TextYellow, "->"),
						content,
					)
				}
			}
		}()
	} else {
		fmt.Println(console.ColorfulText(console.TextRed, "No target host is available"))
	}

	wg.Wait()
}

func getWelcomeMessage() string {
	return `
 ____                      _      _____     _ _
|  _ \ ___ _ __ ___   ___ | |_ __|_   _|_ _(_) |
| |_) / _ \ '_ ' _ \ / _ \| __/ _ \| |/ _' | | |
|  _ <  __/ | | | | | (_) | ||  __/| | (_| | | |
|_| \_\___|_| |_| |_|\___/ \__\___||_|\__,_|_|_|

Author: liuwangchen
Homepage: github.com/liuwangchen/remote-tail
Version: ` + Version + "(" + GitCommit + ")" + `
`
}
