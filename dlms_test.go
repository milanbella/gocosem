package gocosem

import (
	"testing"
)

func TestX_encode_GetRequestNormal(t *testing.T) {
	b := []byte{
		0xC0, 0x01, 0x81,
		0x00, 0x01, 0x00, 0x00, 0x80, 0x00, 0x00, 0xFF, 0x02, 0x00}

	err, pdu := encode_GetRequestNormal(0x81, 1, &tDlmsOid{0, 0, 128, 0, 0, 255}, 2, nil, nil)
	if nil != err {
		t.Fatalf("encode_GetRequestNormal() failed, err: %v", err)
	}

	printBuffer(t, pdu)

	if !byteEquals(t, pdu, b, true) {
		t.Fatalf("bytes don't match")
	}

}

func TestX_decode_GetResponseNormal(t *testing.T) {
	pdu := []byte{
		0xC4, 0x01, 0x81,
		0x00,
		0x09, 0x06,
		0x11, 0x22, 0x33, 0x44, 0x55, 0x66}

	//func decode_GetResponsenormal(pdu []byte) (err error, invokeIdAndPriority tDlmsInvokeIdAndPriority, dataAccessResult tDlmsDataAccessResult, data *tDlmsData) {
	err, invokeIdAndPriority, dataAccessResult, data := decode_GetResponseNormal(pdu)
	if nil != err {
		t.Fatalf("decode_GetResponseNormal() failed, err %v", err)
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
		t.Fatalf("invokeIdAndPriority wrong: %02X", invokeIdAndPriority)
	}
	if 0 != dataAccessResult {
		t.Fatalf("dataAccessResult wrong: %d", dataAccessResult)
	}
	if nil == data {
		t.Fatalf("data is nil")
	}
	if 9 != tag {
		t.Fatalf("data.tag wrong: %d", tag)
	}
	if !byteEquals(t, val, []byte{0x11, 0x22, 0x33, 0x44, 0x55, 0x66}, true) {
		t.Fatalf("bytes don't match")
	}
}

func TestX_encode_GetRequestWithList(t *testing.T) {
	//func encode_GetRequestWithList(invokeIdAndPriority tDlmsInvokeIdAndPriority, classIds []tDlmsClassId, instanceIds []*tDlmsOid, attributeIds []tDlmsAttributeId, accessSelectors []*tDlmsAccessSelector, accessParameters []*tDlmsData) (pdu []byte, err error) {

	b := []byte{
		0xC0, 0x03, 0x81,
		0x02,
		0x00, 0x01, 0x00, 0x00, 0x80, 0x00, 0x00, 0xFF, 0x02, 0x00,
		0x00, 0x01, 0x00, 0x00, 0x80, 0x01, 0x00, 0xFF, 0x02, 0x00}

	err, pdu := encode_GetRequestWithList(0x81, []tDlmsClassId{1, 1}, []*tDlmsOid{&tDlmsOid{0, 0, 128, 0, 0, 255}, &tDlmsOid{0, 0, 128, 1, 0, 255}}, []tDlmsAttributeId{2, 2}, []*tDlmsAccessSelector{nil, nil}, []*tDlmsData{nil, nil})
	if nil != err {
		t.Fatalf("encode_GetRequestWithList() failed, err: %v", err)
	}

	printBuffer(t, pdu)

	if !byteEquals(t, pdu, b, true) {
		t.Fatalf("bytes don't match")
	}
}

func TestX_decode_GetResponseWithList(t *testing.T) {
	b := []byte{
		0xC4, 0x03, 0x81,
		0x02,
		0x00,
		0x09, 0x04,
		0x01, 0x02, 0x03, 0x04,
		0x00,
		0x0A, 0x03,
		0x30, 0x30, 0x30}

	err, invokeIdAndPriority, dataAccessResults, datas := decode_GetResponseWithList(b)
	if nil != err {
		t.Fatalf("decode_GetResponseWithList() failed, err: %v", err)
	}

	if 0x81 != invokeIdAndPriority {
		t.Fatalf("invokeIdAndPriority wrong: %02X", invokeIdAndPriority)
	}

	count := len(dataAccessResults)

	for i := 0; i < count; i += 1 {
		choice := (*tAsn1Choice)(datas[i])
		t.Logf("%d: dataAccessResult: %d tag: %d", i, dataAccessResults[i], choice.getTag())
	}

	if 0 != dataAccessResults[0] {
		t.Fatalf("wrong dataAccessResults[0]: ", dataAccessResults[0])
	}
	choice := (*tAsn1Choice)(datas[0])
	if C_Data_PR_octet_string != choice.getTag() {
		t.Fatalf("wrong tag[0]: ", choice.getTag())
	}
	db := *(choice.getVal().(*tAsn1OctetString))
	printBuffer(t, db)
	if !byteEquals(t, db, []byte{0x01, 0x02, 0x03, 0x04}, true) {
		t.Fatalf("wrong data[0]")
	}

	if 0 != dataAccessResults[1] {
		t.Fatalf("wrong dataAccessResults[1]: ", dataAccessResults[1])
	}
	choice = (*tAsn1Choice)(datas[1])
	if C_Data_PR_visible_string != choice.getTag() {
		t.Fatalf("wrong tag[1]: ", choice.getTag())
	}
	vs := *(choice.getVal().(*tAsn1VisibleString))
	printBuffer(t, vs)
	if !byteEquals(t, vs, []byte{0x30, 0x30, 0x30}, true) {
		t.Fatalf("wrong data[1]")
	}
}

func TestX_decode_GetResponsewithDataBlock(t *testing.T) {
	b := []byte{
		0xC4, 0x02, 0x81,
		0x00,
		0x00, 0x00, 0x00, 0x01,
		0x00,
		0x1E,
		0x09, 0x32, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x10, 0x11, 0x12, 0x13,
		0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x20, 0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27, 0x28}

	err, invokeIdAndPriority, lastBlock, blockNumber, dataAccessResult, rawData := decode_GetResponsewithDataBlock(b)
	if nil != err {
		t.Fatalf("decode_GetResponsewithDataBlock() failed, err %v", err)
	}

	t.Logf("invokeIdAndPriority: %02X", invokeIdAndPriority)
	t.Logf("lastBlock: %t", lastBlock)
	t.Logf("blockNumber: %d", blockNumber)
	t.Logf("dataAccessResult: %d", dataAccessResult)
	printBuffer(t, rawData)

	if 0x81 != invokeIdAndPriority {
		t.Fatalf("wrong invokeIdAndPriority")
	}
	if false != lastBlock {
		t.Fatalf("wrong lastBlock")
	}
	if 1 != blockNumber {
		t.Fatalf("wrong blockNumber")
	}
	if 0 != dataAccessResult {
		t.Fatalf("wrong dataAccessResult")
	}
	if !byteEquals(t, rawData, []byte{0x09, 0x32, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x20, 0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27, 0x28}, true) {
		t.Fatalf("wrong rawData")
	}
}
