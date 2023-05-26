package chain

import (
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/samirshao/itools/ifile"
	"github.com/samirshao/itools/ilog"
	"os"
	"path/filepath"
)

type Evm struct {
	//私钥
	PriKey *ecdsa.PrivateKey
	//公钥
	PubKey *ecdsa.PublicKey
}

// Wallet 生成钱包
func (_this *Evm) Wallet() error {
	fmt.Print("输入钱包密码：")
	var password string
	if _, err := fmt.Scan(&password); err != nil {
		ilog.Logger.Error(err)
		return err
	}

	fmt.Print("确认钱包密码：")
	var repassword string
	if _, err := fmt.Scan(&repassword); err != nil {
		ilog.Logger.Error(err)
		return err
	}

	if password != repassword {
		fmt.Println("密码校验错误")
		return errors.New("密码校验错误")
	}

	ks := keystore.NewKeyStore("./keystore", keystore.StandardScryptN, keystore.StandardScryptP)
	account, err := ks.NewAccount(password)
	if err != nil {
		ilog.Logger.Error(err)
		return err
	}

	fmt.Println("路径：", account.URL.String())
	fmt.Println("地址：", account.Address.Hex())
	fmt.Println("密码：", password)
	return nil
}

func (_this *Evm) Stake() error {
	return nil
}

// DecKeystore 解析钱包
// path 钱包存放路径
// password 钱包密码
func (_this *Evm) DecKeystore(path, password string) error {
	if !ifile.IsExist(path) {
		ilog.Logger.Error("can not find keystore")
		return errors.New("can not find keystore")
	}

	fs, _ := filepath.Glob(path + "/UTC--*")
	if len(fs) == 0 {
		ilog.Logger.Error("can not find keystore file")
		return errors.New("can not find keystore file")
	}

	ksData, err := os.ReadFile(fs[0])
	if err != nil {
		ilog.Logger.Error("read keystore err >>>>> ", err)
		return err
	}

	var key *keystore.Key
	key, err = keystore.DecryptKey(ksData, password)
	if err != nil {
		ilog.Logger.Error("can not decrypt keystore >>>>> ", err)
		return err
	}

	//私钥
	_this.PriKey = key.PrivateKey
	//公钥
	pubKey, ok := key.PrivateKey.Public().(*ecdsa.PublicKey)
	if !ok {
		ilog.Logger.Error("public key is not ecdsa.publickey")
		return errors.New("public key is not ecdsa.publickey")
	}
	_this.PubKey = pubKey

	return nil
}

// Sign 数据签名
func (_this *Evm) Sign(data []byte) (signature []byte, err error) {
	hash := crypto.Keccak256Hash(data)
	signature, err = crypto.Sign(hash.Bytes(), _this.PriKey)
	if err != nil {
		ilog.Logger.Error("签名错误", err)
		return
	}
	return
}

// VerifySign 验签
func (_this *Evm) VerifySign(data, signature []byte) (pubKey string, err error) {
	var edsa *ecdsa.PublicKey
	edsa, err = crypto.SigToPub(crypto.Keccak256Hash(data).Bytes(), signature)
	if err != nil {
		ilog.Logger.Error(err)
		return
	}
	pubKey = crypto.PubkeyToAddress(*edsa).Hex()
	return
}

// 根据钱包查询节点类型
// 0=unkown 1=distributor 2=miner 3=validator
func (_this *Evm) NodeKind(wallet string) (kind int, err error) {
	distributor := map[string]int{
		"0xb2a1a91eA058D7Cd180234E3046E7CB467eF5D26": 0,
		"0xFD0e4dcAd3F1eF2227176c8DE4afAB7a4066886d": 0,
		"0x47b87166eF8Ce8E3350139BDBC8260D7165Fac9C": 0,
		"0xfbE0DF97f183dfF5EC5d95E850d282a3611c63A7": 0,
		"0xB60663C68aFAEd01eE7fE28fa2e3FEd0C0dbe189": 0,
		"0x520c59f2EbD1D0C61998B18C1B14923995749Fab": 0,
	}
	miner := map[string]int{}
	if _, ok := distributor[wallet]; ok {
		kind = 1
	} else if _, ok = miner[wallet]; ok {
		kind = 2
	} else {
		kind = 3
	}
	return
}

// 读取钱包地址
func (_this *Evm) GetAddress() string {
	return crypto.PubkeyToAddress(*_this.PubKey).Hex()
}

func LoadKeystore() {
	KeystorePwdPath := "./keystore/password.txt"
	KeystoreFilePath := "./keystore/UTC--*"

	if !ifile.IsExist(KeystorePwdPath) {
		ilog.Logger.Fatalln("No wallet detected")
		return
	}

	fs, _ := filepath.Glob(KeystoreFilePath)
	if len(fs) == 0 {
		ilog.Logger.Fatalln("No wallet detected")
		return
	}

	ksData, err := os.ReadFile(fs[0])
	if err != nil {
		ilog.Logger.Fatalln("read keystore err >>>>> ", err)
		return
	}

	ksPwd, err := os.ReadFile(KeystorePwdPath)
	if err != nil {
		ilog.Logger.Fatalln("read keystore password err >>>>> ", err)
		return
	}

	var fromKey *keystore.Key
	fromKey, err = keystore.DecryptKey(ksData, string(ksPwd))
	if err != nil {
		ilog.Logger.Fatalln("can not decrypt keystore >>>>> ", err)
		return
	}

	WalletAddr := fromKey.Address.String()
	fmt.Println("钱包地址：", WalletAddr)

	//私钥
	prikey := fromKey.PrivateKey

	//公钥
	pubkey := fromKey.PrivateKey.Public()
	ecdsaPubkey, ok := pubkey.(*ecdsa.PublicKey)
	if !ok {
		ilog.Logger.Error("公钥错误")
		return
	}
	//输出公钥字符串，和钱包地址一致
	address := crypto.PubkeyToAddress(*ecdsaPubkey).Hex()
	fmt.Println("公钥明文", address)

	//私钥签名
	data := []byte("Hello World!Hello World!Hello World!Hello World!Hello World!Hello World!")
	hash := crypto.Keccak256Hash(data)
	signature, err := crypto.Sign(hash.Bytes(), prikey)
	if err != nil {
		ilog.Logger.Error("签名错误", err)
		return
	}
	encSign := hex.EncodeToString(signature)
	fmt.Println("签名", encSign)

	decSign, err := hex.DecodeString(encSign)
	if err != nil {
		ilog.Logger.Error("签名解密错误", err)
		return
	}

	//公钥提取签名
	pubKey, err := crypto.SigToPub(hash.Bytes(), decSign)
	if err != nil {
		ilog.Logger.Error("签名错误", err)
		return
	}
	address = crypto.PubkeyToAddress(*pubKey).Hex()
	fmt.Println("验签公钥", address)

}
