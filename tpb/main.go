package main

import (
	"fmt"
	"text/template"
	//"io/ioutil"
	"log"
	"net/http"
	"regexp"

	"github.com/gorilla/mux"
	"github.com/yhat/scrape"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

var re = regexp.MustCompile("Uploaded ([^,]+), Size ([^,]+), ULed by (.+)")

type Torrent struct {
	Category, Title, Magnet, Size, Uploaded, Uploader, Seeders, Leechers string
}

func ParseRecord(n *html.Node) Torrent {
	tds := scrape.FindAll(n, scrape.ByTag(atom.Td))
	var size, uptime, uploader string
	if len(tds) == 4 {
		cat := scrape.Text(tds[0])[0:3]
		name, magnet, desc := ParseName(tds[1])
		matches := re.FindStringSubmatch(desc)
		uptime, size, uploader = matches[1], matches[2], matches[3]
		seed := scrape.Text(tds[2])
		leech := scrape.Text(tds[3])
		return Torrent{cat, name, magnet, size, uptime, uploader, seed, leech}
	} else {
		fmt.Println("Error: not expected format")
	}
	return Torrent{}
}

func ParseName(n *html.Node) (string, string, string) {
	matcher := func(n *html.Node) bool {
		// must check for nil values
		if n.DataAtom == atom.A && n.Parent.DataAtom == atom.Td {
			return true
		}
		return false
	}

	var name, magnet, desc string

	if detName, ok := scrape.Find(n, scrape.ByClass("detName")); ok {
		name = scrape.Text(detName)
	}
	if anchor, ok := scrape.Find(n, matcher); ok {
		magnet = scrape.Attr(anchor, "href")
	}
	if detDesc, ok := scrape.Find(n, scrape.ByClass("detDesc")); ok {
		desc = scrape.Text(detDesc)
	}
	return name, magnet, desc
}

func main2() {
	TorrentList("https://thepiratebay.la/top/201")
}
func main() {
	r := mux.NewRouter()

	//data, err := Asset("data/tmpl.html")

	//if err != nil {
	//panic(err)
	//}
	funcMap := template.FuncMap{
		"add": func(x, y int) int {
			return x + y
		},
	}

	tmpl := template.Must(template.New("tpb").Funcs(funcMap).Parse(TMPL))

	r.HandleFunc("/{cat}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		category := vars["cat"]
		url := fmt.Sprintf("https://thepiratebay.la/top/%s", category)
		torrents, err := TorrentList(url)
		err = tmpl.Execute(w, torrents)
		if err != nil {
			panic(err)
		}

	})
	r.HandleFunc("/search/{query}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		query := vars["query"]
		url := fmt.Sprintf("https://thepiratebay.la/search/%s", query)
		torrents, err := TorrentList(url)
		err = tmpl.Execute(w, torrents)
		if err != nil {
			panic(err)
		}

	})
	go http.ListenAndServe("0.0.0.0:8080", r)
	log.Println("listening on port 8080")
	select {}
}

func TorrentList(url string) ([]Torrent, error) {
	// request and parse the front page
	resp, err := http.Get(url)
	if err != nil {
		return make([]Torrent, 0), err
	}
	root, err := html.Parse(resp.Body)
	if err != nil {
		return make([]Torrent, 0), err
	}
	var torrents []Torrent
	if content, ok := scrape.Find(root, scrape.ById("searchResult")); ok {
		// define a matcher
		matcher := func(n *html.Node) bool {
			// must check for nil values
			if n.DataAtom == atom.Tr && n.Parent.DataAtom == atom.Tbody {
				return true
			}
			return false
		}
		// grab all articles and print them
		trs := scrape.FindAll(content, matcher)
		for _, tr := range trs {
			torrents = append(torrents, ParseRecord(tr))
		}
	}
	resp.Body.Close()
	return torrents, nil
}
