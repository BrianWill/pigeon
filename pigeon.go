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
	"github.com/BrianWill/pigeon/staticPigeon"
)

func main() {
	if len(os.Args) >= 2 {
		subcommand := os.Args[1]
		if subcommand == "run" {
			if len(os.Args) < 3 {
				fmt.Println("Must specify a file to run.")
				return
			}
			if strings.HasSuffix(os.Args[2], ".spigeon") {
				_, _, err := staticPigeon.Compile(os.Args[2])
				if err != nil {
					fmt.Println(err)
					return
				}
			} else {
				_, err := dynamicPigeon.CompileAndRun(os.Args[2])
				if err != nil {
					fmt.Println(err)
					return
				}
			}

		}
	} else {
		server()
	}
}

type runStateEnum string

const (
	stopped runStateEnum = "stopped"
	running runStateEnum = "running"
	paused  runStateEnum = "paused"
)

type RunState struct {
	state     runStateEnum
	lastBreak int
}

func server() {

	runState := RunState{stopped, 0}
	// TODO use session state
	executablePath := ""
	breakpoints := map[string]bool{}
	validBreakpoints := map[string]bool{}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
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
		runState = RunState{stopped, 0}
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
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "400 - Invalid line number.")
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
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonBytes)
	})

	http.HandleFunc("/getValidBreakpoints", func(w http.ResponseWriter, r *http.Request) {
		jsonBytes, err := json.Marshal(validBreakpoints)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonBytes)
	})

	continueFlag := false

	http.HandleFunc("/checkContinue/", func(w http.ResponseWriter, r *http.Request) {
		// the breakpoint line on which the code is currently paused
		lineStr := strings.TrimPrefix(r.URL.Path, "/checkContinue/")
		line, err := strconv.Atoi(lineStr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "400 - Invalid line number.")
			return
		}
		if continueFlag {
			runState = RunState{running, line}
		} else {
			runState = RunState{paused, line}
		}
		fmt.Fprintf(w, "%v", continueFlag)
		continueFlag = false
	})

	http.HandleFunc("/continue", func(w http.ResponseWriter, r *http.Request) {
		continueFlag = true
		fmt.Fprintf(w, "%v", continueFlag)
	})

	outputBuffer := []string{}
	var outputMux sync.Mutex
	http.HandleFunc("/writeOutput", func(w http.ResponseWriter, r *http.Request) {
		outputMux.Lock()
		defer outputMux.Unlock()
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, err.Error())
			return
		}
		outputBuffer = append(outputBuffer, string(body))
	})

	http.HandleFunc("/readOutput", func(w http.ResponseWriter, r *http.Request) {
		outputMux.Lock()
		defer outputMux.Unlock()
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
		acceptInput = true
		fmt.Fprintf(w, "%v", acceptInput)
	})

	http.HandleFunc("/rejectInput", func(w http.ResponseWriter, r *http.Request) {
		acceptInput = false
		fmt.Fprintf(w, "%v", acceptInput)
	})

	http.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "%v %v", runState.state, runState.lastBreak)
	})

	var runningProgram *exec.Cmd

	http.HandleFunc("/run", func(w http.ResponseWriter, r *http.Request) {
		var err error
		runningProgram, err = dynamicPigeon.Run(executablePath)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Error compiling code: "+err.Error())
			return
		}
		runState = RunState{running, 0}
		fmt.Fprintf(w, "running")
	})

	http.HandleFunc("/stop", func(w http.ResponseWriter, r *http.Request) {
		if runningProgram != nil {
			runningProgram.Process.Kill()
			runState.state = stopped
			fmt.Fprintf(w, "stopped")
		} else {
			fmt.Fprintf(w, "not running")
		}
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
