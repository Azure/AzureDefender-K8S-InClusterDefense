package utils

//TODO add test for this package
import (
	"crypto/tls"
	"crypto/x509"
	"github.com/pkg/errors"
	"io/ioutil"
)

// CreateTlsConfig creates tls.config object.
func CreateTlsConfig(tlsCrtPath string, tlsKeyPath string, caCertPath string, host string) (*tls.Config, error) {
	// Get certificates
	cert, err := tls.LoadX509KeyPair(tlsCrtPath, tlsKeyPath)
	if err != nil {
		err = errors.Wrap(err, "Failed to LoadX509KeyPair")
		return nil, err
	}
	caCert, err := ioutil.ReadFile(caCertPath)
	if err != nil {
		err = errors.Wrap(err, "Failed to read ca.cert file ")
		return nil, err
	}

	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(caCert)

	// Create tlsConfig object with the cert files
	tlsConfig := &tls.Config{
		ServerName:   host,
		Certificates: []tls.Certificate{cert},
		RootCAs:      pool,
	}
	return tlsConfig, nil
}

// GetPasswordFromFile gets a password from a file containing only the password
//TODO - change to get password directly from the secret (no mount)
func GetPasswordFromFile(passwordPath string) (string, error) {
	// Get password
	passwordInBytes, err := ioutil.ReadFile(passwordPath)
	if err != nil {
		err = errors.Wrap(err, "Failed to read password file")
		return "", err
	}
	return string(passwordInBytes), nil
}
