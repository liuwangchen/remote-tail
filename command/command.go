package command

import (
	"bufio"
	"fmt"
	"io"

	"github.com/mylxsw/remote-tail/console"
	"github.com/mylxsw/remote-tail/ssh"
)

type Command struct {
	Host   string
	User   string
	Script string
	Stdout io.Reader
	Stderr io.Reader
	Server Server
}

// Message The message used by channel to transport log line by line
type Message struct {
	Host    string
	Content string
}

// NewCommand Create a new command
func NewCommand(server Server) (cmd *Command) {
	script := fmt.Sprintf("tail -f %s", server.TailFile)
	if server.TailLine > 0 {
		script = fmt.Sprintf("tail -n %d %s", server.TailLine, server.TailFile)
	}
	cmd = &Command{
		Host:   server.Hostname,
		User:   server.User,
		Script: script,
		Server: server,
	}

	//if !strings.Contains(cmd.Host, ":") {
	//	cmd.Host = cmd.Host + ":" + strconv.Itoa(server.Port)
	//}

	return
}

// Execute the remote command
func (cmd *Command) Execute(output chan Message) {

	client := &ssh.Client{
		Host:           cmd.Host + ":22",
		User:           cmd.User,
		Password:       cmd.Server.Password,
		PrivateKeyPath: cmd.Server.PrivateKeyPath,
	}

	if err := client.Connect(); err != nil {
		panic(fmt.Sprintf("[%s] unable to connect: %s", cmd.Host, err))
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		panic(fmt.Sprintf("[%s] unable to create session: %s", cmd.Host, err))
	}
	defer session.Close()

	if err := session.RequestPty("xterm", 80, 40, *ssh.CreateTerminalModes()); err != nil {
		panic(fmt.Sprintf("[%s] unable to create pty: %v", cmd.Host, err))
	}

	cmd.Stdout, err = session.StdoutPipe()
	if err != nil {
		panic(fmt.Sprintf("[%s] redirect stdout failed: %s", cmd.Host, err))
	}

	go bindOutput(cmd.Host, output, &cmd.Stdout, "", 0)

	if err = session.Start(cmd.Script); err != nil {
		panic(fmt.Sprintf("[%s] failed to execute command: %s", cmd.Host, err))
	}

	if err = session.Wait(); err != nil {
		panic(fmt.Sprintf("[%s] failed to wait command: %s", cmd.Host, err))
	}
}

// bing the pipe output for formatted output to channel
func bindOutput(host string, output chan Message, input *io.Reader, prefix string, color int) {
	reader := bufio.NewReader(*input)
	for {
		line, _, err := reader.ReadLine()
		if err != nil || io.EOF == err {
			if err != io.EOF {
				panic(fmt.Sprintf("[%s] faield to execute command: %s", host, err))
			}
			close(output)
			break
		}
		lineStr := string(line)
		lineStr = prefix + lineStr
		if color != 0 {
			lineStr = console.ColorfulText(color, lineStr)
		}

		output <- Message{
			Host:    host,
			Content: lineStr,
		}
	}
}
