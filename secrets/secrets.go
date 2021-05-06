package secrets

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"

	vault "github.com/sosedoff/ansible-vault-go"
)

// Encrypt encrypts the secret using the given password.
func Encrypt(secret string, password string) (string, error) {
	return ansibleVaultEncrypt(secret, password)
}

// Decrypt decrypts the cipher using the given master password.
func Decrypt(cipher string, password string) (string, error) {
	return ansibleVaultDecrypt(cipher, password)
}

func ansibleVaultEncrypt(secret string, password string) (string, error) {
	str, err := vault.Encrypt(secret, password)
	if err != nil {
		return "", err
	}
	parts := strings.SplitN(str, "\n", 2)
	bytes, err := hex.DecodeString(strings.ReplaceAll(parts[1], "\n", ""))
	if err != nil {
		return "", err
	}
	b64 := base64.StdEncoding.EncodeToString(bytes)
	return "AES256:" + b64, nil
}

func ansibleVaultDecrypt(cipher string, password string) (string, error) {
	if !strings.HasPrefix(cipher, "AES256:") {
		return "", fmt.Errorf("Cipher is missing AES256: prefix")
	}

	dec, err := base64.StdEncoding.DecodeString(cipher[7:])
	if err != nil {
		return "", err
	}

	hex := hex.EncodeToString(dec)
	str, err := vault.Decrypt("$ANSIBLE_VAULT;1.1;AES256\n"+hex, password)
	if err != nil {
		return "", err
	}
	return str, nil
}
