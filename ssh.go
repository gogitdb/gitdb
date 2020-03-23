package gitdb

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"os"

	"golang.org/x/crypto/ssh"
)

// generateSSHKeyPair make a pair of public and private keys for SSH access.
// Public key is encoded in the format for inclusion in an OpenSSH authorized_keys file.
// Private Key generated is PEM encoded
func (g *gitdb) generateSSHKeyPair() error {

	if _, err := os.Stat(g.privateKeyFilePath()); err == nil {

		if _, err := os.Stat(g.publicKeyFilePath()); err == nil {
			return nil
		}

		log("Re-generating public key")
		//public key is missing - recreate public key
		pkPem, err := ioutil.ReadFile(g.privateKeyFilePath())
		if err != nil {
			return err
		}

		b, _ := pem.Decode(pkPem)
		privateKey, err := x509.ParsePKCS1PrivateKey(b.Bytes)
		if err != nil {
			return err
		}

		return g.generatePublicKey(privateKey)
	}

	if _, err := os.Stat(g.sshDir()); err != nil {
		if err := os.MkdirAll(g.sshDir(), os.ModePerm); err != nil {
			return err
		}
	}

	log("Generating ssh key pairs")
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}

	err = g.generatePrivateKey(privateKey)
	if err != nil {
		return err
	}

	return g.generatePublicKey(privateKey)
}

func (g *gitdb) generatePrivateKey(pk *rsa.PrivateKey) error {
	// generate and write private key as PEM
	privateKeyFile, err := os.OpenFile(g.privateKeyFilePath(), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0400)
	defer privateKeyFile.Close()
	if err != nil {
		return err
	}
	privateKeyPEM := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(pk)}
	if err := pem.Encode(privateKeyFile, privateKeyPEM); err != nil {
		return err
	}

	return nil
}

func (g *gitdb) generatePublicKey(pk *rsa.PrivateKey) error {
	// generate and write public key
	pub, err := ssh.NewPublicKey(&pk.PublicKey)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(g.publicKeyFilePath(), ssh.MarshalAuthorizedKey(pub), 0655)
}
