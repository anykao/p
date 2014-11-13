package main

import (
	"fmt"
	//"os.signal"
	"github.com/apcera/nats"
	"github.com/skratchdot/open-golang/open"
	//"github.com/rakyll/coop"
)

const(
	TOPIC = "url"
	QUEUE = "worker_group"
)

func main() {
	nc, _ := nats.Connect("nats://172.17.106.112:4222")
	defer nc.Close()
	nc.QueueSubscribe(TOPIC, QUEUE, func(m *nats.Msg) {
		fmt.Println(string(m.Data))
		open.Start( string(m.Data))
	})
	select {
	}
}
