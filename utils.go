package main

import (
	"crypto/rand"
	"github.com/libp2p/go-libp2p-core/crypto"
	"io/ioutil"
	"os"
)

const dataPath = "data"
const privateKeyName = "private.pem"
const privateKeyPath = dataPath + "/" + privateKeyName

func generatePrivateKey() crypto.PrivKey {
	logger.Info("no private key found, generate a random one and save it")
	key, _, err := crypto.GenerateECDSAKeyPair(rand.Reader)
	if err != nil {
		panic(err)
	}
	pem, _ := crypto.MarshalPrivateKey(key)
	err = ioutil.WriteFile(privateKeyPath, pem, 400)
	if err != nil {
		panic(err)
	}
	return key
}

func GetPrivateKey() crypto.PrivKey {
	err := os.MkdirAll(dataPath, 0750)
	if err != nil {
		panic(err)
	}
	pem, err := ioutil.ReadFile(privateKeyPath)
	key, err := crypto.UnmarshalECDSAPrivateKey(pem)
	if err != nil {
		key = generatePrivateKey()
	}
	return key
}
