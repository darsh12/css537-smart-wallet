package lib

import (
	"errors"
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
