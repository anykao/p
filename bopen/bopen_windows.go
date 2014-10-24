package bopen

import (
	"github.com/skratchdot/open-golang/open"
)

func (b Bopen) Open(url string) {
	open.Start(url)
}
