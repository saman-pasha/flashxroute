package flashxroute

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"math/big"
	"encoding/base64"
	
	"github.com/ethereum/go-ethereum/core/types"
)

// ParseInt parse hex string value to int
func ParseInt(value string) (int, error) {
	i, err := strconv.ParseInt(strings.TrimPrefix(value, "0x"), 16, 64)
	if err != nil {
		return 0, err
	}

	return int(i), nil
}

// ParseBigInt parse hex string value to big.Int
func ParseBigInt(value string) (big.Int, error) {
	i := big.Int{}
	_, err := fmt.Sscan(value, &i)

	return i, err
}

// IntToHex convert int to hexadecimal representation
func IntToHex(i int) string {
	return fmt.Sprintf("0x%x", i)
}

// BigToHex covert big.Int to hexadecimal representation
func BigToHex(bigInt big.Int) string {
	if bigInt.BitLen() == 0 {
		return "0x0"
	}

	return "0x" + strings.TrimPrefix(fmt.Sprintf("%x", bigInt.Bytes()), "0")
}

func TxToRlp(tx *types.Transaction) string {
	var buff bytes.Buffer
	tx.EncodeRLP(&buff)
	return fmt.Sprintf("%x", buff.Bytes())
}

func AuthorizationHeader(accountId string, secretHash string) string {
	return base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", accountId, secretHash)))
}
