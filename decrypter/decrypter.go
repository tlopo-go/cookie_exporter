package decrypter

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha1"
	"fmt"
	"github.com/tlopo-go/cookie_exporter/secrets"
	"golang.org/x/crypto/pbkdf2"
	"strings"
)

var creds secrets.Credentials

func unpad(in []byte) []byte {
	if len(in) == 0 {
		return nil
	}

	padding := in[len(in)-1]
	if int(padding) > len(in) || padding > aes.BlockSize {
		return nil
	} else if padding == 0 {
		return nil
	}

	for i := len(in) - 1; i > len(in)-int(padding)-1; i-- {
		if in[i] != padding {
			return nil
		}
	}
	return in[:len(in)-int(padding)]
}

func Decrypt(payload string) (result string, err error) {
	if (creds == secrets.Credentials{}) {
		creds, err = secrets.Get("Chrome Safe Storage")
		if err != nil {
			return
		}
	}

	SALT := []byte("saltysalt")
	IV := []byte(strings.Repeat(" ", 16))
	ITER := 1003
	LEN := 16

	var block cipher.Block

	key := pbkdf2.Key([]byte(creds.Password), SALT, ITER, LEN, sha1.New)

	block, err = aes.NewCipher(key)
	if err != nil {
		return
	}

	ciphertext := []byte(payload)
	ciphertext = ciphertext[3:]

	if len(ciphertext) < aes.BlockSize {
		fmt.Println("key too short")
	}

	cbc := cipher.NewCBCDecrypter(block, IV)

	cbc.CryptBlocks(ciphertext, ciphertext)

	result = strings.TrimSpace(string(unpad(ciphertext)))
	return
}
