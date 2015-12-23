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

func TestX_decode_GetResponsenormal(t *testing.T) {
	pdu := []byte{
		0xC4, 0x01, 0x81,
		0x00,
		0x09, 0x06,
		0x11, 0x22, 0x33, 0x44, 0x55, 0x66}

	//func decode_GetResponsenormal(pdu []byte) (err error, invokeIdAndPriority tDlmsInvokeIdAndPriority, dataAccessResult tDlmsDataAccessResult, data *tDlmsData) {
	err, invokeIdAndPriority, dataAccessResult, data := decode_GetResponsenormal(pdu)
	if nil != err {
		t.Errorf("decode_GetResponsenormal() failed, err %v", err)
	}
	t.Logf("invokeIdAndPriority: %02X", invokeIdAndPriority)
	t.Logf("dataAccessResult: %d", dataAccessResult)
	t.Logf("data: %p", data)

	tag := ((*tAsn1Choice)(data)).getTag()
	val := *(((*tAsn1Choice)(data)).getVal()).(*tAsn1OctetString)
	if 0 == dataAccessResult {
		t.Logf("data.tag: %d", tag)
		printBuffer(t, val)
	}

	if 0x81 != invokeIdAndPriority {
		t.Errorf("invokeIdAndPriority wrong: %02X", invokeIdAndPriority)
	}
	if 0 != dataAccessResult {
		t.Errorf("dataAccessResult wrong: %d", dataAccessResult)
	}
	if nil == data {
		t.Errorf("data is nil")
	}
	if 9 != tag {
		t.Errorf("data.tag wrong: %d", tag)
	}
	if !byteEquals(t, val, []byte{0x11, 0x22, 0x33, 0x44, 0x55, 0x66}, true) {
		t.Errorf("bytes don't match")
	}
}
