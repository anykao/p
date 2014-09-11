package main

import (
	"fmt"
	"time"
	"flag"
	"os"
	"path/filepath"

	"code.google.com/p/go.crypto/ssh"
	"github.com/BurntSushi/toml"
)

var config string

type server struct {
	IP       string
	User     string
	Password string
	Cmd      string
}

type tomlConfig struct {
	Servers  []server `toml:"server"`
	User     string
	Password string
}

func init(){
	flag.StringVar(&config, "conf", "", "config file.")
}

func executeCmd(cmd, hostname string, config *ssh.ClientConfig) string {
	conn, _ := ssh.Dial("tcp", hostname+":22", config)
	session, _ := conn.NewSession()
	defer session.Close()

	output, err := session.CombinedOutput(cmd)

	if err != nil {
		panic(err)
	}

	return "\n\n<<" + hostname + ">>\n" + string(output)
}

func main() {
	flag.Parse()
	var conf tomlConfig
	var command string
	workDir, _ := os.Getwd()

	if config == "" {
		config =filepath.Join(workDir, "config.toml")
	}

	if flag.NArg() > 0 {
		command = flag.Arg(0)
	}

	if _, err := toml.DecodeFile(config, &conf); err != nil {
		panic(err)
	}
	servs := conf.Servers
	var count int = len(servs)
	results := make(chan string, 2)
	timeout := time.After(30 * time.Second)

	for _, serv := range servs {
		go func(serv server) {
			var user, pass, cmd string
			if serv.User == "" {
				user = conf.User
			} else {
				user = serv.User
			}

			if serv.Password == "" {
				pass = conf.Password
			} else {
				pass = serv.Password
			}

			if command == "" {
				cmd = serv.Cmd
			} else {
				cmd = command
			}

			config := &ssh.ClientConfig{
				User: user,
				Auth: []ssh.AuthMethod{
					ssh.Password(pass),
				},
			}
			results <- executeCmd(cmd, serv.IP, config)
		}(serv)
	}

	for i := 0; i < count; i++ {
		select {
		case res := <-results:
			fmt.Print(res)
		case <-timeout:
			fmt.Println("Timed out!")
			return
		}
	}
}
