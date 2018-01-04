package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/BrianWill/pigeon/goPigeon"
)

func Run(filename string) (*exec.Cmd, error) {
	cmd := exec.Command("go", "run", filename)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return nil, err
	}
	return cmd, nil
}

func main() {

	if len(os.Args) < 2 {
		fmt.Println("Must specify a file to run.")
		return
	}
	basedir := os.Getenv("GOPATH") + "/src/pigeon_output/"
	outputFile := basedir + "output.go"
	if _, err := os.Stat(basedir); os.IsNotExist(err) {
		os.Mkdir(basedir, os.ModePerm)
	}
	var code []byte
	if strings.HasSuffix(os.Args[1], ".gopigeon") {
		pkg, err := goPigeon.Compile(os.Args[1], "pigeon_output/")
		if err != nil {
			fmt.Println(err)
			return
		}
		code = []byte(pkg.Code)
	} else if strings.HasSuffix(os.Args[1], ".pigeon") {
		pkg, err := goPigeon.Compile(os.Args[1], "pigeon_output/")
		if err != nil {
			fmt.Println(err)
			return
		}
		code = []byte(pkg.Code)
	} else {
		log.Fatal("File has improper extension.")
	}
	err := ioutil.WriteFile(outputFile, code, os.ModePerm)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = exec.Command("go", "fmt", outputFile).Run()
	if err != nil {
		fmt.Println(err)
		return
	}
	_, err = Run(outputFile)
	if err != nil {
		fmt.Println(err)
		return
	}
}
