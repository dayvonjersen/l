package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"mime"
	"os"
	"path/filepath"
	"strings"

	"camlistore.org/pkg/magic"
	"github.com/lintianzhi/ignore"
	"github.com/petermattis/linguist"
)

func checkErr(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

var langs map[string]int64 = make(map[string]int64)
var total_size int64 = 0
var res map[string]int = make(map[string]int)
var num_files int = 0
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
	if mimetype != "" {
		return mimetype
	}

	mimetyperedux := magic.MIMEType(contents)
	if mimetyperedux != "" {
		return mimetyperedux
	}

	//fmt.Fprintf(os.Stderr, "unknown ext: %s\nfilename: %s\n\n", ext, filename)
	return "(unknown)"
}

var stderr *bufio.Writer = bufio.NewWriter(os.Stderr)

func processDir(dirname string) {
	cwd, err := os.Open(dirname)
	checkErr(err)
	files, err := cwd.Readdir(0)
	checkErr(err)
	checkErr(os.Chdir(dirname))
	for _, file := range files {
		if abs, err := filepath.Abs(file.Name()); err == nil {
			fmt.Fprintf(stderr, "% 80s\r", " ")
			fmt.Fprintf(stderr, "%s...\r", abs)
			stderr.Flush()
		}
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
			res[res1]++
			langs[res1] += file.Size()
			total_size += file.Size()
			num_files++
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
	fmt.Fprintf(stderr, "% 80s\r", " ")
	fmt.Println()
	fmtstr := fmt.Sprintf("%% %ds", max_len)
	fmt.Printf(fmtstr, "Language")
	fmt.Println(" (Size)  (Frequency)\n---")
	fmtstr += ": %07.4f%% %07.4f%%\n"
	for lang, num := range langs {
		fmt.Printf(fmtstr, lang, (float64(num)/float64(total_size))*100.0, (float64(res[lang])/float64(num_files))*100.0)
	}
	fmt.Printf("---\n%d languages detected in %d bytes of %d files\n", len(langs), total_size, num_files)
}
