package utils

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

type Account struct {
	Address string
	Private string
}

// GenerateEthAccount 生成account
func GenerateEthAccount() (*Account, error) {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		return nil, err
	}
	// 获取hex类型的私钥
	privateToBytes := crypto.FromECDSA(privateKey)
	private := common.ToHex(privateToBytes)

	address := crypto.PubkeyToAddress(privateKey.PublicKey)

	return &Account{
		Address: address.String(),
		Private: private,
	}, nil
}

// EncryptPrivate 私钥加密
func EncryptPrivate(private string) (string, error) {
	data := common.FromHex(private)
	result, err := AesEncrypt(data, []byte(CIPHER))
	if err != nil {
		return "", err
	}
	return common.ToHex(result), nil
}

// DecryptPrivate 私钥解密
func DecryptPrivate(val string) (string, error) {
	data := common.FromHex(val)
	result, err := AesDecrypt(data, []byte(CIPHER))
	if err != nil {
		return "", err
	}
	return common.ToHex(result), nil
}
