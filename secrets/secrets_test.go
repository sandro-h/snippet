package secrets

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncryptDecrypt(t *testing.T) {
	cases := []struct {
		plain    string
		password string
	}{
		{"hello world", "password"},
		{"really long really long really long really long really long really long ", "password"},
		{"", "long password long password long password long password long password "},
		{"", "password"},
	}

	for _, c := range cases {
		cipher, err := Encrypt(c.plain, c.password)
		assert.Nil(t, err)
		assert.NotEqual(t, c.plain, cipher)
		dec, err := Decrypt(cipher, c.password)
		assert.Nil(t, err)
		assert.Equal(t, c.plain, dec)
	}
}
