package main

import (
	"flag"
	"fmt"
	"text/template"

	"github.com/skratchdot/open-golang/open"
)

func main() {
	flag.Parse()

	url := template.URLQueryEscaper(flag.Args())

	url = fmt.Sprintf("http://52.68.184.19:8080/search/%s", url)

	fmt.Printf("open %s\n", url)

	open.Start(url)

}
