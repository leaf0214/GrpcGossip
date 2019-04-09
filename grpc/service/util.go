package service

import (
	"crypto/x509"
	"encoding/pem"
)

func AddPemToCertPool(pemCerts []byte, pool *x509.CertPool) error {
	certs, _, err := pemToX509Certs(pemCerts)
	if err != nil {
		return err
	}
	for _, cert := range certs {
		pool.AddCert(cert)
	}
	return nil
}

func pemToX509Certs(pemCerts []byte) ([]*x509.Certificate, []string, error) {
	var certs []*x509.Certificate
	var subjects []string
	for len(pemCerts) > 0 {
		var block *pem.Block
		block, pemCerts = pem.Decode(pemCerts)
		if block == nil {
			break
		}
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, subjects, err
		}
		certs = append(certs, cert)
		subjects = append(subjects, string(cert.RawSubject))
	}
	return certs, subjects, nil
}
