package lib

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
)

const pubPEM = `
-----BEGIN PUBLIC KEY-----
MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQDPngtgG2vZM1YZRw08Iu7RXXO3
1tOutyX/TkWO0T0g1IAn8jAKQ0ZCfo+7MMb2yeeqx7iKs9N2zPWvBeCxiM+h82H4
tbeMTp78laZnsK0m1Vk/yvYpuwmKr8ffb1I9UUUMm3vxpi7k00ZtTWnWtsXoSIpr
wrxwsJ7ZZ1O6JIUWswIDAQAB
-----END PUBLIC KEY-----
`

//Verify the banks EMD to deposit the funds
func VerifySignature(emd string, sig string) (bool, error) {

	block, _ := pem.Decode([]byte(pubPEM))
	if block == nil {
		panic("failed to parse PEM block containing the public key")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		panic("failed to parse DER encoded public key: " + err.Error())
	}

	rsaKey, _ := pub.(*rsa.PublicKey)

	messageDecode, _ := hex.DecodeString(emd)
	str := string(messageDecode)

	hashed := sha1.Sum([]byte(str))

	signature, _ := hex.DecodeString(sig)

	err = rsa.VerifyPKCS1v15(rsaKey, crypto.SHA1, []byte(hashed[:]), signature)
	if err != nil {
		return false, err
	} else {
		return true, nil
	}

}
