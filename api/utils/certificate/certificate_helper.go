package certificate

import (
	"fmt"

	"01cloud-api/api/models"

	"github.com/gosimple/slug"
)

type CaCertificate struct {
	PrivateKey []byte
	PublicKey  []byte
	CSR        string
}

// func genCert(template, parent *x509.Certificate, publicKey *rsa.PublicKey, privateKey *rsa.PrivateKey) (*x509.Certificate, []byte, []byte) {
// 	certBytes, err := x509.CreateCertificate(rand.Reader, template, parent, publicKey, privateKey)
// 	if err != nil {
// 		panic("Failed to create certificate:" + err.Error())
// 	}

// 	cert, err := x509.ParseCertificate(certBytes)
// 	if err != nil {
// 		panic("Failed to parse certificate:" + err.Error())
// 	}

// 	b := pem.Block{Type: "CERTIFICATE", Bytes: certBytes}
// 	certPEM := pem.EncodeToMemory(&b)

// 	return cert, certBytes, certPEM
// }

// func genCARoot(commonName string) (*x509.Certificate, []byte, []byte, *rsa.PrivateKey) {
// 	var rootTemplate = x509.Certificate{
// 		SerialNumber: big.NewInt(1),
// 		Subject: pkix.Name{
// 			Country:      []string{"NE"},
// 			Organization: []string{"01cloud Co."},
// 			CommonName:   commonName,
// 		},
// 		NotBefore:             time.Now().Add(-10 * time.Second),
// 		NotAfter:              time.Now().AddDate(25, 0, 0),
// 		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
// 		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
// 		BasicConstraintsValid: true,
// 		IsCA:                  true,
// 		MaxPathLen:            2,
// 	}
// 	priv, err := rsa.GenerateKey(rand.Reader, 2048)
// 	if err != nil {
// 		panic(err)
// 	}
// 	rootCert, rootBytes, rootPEM := genCert(&rootTemplate, &rootTemplate, &priv.PublicKey, priv)
// 	return rootCert, rootBytes, rootPEM, priv
// }

// func createCertificateAuthority(keys *rsa.PrivateKey, commonName string) (*CaCertificate, error) {
// 	// step: generate a csr template
// 	var csrTemplate = x509.CertificateRequest{
// 		Subject: pkix.Name{
// 			Country:      []string{"NE"},
// 			Organization: []string{"01cloud Co."},
// 			CommonName:   commonName,
// 		},
// 		SignatureAlgorithm: x509.SHA512WithRSA,
// 		ExtraExtensions: []pkix.Extension{
// 			{
// 				Id:       asn1.ObjectIdentifier{2, 5, 29, 19},
// 				Critical: true,
// 			},
// 		},
// 	}
// 	// step: generate the csr request
// 	csrCertificate, err := x509.CreateCertificateRequest(rand.Reader, &csrTemplate, keys)
// 	if err != nil {
// 		return nil, err
// 	}
// 	csr := pem.EncodeToMemory(&pem.Block{
// 		Type: "CERTIFICATE REQUEST", Bytes: csrCertificate,
// 	})

// 	// step: generate a serial number
// 	serial, err := rand.Int(rand.Reader, (&big.Int{}).Exp(big.NewInt(2), big.NewInt(159), nil))
// 	if err != nil {
// 		return nil, err
// 	}

// 	now := time.Now()
// 	// step: create the request template
// 	template := x509.Certificate{
// 		SerialNumber: serial,
// 		Subject: pkix.Name{
// 			Country:      []string{"NE"},
// 			Organization: []string{"01cloud Co."},
// 			CommonName:   commonName,
// 		},
// 		NotBefore:             now.Add(-10 * time.Minute).UTC(),
// 		NotAfter:              now.AddDate(25, 0, 0),
// 		BasicConstraintsValid: true,
// 		IsCA:                  true,
// 		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
// 		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
// 	}

// 	// step: sign the certificate authority
// 	certificate, err := x509.CreateCertificate(rand.Reader, &template, &template, &keys.PublicKey, keys)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to generate certificate, error: %s", err)
// 	}

// 	var request bytes.Buffer
// 	var privateKey bytes.Buffer
// 	if err := pem.Encode(&request, &pem.Block{Type: "CERTIFICATE", Bytes: certificate}); err != nil {
// 		return nil, err
// 	}
// 	if err := pem.Encode(&privateKey, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(keys)}); err != nil {
// 		return nil, err
// 	}
// 	return &CaCertificate{
// 		PrivateKey: privateKey.Bytes(),
// 		PublicKey:  request.Bytes(),
// 		CSR:        string(csr),
// 	}, nil
// }

func GetCertificateFolder(request *models.Cluster) string {
	organizationName := "zerone"
	//if request.OrganizationID == 0 {
	//	return fmt.Sprintf("/data/certificates/zerone-devops-labs")
	//}
	if request.Organization != nil {
		organizationName = slug.Make(request.Organization.Name)
	}

	return fmt.Sprintf("/data/certificates/%s/%s", organizationName, slug.Make(fmt.Sprintf("%s-%d", request.Name, request.ID)))
}
