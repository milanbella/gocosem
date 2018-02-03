package gocosem

import (
	"fmt"
	"gocosem/crypto/aes"
	"gocosem/crypto/cipher"
)

const GCM_TAG_LEN = 12

// Note: Use 'direction' 0 for encrypt, 'direction' 1 for decrypt.

func aesgcm(key []byte, IV []byte, adata []byte, pdu []byte, direction int) (err error, opdu []byte, tag []byte) {
	block, err := aes.NewCipher(key)
	if err != nil {
		errorLog("%s", err)
		return err, nil, nil
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		errorLog("%s", err)
		return err, nil, nil
	}

	if 0 == direction {
		opdu, tag = aesgcm.SealDlms(IV, pdu, adata) // encrypt
	} else if 1 == direction {
		opdu, tag = aesgcm.OpenDlms(IV, pdu, adata) // decrypt
	} else {
		err = fmt.Errorf("wrogn direction parameter: use 0 for encrypt, 1 for decrypt")
		errorLog("%s", err)
		return err, nil, nil
	}

	return nil, opdu, tag
}
