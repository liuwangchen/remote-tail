package main

import (
	"flag"
	"fmt"
	"github.com/spf13/viper"
	"os"
	"strings"
	"sync"

	"github.com/mylxsw/remote-tail/command"
	"github.com/mylxsw/remote-tail/console"
)

var mossSep = ".--. --- .-- . .-. . -..   -... -.--   -- -.-- .-.. -..- ... .-- \n"

var welcomeMessage = getWelcomeMessage() + console.ColorfulText(console.TextMagenta, mossSep)

var configFile = flag.String("conf", "/Users/liuwangchen/work/bash/remotetail/live.conf", "-conf=example.conf")
var label = flag.String("label", "", "-label=test")

var Version = ""
var GitCommit = ""

func usageAndExit(message string) {

	if message != "" {
		_, _ = fmt.Fprintln(os.Stderr, message)
	}

	flag.Usage()
	_, _ = fmt.Fprint(os.Stderr, "\n")

	os.Exit(1)
}

func printWelcomeMessage() {
	fmt.Println(welcomeMessage)

	for _, server := range viper.GetStringSlice(*label) {
		serverInfo := fmt.Sprintf("%s@%s:%s", viper.GetString("user"), server, viper.GetString("tail_file"))
		fmt.Println(console.ColorfulText(console.TextMagenta, serverInfo))
	}
	fmt.Printf("\n%s\n", console.ColorfulText(console.TextCyan, mossSep))
}

func main() {

	flag.Usage = func() {
		_, _ = fmt.Fprint(os.Stderr, welcomeMessage)
		_, _ = fmt.Fprint(os.Stderr, "Options:\n\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	if *configFile == "" || *label == "" {
		usageAndExit("")
	}
	viper.SetConfigFile(*configFile)
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}

	printWelcomeMessage()

	outputs := make(chan command.Message, 255)
	var wg sync.WaitGroup

	for _, server := range viper.GetStringSlice(*label) {
		wg.Add(1)
		go func(server command.Server) {
			defer func() {
				if err := recover(); err != nil {
					fmt.Printf(console.ColorfulText(console.TextRed, "Error: %s\n"), err)
				}
			}()
			defer wg.Done()
			cmd := command.NewCommand(server)
			cmd.Execute(outputs)
		}(command.Server{
			ServerName:     "",
			Hostname:       server,
			Port:           22,
			User:           viper.GetString("user"),
			Password:       viper.GetString("password"),
			PrivateKeyPath: viper.GetString("private_key_path"),
			TailFile:       viper.GetString("tail_file"),
		})
	}
	if len(viper.GetStringSlice(*label)) > 0 {
		go func() {
			for output := range outputs {
				content := strings.Trim(output.Content, "\r\n")
				// 去掉文件名称输出
				if content == "" || (strings.HasPrefix(content, "==>") && strings.HasSuffix(content, "<==")) {
					continue
				}

				fmt.Printf(
					"%s %s %s\n",
					console.ColorfulText(console.TextGreen, output.Host),
					console.ColorfulText(console.TextYellow, "->"),
					content,
				)
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
