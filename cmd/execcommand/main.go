package main

import (
	"bufio"
	"log"
	"log/syslog"
	"os/exec"
)

// Variables to identify the build.
var (
	Version  string
	Build    string
	Identity string
)

func main() {
	logwriter, e := syslog.New(syslog.LOG_NOTICE, Identity)
	if e == nil {
		log.SetOutput(logwriter)
	}
	log.Printf("%s version %s (build %s)", Identity, Version, Build)

	cmdBin := "/usr/bin/find"
	cmdArgs := []string{"/", "-maxdepth", "2", "-name", "*dev*"}
	cmd := exec.Command(cmdBin, cmdArgs...)

	cmdReader, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	cmd.Stderr = cmd.Stdout

	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	scanner := bufio.NewScanner(cmdReader)
	for scanner.Scan() {
		log.Print(scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		log.Printf("Reading cmd.StdoutPipe: %#v", err)
	}

	if err := cmd.Wait(); err != nil {
		log.Fatal(err)
	}
}
