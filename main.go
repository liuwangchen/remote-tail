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

var configFile string
var label string
var file string

var Version = "2.0"

func usageAndExit(message string) {

	if message != "" {
		_, _ = fmt.Fprintln(os.Stderr, message)
	}

	flag.Usage()
	fmt.Println("remote-tail config label file")
	_, _ = fmt.Fprint(os.Stderr, "\n")

	os.Exit(1)
}

func printWelcomeMessage() {
	fmt.Println(welcomeMessage)

	for _, server := range viper.GetStringSlice(label) {
		serverInfo := fmt.Sprintf("%s@%s:%s", viper.GetString("user"), server, viper.GetString("file."+file))
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
	args := flag.Args()
	if len(args) < 2 {
		usageAndExit("")
	}
	configFile = args[0]
	subArgs := args[1]
	label = strings.Split(subArgs, ".")[0]
	file = strings.Split(subArgs, ".")[1]
	viper.SetConfigFile(configFile)
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}

	printWelcomeMessage()

	outputs := make(chan command.Message, 255)
	var wg sync.WaitGroup

	for _, server := range viper.GetStringSlice(label) {
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
			TailFile:       viper.GetString("file." + file),
		})
	}
	if len(viper.GetStringSlice(label)) > 0 {
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
Version: ` + Version + `
`
}
