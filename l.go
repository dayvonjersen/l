package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"

	"github.com/generaltso/linguist"
	"github.com/lintianzhi/ignore"
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

//temporary
var ignore_mimetype []string = []string{
	"application/octet-stream",
}
var ignore_mimetype_start []string = []string{
	"image",
	"audio",
	"video",
}

func getLang(filename string) string {
	res1 := linguist.DetectFromFilename(filename)
	if res1 != "" {

		return res1
	}
	// if we can't guess type by extension
	// before jumping into lexing and parsing things like image files or cat videos
	// or other binary formats which will give erroneous results
	// and unnecessarily waste CPU time reading large files into memory
	mimetype, shouldIgnore, err := linguist.DetectMimeFromFilename(filename)
	checkErr(err)
	if shouldIgnore {
		return mimetype
	}
	contents, err := ioutil.ReadFile(filename)
	checkErr(err)

	res2 := linguist.DetectFromContents(contents)
	if res2 != "" {
		return res2
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
			//			fmt.Fprintf(stderr, "\r\033[H\033[2J")
			fmt.Fprintf(stderr, "%s...\r", abs)
			stderr.Flush()
		}
		if file.Size() == 0 {
			continue
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
	//	fmt.Fprintf(stderr, "\r\033[H\033[2J")

	fmtstr := fmt.Sprintf("%% %ds", max_len)
	fmt.Printf(fmtstr, "Language")
	fmt.Println()
	fmtstr += ": %07.4f%%\n"
	results := []float64{}
	qqq := map[float64]string{}
	for lang, num := range langs {
		res := (float64(num) / float64(total_size)) * 100.0
		results = append(results, res)
		qqq[res] = lang
	}
	sort.Sort(sort.Reverse(sort.Float64Slice(results)))
	for _, percent := range results {
		fmt.Printf(fmtstr, qqq[percent], percent)
	}
	fmt.Printf("---\n%d languages detected in %d bytes of %d files\n", len(langs), total_size, num_files)
}
