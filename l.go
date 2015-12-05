package main

import "github.com/petermattis/linguist"
import "os"
import "io/ioutil"
import "fmt"

func checkErr(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

var res map[string]string = make(map[string]string)
var langs map[string]int = make(map[string]int)
var num_files int = 0
var max_len int = 0

func getLang(filename string) string {
	res1 := linguist.DetectFromFilename(filename)
	if res1 == "" {
		contents, err := ioutil.ReadFile(filename)
		checkErr(err)
		res2 := linguist.DetectFromContents(contents)
		if res2 == "" {
			res2 = "(unknown)"
		}
		res1 = res2
	}
	return res1
}

func processDir(dirname string) {
	cwd, err := os.Open(dirname)
	checkErr(err)
	files, err := cwd.Readdir(0)
	checkErr(err)
	checkErr(os.Chdir(dirname))
	for _, file := range files {
		//println(file.Name())
		if file.IsDir() {
			if file.Name() == ".git" {
				continue
			}
			processDir(file.Name())
		} else if !linguist.IsVendored(file.Name()) {
			res1 := getLang(file.Name())
			res[file.Name()] = res1
			langs[res1]++
			num_files++
			if len(res1) > max_len {
				max_len = len(res1)
			}
		}
	}
	checkErr(os.Chdir(".."))
}

func main() {
	processDir(".")
	fmtstr := fmt.Sprintf("%% %ds: %%04.4f%%%%\n", max_len)
	for lang, num := range langs {
		fmt.Printf(fmtstr, lang, (float64(num)/float64(num_files))*100.0)
	}
}
