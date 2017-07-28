package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/BrianWill/pigeon/dynamicPigeon"
)

func main() {
	if len(os.Args) >= 2 {
		subcommand := os.Args[1]
		if subcommand == "run" {
			if len(os.Args) < 3 {
				fmt.Println("Must specify a file to run.")
				return
			}
			err := dynamicPigeon.CompileAndRun(os.Args[2])
			if err != nil {
				fmt.Println(err)
				return
			}
		}
	} else {
		server()
	}

}

func server() {
	// TODO use session state
	executablePath := ""
	breakpoints := map[string]bool{}
	validBreakpoints := map[string]bool{}

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
		err = t.Execute(w, pigeonFiles)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, err.Error())
			return
		}

	})

	http.HandleFunc("/code/", func(w http.ResponseWriter, r *http.Request) {
		sourceFile := r.URL.Path[6:]
		bytes, err := ioutil.ReadFile(sourceFile)
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
		code := template.HTML(bytes)
		executablePath, validBreakpoints, err = dynamicPigeon.Compile(sourceFile)
		fmt.Println("valid breakpoints", validBreakpoints)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Error compiling code: "+err.Error())
			return
		}
		t, err := template.ParseFiles("templates/code.html")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "500 - error loading template")
			return
		}
		type vals struct {
			FileName string
			Code     template.HTML
		}
		err = t.Execute(w, vals{FileName: sourceFile, Code: code})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, err.Error())
			return
		}
	})

	http.HandleFunc("/setBreakpoint/", func(w http.ResponseWriter, r *http.Request) {
		lineStr := strings.TrimPrefix(r.URL.Path, "/setBreakpoint/")
		line, err := strconv.Atoi(lineStr)
		if err != nil {
			fmt.Fprintf(w, "Invalid line number.")
			return
		}
		if !validBreakpoints[lineStr] {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "400 - invalid line for a breakpoint")
			return
		}
		breakpoints[lineStr] = true
		fmt.Fprintf(w, "Line number: %d", line)
	})

	http.HandleFunc("/clearBreakpoint/", func(w http.ResponseWriter, r *http.Request) {
		lineStr := strings.TrimPrefix(r.URL.Path, "/clearBreakpoint/")
		line, err := strconv.Atoi(lineStr)
		if err != nil {
			fmt.Fprintf(w, "Invalid line number.")
			return
		}
		delete(breakpoints, lineStr)
		fmt.Fprintf(w, "Line number: %d", line)
	})

	http.HandleFunc("/getBreakpoints", func(w http.ResponseWriter, r *http.Request) {
		jsonBytes, err := json.Marshal(breakpoints)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, err.Error())
			return
		}
		w.Write(jsonBytes)
	})

	http.HandleFunc("/getValidBreakpoints", func(w http.ResponseWriter, r *http.Request) {
		jsonBytes, err := json.Marshal(validBreakpoints)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, err.Error())
			return
		}
		w.Write(jsonBytes)
	})

	continueSignal := false

	http.HandleFunc("/checkContinue", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "%v", continueSignal)
		continueSignal = false
	})

	http.HandleFunc("/continue", func(w http.ResponseWriter, r *http.Request) {
		continueSignal = true
		fmt.Fprintf(w, "%v", continueSignal)
	})

	http.HandleFunc("/run", func(w http.ResponseWriter, r *http.Request) {
		err := dynamicPigeon.Run(executablePath)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Error compiling code: "+err.Error())
			return
		}
		fmt.Fprintf(w, "running")
	})

	go func() {
		time.Sleep(300 * time.Millisecond)
		open("http://localhost:7070/")
	}()
	log.Fatal(http.ListenAndServe(":7070", nil))
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
