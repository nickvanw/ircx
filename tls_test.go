package ircx

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"testing"
	"time"
)

func TestTLSConnect(t *testing.T) {
	// Generate self-signed certificate for mock server

	ca := &x509.Certificate{
		SerialNumber: big.NewInt(1234),
		Subject: pkix.Name{
			Country:            []string{"USA"},
			Organization:       []string{"ircxtest"},
			OrganizationalUnit: []string{"test"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		SubjectKeyId:          []byte{1, 2, 3, 4, 5},
		BasicConstraintsValid: true,
		IsCA:        true,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
	}

	priv, _ := rsa.GenerateKey(rand.Reader, 1024)
	pub := &priv.PublicKey

	ca_b, err := x509.CreateCertificate(rand.Reader, ca, ca, pub, priv)
	if err != nil {
		t.Fatalf("error generating self-cert: %v", err)
	}

	cert := tls.Certificate{
		Certificate: [][]byte{ca_b},
		PrivateKey:  priv,
	}

	// prep config for mock server
	serverConfig := tls.Config{Certificates: []tls.Certificate{cert}}

	l, err := tls.Listen("tcp", "127.0.0.1:0", &serverConfig)
	if err != nil {
		t.Fatalf("Wanted listener, got err: %v", err)
	}

	// prep bot/bot config
	botConfig := &tls.Config{InsecureSkipVerify: true}
	b := WithLoginTLS(l.Addr().String(), "test-bot", "test-user", "test-password", botConfig)

	not := make(chan string)
	go echoHelper(l, not)
	err = b.Connect()
	if err != nil {
		t.Fatalf("error connecting to mock TLS server: %v", err)
	}

	// We should get back the connect info. If 500ms has happened and we haven't gotten anything
	// we're either not connected right, or all of the data has been sent.
	data := []string{}
	for {
		select {
		case d := <-not:
			data = append(data, d)
		case <-time.After(250 * time.Millisecond):
			goto DONE
		}
	}
DONE:
	d := b.connectMessages()
	for k, v := range d {
		if v.String() != data[k] {
			t.Fatalf("Should have recieved %s, got %s", d[k], v)
		}
	}
}
