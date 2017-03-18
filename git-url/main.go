package main

import (
	"bufio"
	"net"
	"os"

	log "github.com/Sirupsen/logrus"
	giturl "github.com/neechbear/gogiturl"
)

// Variables to identify the build.
var (
	Version  string
	Build    string
	Identity string
)

func main() {
	log.SetLevel(log.DebugLevel)
	log.WithFields(log.Fields{
		"version": Version,
		"build":   Build,
	}).Debugf("Starting %s", Identity)

	file, err := os.Open("testdata.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		rawurl := scanner.Text()

		u, err := giturl.Parse(rawurl)
		if err != nil {
			log.WithFields(log.Fields{
				"err": err,
			}).Errorf("%s", err.Error())
			continue
		}

		user := ""
		if u.User != nil {
			user = u.User.Username()
		}
		_, port, _ := net.SplitHostPort(u.Host)

		log.WithFields(log.Fields{
			"scheme": u.Scheme,
			"user":   user,
			"host":   u.Hostname(),
			"port":   port,
			"path":   u.Path,
		}).Info(rawurl)
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}