package lib

import (
	"crypto/aes"
	"encoding/hex"
	"github.com/pkg/errors"
	"log"
)

//const PrivateKey = "752EF0D8FB4958670DBA40AB1F3C1D0F8FB4958670DBA40AB1F3752EF0DC1D0F"

//func main() {
//	key := DecodeString(PrivateKey)
//	plainText := []byte("1234567891234567")
//
//	Encrypt(key, plainText)
//
//	cipherText := DecodeString("5f87ba11faf8adea137ed6bd3cbd902a")
//	Decrypt(key, cipherText)
//
//}

//Decode the string from hex to bytes
func DecodeString(data string) []byte {
	decodedString, err := hex.DecodeString(data)
	if err != nil {
		log.Fatal(err)
		return nil
	}

	return decodedString
}

//Encrypt the plaintext
func Encrypt(key []byte, plaintext []byte) (cipher string, err error) {

	//Decode the key from a hex to a 32byte
	//It is assumed that the block size will always be 16bits so padding is not needed
	if len(plaintext)%aes.BlockSize != 0 {
		return "", errors.New("Input cannot be encrypted")
	}

	//Create a new cipher block
	block, err := aes.NewCipher(key)

	if err != nil {
		return "", err
	}

	//Create a byte of length 16 bytes
	cipherText := make([]byte, block.BlockSize())

	//Encrypt the plaintext
	block.Encrypt(cipherText, plaintext)

	//Encode cipher to hex and print
	//fmt.Printf("%x\n", cipherText)

	return hex.EncodeToString(cipherText), nil
}

//Decrypt the ciphertext
func Decrypt(key []byte, cipherText []byte) ([]byte, error) {

	//Create a new cipher block
	block, err := aes.NewCipher(key)

	if err != nil {
		//log.Fatal(err)
		return nil, errors.New("Could not create a cipher block")
	}

	//Create a byte of length 16 bytes
	plainText := make([]byte, block.BlockSize())

	//Decrypt the text
	block.Decrypt(plainText, cipherText)

	//Print the plaintext in a string format
	//fmt.Printf("%s\n", plainText)
	return plainText, nil
}
