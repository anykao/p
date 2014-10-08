package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"

	//"github.com/olekukonko/tablewriter"

	"github.com/skratchdot/open-golang/open"
	"github.com/wsxiaoys/terminal/color"
)

var (
	next    bool
	comment bool
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

//func showNewsList(news []Item) {

//table := tablewriter.NewWriter(os.Stdout)
//table.SetColWidth(158)
//table.SetHeader([]string{"Index", "Cmts", "Domain", "Title"})
//for i, item := range news {
//table.Append([]string{strconv.Itoa(i + 1), strconv.Itoa(item.CommentsCount), item.Domain, item.Title})
//}
//table.Render()
//}

func showNewsList(news []Item) {
	for i, item := range news {
		color.Print("@b", fmt.Sprintf("%2d", i+1))
		fmt.Print(" ")
		color.Print("@{!g}", item.Title)
		fmt.Print(" ")
		color.Print("@y", fmt.Sprintf("%dc", item.CommentsCount))
		fmt.Print(" ")
		color.Print("@m", fmt.Sprintf("%dp", item.Points))
		fmt.Print(" ")
		color.Print("@w", item.Domain)
		fmt.Println()
	}
}
func init() {
	flag.BoolVar(&next, "n", false, "show new2.")
	flag.BoolVar(&comment, "c", false, "show comment.")
}
func main() {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	cache := filepath.Join(usr.HomeDir, ".ak", "news")

	flag.Parse()
	if flag.NArg() > 0 {
		idx, e := strconv.ParseInt(flag.Arg(0), 0, 0)
		if e != nil {
			// not int
		}
		item := getItem(cache, idx)
		if comment {
			open.Start(HACKERWEB + "/#/item/" + item.ID)
		} else {
			open.Start(item.URL)
		}

	} else {
		if next {
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
		} else {
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
	}
}
