package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"strings"

	"github.com/BrianWill/pigeon/dynamicPigeon"
)

func main() {
	if len(os.Args) >= 2 {
		subcommand := os.Args[1]
		if subcommand == "run" {
			err := dynamicPigeon.Run(os.Args)
			if err != nil {
				fmt.Println(err)
				return
			}
		}
	} else {
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

			pigeonFiles := []string{}
			files, err := ioutil.ReadDir(".")
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "500 - cannot read directory")
				return
			}
			for _, file := range files {
				name := file.Name()
				if strings.HasSuffix(name, ".pigeon") {
					pigeonFiles = append(pigeonFiles, name)
				}
			}
			sort.Strings(pigeonFiles)
			t, err := template.ParseFiles("templates/main.html")
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "500 - error loading template")
				return
			}
			code := template.HTML("")
			if r.URL.Path != "/" {
				// read code file from path
				filePath := r.URL.Path[1:]
				bytes, err := ioutil.ReadFile(filePath)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					fmt.Fprintf(w, "500 - error reading code file")
					return
				}
				bytes, err = dynamicPigeon.Highlight(bytes)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					fmt.Fprintf(w, "500 - error highlighting code file")
					return
				}
				code = template.HTML(bytes)
			}
			type vals struct {
				Files []string
				Code  template.HTML
			}
			err = t.Execute(w, vals{Files: pigeonFiles, Code: code})
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, err.Error())
				return
			}

		})
		log.Fatal(http.ListenAndServe(":8080", nil))
	}

}

// open opens the specified URL in the default browser of the user.
func open(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}
	args = append(args, url)
	return exec.Command(cmd, args...).Start()
}
