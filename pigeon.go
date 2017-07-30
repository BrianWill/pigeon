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
	"sync"
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
		defer r.Body.Close()
		pigeonFiles := []string{}
		files, err := ioutil.ReadDir(".")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, err.Error())
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
			fmt.Fprintf(w, err.Error())
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
		defer r.Body.Close()
		sourceFile := r.URL.Path[6:]
		bytes, err := ioutil.ReadFile(sourceFile)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, err.Error())
			return
		}
		bytes, err = dynamicPigeon.Highlight(bytes)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, err.Error())
			return
		}
		code := template.HTML(bytes)
		executablePath, validBreakpoints, err = dynamicPigeon.Compile(sourceFile)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, err.Error())
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
		defer r.Body.Close()
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
		defer r.Body.Close()
		lineStr := strings.TrimPrefix(r.URL.Path, "/clearBreakpoint/")
		line, err := strconv.Atoi(lineStr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "400 - Invalid line number.")
			return
		}
		delete(breakpoints, lineStr)
		fmt.Fprintf(w, "Line number: %d", line)
	})

	http.HandleFunc("/getBreakpoints", func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		jsonBytes, err := json.Marshal(breakpoints)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonBytes)
	})

	http.HandleFunc("/getValidBreakpoints", func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		jsonBytes, err := json.Marshal(validBreakpoints)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonBytes)
	})

	continueSignal := false

	http.HandleFunc("/checkContinue", func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		fmt.Fprintf(w, "%v", continueSignal)
		continueSignal = false
	})

	http.HandleFunc("/continue", func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		continueSignal = true
		fmt.Fprintf(w, "%v", continueSignal)
	})

	outputBuffer := []string{}
	var outputMux sync.Mutex
	http.HandleFunc("/writeOutput", func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		outputMux.Lock()
		defer outputMux.Unlock()
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, err.Error())
			return
		}
		outputBuffer = append(outputBuffer, string(body))
		fmt.Println("writeOutput: ", outputBuffer)
	})

	http.HandleFunc("/readOutput", func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		outputMux.Lock()
		defer outputMux.Unlock()
		fmt.Println("readOutput: ", outputBuffer)
		jsonBytes, err := json.Marshal(outputBuffer)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, err.Error())
			return
		}
		outputBuffer = []string{}
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonBytes)
	})

	acceptInput := false
	inputBuffer := []string{}
	var inputMux sync.Mutex
	http.HandleFunc("/writeInput", func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		if !acceptInput {
			fmt.Fprintf(w, "not ready")
			return
		}
		inputMux.Lock()
		defer inputMux.Unlock()
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, err.Error())
			return
		}
		var line string
		err = json.Unmarshal(body, &line)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, err.Error())
			return
		}
		inputBuffer = append(inputBuffer, line)
		acceptInput = false
	})

	http.HandleFunc("/readInput", func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		inputMux.Lock()
		defer inputMux.Unlock()
		jsonBytes, err := json.Marshal(inputBuffer)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, err.Error())
			return
		}
		inputBuffer = []string{}
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonBytes)
	})

	http.HandleFunc("/acceptInput", func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		acceptInput = true
		fmt.Fprintf(w, "%v", acceptInput)
		fmt.Println("accept input: ", acceptInput)
	})

	http.HandleFunc("/rejectInput", func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		acceptInput = false
		fmt.Fprintf(w, "%v", acceptInput)
	})

	http.HandleFunc("/run", func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
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
