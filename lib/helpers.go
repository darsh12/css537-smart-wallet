package lib

import (
	"encoding/hex"
	"errors"
	"math/big"
	"strconv"
)

/*
Pad zeros to left of the string in order to fill byte array
*/
func PadStringLeft(str string) (padded string, err error) {
	i, err := strconv.Atoi(str)
	if err != nil {
		return "", errors.New("could not convert string to int")
	}

	padded = strconv.Itoa(i)
	switch len(padded) {
	case 1:
		return "000" + padded, nil
	case 2:
		return "00" + padded, nil
	case 3:
		return "0" + padded, nil
	default:
		return padded, nil
	}
}

//Extract hex encoded values to public exponent and public modulus
func DecodePublicKeyValues(n string, e string) (*big.Int, int) {

	//eDecode, _ := hex.DecodeString(e)
	//eDecimal := big.NewInt(0)
	//eDecimal.SetBytes(eDecode)
	//Separate
	eDecimal, _ := strconv.ParseInt(e, 16, 32)

	nDecode, _ := hex.DecodeString(n)
	nDecimal := big.NewInt(0)
	nDecimal.SetBytes(nDecode)

	return nDecimal, int(eDecimal)

}
