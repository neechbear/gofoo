
package main

import (
	"log"
	"net"
	"os"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

const (
	username = "nicolaw"
	server   = "localhost:1009"
)

func main() {
	conn, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK"))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	ag := agent.NewClient(conn)
	auths := []ssh.AuthMethod{ssh.PublicKeysCallback(ag.Signers)}

	config := &ssh.ClientConfig{
		User: username,
		Auth: auths,
	}
	client, err := ssh.Dial("tcp", server, config)
	if err != nil {
		log.Fatalln("Failed to dial:", err)
	}

	session, err := client.NewSession()
	if err != nil {
		log.Fatalln("Failed to create session:", err)
	}
	defer session.Close()

	session.Stdout = os.Stdout
	session.Stderr = os.Stderr
	if err := session.Run("ls"); err != nil {
		log.Fatalln("Failed to run:", err)
	}
}


