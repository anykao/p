package main

import (
	"log"
	"os"
	"time"
	"github.com/apcera/nats"
	"github.com/skratchdot/open-golang/open"
	//"github.com/rakyll/coop"
)

const(
	TOPIC = "url"
	QUEUE = "worker_group"
)

func main() {
	f,err := os.Create("opener.log")
	if err != nil{
		log.Panic(err)
	}
	logger := log.New(f, "logger: ", log.Lshortfile)
	nc, _ := nats.Connect("nats://yourhost:4222")
	defer nc.Close()
	logger.Println("start polling url")
	nc.QueueSubscribe(TOPIC, QUEUE, func(m *nats.Msg) {
		open.Start( string(m.Data))
	})
	time.Sleep(8 * time.Hour)
	logger.Println("end polling")
}
