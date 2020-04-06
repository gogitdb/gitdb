package main

import (
	"bytes"
	"encoding/base64"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

var (
	packageRoot = "../../"

	embedCommand = flag.NewFlagSet("embed", flag.ExitOnError)
	output       = embedCommand.String("o", "./ui_static.go", "output file name; default ./ui_static.go")

	// dbpath      = flag.String("p", "", "path do gitdb")
)

func main() {

	command := os.Args[1]
	switch command {
	case "embed-ui":
		embedCommand.Parse(os.Args[2:])
		err := embedUI()
		if err != nil {
			fmt.Println(err.Error())
		}
	default:
		fmt.Println("invalid command; try gitdb embed-ui")
		//future commands
		//clean-db i.e git gc
		//repair
		//dataset
		//dataset <name> blocks
		//dataset <name> records
		return
	}
}

type staticFile struct {
	Name    string
	Content string
}

func embedUI() error {
	_, filename, _, ok := runtime.Caller(0)
	if ok {
		packageRoot = path.Dir(path.Dir(path.Dir(filename))) + "/"
	}

	var files []staticFile
	if err := readAllStaticFiles(filepath.Join(packageRoot, "static"), &files); err != nil {
		return err
	}

	_, err := os.Stat(path.Dir(*output))
	if err != nil {
		return err
	}

	w, err := os.Create(*output)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return packageTmpl.Execute(w, struct {
		Files []staticFile
		Date  string
	}{
		Files: files,
		Date:  time.Now().Format(time.RFC1123),
	})
}

func readAllStaticFiles(path string, files *[]staticFile) error {

	dirs, err := ioutil.ReadDir(path)
	if err != nil {
		return err
	}

	for _, dir := range dirs {
		fileName := filepath.Join(path, dir.Name())
		fmt.Println(fileName)
		if !dir.IsDir() {
			b, err := ioutil.ReadFile(fileName)
			if err != nil {
				return err
			}

			b = bytes.Replace(b, []byte("  "), []byte(""), -1)
			b = bytes.Replace(b, []byte("\n"), []byte(""), -1)

			content := base64.StdEncoding.EncodeToString(b)

			fileKey := strings.Replace(fileName, packageRoot, "", 1)
			*files = append(*files, staticFile{fileKey, content})
		}

		if !strings.HasPrefix(dir.Name(), ".") && dir.IsDir() {
			readAllStaticFiles(fileName, files)
		}
	}

	return nil
}
