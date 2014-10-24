package bopen

import (
	"github.com/apcera/nats"
)

func (b Bopen) Open(url string) {
	nc, _ := nats.Connect(nats.DefaultURL)
	defer nc.Close()
	if b.Topic != "" {
		nc.Publish(b.Topic, []byte(url))
	} else {
		nc.Publish("url", []byte(url))
	}
}
