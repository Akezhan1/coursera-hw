package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

func main() {
	out := os.Stdout
	if !(len(os.Args) == 2 || len(os.Args) == 3) {
		panic("usage go run main.go . [-f]")
	}
	path := os.Args[1]
	printFiles := len(os.Args) == 3 && os.Args[2] == "-f"
	err := dirTree(out, path, printFiles)
	if err != nil {
		panic(err.Error())
	}
}

func dirTree(out io.Writer, path string, flag bool) error {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return err
	}

	if !flag {
		files = removeFiles(files)
	}

	for i, file := range files {

		fmt.Printf("%s%s\n", getIndent(path, file.Name(), i == len(files)-1), file.Name())

		if file.IsDir() {
			if err = dirTree(out, path+"/"+file.Name(), flag); err != nil {
				return err
			}

		}
	}
	return nil
}

func getIndent(path string, curFile string, isLast bool) string {
	deep := strings.Count(path, "/")
	str := ""
	str += strings.Repeat("\t", deep)
	//fmt.Println(path, removeLastDirInPath(path), curFile, checkParentToLastDir(path, curFile))
	if isLast {
		str += "└───"
	} else {
		str += "├───"
	}
	return str
}

func checkParentToLastDir(path string, curFile string) bool {
	p := removeLastDirInPath(path)
	if p == "" {
		return true
	}

	files, _ := ioutil.ReadDir(path)
	if files[len(files)-1].Name() == curFile {
		return true
	}

	return false
}

func getSizeInString(b int64) string {
	if b == 0 {
		return "empty"
	}
	return fmt.Sprintf("%db", b)
}

func removeFiles(files []os.FileInfo) []os.FileInfo {
	var tmp []os.FileInfo
	for _, f := range files {
		if f.IsDir() {
			tmp = append(tmp, f)
		}
	}
	return tmp
}

func removeLastDirInPath(path string) string {
	newPath := ""
	for i := len(path) - 1; i > 0; i-- {
		if path[i] == '/' {
			newPath = path[:i]
			break
		}
	}
	return newPath
}
