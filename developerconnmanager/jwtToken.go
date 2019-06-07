package developerconnmanager

import (
	"code.byted.org/gopkg/logs"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"github.com/dgrijalva/jwt-go"
	"io/ioutil"
	"os"
	"time"
	"errors"
)
var (
	ErrAuthKeyNotPem   = errors.New("token: AuthKey must be a valid .p8 PEM file")
	ErrAuthKeyNotECDSA = errors.New("token: AuthKey must be of type ecdsa.PrivateKey")
)


func AuthKeyFromFile(filename string) (*ecdsa.PrivateKey, error) {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return AuthKeyFromBytes(bytes)
}

func AuthKeyFromBytes(bytes []byte) (*ecdsa.PrivateKey, error) {
	block, _ := pem.Decode(bytes)
	if block == nil {
		return nil, ErrAuthKeyNotPem
	}
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	switch pk := key.(type) {
	case *ecdsa.PrivateKey:
		return pk, nil
	default:
		return nil, ErrAuthKeyNotECDSA
	}
}

func GetTokenString() string{
	pre_dir, _ := os.Getwd()
	p8filepath := pre_dir + "/developerconnmanager/AuthKey_52M3TTTA75.p8"
	authKey, error := AuthKeyFromFile(p8filepath)
	if error != nil{
		logs.Info("读取authKey失败")
	}
	token := jwt.New(jwt.SigningMethodES256)
	claims := make(jwt.MapClaims)
	claims["exp"] = time.Now().Add(5 * time.Minute).Unix()
	claims["iss"] = "69a6de75-923e-47e3-e053-5b8c7c11a4d1"
	claims["aud"] = "appstoreconnect-v1"
	token.Claims = claims
	token.Header["kid"] = "52M3TTTA75"
	tokenString, err := token.SignedString(authKey)
	if err != nil{
		logs.Info("签token失败")
	}
	return tokenString
}