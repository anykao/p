package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	_ "net/http/pprof"
	log "github.com/Sirupsen/logrus"
	//"github.com/kr/pretty"
	"github.com/mholt/binding"
	"github.com/pmezard/go-difflib/difflib"
	"github.com/zenazn/goji"
	"github.com/zenazn/goji/web"
	"gopkg.in/unrolled/render.v1"
	"github.com/skratchdot/open-golang/open"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

var wd string
var tempDir string
var err error

const (
	// Version number
	VERSION = "0.01"

	// Scan at up to this size in file for '\0' in test for binary file
	BINARY_CHECK_SIZE = 65536
	// default number of context lines to display
	CONTEXT_LINES = 3
	MSG_BIN_FILE_DIFFERS = "File differs. This is a binary file"
)

type Config struct {
	FileA *multipart.FileHeader
	FileB *multipart.FileHeader
	Verbose *bool
}

// Then provide a field mapping (pointer receiver is vital)
func (conf *Config) FieldMap() binding.FieldMap {
	return binding.FieldMap{
		&conf.FileA: "file-a",
		&conf.FileB: "file-b",
		&conf.Verbose: "verbose",
	}
}

func init() {
	wd, err = os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	tempDir = filepath.Join(wd, "tmp")
	err = os.MkdirAll(tempDir, 0755)
	if err != nil {
		log.Fatal(err)
	}
}
func DiffContent(zipa, zipb *zip.Reader) string {
	var buf bytes.Buffer
	rightFileSet := make(map[string]struct{})
	for _, f := range zipb.File {
		rightFileSet[f.Name] = struct{}{}
	}
	var hitFlag bool
	for _, a := range zipa.File {
		hitFlag = false
		for _, b := range zipb.File {
			if a.Name == b.Name {
				hitFlag = true
				delete(rightFileSet, b.Name)
				a1, _ := a.Open()
				b1, _ := b.Open()
				data1, _ := ioutil.ReadAll(a1)
				data2, _ := ioutil.ReadAll(b1)
				if !bytes.Equal(data1, data2) {
					fmt.Fprintln(&buf,a.Name)
					if strings.HasSuffix(a.Name,"jar") {
						zipa, err := zip.NewReader(bytes.NewReader(data1), int64(len(data1)))
						if (err != nil){
							log.Fatal(err)
						}
						zipb, err := zip.NewReader(bytes.NewReader(data2), int64(len(data2)))
						if (err != nil){
							log.Fatal(err)
						}
						diff := DiffContent(zipa, zipb)
						fmt.Fprintf(&buf, diff)
					} else if (!strings.HasSuffix(a.Name, "class") && check_binary(data1)) {
						fmt.Fprintln(&buf, MSG_BIN_FILE_DIFFERS)
					} else{
						var leftLines, rightLines []string
						if strings.HasSuffix(a.Name, "class") {
							leftLines = difflib.SplitLines(string(jadfile(data1)))
							rightLines = difflib.SplitLines(string(jadfile(data2)))
						} else {
							leftLines = difflib.SplitLines(string(data1))
							rightLines = difflib.SplitLines(string(data2))
						}
						diff := difflib.UnifiedDiff{
							A:        leftLines,
							B:        rightLines,
							FromFile: "File A",
							ToFile:   "File B",
							Context:  3,
						}
						text, _ := difflib.GetUnifiedDiffString(diff)
						fmt.Fprint(&buf,text)
					}
				}
				break
			}
		}
		if !hitFlag {
			fmt.Fprintf(&buf, "%s only in the left file.\n", a.Name)
		}
	}
	if len(rightFileSet) != 0 {
		for f := range rightFileSet {
			fmt.Fprintf(&buf,"%s only in the right file.\n", f)
		}
	}
	return buf.String()
}

func console() {
	if len(os.Args) < 3 {
		fmt.Println(`Usage: diffjar left-file right-file`)
		os.Exit(1)
	}
	jar1, _ := zip.OpenReader(os.Args[1])
	jar2, _ := zip.OpenReader(os.Args[2])
	defer jar1.Close()
	defer jar2.Close()
	DiffContent(&jar1.Reader, &jar2.Reader)
}
func jadfile(data []byte) []byte {
	fclass := filepath.Join(tempDir, "tmp.class")
	tmpClass, err := os.Create(fclass)
	if err != nil {
		log.Fatal(err)
	}
	tmpClass.Write(data)
	tmpClass.Close()
	cmdjad := filepath.Join(wd, "tools", "jad.exe")
	cmd := exec.Command(cmdjad, "-p", fclass)
	output, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
	}
	return output
}

func index(c web.C, w http.ResponseWriter, r *http.Request) {
	ren := render.New(render.Options{})
	ren.HTML(w, http.StatusOK, "index", nil)
}
func PostDiff(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	//fmt.Printf("%# v", pretty.Formatter(r))
	conf := new(Config)
	errs := binding.Bind(r, conf)
	if errs.Handle(w) {
		return
	}
	fa, err := conf.FileA.Open()
	if (err != nil){
		log.Fatal(err)
	}
	fb, err := conf.FileB.Open()
	if (err != nil){
		log.Fatal(err)
	}
	readFromA, err := ioutil.ReadAll(fa)
	if (err != nil){
		log.Fatal(err)
	}
	readFromB, err := ioutil.ReadAll(fb)
	if (err != nil){
		log.Fatal(err)
	}

	zipa, err := zip.NewReader(fa, int64(len(readFromA)))
	if (err != nil){
		log.Fatal(err)
	}
	zipb, err := zip.NewReader(fb, int64(len(readFromB)))
	if (err != nil){
		log.Fatal(err)
	}
	diff := DiffContent(zipa, zipb)
	fmt.Fprint(w, "verbose: ", conf.Verbose)
	fmt.Fprintf(w, diff)
}

func min_int(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// check if file is binary
func check_binary(data []byte) bool {
	if data == nil {
		return false
	}
	if len(data) == 0 {
		return false
	}
	if bytes.IndexByte(data[0:min_int(len(data), BINARY_CHECK_SIZE)], 0) >= 0 {
		return true
	}
	return false
}

func site() {
	goji.Get("/", index)
	goji.Post("/diff", PostDiff)
	//Fully backwards compatible with net/http's Handlers
	//goji.Get("/result", http.RedirectHandler("/", 301))
	time.AfterFunc(500*time.Millisecond, func(){
		open.Start("http://localhost:8000")
	})
	goji.Serve()
}

func main() {
	site()
}
