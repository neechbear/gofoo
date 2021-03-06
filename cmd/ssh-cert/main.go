package main

/*
https://github.com/neechbear/teleport/blob/master/lib/auth/native/native.go
https://github.com/gravitational/teleport/blob/master/lib/auth/native/native.go

Copyright 2015 Gravitational, Inc.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

var (
	// this global configures how many pre-caluclated keypairs to keep in the
	// background (perform key genreation in a separate goroutine, useful for
	// web sesssion for snappy UI)
	PrecalculatedKeysNum = 10

	// only one global copy of 'nauth' exists
	singleton nauth = nauth{
		closeC: make(chan bool),
	}
)

type keyPair struct {
	privPem  []byte
	pubBytes []byte
}

type nauth struct {
	generatedKeysC chan keyPair
	closeC         chan bool
	mutex          sync.Mutex
}

// New returns a pointer to a key generator for production purposes
func New() *nauth {
	singleton.mutex.Lock()
	defer singleton.mutex.Unlock()

	if singleton.generatedKeysC == nil && PrecalculatedKeysNum > 0 {
		singleton.generatedKeysC = make(chan keyPair, PrecalculatedKeysNum)
		go singleton.precalculateKeys()
	}
	return &singleton
}

// Close() closes and re-sets the key generator (better to call it only once,
// when the process is stopping, to avoid costly re-initialization)
func (n *nauth) Close() {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	close(n.closeC)
	n.generatedKeysC = nil
	n.closeC = make(chan bool)
}

// GetNewKeyPairFromPool returns pre-generated keypair from a channel, which
// gets replenished by `precalculateKeys` goroutine
func (n *nauth) GetNewKeyPairFromPool() ([]byte, []byte, error) {
	select {
	case key := <-n.generatedKeysC:
		return key.privPem, key.pubBytes, nil
	default:
		return n.GenerateKeyPair("")
	}
}

func (n *nauth) precalculateKeys() {
	for {
		privPem, pubBytes, err := n.GenerateKeyPair("")
		if err != nil {
			log.Errorf(err.Error())
			continue
		}
		key := keyPair{
			privPem:  privPem,
			pubBytes: pubBytes,
		}

		select {
		case <-n.closeC:
			log.Infof("[KEYS] precalculateKeys() exited")
			return
		case n.generatedKeysC <- key:
			continue
		}
	}
}

// GenerateKeyPair returns fresh priv/pub keypair, takes about 300ms to execute
func (n *nauth) GenerateKeyPair(passphrase string) ([]byte, []byte, error) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}
	privDer := x509.MarshalPKCS1PrivateKey(priv)
	privBlock := pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   privDer,
	}
	privPem := pem.EncodeToMemory(&privBlock)

	pub, err := ssh.NewPublicKey(&priv.PublicKey)
	if err != nil {
		return nil, nil, err
	}
	pubBytes := ssh.MarshalAuthorizedKey(pub)
	return privPem, pubBytes, nil
}

func (n *nauth) GenerateHostCert(c services.CertParams) ([]byte, error) {
	if err := c.Check(); err != nil {
		return nil, trace.Wrap(err)
	}

	pubKey, _, _, _, err := ssh.ParseAuthorizedKey(c.PublicHostKey)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	signer, err := ssh.ParsePrivateKey(c.PrivateCASigningKey)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	principals := buildPrincipals(c.HostID, c.NodeName, c.ClusterName, c.Roles)

	// create certificate
	validBefore := uint64(ssh.CertTimeInfinity)
	if c.TTL != 0 {
		b := time.Now().Add(c.TTL)
		validBefore = uint64(b.Unix())
	}
	cert := &ssh.Certificate{
		ValidPrincipals: principals,
		Key:             pubKey,
		ValidBefore:     validBefore,
		CertType:        ssh.HostCert,
	}
	cert.Permissions.Extensions = make(map[string]string)
	cert.Permissions.Extensions[utils.CertExtensionRole] = c.Roles.String()
	cert.Permissions.Extensions[utils.CertExtensionAuthority] = string(c.ClusterName)

	// sign host certificate with private signing key of certificate authority
	if err := cert.SignCert(rand.Reader, signer); err != nil {
		return nil, trace.Wrap(err)
	}

	return ssh.MarshalAuthorizedKey(cert), nil
}

func (n *nauth) GenerateUserCert(pkey, key []byte, teleportUsername string, allowedLogins []string, ttl time.Duration, permitAgentForwarding bool) ([]byte, error) {
	if ttl < defaults.MinCertDuration {
		return nil, trace.BadParameter("wrong certificate TTL")
	}
	if len(allowedLogins) == 0 {
		return nil, trace.BadParameter("allowedLogins: need allowed OS logins")
	}
	pubKey, _, _, _, err := ssh.ParseAuthorizedKey(key)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	validBefore := uint64(ssh.CertTimeInfinity)
	if ttl != 0 {
		b := time.Now().Add(ttl)
		validBefore = uint64(b.Unix())
		log.Infof("generated user key for %v with expiry on (%v) %v", allowedLogins, validBefore, b)
	}
	// we do not use any extensions in users certs because of this:
	// https://bugzilla.mindrot.org/show_bug.cgi?id=2387
	cert := &ssh.Certificate{
		KeyId:           teleportUsername, // we have to use key id to identify teleport user
		ValidPrincipals: allowedLogins,
		Key:             pubKey,
		ValidBefore:     validBefore,
		CertType:        ssh.UserCert,
	}
	cert.Permissions.Extensions = map[string]string{
		teleport.CertExtensionPermitPTY:            "",
		teleport.CertExtensionPermitPortForwarding: "",
	}
	if permitAgentForwarding {
		cert.Permissions.Extensions[teleport.CertExtensionPermitAgentForwarding] = ""
	}
	signer, err := ssh.ParsePrivateKey(pkey)
	if err != nil {
		return nil, err
	}
	if err := cert.SignCert(rand.Reader, signer); err != nil {
		return nil, err
	}
	return ssh.MarshalAuthorizedKey(cert), nil
}

func buildPrincipals(host string, domain string) []string {
	principals := []string{
		host,
	}
	if domain != "" {
		principals = append(principals, fmt.Sprintf("%s.%s", host, domain))
	}
	return utils.Deduplicate(principals)
}

func main() {
	MakeSSHKeyPair("mykey.pub", "mykey")
}
