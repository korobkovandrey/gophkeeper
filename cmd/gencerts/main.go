package main

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"flag"
	"log"
	"math/big"
	"net"
	"os"
	"strings"
	"time"
)

func main() {
	org := flag.String("org", "Yandex.Praktikum", "Organization name")
	country := flag.String("country", "RU", "Country code")
	years := flag.Int("years", 10, "Certificate validity period in years")
	ips := flag.String("ips", "127.0.0.1,::1", "Comma-separated list of IP addresses")
	serial := flag.Int64("serial", 1658, "Certificate serial number")
	caSerial := flag.Int64("ca-serial", 1659, "CA certificate serial number")
	dns := flag.String("dns", "localhost", "Comma-separated list of DNS names")
	certFile := flag.String("cert", "./certs/server.crt", "Output certificate file")
	keyFile := flag.String("key", "./certs/server.key", "Output private key file")
	caFile := flag.String("ca", "./certs/ca.crt", "Output CA certificate file")
	flag.Parse()

	ipList := strings.Split(*ips, ",")
	ipAddresses := make([]net.IP, len(ipList))
	for i := range ipList {
		ipAddresses[i] = net.ParseIP(strings.TrimSpace(ipList[i]))
		if ipAddresses[i] == nil {
			log.Fatalf("Invalid IP address: %s", ipList[i])
		}
	}
	dnsNames := strings.Split(*dns, ",")
	for i, name := range dnsNames {
		dnsNames[i] = strings.TrimSpace(name)
		if dnsNames[i] == "" {
			log.Fatalf("Invalid DNS name: %s", name)
		}
	}
	caCert := &x509.Certificate{
		SerialNumber: big.NewInt(*caSerial),
		Subject: pkix.Name{
			Organization: []string{*org + " CA"},
			Country:      []string{*country},
			CommonName:   "Root CA",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(*years, 0, 0),
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
	}
	caPrivateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		log.Fatalf("Failed to generate CA private key: %v", err)
	}
	caCertBytes, err := x509.CreateCertificate(rand.Reader, caCert, caCert, &caPrivateKey.PublicKey, caPrivateKey)
	if err != nil {
		log.Fatalf("Failed to create CA certificate: %v", err)
	}
	var caCertPEM bytes.Buffer
	err = pem.Encode(&caCertPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caCertBytes,
	})
	if err != nil {
		log.Fatalf("Failed to encode CA certificate: %v", err)
	}

	err = os.WriteFile(*caFile, caCertPEM.Bytes(), 0600)
	if err != nil {
		log.Fatalf("Failed to write CA certificate to %s: %v", *caFile, err)
	}

	serverCert := &x509.Certificate{
		SerialNumber: big.NewInt(*serial),
		Subject: pkix.Name{
			Organization: []string{*org},
			Country:      []string{*country},
			CommonName:   "Server Certificate",
		},
		IPAddresses:  ipAddresses,
		DNSNames:     dnsNames,
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(*years, 0, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}
	serverPrivateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		log.Fatalf("Failed to generate server private key: %v", err)
	}
	caCertParsed, err := x509.ParseCertificate(caCertBytes)
	if err != nil {
		log.Fatalf("Failed to parse CA certificate: %v", err)
	}
	serverCertBytes, err := x509.CreateCertificate(rand.Reader, serverCert, caCertParsed, &serverPrivateKey.PublicKey, caPrivateKey)
	if err != nil {
		log.Fatalf("Failed to create server certificate: %v", err)
	}
	var certPEM bytes.Buffer
	err = pem.Encode(&certPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: serverCertBytes,
	})
	if err != nil {
		log.Fatalf("Failed to encode server certificate: %v", err)
	}
	var privateKeyPEM bytes.Buffer
	err = pem.Encode(&privateKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(serverPrivateKey),
	})
	if err != nil {
		log.Fatalf("Failed to encode server private key: %v", err)
	}
	err = os.WriteFile(*certFile, certPEM.Bytes(), 0600)
	if err != nil {
		log.Fatalf("Failed to write server certificate to %s: %v", *certFile, err)
	}
	err = os.WriteFile(*keyFile, privateKeyPEM.Bytes(), 0600)
	if err != nil {
		log.Fatalf("Failed to write server private key to %s: %v", *keyFile, err)
	}

	log.Printf("CA certificate saved to %s", *caFile)
	log.Printf("Server certificate saved to %s", *certFile)
	log.Printf("Server private key saved to %s", *keyFile)
}
