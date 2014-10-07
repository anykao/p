package main

import (
	"encoding/json"
	"github.com/codegangsta/cli"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/skratchdot/open-golang/open"
)

type Item struct {
	CommentsCount int    `json:"comments_count"`
	Domain        string `json:"domain"`
	ID            string `json:"id"`
	Points        int    `json:"points"`
	TimeAgo       string `json:"time_ago"`
	Title         string `json:"title"`
	Type          string `json:"type"`
	URL           string `json:"url"`
	User          string `json:"user"`
}

type Comment struct {
	Comments []Comment `json:"comments"`
	Content  string    `json:"content"`
	ID       string    `json:"id"`
	Level    float64   `json:"level"`
	TimeAgo  string    `json:"time_ago"`
	User     string    `json:"user"`
}

const (
	NEWS      = "http://node-hnapi.herokuapp.com/news"
	NEWS2     = "http://node-hnapi.herokuapp.com/news2"
	ITEM      = "http://node-hnapi.herokuapp.com/item"
	HACKERWEB = "http://cheeaun.github.io/hackerweb"
)

func populateCache(name string, contents []byte) {
	e := os.MkdirAll(filepath.Dir(name), 0755)
	if e != nil {
		log.Fatal(e)
	}
	f, e := os.Create(name)
	if e != nil {
		log.Fatal(e)
	}
	defer f.Close()

	f.Write(contents)
}
func getItem(cache string, idx int64) Item {
	f, e := os.Open(cache)
	if e != nil {
		log.Fatal(e)
	}
	buf, e := ioutil.ReadAll(f)
	if e != nil {
		log.Fatal(e)
	}
	var items []Item
	e = json.Unmarshal(buf, &items)
	if e != nil {
		log.Fatal(e)
	}
	return items[idx-1]
}
func showNewsList(news []Item) {

	table := tablewriter.NewWriter(os.Stdout)
	table.SetColWidth(158)
	table.SetHeader([]string{"Index", "Cmts", "Domain", "Title"})
	for i, item := range news {
		table.Append([]string{strconv.Itoa(i + 1), strconv.Itoa(item.CommentsCount), item.Domain, item.Title})
	}
	table.Render()
}

func main() {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	cache := filepath.Join(usr.HomeDir, ".ak", "news")

	app := cli.NewApp()
	app.Name = "hn"
	app.Usage = "hacker news under your finger."
	app.Action = func(c *cli.Context) {
		var news []Item
		res, err := http.Get(NEWS)
		if err != nil {
			log.Fatal(err)
		}
		bytes, err := ioutil.ReadAll(res.Body)
		if err != nil {
			log.Fatal(err)
		}
		res.Body.Close()
		populateCache(cache, bytes)
		contents := string(bytes)
		err = json.NewDecoder(strings.NewReader(contents)).Decode(&news)
		showNewsList(news)
	}
	app.Commands = []cli.Command{
		{
			Name:      "view",
			ShortName: "v",
			Usage:     "view news",
			Action: func(c *cli.Context) {
				idx, e := strconv.ParseInt(c.Args().First(), 0, 0)
				if e != nil {
					log.Fatal(e)
				}
				item := getItem(cache, idx)
				open.Start(item.URL)
			},
		},
		{
			Name:      "news2",
			ShortName: "n",
			Usage:     "show news2",
			Action: func(c *cli.Context) {
				var news []Item
				res, err := http.Get(NEWS2)
				if err != nil {
					log.Fatal(err)
				}
				bytes, err := ioutil.ReadAll(res.Body)
				if err != nil {
					log.Fatal(err)
				}
				res.Body.Close()
				populateCache(cache, bytes)
				contents := string(bytes)
				err = json.NewDecoder(strings.NewReader(contents)).Decode(&news)
				if err != nil {
					log.Fatal(err)
				}
				showNewsList(news)
			},
		},
		{
			Name:      "comment",
			ShortName: "c",
			Usage:     "show comment",
			Action: func(c *cli.Context) {
				idx, e := strconv.ParseInt(c.Args().First(), 0, 0)
				if e != nil {
					log.Fatal(e)
				}
				item := getItem(cache, idx)
				open.Start(HACKERWEB + "/#/item/" + item.ID)
			},
		}}
	app.Run(os.Args)
}
