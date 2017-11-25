package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/BrianWill/pigeon/dynamicPigeon"
	"github.com/BrianWill/pigeon/staticPigeon"
)

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

func Run(filename string) (*exec.Cmd, error) {
	cmd := exec.Command("go", "run", filename)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Start()
	if err != nil {
		return nil, err
	}
	return cmd, nil
}

func main() {
	if len(os.Args) >= 2 {
		subcommand := os.Args[1]
		if subcommand == "run" {
			if len(os.Args) < 3 {
				fmt.Println("Must specify a file to run.")
				return
			}
			gopath := os.Getenv("GOPATH")
			basedir := gopath + "/src/pigeon_output/"
			outputFile := "output.go"
			if _, err := os.Stat(basedir); os.IsNotExist(err) {
				os.Mkdir(basedir, os.ModePerm)
			}
			if strings.HasSuffix(os.Args[2], ".sp") {
				packages, err := staticPigeon.Compile(os.Args[2], "pigeon_output/")
				if err != nil {
					fmt.Println(err)
					return
				}
				for _, p := range packages {
					pkgDir := basedir + p.Prefix
					if _, err := os.Stat(pkgDir); os.IsNotExist(err) {
						os.Mkdir(pkgDir, os.ModePerm)
					}
					outputFilename := pkgDir + "/" + p.Prefix + ".go"
					err = ioutil.WriteFile(outputFilename, []byte(p.Code), os.ModePerm)
					if err != nil {
						fmt.Println(err)
						return
					}
					err = exec.Command("go", "fmt", outputFilename).Run()
					if err != nil {
						fmt.Println(err)
						return
					}
				}
			} else if strings.HasSuffix(os.Args[2], ".dp") {
				pkg, err := dynamicPigeon.Compile(os.Args[2], "pigeon_output/")
				if err != nil {
					fmt.Println(err)
					return
				}
				outputFilename := basedir + outputFile
				err = ioutil.WriteFile(outputFilename, []byte(pkg.Code), os.ModePerm)
				if err != nil {
					fmt.Println(err)
					return
				}
				err = exec.Command("go", "fmt", outputFilename).Run()
				if err != nil {
					fmt.Println(err)
					return
				}
			} else {
				log.Fatal("File has improper extension.")
			}
			_, err := Run(basedir + outputFile)
			if err != nil {
				fmt.Println(err)
				return
			}
		}
	}
}
