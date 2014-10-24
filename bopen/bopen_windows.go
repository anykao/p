package bopen

import (
	"github.com/skratchdot/open-golang/open"
)

func (b Bopen) Open(string url) {
	open.Start(url)
}
