
// agent.go
package main

import (
    "io"
    "net"

    "golang.org/x/crypto/ssh/agent"
)

func main() {
    keyring := agent.NewKeyring()

    l, err := net.Listen("unix", "agent_test.sock")
    if err != nil {
        panic(err)
    }

    for {
        c, err := l.Accept()
        if err != nil {
            panic(err)
        }

        err = agent.ServeAgent(keyring, c)
        if err != nil && err != io.EOF {
            panic(err)
        }
    }
}

