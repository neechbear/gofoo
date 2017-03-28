
package main

import (
  "os"
  "crypto/rsa"
  "crypto/rand"
  "encoding/pem"
  "crypto/x509"
  "golang.org/x/crypto/ssh"
  "io/ioutil"
)

// MakeSSHKeyPair make a pair of public and private keys for SSH access.
// Public key is encoded in the format for inclusion in an OpenSSH authorized_keys file.
// Private Key generated is PEM encoded
func MakeSSHKeyPair(pubKeyPath, privateKeyPath string) error {
    privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
    if err != nil {
        return err
    }

    // generate and write private key as PEM
    privateKeyFile, err := os.Create(privateKeyPath)
    defer privateKeyFile.Close()
    if err != nil {
        return err
    }
    privateKeyPEM := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)}
    if err := pem.Encode(privateKeyFile, privateKeyPEM); err != nil {
        return err
    }

    // generate and write public key
    pub, err := ssh.NewPublicKey(&privateKey.PublicKey)
    if err != nil {
        return err
    }
    return ioutil.WriteFile(pubKeyPath, ssh.MarshalAuthorizedKey(pub), 0644)
}

func main() {
	MakeSSHKeyPair("mykey.pub", "mykey")
}



