package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"code.google.com/p/go.crypto/ssh"
	"github.com/BurntSushi/toml"
)

var config string

type logInfo struct {
	timestamp time.Time
	host      string
}

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

//var logChan = make(chan, logInfo)

func init() {
	flag.StringVar(&config, "conf", "", "config file.")
}

func executeCmd(cmd, hostname string, config *ssh.ClientConfig) string {
	conn, err := ssh.Dial("tcp", hostname+":22", config)
	if err != nil {
		log.Fatal(err)
	}
	//fmt.Println(time.Now(), hostname, "start new session")
	session, err := conn.NewSession()
	if err != nil {
		log.Fatal(err)
	}
	defer session.Close()

	output, err := session.CombinedOutput(cmd)

	if err != nil {
		log.Print("ERR: ",err)
	}

	return "\n\n<<" + hostname + ">>\n" + string(output)
}

func main() {
	//go func() {
		//log.Println(http.ListenAndServe(":6060", nil))
	//}()
	log.SetFlags(log.LstdFlags | log.Llongfile)

	flag.Parse()
	var conf tomlConfig
	var command string
	workDir, _ := os.Getwd()

	if config == "" {
		config = filepath.Join(workDir, "config.toml")
	}

	if flag.NArg() > 0 {
		command = flag.Arg(0)
	}

	if _, err := toml.DecodeFile(config, &conf); err != nil {
		log.Fatal(err)
	}
	servs := conf.Servers
	var count int = len(servs)
	results := make(chan string)
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
