package main

import (
	"archive/zip"
	"bytes"
	"errors"
	//_ "expvar"
	"flag"
	"fmt"
	_ "net/http/pprof"
	_ "github.com/rakyll/gometry/http"

	log "github.com/Sirupsen/logrus"
	//"github.com/kr/pretty"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/mholt/binding"
	"github.com/pmezard/go-difflib/difflib"
	"github.com/skratchdot/open-golang/open"
	"github.com/zenazn/goji"
	"github.com/zenazn/goji/web"
	"gopkg.in/unrolled/render.v1"
)

var wd string
var err error
var diffs OutputData

const (
	// Version number
	VERSION = "0.05"
	// Scan at up to this size in file for '\0' in test for binary file
	BINARY_CHECK_SIZE = 65536
	// default number of context lines to display
	CONTEXT_LINES = 3
)

type Config struct {
	FileA     *multipart.FileHeader
	FileB     *multipart.FileHeader
	FilePathA string
	FilePathB string
	Brief     bool
	Skip      string
}

type DiffResult struct {
	Title    string
	Diff     string
	IsZip    bool
	IsBinary bool
}

type OutputData struct{
	FileA string
	FileB string
	Diffs []DiffResult
}

type ZipSrc interface {
	io.Reader
	io.ReaderAt
}

func DisplayAsText(w io.Writer, dfs []DiffResult) {
	for _, dr := range dfs {
		fmt.Fprintln(w, dr.Title)
		if dr.Diff != "" {
			if dr.IsZip {
				fmt.Fprintln(w, "#####################################################################")
				fmt.Fprint(w, dr.Diff)
				fmt.Fprintln(w, "#####################################################################")
				fmt.Fprintln(w)
			} else {
				fmt.Fprint(w, dr.Diff)
				fmt.Fprintln(w)
			}
		} else if dr.IsBinary {
			fmt.Fprintln(w, "<<binary>>")
			fmt.Fprintln(w)
		}
	}
}

// Then provide a field mapping (pointer receiver is vital)
func (conf *Config) FieldMap() binding.FieldMap {
	return binding.FieldMap{
		&conf.FileA:     "file-a",
		&conf.FileB:     "file-b",
		&conf.FilePathA: "filepath-a",
		&conf.FilePathB: "filepath-b",
		&conf.Brief:     "brief",
		&conf.Skip:      "skip",
	}
}

func init() {
	wd, err = os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
}

func isSkiping(name string, skipstr string) bool {
	var prefix []string
	skip := strings.TrimSpace(skipstr)
	if skip != "" {
		skips := strings.Split(skip, "\n")
		for _, s := range skips {
			s = strings.TrimSpace(s)
			if s != "" {
				prefix = append(prefix, s)
			}
		}
	}

	for _, p := range prefix {
		if strings.Contains(name, p) {
			return true
		}
	}
	return false
}

func DirDiffContent(conf *Config) []DiffResult {
	log.Warn("未実装")
	return make([]DiffResult, 0, 10)
}

func ProcessDiff(fname string, data1, data2 []byte, brief bool, skips string) (DiffResult, error) {
	diffResult := DiffResult{}
	diffResult.Title = fmt.Sprintf("ファイル %s は異なります", fname)
	if brief {
		return diffResult, nil
	}
	if isSkiping(fname, skips) {
		return diffResult, nil
	}
	if strings.HasSuffix(fname, "jar") || strings.HasSuffix(fname, "war") || strings.HasSuffix(fname, "zip") {
		diffResult.IsZip = true
		zipa, err := zip.NewReader(bytes.NewReader(data1), int64(len(data1)))
		if err != nil {
			return diffResult, err
		}
		zipb, err := zip.NewReader(bytes.NewReader(data2), int64(len(data2)))
		if err != nil {
			return diffResult, err
		}
		var buf bytes.Buffer
		DisplayAsText(&buf, ZipDiffContent(zipa, zipb, brief, skips))
		diffResult.Diff = buf.String()
	} else if strings.HasSuffix(fname, "class") {
		diffResult.Diff = godiff(string(jadfile(data1)), string(jadfile(data2)))
	} else if checkBinary(data1) {
		diffResult.IsBinary = true
	} else {
		diffResult.Diff = godiff(string(data1), string(data2))
	}
	return diffResult, nil
}

func ZipDiffContent(zipa, zipb *zip.Reader, brief bool, skips string) []DiffResult {
	var result []DiffResult
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
					diffResult, err := ProcessDiff(a.Name, data1, data2, brief, skips)
					if err != nil {
						log.Error(err)
					}
					result = append(result, diffResult)
				}
				break // break zipb.File loop
			}
		}
		if !hitFlag {
			result = append(result, DiffResult{Title: fmt.Sprintf("Aだけに発見: %s\n", a.Name)})
		}
	}
	if len(rightFileSet) != 0 {
		for f := range rightFileSet {
			result = append(result, DiffResult{Title: fmt.Sprintf("Bだけに発見: %s\n", f)})
		}
	}
	return result
}

func godiff(a, b string) string {
	diff := difflib.UnifiedDiff{
		A:        difflib.SplitLines(a),
		B:        difflib.SplitLines(b),
		FromFile: "File A",
		ToFile:   "File B",
		Context:  3,
	}
	text, err := difflib.GetUnifiedDiffString(diff)
	if err != nil {
		log.Fatal(err)
	}
	return text
}

func jadfile(data []byte) []byte {
	tmpClass, err := ioutil.TempFile("", "")
	defer os.Remove(tmpClass.Name())
	defer tmpClass.Close()

	if err != nil {
		log.Fatal(err)
	}
	tmpClass.Write(data)
	tmpClass.Close()
	cmdjad := filepath.Join(wd, "tools", "jad.exe")
	cmd := exec.Command(cmdjad, "-p", tmpClass.Name())
	var bufOut bytes.Buffer
	cmd.Stdout = &bufOut
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
	return bufOut.Bytes()
}

// readerのレングス取得
func getReaderLength(r io.Reader) (length int64){

	bufsize := 32 << 10 // 1M

	var buf = make([]byte, bufsize)

	defer func(){
		log.Info(length)
	}()

	for {
		n, err := r.Read(buf)
		length = length + int64(n)
		if err == io.EOF{
			return
		} else if err != nil{
			log.Fatal(err)
		}
	}

}

func ReadZip(data ZipSrc) (*zip.Reader, error) {
	//content, err := ioutil.ReadAll(data)
	length := getReaderLength(data)
	if err != nil {
		return nil, err
	}

	zr, err := zip.NewReader(data, length)
	if err != nil {
		return nil, err
	}
	return zr, nil
}

func ErrorPage(err error, w http.ResponseWriter) {
	log.Error(err)
	w.WriteHeader(http.StatusInternalServerError)
	io.WriteString(w, "Something go wrong\n\n")
	io.WriteString(w, err.Error())
	return
}

func IsDir(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return info.IsDir(), nil
}

func index(c web.C, w http.ResponseWriter, r *http.Request) {
	ren := render.New(render.Options{})
	ren.HTML(w, http.StatusOK, "index", nil)
}
func PostDiff(w http.ResponseWriter, r *http.Request) {
	diffs = OutputData{}
	conf := new(Config)
	errs := binding.Bind(r, conf)
	if errs.Handle(w) {
		return
	}

	if conf.FileA != nil && conf.FileB != nil {
		fa, err := conf.FileA.Open()
		if err != nil {
			ErrorPage(err, w)
			return
		}
		defer fa.Close()

		fb, err := conf.FileB.Open()
		if err != nil {
			ErrorPage(err, w)
			return
		}
		defer fb.Close()

		zra, err := ReadZip(fa)
		if err != nil {
			log.Fatal(err)
		}
		zrb, err := ReadZip(fb)
		if err != nil {
			log.Fatal(err)
		}
		diffs.FileA = conf.FileA.Filename
		diffs.FileB = conf.FileB.Filename
		diffs.Diffs = ZipDiffContent(zra, zrb, conf.Brief, conf.Skip)
	} else {
		if conf.FilePathA == "" || conf.FilePathB == "" {
			err = errors.New("No input data")
			ErrorPage(err, w)
			return
		}
		isDirA, err := IsDir(conf.FilePathA)
		if err != nil {
			ErrorPage(err, w)
			return
		}
		isDirB, err := IsDir(conf.FilePathB)
		if err != nil {
			ErrorPage(err, w)
			return
		}

		if isDirA && isDirB {
			// Got Two Directory
			diffs.Diffs = DirDiffContent(conf)
		} else if !isDirB && !isDirB {
			// Got Two File
			fa, err := os.Open(conf.FilePathA)
			if err != nil {
				ErrorPage(err, w)
				return
			}
			defer fa.Close()

			fb, err := os.Open(conf.FilePathB)
			if err != nil {
				ErrorPage(err, w)
				return
			}
			defer fb.Close()

			zra, err := ReadZip(fa)
			if err != nil {
				ErrorPage(err, w)
				return
			}
			zrb, err := ReadZip(fb)
			if err != nil {
				ErrorPage(err, w)
				return
			}
			diffs.FileA = conf.FilePathA
			diffs.FileB = conf.FilePathB
			diffs.Diffs = ZipDiffContent(zra, zrb, conf.Brief, conf.Skip)
		} else {
			err = errors.New("I dont know how to compare different type.")
			ErrorPage(err, w)
			return
		}
	}

	ren := render.New(render.Options{})
	ren.HTML(w, http.StatusOK, "richdiff", diffs)
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// check if file is binary
func checkBinary(data []byte) bool {
	if data == nil {
		return false
	}
	if len(data) == 0 {
		return false
	}
	if bytes.IndexByte(data[0:minInt(len(data), BINARY_CHECK_SIZE)], 0) >= 0 {
		return true
	}
	return false
}
func generateResult(w http.ResponseWriter, r *http.Request){
	w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
	io.WriteString(w, fmt.Sprintf("File A: %s\n", diffs.FileA))
	io.WriteString(w, fmt.Sprintf("File B: %s\n", diffs.FileB))
	io.WriteString(w, "\n")
	DisplayAsText(w, diffs.Diffs)
}

func webui() {
	goji.Get("/", index)
	goji.Get("/result/diff.txt", generateResult)
	goji.Post("/diff", PostDiff)
	//Fully backwards compatible with net/http's Handlers
	//goji.Get("/result", http.RedirectHandler("/", 301))
	if os.Getenv("DEBUG") == "" {
		time.AfterFunc(500*time.Millisecond, func() {
			open.Start("http://localhost:8000")
		})
	}
	goji.Serve()
}

func main() {
	var output string
	flag.StringVar(&output, "file", "", "Output result to this file")
	flag.Parse()

	if flag.NArg() < 2 {
		webui()
		os.Exit(1)
	} else {
		zip1, err := zip.OpenReader(flag.Arg(0))
		if err != nil {
			log.Fatal(err)
		}
		zip2, err := zip.OpenReader(flag.Arg(1))
		if err != nil {
			log.Fatal(err)
		}
		defer zip1.Close()
		defer zip2.Close()
		result := ZipDiffContent(&zip1.Reader, &zip2.Reader, false, "")
		if output != "" {
			f, err := os.Create(output)
			if err != nil {
				log.Fatal(err)
			}
			defer f.Close()
			DisplayAsText(f, result)
		} else {
			DisplayAsText(os.Stdout, result)
		}
	}
}
