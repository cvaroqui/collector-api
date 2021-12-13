package main

import (
	"fmt"
	"io/ioutil"
	"log"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-chi/jwtauth/v5"
	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh"
)

func initJWT() error {
	log.Println("init token factory")
	jwtSignKey := viper.GetString("jwt.sign_key")
	jwtSignKeyFile = viper.GetString("jwt.sign_key_file")
	jwtVerifyKeyFile = viper.GetString("jwt.verify_key_file")

	if jwtSignKeyFile == "" && jwtVerifyKeyFile == "" && jwtSignKey == "" {
		return fmt.Errorf("API_JWT_SIGN_KEY or API_JWT_SIGN_KEY + API_JWT_VERIFY_KEY must be set.")
	} else if jwtSignKeyFile != "" {
		signBytes, err := ioutil.ReadFile(jwtSignKeyFile)
		if err != nil {
			return err
		}
		if signKey, err = jwt.ParseRSAPrivateKeyFromPEM(signBytes); err != nil {
			return err
		}
		if jwtVerifyKeyFile == "" {
			return fmt.Errorf("API_JWT_SIGN_KEY is set to the path of a RSA key. In this case, API_JWT_VERIFY_KEY must also be set to the path of the RSA public key.")
		}
		if verifyBytes, err = ioutil.ReadFile(jwtVerifyKeyFile); err != nil {
			return err
		}
		if verifyKey, err = jwt.ParseRSAPublicKeyFromPEM(verifyBytes); err != nil {
			return err
		} else {
			if pk, err := ssh.NewPublicKey(verifyKey); err != nil {
				log.Printf("  load verify key: %s", err)
			} else {
				finger := ssh.FingerprintLegacyMD5(pk)
				log.Printf("  verify key sig: %s", finger)
			}
			tokenAuth = jwtauth.New("RS256", signKey, verifyKey)
		}
	} else {
		log.Printf("Using JWT HMAC signature. This is less secure than RS256 signature and verification. Set both JWT_SIGN_KEY and JWT_VERIFY_KEY to paths of a RSA key-pair.")
		tokenAuth = jwtauth.New("HMAC", []byte(jwtSignKey), nil)
	}
	return nil
}
