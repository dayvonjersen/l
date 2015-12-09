package main

import (
	"flag"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/generaltso/linguist"
)

func getFiles(tree_id string) {
	git := exec.Command("sh", "-c", "git ls-tree "+tree_id)
	out, _ := git.CombinedOutput()

	//	results := map[string]int{}
	//	total_size := 0

	for _, ln := range strings.Split(string(out), "\n") {
		fields := strings.Split(ln, " ")
		if len(fields) != 3 {
			continue
		}
		fmode := fields[0]
		ftype := fields[1]
		fields = strings.Split(fields[2], "\t")
		if len(fields) != 2 {
			continue
		}
		fhash := fields[0]
		fname := fields[1]

		fmt.Println(fmode, ftype, fhash, fname)

		switch ftype {
		case "tree":
			getFiles(fhash)
		case "blob":
			cats := exec.Command("sh", "-c", "git cat-file -s "+fhash)
			dats, _ := cats.CombinedOutput()
			size, _ := strconv.Atoi(strings.TrimSpace(string(dats)))
			fmt.Printf("filesize: %d bytes\n", size)
			if linguist.IsVendored(fname) {
				fmt.Println("(vendored)")
				continue
			}
			by_ext := linguist.DetectFromFilename(fname)
			if by_ext != "" {
				fmt.Println(by_ext)
				continue
			}
			cat := exec.Command("sh", "-c", "git cat-file blob "+fhash)
			dat, _ := cat.CombinedOutput()

			fmt.Println(linguist.DetectFromContents(dat))
		default:
			println("unsupported ftype:" + ftype)
		}
	}
}

func main() {
	flag.Parse()
	getFiles(flag.Args()[0])
}
