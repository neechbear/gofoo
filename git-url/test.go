package main

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"strings"

	log "github.com/Sirupsen/logrus"
)

// Variables to identify the build.
var (
	Version  string
	Build    string
	Identity string
)

func getScheme(rawurl string) (scheme, path string, err error) {
	for i := 0; i < len(rawurl); i++ {
		c := rawurl[i]
		switch {
		case 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z':
		// do nothing
		case '0' <= c && c <= '9' || c == '+' || c == '-' || c == '.':
			if i == 0 {
				return "", rawurl, errors.New("missing protocol scheme")
			}
		case c == ':':
			if i == 0 {
				return "", "", errors.New("missing protocol scheme")
			}
			return rawurl[:i], rawurl[i+1:], nil
		default:
			// we have encountered an invalid character,
			// so there is no valid scheme
			return "", rawurl, errors.New("missing protocol scheme")
		}
	}
	return "", rawurl, errors.New("missing protocol scheme")
}

func parseGitURL(s string, orig string) error {
	u, err := url.Parse(s)
	if err != nil {
		scheme, _, err := getScheme(s)

		// Edge case git@/scp type "URL" syntax.
		if err != nil && scheme == "" {
			i := strings.Index(s, "]") // Does it look like we have an IPv6 host
			j := 0
			if i < 0 {
				j = strings.Index(s, ":")
			} else { // Probably IPv6 hostname.
				j = strings.Index(s[i:], ":")
			}

			if j < 0 { // No colon in URL, so probably bogus.
				err = errors.New("No colon (:) in URL; unable to munge git " +
					"edge case")
				log.WithFields(log.Fields{
					"i":    i,
					"j":    j,
					"s":    s,
					"orig": orig,
				}).Error(err.Error())
				return err
			}

			if i >= 0 { // Probably IPv6 hostname.
				j = j + i // Add IPv6 hostname index offset to colon (:) offset
			}

			// Munge a URL with forced ssh:// scheme, replacing the delimiting
			// hostname:path colon (:) with a slash (/).
			s2 := "ssh://" + s[:j] + strings.Replace(s[j:], ":", "/", 1)

			log.WithFields(log.Fields{
				"i":    i,
				"j":    j,
				"s2":   s2,
				"s":    s,
				"orig": orig,
			}).Debug(orig)

			// Reparse our munged URL.
			return parseGitURL(s2, s)
		}
		log.WithFields(log.Fields{
			"err": fmt.Sprintf("%#v", err),
		}).Error(err.Error())
		return err
	}

	path := u.RequestURI()
	if s != orig {
		path = path[1:]
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
		"path":   path,
	}).Info(orig)

	return err
}

func main() {
	log.SetLevel(log.InfoLevel)
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
		url := scanner.Text()
		parseGitURL(url, url)
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}
