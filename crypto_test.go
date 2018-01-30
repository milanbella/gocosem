package gocosem

import (
	"gocosem/crypto/aes"
	"gocosem/crypto/cipher"
	"testing"
)

func TestCrypto_aesgcm(t *testing.T) {
	IV := []byte{0x00, 0x00, 0x00, 0x10, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
	ADD := []byte{0x30, 0xD0, 0xD1, 0xD2, 0xD3, 0xD4, 0xD5, 0xD6, 0xD7, 0xD8, 0xD9, 0xDA, 0xDB, 0xDC, 0xDD, 0xDE, 0xDF}
	EK := []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F}
	plaintext := []byte("1234567890123456789")

	err, ciphertext, tag := aesgcm(EK, IV, ADD, ([]byte)(plaintext), 0)
	if nil != err {
		t.Fatal(err)
	}

	t.Logf("plaintext: %s", string(plaintext))
	t.Logf("ciphertext: % 0X", ciphertext)
	t.Logf("tag: % 0X", tag)

	plaintext = []byte("0000000000000000000")
	tag = []byte{}
	err, plaintext, tag = aesgcm(EK, IV, ADD, ([]byte)(ciphertext), 0)
	if nil != err {
		t.Fatal(err)
	}

	t.Logf("plaintext: %s", string(plaintext))
	t.Logf("ciphertext: % 0X", ciphertext)
	t.Logf("tag: % 0X", tag)
}

func TestCrypto_goAesgcm(t *testing.T) {
	IV := []byte{0x00, 0x00, 0x00, 0x10, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
	ADD := []byte{0x30, 0xD0, 0xD1, 0xD2, 0xD3, 0xD4, 0xD5, 0xD6, 0xD7, 0xD8, 0xD9, 0xDA, 0xDB, 0xDC, 0xDD, 0xDE, 0xDF}
	EK := []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F}
	plaintext := []byte("1234567890123456789")

	block, err := aes.NewCipher(EK)
	if err != nil {
		panic(err.Error())
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}

	ciphertext, tag := aesgcm.SealDlms(IV, plaintext, ADD)
	t.Logf("plaintext: %s", string(plaintext))
	t.Logf("ciphertext: % 0X", ciphertext)
	t.Logf("tag: % 0X", tag)

	plaintext, tag = aesgcm.OpenDlms(IV, ciphertext, ADD)
	t.Logf("plaintext: %s", string(plaintext))
	t.Logf("ciphertext: % 0X", ciphertext)
	t.Logf("tag: % 0X", tag)
}

func TestCrypto_goAesgcm_noplaintext(t *testing.T) {
	IV := []byte{0x00, 0x00, 0x00, 0x10, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
	ADD := []byte{0x30, 0xD0, 0xD1, 0xD2, 0xD3, 0xD4, 0xD5, 0xD6, 0xD7, 0xD8, 0xD9, 0xDA, 0xDB, 0xDC, 0xDD, 0xDE, 0xDF}
	EK := []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F}
	plaintext := []byte{}

	block, err := aes.NewCipher(EK)
	if err != nil {
		panic(err.Error())
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}

	ciphertext, tag := aesgcm.SealDlms(IV, plaintext, ADD)
	t.Logf("plaintext: %s", string(plaintext))
	t.Logf("ciphertext: % 0X", ciphertext)
	t.Logf("tag: % 0X", tag)

	plaintext, tag = aesgcm.OpenDlms(IV, ciphertext, ADD)
	t.Logf("plaintext: %s", string(plaintext))
	t.Logf("ciphertext: % 0X", ciphertext)
	t.Logf("tag: % 0X", tag)
}
