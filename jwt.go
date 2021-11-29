package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-chi/jwtauth/v5"
)

func initJWT() error {
	jwtSignKeyPath = os.Getenv("JWT_SIGN_KEY")
	jwtVerifyKeyPath = os.Getenv("JWT_VERIFY_KEY")

	if jwtSignKeyPath == "" && jwtVerifyKeyPath == "" {
		return fmt.Errorf("JWT_SIGN_KEY must be to either a secret string, or the path of a RSA key. In the later case, JWT_VERIFY_KEY must also be set to the path of the RSA public key.")
	} else if strings.HasPrefix(jwtSignKeyPath, "/") {
		signBytes, err := ioutil.ReadFile(jwtSignKeyPath)
		if err != nil {
			return err
		}
		if signKey, err = jwt.ParseRSAPrivateKeyFromPEM(signBytes); err != nil {
			return err
		}
		if jwtVerifyKeyPath == "" {
			return fmt.Errorf("JWT_SIGN_KEY is set to the path of a RSA key. In this case, JWT_VERIFY_KEY must also be set to the path of the RSA public key.")
		}
		if verifyBytes, err = ioutil.ReadFile(jwtVerifyKeyPath); err != nil {
			return err
		} else {
			log.Printf("Verify key:\n%s", string(verifyBytes))
		}

		if verifyKey, err = jwt.ParseRSAPublicKeyFromPEM(verifyBytes); err != nil {
			return err
		}

		tokenAuth = jwtauth.New("RS256", signKey, verifyKey)
	} else {
		log.Printf("Using JWT HMAC signature. This is less secure than RS256 signature and verification. Set both JWT_SIGN_KEY and JWT_VERIFY_KEY to paths of a RSA key-pair.")
		tokenAuth = jwtauth.New("HMAC", signKey, nil)
	}
	return nil
}
