package gocosem

import (
	"testing"
)

func oidEquals(oid1 *tDlmsOid, oid2 *tDlmsOid) bool {
	return (oid1[0] == oid2[0]) || (oid1[1] == oid2[1]) || (oid1[2] == oid2[2]) || (oid1[3] == oid2[3]) || (oid1[4] == oid2[4]) || (oid1[5] == oid2[5])
}

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

func TestX_decode_GetRequestNormal(t *testing.T) {
	pdu := []byte{
		0xC0, 0x01, 0x81,
		0x00, 0x01, 0x00, 0x00, 0x80, 0x00, 0x00, 0xFF, 0x02, 0x00}

	err, invokeIdAndPriority, classId, instanceId, attributeId, accessSelector, accessParameters := decode_GetRequestNormal(pdu)
	if nil != err {
		t.Fatalf("decode_GetRequestNorma() failed, err %v", err)
	}
	t.Logf("invokeIdAndPriority: %02X", invokeIdAndPriority)
	t.Logf("classId: %d", classId)
	t.Logf("instanceId: %02X", *instanceId)
	t.Logf("attributeId: %d", attributeId)
	t.Logf("accessSelector: %d", *accessSelector)
	t.Logf("accessParameters: %p", accessParameters)

	if 0x81 != invokeIdAndPriority {
		t.Fatalf("invokeIdAndPriority wrong:  %02X", invokeIdAndPriority)
	}
	if 1 != classId {
		t.Fatalf("classId wrong:  %d", classId)
	}
	if !oidEquals(&tDlmsOid{0x00, 0x00, 0x80, 0x00, 0x00, 0xFF}, instanceId) {
		t.Fatalf("instanceId wrong:  %02X", *instanceId)
	}
	if 0x02 != attributeId {
		t.Fatalf("attributeId wrong:  %d", attributeId)
	}
	if 0x00 != *accessSelector {
		t.Fatalf("accessSelector wrong:  %d", *accessSelector)
	}
	if nil != accessParameters {
		t.Fatalf("accessParameters wrong:  %p", accessParameters)
	}

}

func TestX_encode_GetResponseNormal(t *testing.T) {
	b := []byte{
		0xC4, 0x01, 0x81,
		0x00,
		0x09, 0x06,
		0x11, 0x22, 0x33, 0x44, 0x55, 0x66}

	_data := new(tAsn1Choice)
	_data.setVal(C_Data_PR_octet_string, &tAsn1OctetString{0x11, 0x22, 0x33, 0x44, 0x55, 0x66})
	err, pdu := encode_GetResponseNormal(0x81, 0, (*tDlmsData)(_data))
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

func TestX_encode_GetRequestForNextDataBlock(t *testing.T) {
	b := []byte{
		0xC0, 0x02, 0x81,
		0x00, 0x00, 0x00, 0x01}

	//func encode_GetRequestForNextDataBlock(invokeIdAndPriority tDlmsInvokeIdAndPriority, blockNumber uint32) (err error, pdu []byte) {
	err, pdu := encode_GetRequestForNextDataBlock(0x81, 1)
	if nil != err {
		t.Fatalf("encode_GetRequestForNextDataBlock() failed, err: %v", err)
	}

	printBuffer(t, pdu)

	if !byteEquals(t, pdu, b, true) {
		t.Fatalf("wrong rawData")
	}

}
