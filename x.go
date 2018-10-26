package main

import (
	"SmartWallet/lib"
	"crypto"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"log"
)

func main() {

	const pubPEM = `
-----BEGIN PUBLIC KEY-----
MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQDPngtgG2vZM1YZRw08Iu7RXXO3
1tOutyX/TkWO0T0g1IAn8jAKQ0ZCfo+7MMb2yeeqx7iKs9N2zPWvBeCxiM+h82H4
tbeMTp78laZnsK0m1Vk/yvYpuwmKr8ffb1I9UUUMm3vxpi7k00ZtTWnWtsXoSIpr
wrxwsJ7ZZ1O6JIUWswIDAQAB
-----END PUBLIC KEY-----
`

	block, _ := pem.Decode([]byte(pubPEM))
	if block == nil {
		panic("failed to parse PEM block containing the public key")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		panic("failed to parse DER encoded public key: " + err.Error())
	}

	//switch pub := pub.(type) {
	//case *rsa.PublicKey:
	//	fmt.Println("pub is of type RSA:", pub)
	//case *dsa.PublicKey:
	//	fmt.Println("pub is of type DSA:", pub)
	//case *ecdsa.PublicKey:
	//	fmt.Println("pub is of type ECDSA:", pub)
	//default:
	//	panic("unknown type of public key")
	//}
	rsaPublicKey := pub.(*rsa.PublicKey)

	secret := lib.DecodeString("752EF0D8FB4958670DBA40AB1F3C1D0F8FB4958670DBA40AB1F3752EF0DC1D0F")

	//cipher := lib.DecodeString("13758b480bd8d0ce5089aa23b16fde28")
	//cipher := lib.DecodeString("2CD47C6D98EE8FD57A09F9398612082B")
	cipher:=lib.DecodeString("8862FE16D56F325584107749487E2F78")
	plainText, err := lib.Decrypt(secret, cipher)
	if err != nil {
		log.Fatal(err)
	}

	//TODO: Decode all plaintext from base16
	fmt.Printf("%x\n",plainText)


	signature, _ := hex.DecodeString("4DA6833A21AB0051E6021305DCA949770B2100E0620C8C9C77E47495CC5871BE98115306FEEB6132D099005C20D98FC57E812F665BC9948BE953328EC651F794DE66E1EA3D8657C3F276489E6026ED2D6C16C6CA683A42487AA0C26DB76BD538D873FDFD5C045FAF58739272D06088C4AC4452DC90692ED16916D5B8AB8AC5F1")

	hashed := sha1.Sum([]byte("2CD47C6D98EE8FD57A09F9398612082B"))

	err = rsa.VerifyPKCS1v15(rsaPublicKey, crypto.SHA1, hashed[:], signature)
	if err != nil {
		log.Fatal(err)
	}

}
