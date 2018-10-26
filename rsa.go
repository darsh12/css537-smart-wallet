package main

import (
	"encoding/hex"
	"fmt"
	"math/big"
)

func main() {
	c := "4674726773D7A63748F15974BD70EEE545B6A8E52EBC2DE8DA03A28E539073EA9B8889D2467FA21C4AB2D75D18BB65D5CC1BBB5E0471759987E65012D850D28CBF1D1D31862895383E4C5EB9F75C7AF009BB5DD0D700880F009BB7E502B80E58C91CD5F95DF607BA1375EFDBBE8E919A65DE4801C0AE1D7C3316BD9F874AC1E1"
	n := "CF9E0B601B6BD9335619470D3C22EED15D73B7D6D3AEB725FF4E458ED13D20D48027F2300A4346427E8FBB30C6F6C9E7AAC7B88AB3D376CCF5AF05E0B188CFA1F361F8B5B78C4E9EFC95A667B0AD26D5593FCAF629BB098AAFC7DF6F523D51450C9B7BF1A62EE4D3466D4D69D6B6C5E8488A6BC2BC70B09ED96753BA248516B3"
	//e := "010001"

	//eDecode, _ := hex.DecodeString(e)
	//eDecimal := big.NewInt(0)
	//eDecimal.SetBytes(eDecode)
	//eDecimal,_:=strconv.ParseInt(e,16,32)


	cDecode, _ := hex.DecodeString(c)
	cDecimal := big.NewInt(0)
	cDecimal.SetBytes(cDecode)


	nDecode, _ := hex.DecodeString(n)
	nDecimal := big.NewInt(0)
	nDecimal.SetBytes(nDecode)

	fmt.Println(nDecimal)


}
