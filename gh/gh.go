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
	"bufio"

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

var cache string
var name string

const (
	BASE_URL = "https://github.com/"
	URL      = "https://github.com/trending"
	CACHE    = "cache.json"
)

func prettyPrint(repo github.Repository, index int) {
	fmt.Print("[")
	color.Print("@b", index)
	fmt.Print("]")
	fmt.Print(*repo.FullName)
	if repo.StargazersCount != nil {
		color.Print("@r"+" "+strconv.Itoa(*repo.StargazersCount), "☆")
	}
	if repo.Language != nil {
		color.Print("@c" + " " + *repo.Language)
	}
	if repo.Description != nil {
		color.Print("@g" + " " + *repo.Description)
	}
	fmt.Println()
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

func populateCache(cache string, repos []github.Repository) {
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
func trimMultiLineString(origin string)[]string{
	var trimed [] string
	scanner := bufio.NewScanner(strings.NewReader(origin))
	for scanner.Scan() {
		line := strings.Trim(scanner.Text(), " ")
		if (line != ""){
			trimed = append(trimed, line)
		}
	}
	return trimed
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
				if time.Now().Sub(modtime) > 5*time.Minute {
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
		repos := make(chan github.Repository)
		var url string
		if query != "" {
			Debug(query)
			client := github.NewClient(nil)

			opts := &github.SearchOptions{}
			result, _, err := client.Search.Repositories(query, opts)
			if err != nil {
				panic(err)
			}
			go func(){
				for _, repo := range result.Repositories {
					repos <- repo
				}
				close(repos)
			}()
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
			go func(){
				doc.Find("li.repo-list-item").Each(func(i int, s *goquery.Selection) {
					name := strings.Join(trimMultiLineString(s.Find("h3.repo-list-name a").Text()),"")
					lang := trimMultiLineString(s.Find(".repo-list-meta").Text())[0]
					desc := strings.Trim(s.Find(".repo-list-description").Text(), "\n ")
					repo := github.Repository{FullName: &name, Description: &desc, Language: &lang}
					repos <- repo
				})
				close(repos)
			}()
		}
		var repoArr []github.Repository
		var counter int
		for repo := range repos {
			counter++
			prettyPrint(repo, counter)
			repoArr = append(repoArr, repo)
		}
		populateCache(cache, repoArr)
	}
}

