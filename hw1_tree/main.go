package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	path2 "path"
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

func dirTree(out io.Writer, path string, printFiles bool) error {
	if err := printDirTree(out, "", path, printFiles); err != nil {
		return err
	}

	return nil
}

func printDirTree(out io.Writer, prefix string, path string, printFiles bool) error {
	dirContent, err := ioutil.ReadDir(path)
	if err != nil {
		return err
	}

	if !printFiles {
		dirContent = removeFiles(dirContent)
	}

	dirContentLen := len(dirContent)
	if dirContentLen > 0 {
		for i, entry := range dirContent {
			entryName := entry.Name()
			isLast := i == len(dirContent)-1
			tmpForPrefix := ""
			if isLast {
				tmpForPrefix = "└───"
			} else {
				tmpForPrefix = "├───"
			}
			if entry.IsDir() {
				fmt.Fprintf(out, "%v%v\n", prefix+tmpForPrefix, entryName)

				nextPath := path2.Join(path, entryName)
				if isLast {
					tmpForPrefix = "\t"
				} else {
					tmpForPrefix = "│\t"
				}
				if err := printDirTree(out, prefix+tmpForPrefix, nextPath, printFiles); err != nil {
					return err
				}
			} else {
				fmt.Fprintf(out, "%v%v (%v)\n", prefix+tmpForPrefix, entryName, formatSize(entry.Size()))
			}
		}
	}
	return nil
}

func formatSize(size int64) string {
	if size == 0 {
		return "empty"
	}
	return fmt.Sprintf("%vb", size)
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
