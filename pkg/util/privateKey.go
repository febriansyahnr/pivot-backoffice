package util

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
)

func RSAPrivateKey(val string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(val))
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing the key")
	}

	if block.Type == "RSA PRIVATE KEY" {
		pkcs1, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		return pkcs1, nil
	}

	pkcs8, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	rsaKey, ok := pkcs8.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("key is not an RSA private key")
	}
	return rsaKey, nil
}

func GetMockPrivateKey() *rsa.PrivateKey {
	pemData := `-----BEGIN PRIVATE KEY-----
MIIEvwIBADANBgkqhkiG9w0BAQEFAASCBKkwggSlAgEAAoIBAQDI/t4aI3sTFTTH
Lxpdd1HFqKoMUyeHFa9Oiq//OGlMODsR5JHsF+IuRfar/U8HXe6jg0j5sLp/5D8a
i5MiqvyXJxd1j8EuKKs/qeQrtS0usqqJYJDMze3YX1eIHrD5LO2CBiJhPp6gIhUO
SfhImxJEMaMDiounZmoK9P66JrEKjedRbuVgDn4PtxSLr3VzYYRZPa65D5HzjdIm
OuotubMN3mbWRWhicT2oM29fTe3cJlzRAcaOmicPdRtNTf0/Wk+jCDeiJxoLuKyZ
/IUSVuzIl4gnMGKM/KheSPOlTRwVm+eHWjr2+xB4KZf00FNoJTu+TXUEe0pdYVfw
ioeavUHdAgMBAAECggEBAJNI8EgHJ/Db4Uj0Y0WKYgmNhs5xQM3kPgo35rAHDmIj
8mUyMRvohH2UFyYBASBM3MpFMfyGXKPLBdLV5IPK+D1rD+294bmJY7PLMsA0i19k
3UK92F27qUac1u+QTe7J1WEqTZck4+hEEVnfKmlJ+SCvntzBcYTBr4NH9EFEiQdJ
l1fMfMxS6HgpIcY2wxnypqjpeC/LdG4543iGpP8zbnunxW43yFgkNC8AJ/KuY9eL
dX/bES25F0mmhQRruzpFczre+4S2tiVhMq73VuyXJiYIvaCxA1JHQjQBCjmkJ7r7
lNKZYlf7Qfcg5FZ9njGdFWpDeM40vyiGYZbKVV6X4aECgYEA9n84ldyPD4U8Kmnz
56pJ5lK+2ztC90pXCNS+vpCqg3VdLf7eE1nUAp2f6JBA/qm3kMIfju49bpTFM6ZD
Kj4jI8WI7JdBy1gfNls9HF5nVtoHzTHG9E3wT0YAOPHmyCfohpfFIFaWX63VwZ15
/jKP9OqW7aeMLoOWv9Mg+WQbVrMCgYEA0L6TKuttEZmyLDFhb+4T52Yl8ofNLbxe
6BhqDy44xwjBsi2a0BDc9l1/kv2ACsHEfMNFO/jW1IZ4C+2S3WDSoH+ff7uFVxS0
1wGrNNM/ch/zX3oLzX+ZevAYSys95IshMQMfoePp533ZYk504nRsxM4JVGSUoSHK
MvmnlRMJzS8CgYEAiJOE7sP+IENaSsXZ9opL1+oRBbeYKxxtjN8TsNLHJ39n2YxV
z7L93VUovNrwqCmxI+vrQG6QayzS9wMwQ7+aCL/yVeSY9+ojoSJ8gbNs3pp/qBnk
eoiUldfbV7HwhQZXt/tvpbNULj9LKLPwW//382PnrFYhPcR7Sl3Y71WgMDECgYBv
DzXVe/RHjPJSuOMSXiSQ1LQT2VS8pKAJ9BNZiEoE+w+y8LiRQqeNHCmn1t+s2XLk
vi+zvKzv3as5DWk6By2I3t3JY8eJkSa1zdl8/XegDIe7oH9vEhhiZCNIuvTvB2bd
YMAPrebglwB1YTCm2zKTcttb3zeEkym0/Ua/9aUdWQKBgQC43+cW1iWcw3U8LhO7
3Wkys/d07xkZEoEGJRFymqhyUNIBRY3fe4ukQOp3Qq0bff0Oj5YBDD6n+TES9nSh
wrKPt/kduqwqr9Ob4SwaUwX18rQlQwRxo2O3EQLcIIMiMVSuDKdc68AT8uc1Dbqn
gmyBqtD/AVn+rKit0f7HDuOrcw==
-----END PRIVATE KEY-----`

	block, _ := pem.Decode([]byte(pemData))
	if block == nil {
		return nil
	}

	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil
	}

	return privateKey.(*rsa.PrivateKey)
}
