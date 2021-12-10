package w2pcrypt

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"strings"
)

var (
	digestAlgBySize = map[int]string{
		128 / 4: "md5",
		160 / 4: "sha1",
		224 / 4: "sha224",
		256 / 4: "sha256",
		384 / 4: "sha384",
		512 / 4: "sha512",
	}
)

type Crypt struct {
	hmacKey string
	hmacAlg string
}

func NewCrypt(hmacKey, hmacAlg string) *Crypt {
	return &Crypt{
		hmacKey: hmacKey,
		hmacAlg: hmacAlg,
	}
}

func guessAlg(s string) string {
	n := len(s)
	alg, ok := digestAlgBySize[n]
	if ok {
		return alg
	}
	return ""
}

func toMD5(b []byte) []byte {
	a := md5.Sum(b)
	return a[:]
}

func toHMACSHA512(secret, b []byte) []byte {
	h := hmac.New(sha512.New, secret)
	h.Write(b)
	return h.Sum(nil)
}

func toSHA512(b []byte) []byte {
	a := sha512.Sum512(b)
	return a[:]
}

func toSHA384(b []byte) []byte {
	a := sha512.Sum384(b)
	return a[:]
}

func toSHA256(b []byte) []byte {
	a := sha256.Sum256(b)
	return a[:]
}

func toSHA224(b []byte) []byte {
	a := sha256.Sum224(b)
	return a[:]
}

func toSHA1(b []byte) []byte {
	a := sha1.Sum(b)
	return a[:]
}

func (t Crypt) IsEqual(s, stored string) (bool, error) {
	var alg, salt, prefix, h string
	if stored == "" {
		return false, fmt.Errorf("empty password")
	}
	l := strings.SplitN(stored, "$", 3)
	switch len(l) {
	case 3:
		alg = l[0]
		salt = l[1]
		prefix = l[0] + "$" + l[1] + "$"
	case 2:
		alg = l[0]
		prefix = l[0] + "$"
	default:
		alg = guessAlg(l[0])
	}

	var digestBytes []byte
	if t.hmacKey != "" {
		secretBytes := []byte(t.hmacKey + salt)
		textBytes := []byte(s)
		switch t.hmacAlg {
		case "sha512":
			digestBytes = toHMACSHA512(secretBytes, textBytes)
		default:
			return false, fmt.Errorf("unsupported hmac alg %s", t.hmacAlg)
		}
	} else {
		text := s + salt
		textBytes := []byte(text)
		switch alg {
		case "sha512":
			digestBytes = toSHA512(textBytes)
		case "sha384":
			digestBytes = toSHA384(textBytes)
		case "sha256":
			digestBytes = toSHA256(textBytes)
		case "sha224":
			digestBytes = toSHA224(textBytes)
		case "sha1":
			digestBytes = toSHA1(textBytes)
		case "md5":
			digestBytes = toMD5(textBytes)
		default:
			return false, fmt.Errorf("unsupported digest alg %s", alg)
		}
	}
	h = prefix + hex.EncodeToString(digestBytes)
	return h == stored, nil
}
