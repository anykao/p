package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	. "github.com/azer/debug"
	"github.com/google/go-github/github"
	"github.com/skratchdot/open-golang/open"
	"github.com/wsxiaoys/terminal/color"
)

var (
	lang   string
	query  string
	clone  bool
	browse bool
)

var repos []github.Repository
var cache string
var name string

const (
	BASE_URL = "https://github.com/"
	URL      = "https://github.com/trending"
	README   = "https://raw.githubusercontent.com/"
	CACHE    = "cache.json"
)

func prettyPrint(repos []github.Repository) {
	for i, repo := range repos {
		fmt.Print("[")
		color.Print("@b", i+1)
		fmt.Print("]")
		fmt.Print(*repo.FullName)
		if repo.StargazersCount != nil {
			color.Print("@r" + " " + strconv.Itoa(*repo.StargazersCount), "☆")
		}
		if repo.Language != nil {
			color.Print("@c" + " " + *repo.Language)
		}
		if repo.Description != nil {
			color.Print("@g" + " " + *repo.Description)
		}
		fmt.Println()
	}
}

func f1(i int, s *goquery.Selection) {
	name := s.Find(".repository-name").Text()
	lang := s.Find(".title-meta").Text()
	desc := s.Find(".repo-leaderboard-description").Text()
	r := github.Repository{FullName: &name, Description: &desc, Language: &lang}
	repos = append(repos, r)

}

func view(repo string) {
	name = repo
	url := BASE_URL + name
	Debug(url)
	if clone {
		cmd := exec.Command("git", "clone", url)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Run()
	} else {
		open.Start(url)
	}
}

func doQuery(querystr string) {
	Debug(querystr)
	client := github.NewClient(nil)

	opts := &github.SearchOptions{}
	result, _, err := client.Search.Repositories(querystr, opts)
	if err != nil {
		panic(err)
	}
	repos = result.Repositories
}

func populateCache(cache string) {
	e := os.MkdirAll(filepath.Dir(cache), 0666)
	if e != nil {
		log.Fatal(e)
	}
	f, e := os.Create(cache)
	if e != nil {
		log.Fatal(e)
	}
	defer f.Close()
	contents, e := json.Marshal(repos)
	if e != nil {
		log.Fatal(e)
	}
	f.Write(contents)
}

func main() {
	flag.StringVar(&lang, "l", "", "language")
	flag.StringVar(&query, "q", "", "query")
	flag.BoolVar(&clone, "c", false, "clone")
	flag.BoolVar(&browse, "b", false, "browse")
	flag.Parse()
	usr, e := user.Current()
	if e != nil {
		log.Fatal(e)
	}
	cache = filepath.Join(usr.HomeDir, ".gh", CACHE)
	if flag.NArg() > 0 {
		repo := flag.Arg(0)
		if strings.Index(repo, "/") > 0 {
			view(repo)
		} else {
			id, e := strconv.ParseInt(repo, 0, 0)
			if e == nil && id < 26 && id > 0 {
				fi, e := os.Stat(cache)
				if os.IsNotExist(e) {
					log.Fatal("Please run 'gh' to populate the cache.")
				}
				modtime := fi.ModTime()
				if e != nil {
					log.Fatal(e)
				}
				// 5分後無効する
				if time.Now().Sub(modtime) > 5 * time.Minute {
					fmt.Println("Cache is too old.")
					fmt.Println("Please rerun 'gh' to populate the cache.")
					os.Exit(0)
				}
				Debug(modtime.Format(time.ANSIC))
				f, e := os.Open(cache)
				if e != nil {
					log.Fatal(e)
				}
				buf, e := ioutil.ReadAll(f)
				if e != nil {
					log.Fatal(e)
				}
				var cachedRepos []github.Repository
				e = json.Unmarshal(buf, &cachedRepos)
				if e != nil {
					log.Fatal(e)
				}
				repoName := *cachedRepos[id-1].FullName
				view(repoName)
			}
		}

	} else {

		var url string
		if query != "" {
			doQuery(query)
		} else {
			if lang != "" {
				url = URL + "?l=" + lang
			} else {
				url = URL
			}
			Debug(url)
			doc, e := goquery.NewDocument(url)
			if e != nil {
				log.Fatal(e)
			}
			doc.Find(".leaderboard-list-content").Each(f1)
		}
		prettyPrint(repos)
		populateCache(cache)
	}
}
