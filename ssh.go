package gitdb

import (
	"crypto/rsa"
	"os"
	"encoding/pem"
	"crypto/x509"
	"crypto/rand"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
)

// generateSSHKeyPair make a pair of public and private keys for SSH access.
// Public key is encoded in the format for inclusion in an OpenSSH authorized_keys file.
// Private Key generated is PEM encoded
func generateSSHKeyPair() error {

	if _, err := os.Stat(privateKeyFilePath()); err == nil {

		if _, err := os.Stat(publicKeyFilePath()); err == nil {
			return nil
		}

		log("Re-generating public key")
		//public key is missing - recreate public key
		pkPem, err := ioutil.ReadFile(privateKeyFilePath())
		if err != nil {
			return err
		}

		b, _ := pem.Decode(pkPem)
		privateKey, err := x509.ParsePKCS1PrivateKey(b.Bytes)
		if err != nil {
			return err
		}

		return generatePublicKey(privateKey)
	}

	if _, err := os.Stat(sshDir()); err != nil {
		if err := os.MkdirAll(sshDir(), os.ModePerm); err != nil {
			return err
		}
	}

	log("Generating ssh key pairs")
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}

	err = generatePrivateKey(privateKey)
	if err != nil {
		return err
	}

	return generatePublicKey(privateKey)
}

func generatePrivateKey(pk *rsa.PrivateKey) error {
	// generate and write private key as PEM
	privateKeyFile, err := os.OpenFile(privateKeyFilePath(), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0400)
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

func generatePublicKey(pk *rsa.PrivateKey) error {
	// generate and write public key
	pub, err := ssh.NewPublicKey(&pk.PublicKey)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(publicKeyFilePath(), ssh.MarshalAuthorizedKey(pub), 0655)
}