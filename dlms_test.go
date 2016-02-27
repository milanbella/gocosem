package gocosem

import (
	"bytes"
	"testing"
)

func oidEquals(oid1 *DlmsOid, oid2 *DlmsOid) bool {
	return (oid1[0] == oid2[0]) || (oid1[1] == oid2[1]) || (oid1[2] == oid2[2]) || (oid1[3] == oid2[3]) || (oid1[4] == oid2[4]) || (oid1[5] == oid2[5])
}

func TestX_encode_decode_DlmsData_array(t *testing.T) {

	data := new(DlmsData)
	data.Typ = DATA_TYPE_ARRAY
	data.Arr = make([]*DlmsData, 21)

	i := 0
	d := new(DlmsData)
	d.SetNULL()
	data.Arr[i] = d

	i = 1
	d = new(DlmsData)
	d.SetBoolean(true)
	data.Arr[i] = d

	i = 2
	d = new(DlmsData)
	d.SetBitString([]byte{0xFF, 0x80}, 9)
	data.Arr[i] = d

	i = 3
	d = new(DlmsData)
	d.SetDoubleLong(0x11223344)
	data.Arr[i] = d

	i = 4
	d = new(DlmsData)
	d.SetDoubleLongUnsigned(0x11223344)
	data.Arr[i] = d

	i = 5
	d = new(DlmsData)
	d.SetFloatingPoint(0.25)
	data.Arr[i] = d

	i = 6
	d = new(DlmsData)
	d.SetOctetString([]byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09})
	data.Arr[i] = d

	i = 7
	d = new(DlmsData)
	d.SetVisibleString([]byte{0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19})
	data.Arr[i] = d

	i = 8
	d = new(DlmsData)
	d.SetBcd(5)
	data.Arr[i] = d

	i = 9
	d = new(DlmsData)
	d.SetInteger(-11)
	data.Arr[i] = d

	i = 10
	d = new(DlmsData)
	d.SetLong(-23457)
	data.Arr[i] = d

	i = 11
	d = new(DlmsData)
	d.SetUnsigned(254)
	data.Arr[i] = d

	i = 12
	d = new(DlmsData)
	d.SetLong64(-1254999)
	data.Arr[i] = d

	i = 13
	d = new(DlmsData)
	d.SetUnsignedLong64(91254999)
	data.Arr[i] = d

	i = 14
	d = new(DlmsData)
	d.SetEnum(0x70)
	data.Arr[i] = d

	i = 15
	d = new(DlmsData)
	d.SetReal32(100.57)
	data.Arr[i] = d

	i = 16
	d = new(DlmsData)
	d.SetReal64(1105.9)
	data.Arr[i] = d

	i = 17
	d = new(DlmsData)
	d.SetDateTime([]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x08, 0x10, 0x11, 0x12})
	data.Arr[i] = d

	i = 18
	d = new(DlmsData)
	d.SetDate([]byte{0x01, 0x02, 0x03, 0x04, 0x05})
	data.Arr[i] = d

	i = 19
	d = new(DlmsData)
	d.SetTime([]byte{0x01, 0x02, 0x03, 0x04})
	data.Arr[i] = d

	i = 20
	d = new(DlmsData)
	d.Typ = DATA_TYPE_ARRAY
	d.Arr = make([]*DlmsData, 2)
	data.Arr[20] = d

	i = 0
	d = new(DlmsData)
	d.SetOctetString([]byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05})
	data.Arr[20].Arr[i] = d

	i = 1
	d = new(DlmsData)
	d.SetLong64(-9254799)
	data.Arr[20].Arr[i] = d

	var ebuf bytes.Buffer

	err := data.Encode(&ebuf)
	if nil != err {
		t.Fatalf("DlmsData.Encode() failed, err: %v", err)
	}

	dbuf := bytes.NewBuffer(ebuf.Bytes())
	ddata := new(DlmsData)
	err = ddata.Decode(dbuf)
	if nil != err {
		t.Fatalf("DlmsData.Decode() failed, err: %v", err)
	}

	if DATA_TYPE_ARRAY != ddata.GetType() {
		t.Fatalf("failed")
	}

	if DATA_TYPE_NULL != ddata.Arr[0].Typ {
		t.Fatalf("failed: decoded wrong type")
	}

	if DATA_TYPE_BOOLEAN != ddata.Arr[1].Typ {
		t.Fatalf("failed: decoded wrong type")
	}
	if true != ddata.Arr[1].GetBoolean() {
		t.Fatalf("failed: decoded wrong value")
	}

	if DATA_TYPE_BIT_STRING != ddata.Arr[2].Typ {
		t.Fatalf("failed: decoded wrong type")
	}
	bitString, bitStringLen := ddata.Arr[2].GetBitString()
	if !bytes.Equal(bitString, []byte{0xFF, 0x80}) {
		t.Fatalf("failed: decoded wrong value")
	}
	if 9 != bitStringLen {
		t.Fatalf("failed: decoded wrong value")
	}

	if DATA_TYPE_DOUBLE_LONG != ddata.Arr[3].Typ {
		t.Fatalf("failed: decoded wrong type")
	}
	if 0x11223344 != ddata.Arr[3].GetDoubleLong() {
		t.Fatalf("failed: decoded wrong value")
	}

	if DATA_TYPE_DOUBLE_LONG_UNSIGNED != ddata.Arr[4].Typ {
		t.Fatalf("failed: decoded wrong type")
	}
	if 0x11223344 != ddata.Arr[4].GetDoubleLongUnsigned() {
		t.Fatalf("failed: decoded wrong value")
	}

	if DATA_TYPE_FLOATING_POINT != ddata.Arr[5].Typ {
		t.Fatalf("failed: decoded wrong type")
	}
	if 0.25 != ddata.Arr[5].GetFloatingPoint() {
		t.Fatalf("failed: decoded wrong value")
	}

	if DATA_TYPE_OCTET_STRING != ddata.Arr[6].Typ {
		t.Fatalf("failed: decoded wrong type")
	}
	if !bytes.Equal(ddata.Arr[6].GetOctetString(), []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09}) {
		t.Fatalf("failed: decoded wrong value")
	}

	if DATA_TYPE_VISIBLE_STRING != ddata.Arr[7].Typ {
		t.Fatalf("failed: decoded wrong type")
	}
	if !bytes.Equal(ddata.Arr[7].GetVisibleString(), []byte{0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19}) {
		t.Fatalf("failed: decoded wrong value")
	}

	if DATA_TYPE_BCD != ddata.Arr[8].Typ {
		t.Fatalf("failed: decoded wrong type")
	}
	if 5 != ddata.Arr[8].GetBcd() {
		t.Fatalf("failed: decoded wrong value")
	}

	if DATA_TYPE_INTEGER != ddata.Arr[9].Typ {
		t.Fatalf("failed: decoded wrong type")
	}
	if -11 != ddata.Arr[9].GetInteger() {
		t.Fatalf("failed: decoded wrong value")
	}

	if DATA_TYPE_LONG != ddata.Arr[10].Typ {
		t.Fatalf("failed: decoded wrong type")
	}
	if -23457 != ddata.Arr[10].GetLong() {
		t.Fatalf("failed: decoded wrong value")
	}

	if DATA_TYPE_UNSIGNED != ddata.Arr[11].Typ {
		t.Fatalf("failed: decoded wrong type")
	}
	if 254 != ddata.Arr[11].GetUnsigned() {
		t.Fatalf("failed: decoded wrong value")
	}

	if DATA_TYPE_LONG64 != ddata.Arr[12].Typ {
		t.Fatalf("failed: decoded wrong type")
	}
	if -1254999 != ddata.Arr[12].GetLong64() {
		t.Fatalf("failed: decoded wrong value")
	}

	if DATA_TYPE_UNSIGNED_LONG64 != ddata.Arr[13].Typ {
		t.Fatalf("failed: decoded wrong type")
	}
	if 91254999 != ddata.Arr[13].GetUnsignedLong64() {
		t.Fatalf("failed: decoded wrong value")
	}

	if DATA_TYPE_ENUM != ddata.Arr[14].Typ {
		t.Fatalf("failed: decoded wrong type")
	}
	if 0x70 != ddata.Arr[14].GetEnum() {
		t.Fatalf("failed: decoded wrong value")
	}

	if DATA_TYPE_REAL32 != ddata.Arr[15].Typ {
		t.Fatalf("failed: decoded wrong type")
	}
	if 100.57 != ddata.Arr[15].GetReal32() {
		t.Fatalf("failed: decoded wrong value")
	}

	if DATA_TYPE_REAL64 != ddata.Arr[16].Typ {
		t.Fatalf("failed: decoded wrong type")
	}
	if 1105.9 != ddata.Arr[16].GetReal64() {
		t.Fatalf("failed: decoded wrong value")
	}

	if DATA_TYPE_DATETIME != ddata.Arr[17].Typ {
		t.Fatalf("failed: decoded wrong type")
	}
	if !bytes.Equal(ddata.Arr[17].GetDateTime(), []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x08, 0x10, 0x11, 0x12}) {
		t.Fatalf("failed: decoded wrong value")
	}

	if DATA_TYPE_DATE != ddata.Arr[18].Typ {
		t.Fatalf("failed: decoded wrong type")
	}
	if !bytes.Equal(ddata.Arr[18].GetDate(), []byte{0x01, 0x02, 0x03, 0x04, 0x05}) {
		t.Fatalf("failed: decoded wrong value")
	}

	if DATA_TYPE_TIME != ddata.Arr[19].Typ {
		t.Fatalf("failed: decoded wrong type")
	}
	if !bytes.Equal(ddata.Arr[19].GetTime(), []byte{0x01, 0x02, 0x03, 0x04}) {
		t.Fatalf("failed: decoded wrong value")
	}

	if DATA_TYPE_ARRAY != ddata.Arr[20].Typ {
		t.Fatalf("failed: decoded wrong type")
	}

	if DATA_TYPE_OCTET_STRING != ddata.Arr[20].Arr[0].Typ {
		t.Fatalf("failed: decoded wrong type")
	}
	if !bytes.Equal(ddata.Arr[20].Arr[0].GetOctetString(), []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05}) {
		t.Fatalf("failed: decoded wrong value")
	}

	if DATA_TYPE_LONG64 != ddata.Arr[20].Arr[1].Typ {
		t.Fatalf("failed: decoded wrong type")
	}
	if -9254799 != ddata.Arr[20].Arr[1].GetLong64() {
		t.Fatalf("failed: decoded wrong value")
	}

}

func TestX_encode_GetRequestNormal(t *testing.T) {
	b := []byte{
		0x00, 0x01, 0x00, 0x00, 0x80, 0x00, 0x00, 0xFF, 0x02, 0x00}

	var buf bytes.Buffer
	err := encode_GetRequestNormal(&buf, 1, &DlmsOid{0, 0, 128, 0, 0, 255}, 2, 0, nil)
	if nil != err {
		t.Fatalf("encode_GetRequestNormal() failed, err: %v", err)
	}

	if !bytes.Equal(buf.Bytes(), b) {
		t.Fatalf("bytes don't match")
	}

}

func TestX_decode_GetRequestNormal(t *testing.T) {
	pdu := []byte{
		0x00, 0x01, 0x00, 0x00, 0x80, 0x00, 0x00, 0xFF, 0x02, 0x00}
	buf := bytes.NewBuffer(pdu)

	err, classId, instanceId, attributeId, accessSelector, accessParameters := decode_GetRequestNormal(buf)
	if nil != err {
		t.Fatalf("decode_GetRequestNorma() failed, err %v", err)
	}

	if 1 != classId {
		t.Fatalf("classId wrong:  %d", classId)
	}
	if !oidEquals(&DlmsOid{0x00, 0x00, 0x80, 0x00, 0x00, 0xFF}, instanceId) {
		t.Fatalf("instanceId wrong:  %02X", *instanceId)
	}
	if 0x02 != attributeId {
		t.Fatalf("attributeId wrong:  %d", attributeId)
	}
	if 0x00 != accessSelector {
		t.Fatalf("accessSelector wrong:  %d", accessSelector)
	}
	if nil != accessParameters {
		t.Fatalf("accessParameters wrong:  %p", accessParameters)
	}

}

func TestX_encode_GetResponseNormal(t *testing.T) {
	b := []byte{
		0x00,
		0x09, 0x06,
		0x11, 0x22, 0x33, 0x44, 0x55, 0x66}

	data := new(DlmsData)
	data.SetOctetString([]byte{0x11, 0x22, 0x33, 0x44, 0x55, 0x66})

	var buf bytes.Buffer
	err := encode_GetResponseNormal(&buf, 0, data)
	if nil != err {
		t.Fatalf("encode_GetRequestNormal() failed, err: %v", err)
	}

	if !bytes.Equal(buf.Bytes(), b) {
		t.Fatalf("bytes don't match")
	}
}

func TestX_encode_GetResponseNormalBlock(t *testing.T) {
	b := []byte{
		0x09, 0x06,
		0x11, 0x22, 0x33, 0x44, 0x55, 0x66}

	data := new(DlmsData)
	data.SetOctetString([]byte{0x11, 0x22, 0x33, 0x44, 0x55, 0x66})

	var buf bytes.Buffer
	err := encode_GetResponseNormalBlock(&buf, data)
	if nil != err {
		t.Fatalf("encode_GetRequestNormalBlock() failed, err: %v", err)
	}

	if !bytes.Equal(buf.Bytes(), b) {
		t.Fatalf("bytes don't match")
	}
}

func TestX_decode_GetResponseNormal(t *testing.T) {
	pdu := []byte{
		0x00,
		0x09, 0x06,
		0x11, 0x22, 0x33, 0x44, 0x55, 0x66}
	buf := bytes.NewBuffer(pdu)

	err, dataAccessResult, data := decode_GetResponseNormal(buf)
	if nil != err {
		t.Fatalf("decode_GetResponseNormal() failed, err %v", err)
	}

	tag := data.GetType()
	val := data.GetOctetString()

	if 0 != dataAccessResult {
		t.Fatalf("dataAccessResult wrong: %d", dataAccessResult)
	}
	if nil == data {
		t.Fatalf("data is nil")
	}
	if 9 != tag {
		t.Fatalf("data.tag wrong: %d", tag)
	}
	if !bytes.Equal(val, []byte{0x11, 0x22, 0x33, 0x44, 0x55, 0x66}) {
		t.Fatalf("bytes don't match")
	}
}

func TestX_decode_GetResponseNormalBlock(t *testing.T) {
	pdu := []byte{
		0x09, 0x06,
		0x11, 0x22, 0x33, 0x44, 0x55, 0x66}
	buf := bytes.NewBuffer(pdu)

	err, data := decode_GetResponseNormalBlock(buf)
	if nil != err {
		t.Fatalf("decode_GetResponseNormalBlock() failed, err %v", err)
	}

	tag := data.GetType()
	val := data.GetOctetString()

	if nil == data {
		t.Fatalf("data is nil")
	}
	if 9 != tag {
		t.Fatalf("data.tag wrong: %d", tag)
	}
	if !bytes.Equal(val, []byte{0x11, 0x22, 0x33, 0x44, 0x55, 0x66}) {
		t.Fatalf("bytes don't match")
	}
}

func TestX_encode_GetRequestWithList(t *testing.T) {

	b := []byte{
		0x02,
		0x00, 0x01, 0x00, 0x00, 0x80, 0x00, 0x00, 0xFF, 0x02, 0x00,
		0x00, 0x01, 0x00, 0x00, 0x80, 0x01, 0x00, 0xFF, 0x02, 0x00}

	var buf bytes.Buffer
	err := encode_GetRequestWithList(&buf, []DlmsClassId{1, 1}, []*DlmsOid{&DlmsOid{0, 0, 128, 0, 0, 255}, &DlmsOid{0, 0, 128, 1, 0, 255}}, []DlmsAttributeId{2, 2}, []DlmsAccessSelector{0, 0}, []*DlmsData{nil, nil})
	if nil != err {
		t.Fatalf("encode_GetRequestWithList() failed, err: %v", err)
	}

	bd := buf.Bytes()

	if !bytes.Equal(bd, b) {
		t.Fatalf("bytes don't match")
	}
}

func TestX_decode_GetRequestWithList(t *testing.T) {
	b := []byte{
		0x02,
		0x00, 0x01, 0x00, 0x00, 0x80, 0x00, 0x00, 0xFF, 0x02, 0x00,
		0x00, 0x01, 0x00, 0x00, 0x80, 0x01, 0x00, 0xFF, 0x02, 0x00}
	buf := bytes.NewBuffer(b)

	err, classIds, instanceIds, attributeIds, accessSelectors, accessParameters := decode_GetRequestWithList(buf)
	if nil != err {
		t.Fatalf("decode_GetRequestWithList() failed, err: %v", err)
	}

	if 2 != len(classIds) {
		t.Fatalf("wrong count")
	}

	if 0x0001 != classIds[0] {
		t.Fatalf("wrong classId[0] ")
	}
	if !oidEquals(&DlmsOid{0x00, 0x00, 0x80, 0x00, 0x00, 0xFF}, instanceIds[0]) {
		t.Fatalf("wrong instanceId[0] ")
	}
	if 0x02 != attributeIds[0] {
		t.Fatalf("wrong attributeId[0] ")
	}
	if 0x00 != accessSelectors[0] {
		t.Fatalf("wrong accessSelector[0] ")
	}
	if nil != accessParameters[0] {
		t.Fatalf("wrong accessParameters[0]")
	}

	if 0x0001 != classIds[1] {
		t.Fatalf("wrong classId[1] ")
	}
	if !oidEquals(&DlmsOid{0x00, 0x00, 0x80, 0x01, 0x00, 0xFF}, instanceIds[1]) {
		t.Fatalf("wrong instanceId[1] ")
	}
	if 0x02 != attributeIds[1] {
		t.Fatalf("wrong attributeId[1] ")
	}
	if 0x00 != accessSelectors[1] {
		t.Fatalf("wrong accessSelector[1] ")
	}
	if nil != accessParameters[1] {
		t.Fatalf("wrong accessParameters[0]")
	}
}

func TestX_encode_GetResponseWithList(t *testing.T) {
	b := []byte{
		0x02,
		0x00,
		0x09, 0x04,
		0x01, 0x02, 0x03, 0x04,
		0x00,
		0x0A, 0x03,
		0x30, 0x30, 0x30}

	data1 := new(DlmsData)
	data1.SetOctetString([]byte{0x01, 0x02, 0x03, 0x04})

	data2 := new(DlmsData)
	data2.SetVisibleString([]byte{0x30, 0x30, 0x30})

	dataAccessResults := make([]DlmsDataAccessResult, 2)
	datas := make([]*DlmsData, 2)

	dataAccessResults[0] = 0
	datas[0] = data1

	dataAccessResults[1] = 0
	datas[1] = data2

	var buf bytes.Buffer
	err := encode_GetResponseWithList(&buf, dataAccessResults, datas)
	if nil != err {
		t.Fatalf("encode_GetResponseWithList() failed, err: %v", err)
	}

	if !bytes.Equal(buf.Bytes(), b) {
		t.Fatalf("bytes don't match")
	}
}

func TestX_decode_GetResponseWithList(t *testing.T) {
	b := []byte{
		0x02,
		0x00,
		0x09, 0x04,
		0x01, 0x02, 0x03, 0x04,
		0x00,
		0x0A, 0x03,
		0x30, 0x30, 0x30}
	buf := bytes.NewBuffer(b)

	err, dataAccessResults, datas := decode_GetResponseWithList(buf)
	if nil != err {
		t.Fatalf("decode_GetResponseWithList() failed, err: %v", err)
	}

	if 0 != dataAccessResults[0] {
		t.Fatalf("wrong dataAccessResults[0]: ", dataAccessResults[0])
	}
	data := datas[0]
	if DATA_TYPE_OCTET_STRING != data.GetType() {
		t.Fatalf("wrong tag[0]: ", data.GetType)
	}
	if !bytes.Equal(data.GetOctetString(), []byte{0x01, 0x02, 0x03, 0x04}) {
		t.Fatalf("wrong data[0]")
	}

	if 0 != dataAccessResults[1] {
		t.Fatalf("wrong dataAccessResults[1]: ", dataAccessResults[1])
	}
	data = datas[1]
	if DATA_TYPE_VISIBLE_STRING != data.GetType() {
		t.Fatalf("wrong tag[1]: ", data.GetType())
	}
	if !bytes.Equal(data.GetVisibleString(), []byte{0x30, 0x30, 0x30}) {
		t.Fatalf("wrong data[1]")
	}
}

func TestX_encode_GetResponsewithDataBlock(t *testing.T) {
	b := []byte{
		0x00,
		0x00, 0x00, 0x00, 0x01,
		0x00,
		0x1E,
		0x09, 0x32, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x10, 0x11, 0x12, 0x13,
		0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x20, 0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27, 0x28}

	var buf bytes.Buffer
	err := encode_GetResponsewithDataBlock(&buf, false, 1, 0, []byte{0x09, 0x32, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x20, 0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27, 0x28})
	if nil != err {
		t.Fatalf("encode_GetResponsewithDataBlock() failed, err: %v", err)
	}

	if !bytes.Equal(buf.Bytes(), b) {
		t.Fatalf("bytes don't match")
	}
}

func TestX_decode_GetResponsewithDataBlock(t *testing.T) {
	b := []byte{
		0x00,
		0x00, 0x00, 0x00, 0x01,
		0x00,
		0x1E,
		0x09, 0x32, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x10, 0x11, 0x12, 0x13,
		0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x20, 0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27, 0x28}

	buf := bytes.NewBuffer(b)
	err, lastBlock, blockNumber, dataAccessResult, rawData := decode_GetResponsewithDataBlock(buf)
	if nil != err {
		t.Fatalf("decode_GetResponsewithDataBlock() failed, err %v", err)
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
	if !bytes.Equal(rawData, []byte{0x09, 0x32, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x20, 0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27, 0x28}) {
		t.Fatalf("wrong rawData")
	}
}

func TestX_encode_GetRequestForNextDataBlock(t *testing.T) {
	b := []byte{
		0x00, 0x00, 0x00, 0x01}

	var buf bytes.Buffer
	err := encode_GetRequestForNextDataBlock(&buf, 1)
	if nil != err {
		t.Fatalf("encode_GetRequestForNextDataBlock() failed, err: %v", err)
	}

	if !bytes.Equal(buf.Bytes(), b) {
		t.Fatalf("wrong rawData")
	}

}

func TestX_decode_GetRequestForNextDataBlock(t *testing.T) {
	b := []byte{
		0x00, 0x00, 0x00, 0x01}
	buf := bytes.NewBuffer(b)

	err, blockNumber := decode_GetRequestForNextDataBlock(buf)
	if nil != err {
		t.Fatalf("decode_GetRequestForNextDataBlock() failed, err: %v", err)
	}

	if 1 != blockNumber {
		t.Fatalf("wrong blockNumber")
	}
}

func TestX_encode_SetRequestNormal(t *testing.T) {
	b := []byte{
		0x00, 0x01, 0x00, 0x00, 0x80, 0x00, 0x00, 0xFF, 0x02, 0x00,
		0x09, 0x32,
		0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16,
		0x17, 0x18, 0x19, 0x20, 0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27, 0x28, 0x29, 0x30, 0x31, 0x32,
		0x33, 0x34, 0x35, 0x36, 0x37, 0x38, 0x39, 0x40, 0x41, 0x42, 0x43, 0x44, 0x45, 0x46, 0x47, 0x48,
		0x49, 0x50}

	data := new(DlmsData)
	data.SetOctetString([]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16,
		0x17, 0x18, 0x19, 0x20, 0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27, 0x28, 0x29, 0x30, 0x31, 0x32,
		0x33, 0x34, 0x35, 0x36, 0x37, 0x38, 0x39, 0x40, 0x41, 0x42, 0x43, 0x44, 0x45, 0x46, 0x47, 0x48,
		0x49, 0x50})

	var buf bytes.Buffer
	err := encode_SetRequestNormal(&buf, 1, &DlmsOid{0, 0, 128, 0, 0, 255}, 2, 0, nil, data)
	if nil != err {
		t.Fatalf("encode_GetRequestNormal() failed, err: %v", err)
	}

	if !bytes.Equal(buf.Bytes(), b) {
		t.Fatalf("bytes don't match")
	}

}

func TestX_decode_SetRequestNormal(t *testing.T) {
	pdu := []byte{
		0x00, 0x01, 0x00, 0x00, 0x80, 0x00, 0x00, 0xFF, 0x02, 0x00,
		0x09, 0x32,
		0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16,
		0x17, 0x18, 0x19, 0x20, 0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27, 0x28, 0x29, 0x30, 0x31, 0x32,
		0x33, 0x34, 0x35, 0x36, 0x37, 0x38, 0x39, 0x40, 0x41, 0x42, 0x43, 0x44, 0x45, 0x46, 0x47, 0x48,
		0x49, 0x50}
	buf := bytes.NewBuffer(pdu)

	err, classId, instanceId, attributeId, accessSelector, accessParameters, data := decode_SetRequestNormal(buf)
	if nil != err {
		t.Fatalf("decode_GetRequestNorma() failed, err %v", err)
	}

	if 1 != classId {
		t.Fatalf("classId wrong:  %d", classId)
	}
	if !oidEquals(&DlmsOid{0x00, 0x00, 0x80, 0x00, 0x00, 0xFF}, instanceId) {
		t.Fatalf("instanceId wrong:  %02X", *instanceId)
	}
	if 0x02 != attributeId {
		t.Fatalf("attributeId wrong:  %d", attributeId)
	}
	if 0x00 != accessSelector {
		t.Fatalf("accessSelector wrong:  %d", accessSelector)
	}
	if nil != accessParameters {
		t.Fatalf("accessParameters wrong:  %p", accessParameters)
	}
	if DATA_TYPE_OCTET_STRING != data.Typ {
		t.Fatalf("data type wrong:  %d", data.Typ)
	}
	if !bytes.Equal(data.GetOctetString(), []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16,
		0x17, 0x18, 0x19, 0x20, 0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27, 0x28, 0x29, 0x30, 0x31, 0x32,
		0x33, 0x34, 0x35, 0x36, 0x37, 0x38, 0x39, 0x40, 0x41, 0x42, 0x43, 0x44, 0x45, 0x46, 0x47, 0x48,
		0x49, 0x50}) {
		t.Fatalf("data does nit match")
	}

}
