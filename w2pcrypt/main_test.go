package w2pcrypt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKey(t *testing.T) {
	tests := []struct {
		Password     string
		HMACKey      string
		StoredDigest string
	}{
		{
			Password:     "fooooooo",
			HMACKey:      "sha512:7755f108-1b83-45dc-8302-54be8f3616a1",
			StoredDigest: "sha512$bb9ef128cf2844a5$85680adac9e462f8c0419bce00c803a0728bb9b1db20941cdf2f775cadf68606efaa7bf31cafc0349e429ab2d4fed3e58e526d03a2cb3c87ba6b78733f61ae48",
		},
	}
	for _, test := range tests {
		hmacKey, hmacAlg = parseHMACKey(test.HMACKey)
		if test.HMACKey != "" {
			t.Logf("hmac %s digest of %s is %s", test.HMACKey, test.Password, test.StoredDigest)
		} else {
			t.Logf("digest of %s is %s", test.Password, test.StoredDigest)
		}
		v, err := IsEqual(test.Password, test.StoredDigest)
		assert.NoError(t, err)
		assert.True(t, v)
	}
}
