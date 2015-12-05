package main

import (
	"fmt"
	"github.com/lintianzhi/ignore"
	"github.com/petermattis/linguist"
	"io/ioutil"
	"mime"
	"os"
	"strings"
)

func checkErr(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

var res map[string]string = make(map[string]string)
var langs map[string]int64 = make(map[string]int64)
var num_files int64 = 0
var max_len int = 0

func getLang(filename string) string {
	res1 := linguist.DetectFromFilename(filename)
	if res1 != "" {
		return res1
	}

	contents, err := ioutil.ReadFile(filename)
	checkErr(err)
	res2 := linguist.DetectFromContents(contents)
	if res2 != "" {
		return res2
	}

	parts := strings.Split(filename, ".")
	ext := parts[len(parts)-1]
	mimetype := mime.TypeByExtension("." + ext)
	if mimetype == "" {
		fmt.Printf("unknown ext: %s\nfilename: %s\n\n", ext, filename)
	}
	return mimetype
}

func processDir(dirname string) {
	cwd, err := os.Open(dirname)
	checkErr(err)
	files, err := cwd.Readdir(0)
	checkErr(err)
	checkErr(os.Chdir(dirname))
	for _, file := range files {
		if isIgnored(dirname + string(os.PathSeparator) + file.Name()) {
			continue
		}
		if file.IsDir() {
			if file.Name() == ".git" {
				continue
			}
			processDir(file.Name())
		} else if !linguist.IsVendored(file.Name()) {
			//println(file.Name())
			res1 := getLang(file.Name())
			res[file.Name()] = res1
			langs[res1] += file.Size()
			num_files += file.Size()
			if len(res1) > max_len {
				max_len = len(res1)
			}
		}
	}
	checkErr(os.Chdir(".."))
}

type ignorer func(string) bool

var isIgnored ignorer

func main() {
	g, err := os.Open(".gitignore")
	g.Close()
	if os.IsNotExist(err) {
		isIgnored = func(filename string) bool {
			return false
		}
	} else {
		gitIgn, err := ignore.NewGitIgn(".gitignore")
		checkErr(err)
		gitIgn.Start(".")
		isIgnored = func(filename string) bool {
			return gitIgn.TestIgnore(filename)
		}
	}
	processDir(".")
	fmtstr := fmt.Sprintf("%% %ds: %%04.4f%%%%\n", max_len)
	for lang, num := range langs {
		fmt.Printf(fmtstr, lang, (float64(num)/float64(num_files))*100.0)
	}
}
