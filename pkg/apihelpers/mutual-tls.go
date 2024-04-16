package apihelpers

import (
	"crypto/tls"
	"crypto/x509"
	"os"
)

type CertificatePaths struct {
	ServerCertPath string `yaml:"server_cert_path"`
	ServerKeyPath  string `yaml:"server_key_path"`
	CACertPath     string `yaml:"ca_cert_path"`
}

func LoadTLSConfig(paths CertificatePaths) (*tls.Config, error) {
	serverCert, err := tls.LoadX509KeyPair(paths.ServerCertPath, paths.ServerKeyPath)
	if err != nil {
		return nil, err
	}

	caCert, err := os.ReadFile(paths.CACertPath)
	if err != nil {
		return nil, err
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	return &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    caCertPool,
	}, nil
}
