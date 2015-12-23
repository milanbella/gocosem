package gocosem

import (
	"testing"
)

func TestX_encode_GetRequestNormal(t *testing.T) {
	b := []byte{
		0xC0, 0x01, 0x81,
		0x00, 0x01, 0x00, 0x00, 0x80, 0x00, 0x00, 0xFF, 0x02, 0x00}

	pdu, err := encode_GetRequestNormal(0x81, 1, &tDlmsOid{0, 0, 128, 0, 0, 255}, 2, nil, nil)
	if nil != err {
		t.Errorf("encode_GetRequestNormal() failed, err: %v", err)
	}

	printBuffer(t, pdu)

	if !byteEquals(t, pdu, b, true) {
		t.Errorf("bytes don't match")
	}

}
