package gocosem

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"time"
	"unsafe"
)

const (
	C_Dlms_Data_Nothing       = iota
	C_Dlms_Data_Nil           = iota
	C_Dlms_Data_Bool          = iota
	C_Dlms_Data_BitString     = iota
	C_Dlms_Data_Int32         = iota
	C_Dlms_Data_Uint32        = iota
	C_Dlms_Data_Bytes         = iota
	C_Dlms_Data_VisibleString = iota
	C_Dlms_Data_Bcd           = iota
	C_Dlms_Data_Int8          = iota
	C_Dlms_Data_Int16         = iota
	C_Dlms_Data_Uint8         = iota
	C_Dlms_Data_Uint16        = iota
	C_Dlms_Data_Int64         = iota
	C_Dlms_Data_Uint64        = iota
	C_Dlms_Data_Float32       = iota
	C_Dlms_Data_Float64       = iota
	C_Dlms_Data_DateTime      = iota
	C_Dlms_Data_Date          = iota
	C_Dlms_Data_Time          = iota
	C_Dlms_Data_DontCare      = iota
)

const (
	dataAccessResult_success                 = 0
	dataAccessResult_hardwareFault           = 1
	dataAccessResult_temporaryFailure        = 2
	dataAccessResult_readWriteDenied         = 3
	dataAccessResult_objectUndefined         = 4
	dataAccessResult_objectClassInconsistent = 9
	dataAccessResult_objectUnavailable       = 11
	dataAccessResult_typeUnmatched           = 12
	dataAccessResult_scopeOfAccessViolated   = 13
	dataAccessResult_dataBlockUnavailable    = 14
	dataAccessResult_longGetAborted          = 15
	dataAccessResult_noLongGetInProgress     = 16
	dataAccessResult_longSetAborted          = 17
	dataAccessResult_noLongSetInProgress     = 18
	dataAccessResult_dataBlockNumberInvalid  = 19
	dataAccessResult_otherReason             = 250
)

type tDlmsInvokeIdAndPriority uint8
type tDlmsClassId uint16
type tDlmsOid [6]uint8
type tDlmsAttributeId uint8
type tDlmsAccessSelector uint8

type tDlmsData tAsn1Choice
type tDlmsDataAccessResult uint8

type tDlmsDate struct {
	year       uint16
	month      uint8
	dayOfMonth uint8
	dayOfWeek  uint8
}

type tDlmsTime struct {
	hour       uint8
	minute     uint8
	second     uint8
	hundredths uint8
}

type tDlmsDateTime struct {
	tDlmsDate
	tDlmsTime
	deviation   int16
	clockStatus uint8
}

type tDlmsAsn1Data struct {
	adata *tAsn1Choice
}

var errorLog *log.Logger = getErrorLogger()
var debugLog *log.Logger = getDebugLogger()

func DlmsDateFromBytes(b []byte) (date *tDlmsDate) {
	date = new(tDlmsDate)
	by := (*[2]byte)(unsafe.Pointer(&date.year))
	by[0] = b[0]
	by[1] = b[1]
	date.month = b[2]
	date.dayOfMonth = b[3]
	date.dayOfWeek = b[4]
	return date
}

func (date *tDlmsDate) toBytes() []byte {
	b := make([]byte, 5)
	b[0] = byte((date.year & 0xFF00) >> 8)
	b[1] = byte(date.year & 0x00FF)
	b[2] = date.month
	b[3] = date.dayOfMonth
	b[4] = date.dayOfWeek
	return b
}

func (date *tDlmsDate) setYearWildcard() {
	date.year = 0xFFFF
}

func (date *tDlmsDate) isYearWildcard() bool {
	return date.year == 0xFFFF
}

func (date *tDlmsDate) setMonthWildcard() {
	date.month = 0xFF
}

func (date *tDlmsDate) isMonthWildcard() bool {
	return date.month == 0xFF
}

func (date *tDlmsDate) setDaylightSavingsEnd() {
	date.month = 0xFD
}

func (date *tDlmsDate) isDaylightSavingsEnd() bool {
	return date.month == 0xFD
}

func (date *tDlmsDate) setDaylightSavingsBegin() {
	date.month = 0xFE
}

func (date *tDlmsDate) isDaylightSavingsBegin() bool {
	return date.month == 0xFE
}

func (date *tDlmsDate) setDayOfWeekWildcard() {
	date.dayOfWeek = 0xFF
}

func (date *tDlmsDate) isDayOfWeekWildcard() bool {
	return date.dayOfWeek == 0xFF
}

func DlmsTimeFromBytes(b []byte) (tim *tDlmsTime) {
	tim = new(tDlmsTime)
	tim.hour = b[0]
	tim.minute = b[1]
	tim.second = b[2]
	tim.hundredths = b[3]
	return tim
}

func (tim *tDlmsTime) toBytes() []byte {
	b := make([]byte, 4)
	b[0] = tim.hour
	b[1] = tim.minute
	b[2] = tim.second
	b[3] = tim.hundredths
	return b
}

func (tim *tDlmsTime) setHourWildcard() {
	tim.hour = 0xFF
}

func (tim *tDlmsTime) isHourWildcard() bool {
	return tim.hour == 0xFF
}

func (tim *tDlmsTime) setMinuteWildcard() {
	tim.minute = 0xFF
}

func (tim *tDlmsTime) isMinuteWildcard() bool {
	return tim.minute == 0xFF
}

func (tim *tDlmsTime) setSecondWildcard() {
	tim.second = 0xFF
}

func (tim *tDlmsTime) isSecondWildcard() bool {
	return tim.second == 0xFF
}

func (tim *tDlmsTime) setHundredthsWildcard() {
	tim.hundredths = 0xFF
}

func (tim *tDlmsTime) isHundredthsWildcard() bool {
	return tim.hundredths == 0xFF
}

func DlmsDateTimeFromBytes(b []byte) (dateTime *tDlmsDateTime) {

	dateTime = new(tDlmsDateTime)
	b2 := (*[2]byte)(unsafe.Pointer(&dateTime.year))
	b2[0] = b[0]
	b2[1] = b[1]
	dateTime.month = b[2]
	dateTime.dayOfMonth = b[3]
	dateTime.dayOfWeek = b[4]
	dateTime.hour = b[5]
	dateTime.minute = b[6]
	dateTime.second = b[7]
	dateTime.hundredths = b[8]
	b2 = (*[2]byte)(unsafe.Pointer(&dateTime.deviation))
	b2[0] = b[9]
	b2[1] = b[10]
	dateTime.clockStatus = b[11]

	return dateTime
}

func (dateTime *tDlmsDateTime) toBytes() []byte {
	b := make([]byte, 12)
	b2 := (*[2]byte)(unsafe.Pointer(&dateTime.year))
	b[0] = b2[0]
	b[1] = b2[1]
	b[2] = dateTime.month
	b[3] = dateTime.dayOfMonth
	b[4] = dateTime.dayOfWeek
	b[5] = dateTime.hour
	b[6] = dateTime.minute
	b[7] = dateTime.second
	b[8] = dateTime.hundredths
	b2 = (*[2]byte)(unsafe.Pointer(&dateTime.deviation))
	b[9] = b[0]
	b[10] = b[1]
	b[11] = dateTime.clockStatus
	return b
}

func (dateTime *tDlmsDateTime) setDeviationWildcard() {
	b := (*[2]byte)(unsafe.Pointer(&dateTime.deviation))
	b[0] = 0x80
	b[1] = 0x00
}

func (dateTime *tDlmsDateTime) isDeviationWildcard() bool {
	b := (*[2]byte)(unsafe.Pointer(&dateTime.deviation))
	return (b[0] == 0x80) && (b[1] == 0x00)
}

func (dateTime *tDlmsDateTime) setClockStatusInvalid() {
	dateTime.clockStatus |= 0x01
}

func (dateTime *tDlmsDateTime) isClockStatusInvalid() bool {
	return dateTime.clockStatus&0x01 > 0
}

func (dateTime *tDlmsDateTime) setClockStatusDoubtful() {
	dateTime.clockStatus |= 0x02
}

func (dateTime *tDlmsDateTime) isClockStatusDoubtful() bool {
	return dateTime.clockStatus&0x02 > 0
}

func (dateTime *tDlmsDateTime) setClockStatusDifferentClockBase() {
	dateTime.clockStatus |= 0x04
}

func (dateTime *tDlmsDateTime) isClockStatusDifferentClockBase() bool {
	return dateTime.clockStatus&0x04 > 0
}

func (dateTime *tDlmsDateTime) setClockStatusDaylightSavingActive() {
	dateTime.clockStatus |= 0x80
}

func (dateTime *tDlmsDateTime) isClockStatusDaylightSavingActive() bool {
	return dateTime.clockStatus&0x80 > 0
}

func NewDlmsData() (data *tDlmsData) {

	return (*tDlmsData)(new(tAsn1Choice))
}

func (data *tDlmsData) setNothing() {
	adata := (*tAsn1Choice)(data)
	adata.setVal(C_Data_PR_NOTHING, nil)
}

func (data *tDlmsData) setNil() {
	adata := (*tAsn1Choice)(data)
	adata.setVal(C_Data_PR_null_data, nil)
}

func (data *tDlmsData) setBool(b bool) {
	adata := (*tAsn1Choice)(data)
	adata.setVal(C_Data_PR_boolean, (*tAsn1Boolean)(&b))
}

func (data *tDlmsData) getBool() bool {
	adata := (*tAsn1Choice)(data)
	v := adata.getVal().(*tAsn1Boolean)
	return bool(*v)
}

// 'unusedBits' is number of unused bits in last octet of 'b'
func (data *tDlmsData) setBitString(b []byte, unusedBits int) {

	if !((unusedBits >= 0) && (unusedBits <= 7)) {
		panic("assertion failed")
	}

	adata := (*tAsn1Choice)(data)

	abitString := new(tAsn1BitString)
	abitString.buf = b
	abitString.bitsUnused = unusedBits

	adata.setVal(C_Data_PR_bit_string, abitString)
}

func (data *tDlmsData) getBitString() (b []byte, unusedBits int) {
	adata := (*tAsn1Choice)(data)
	v := adata.getVal().(*tAsn1BitString)
	b = v.buf
	unusedBits = v.bitsUnused
	return b, unusedBits
}

func (data *tDlmsData) setInt32(i int32) {

	adata := (*tAsn1Choice)(data)
	adata.setVal(C_Data_PR_double_long, (*tAsn1Integer32)(&i))
}

func (data *tDlmsData) getInt32() int32 {
	adata := (*tAsn1Choice)(data)
	v := adata.getVal().(*tAsn1Integer32)
	return int32(*v)
}

func (data *tDlmsData) setUint32(i uint32) {

	adata := (*tAsn1Choice)(data)
	adata.setVal(C_Data_PR_double_long_unsigned, (*tAsn1Unsigned32)(&i))
}

func (data *tDlmsData) getUint32() uint32 {
	adata := (*tAsn1Choice)(data)
	v := adata.getVal().(*tAsn1Unsigned32)
	return uint32(*v)
}

func (data *tDlmsData) setBytes(b []byte) {

	adata := (*tAsn1Choice)(data)
	adata.setVal(C_Data_PR_octet_string, (*tAsn1OctetString)(&b))
}

func (data *tDlmsData) getBytes() []byte {
	adata := (*tAsn1Choice)(data)
	v := adata.getVal().(*tAsn1OctetString)
	return ([]byte)(*v)
}

func (data *tDlmsData) setVisibleString(s string) {

	adata := (*tAsn1Choice)(data)
	b := ([]byte)(s)
	adata.setVal(C_Data_PR_visible_string, (*tAsn1VisibleString)(&b))
}

func (data *tDlmsData) getVisibleString() string {
	adata := (*tAsn1Choice)(data)
	v := adata.getVal().(*tAsn1VisibleString)
	return string(*v)
}

func (data *tDlmsData) setBcd(bcd int8) {

	adata := (*tAsn1Choice)(data)
	adata.setVal(C_Data_PR_bcd, (*tAsn1Integer8)(&bcd))
}

func (data *tDlmsData) getBcd() int8 {
	adata := (*tAsn1Choice)(data)
	v := adata.getVal().(*tAsn1Integer8)
	return int8(*v)
}

func (data *tDlmsData) setInt8(i int8) {

	adata := (*tAsn1Choice)(data)
	adata.setVal(C_Data_PR_integer, (*tAsn1Integer8)(&i))
}

func (data *tDlmsData) getInt8() int8 {
	adata := (*tAsn1Choice)(data)
	v := adata.getVal().(*tAsn1Integer8)
	return int8(*v)
}

func (data *tDlmsData) setIn16(i int16) {

	adata := (*tAsn1Choice)(data)
	adata.setVal(C_Data_PR_long, (*tAsn1Integer16)(&i))
}

func (data *tDlmsData) getInt16() int16 {
	adata := (*tAsn1Choice)(data)
	v := adata.getVal().(*tAsn1Integer16)
	return int16(*v)
}

func (data *tDlmsData) setUint8(i uint8) {

	adata := (*tAsn1Choice)(data)
	adata.setVal(C_Data_PR_unsigned, (*tAsn1Unsigned8)(&i))
}

func (data *tDlmsData) getUint8() uint8 {
	adata := (*tAsn1Choice)(data)
	v := adata.getVal().(*tAsn1Unsigned8)
	return uint8(*v)
}

func (data *tDlmsData) setUint16(i uint16) {

	adata := (*tAsn1Choice)(data)
	adata.setVal(C_Data_PR_long_unsigned, (*tAsn1Unsigned16)(&i))
}

func (data *tDlmsData) getUint16() uint16 {
	adata := (*tAsn1Choice)(data)
	v := adata.getVal().(*tAsn1Unsigned16)
	return uint16(*v)
}

func (data *tDlmsData) setInt64(i int64) {

	adata := (*tAsn1Choice)(data)
	adata.setVal(C_Data_PR_long64, (*tAsn1Long64)(&i))
}

func (data *tDlmsData) getInt64() int64 {
	adata := (*tAsn1Choice)(data)
	v := adata.getVal().(*tAsn1Long64)
	return int64(*v)
}

func (data *tDlmsData) setUint64(i uint64) {

	adata := (*tAsn1Choice)(data)
	adata.setVal(C_Data_PR_long64_unsigned, (*tAsn1UnsignedLong64)(&i))
}

func (data *tDlmsData) getUint64() uint64 {
	adata := (*tAsn1Choice)(data)
	v := adata.getVal().(*tAsn1UnsignedLong64)
	return uint64(*v)
}

func (data *tDlmsData) setFloat32(f float32) {

	adata := (*tAsn1Choice)(data)
	adata.setVal(C_Data_PR_float32, (*tAsn1Float32)(&f))
}

func (data *tDlmsData) getFloat32() (f float32) {
	adata := (*tAsn1Choice)(data)
	v := adata.getVal().(*tAsn1Float32)
	return float32(*v)
}

func (data *tDlmsData) setFloat64(f float64) {

	adata := (*tAsn1Choice)(data)
	adata.setVal(C_Data_PR_float64, (*tAsn1Float64)(&f))
}

func (data *tDlmsData) getFloat64() float64 {
	adata := (*tAsn1Choice)(data)
	v := adata.getVal().(*tAsn1Float64)
	return float64(*v)
}

func (data *tDlmsData) setDateTime(dateTime *tDlmsDateTime) {

	adata := (*tAsn1Choice)(data)
	b := dateTime.toBytes()
	adata.setVal(C_Data_PR_date_time, (*tAsn1DateTime)(&b))
}

func (data *tDlmsData) getDateTime() *tDlmsDateTime {
	adata := (*tAsn1Choice)(data)
	v := adata.getVal().(*tAsn1DateTime)
	return DlmsDateTimeFromBytes(([]byte)(*v))
}

func (data *tDlmsData) setDate(date *tDlmsDate) {

	adata := (*tAsn1Choice)(data)
	b := date.toBytes()
	adata.setVal(C_Data_PR_date, (*tAsn1Date)(&b))
}

func (data *tDlmsData) getDate() *tDlmsDate {
	adata := (*tAsn1Choice)(data)
	v := adata.getVal().(*tAsn1Date)
	return DlmsDateFromBytes(([]byte)(*v))
}

func (data *tDlmsData) setTime(tim *tDlmsTime) {

	adata := (*tAsn1Choice)(data)
	b := tim.toBytes()
	adata.setVal(C_Data_PR_time, (*tAsn1Time)(&b))
}

func (data *tDlmsData) getTime() *tDlmsTime {
	adata := (*tAsn1Choice)(data)
	v := adata.getVal().(*tAsn1Time)
	return DlmsTimeFromBytes(([]byte)(*v))
}

func (data *tDlmsData) setDontCare() {

	adata := (*tAsn1Choice)(data)
	adata.setVal(C_Data_PR_dont_care, nil)
}

func (data *tDlmsData) getTag() int {

	adata := (*tAsn1Choice)(data)

	switch adata.getTag() {
	case C_Data_PR_NOTHING:
		return C_Dlms_Data_Nothing
	case C_Data_PR_null_data:
		return C_Dlms_Data_Nil
	case C_Data_PR_boolean:
		return C_Dlms_Data_Bool
	case C_Data_PR_bit_string:
		return C_Dlms_Data_BitString
	case C_Data_PR_double_long:
		return C_Dlms_Data_Int32
	case C_Data_PR_double_long_unsigned:
		return C_Dlms_Data_Uint32
	case C_Data_PR_octet_string:
		return C_Dlms_Data_Bytes
	case C_Data_PR_visible_string:
		return C_Dlms_Data_VisibleString
	case C_Data_PR_bcd:
		return C_Dlms_Data_Bcd
	case C_Data_PR_integer:
		return C_Dlms_Data_Int8
	case C_Data_PR_long:
		return C_Dlms_Data_Int16
	case C_Data_PR_unsigned:
		return C_Dlms_Data_Uint8
	case C_Data_PR_long_unsigned:
		return C_Dlms_Data_Uint16
	case C_Data_PR_long64:
		return C_Dlms_Data_Int64
	case C_Data_PR_long64_unsigned:
		return C_Dlms_Data_Uint64
	case C_Data_PR_float32:
		return C_Dlms_Data_Float32
	case C_Data_PR_float64:
		return C_Dlms_Data_Float64
	case C_Data_PR_date_time:
		return C_Dlms_Data_DateTime
	case C_Data_PR_date:
		return C_Dlms_Data_Date
	case C_Data_PR_time:
		return C_Dlms_Data_Time
	case C_Data_PR_dont_care:
		return C_Dlms_Data_DontCare
	default:
		panic("assertion failed")
	}
}

func encode_getRequest(classId tDlmsClassId, instanceId *tDlmsOid, attributeId tDlmsAttributeId, accessSelector *tDlmsAccessSelector, accessParameters *tDlmsData) (err error, pdu []byte) {
	var FNAME = "encode_getRequest()"

	var w bytes.Buffer

	var buf bytes.Buffer
	err = binary.Write(&buf, binary.BigEndian, classId)
	if nil != err {
		errorLog.Printf(fmt.Sprintf("%s: binary.Write() failed, err: %s\n", FNAME, err))
		return err, nil
	}
	b := buf.Bytes()
	_, err = w.Write(b)
	if nil != err {
		errorLog.Printf("%s: w.Wite() failed, err: %v\n", FNAME, err)
		return err, nil
	}

	_, err = w.Write((*instanceId)[0:6])
	if nil != err {
		errorLog.Printf("%s: w.Wite() failed, err: %v\n", FNAME, err)
		return err, nil
	}

	_, err = w.Write([]byte{byte(attributeId)})
	if nil != err {
		errorLog.Printf("%s: w.Wite() failed, err: %v\n", FNAME, err)
		return err, nil
	}

	if 0 != attributeId {
		var as []byte
		var ap []byte
		if nil == accessSelector {
			as = []byte{0}
		} else {
			as = []byte{byte(*accessSelector)}
		}
		if nil != accessParameters {
			err, ap = encode_Data((*tAsn1Choice)(accessParameters))
			if nil != err {
				return err, nil
			}
		} else {
			ap = make([]byte, 0)
		}

		_, err = w.Write(as)
		if nil != err {
			errorLog.Printf("%s: w.Wite() failed, err: %v\n", FNAME, err)
			return err, nil
		}

		_, err = w.Write(ap)
		if nil != err {
			errorLog.Printf("%s: w.Wite() failed, err: %v\n", FNAME, err)
			return err, nil
		}
	}

	return nil, w.Bytes()
}

func decode_getRequest(pdu []byte) (err error, n int, classId tDlmsClassId, instanceId *tDlmsOid, attributeId tDlmsAttributeId, accessSelector *tDlmsAccessSelector, accessParameters *tDlmsData) {
	var FNAME = "decode_getRequest()"
	var serr string

	b := pdu[0:]
	n = 0

	if len(b) < 2 {
		serr = fmt.Sprintf("%s: short pdu", FNAME)
		errorLog.Println(serr)
		return errors.New(serr), 0, 0, nil, 0, nil, nil
	}
	err = binary.Read(bytes.NewBuffer(b[0:2]), binary.BigEndian, &classId)
	if nil != err {
		errorLog.Println("%s: binary.Read() failed, err: %v", err)
		return err, 0, 0, nil, 0, nil, nil
	}
	b = b[2:]
	n += 2

	if len(b) < 6 {
		serr = fmt.Sprintf("%s: short pdu", FNAME)
		errorLog.Println(serr)
		return errors.New(serr), 0, 0, nil, 0, nil, nil
	}
	instanceId = new(tDlmsOid)
	err = binary.Read(bytes.NewBuffer(b[0:6]), binary.BigEndian, instanceId)
	if nil != err {
		errorLog.Println("%s: binary.Read() failed, err: %v", err)
		return err, 0, 0, nil, 0, nil, nil
	}
	b = b[6:]
	n += 6

	if len(b) < 1 {
		serr = fmt.Sprintf("%s: short pdu", FNAME)
		errorLog.Println(serr)
		return errors.New(serr), 0, 0, nil, 0, nil, nil
	}
	err = binary.Read(bytes.NewBuffer(b[0:1]), binary.BigEndian, &attributeId)
	if nil != err {
		errorLog.Println("%s: binary.Read() failed, err: %v", err)
		return err, 0, 0, nil, 0, nil, nil
	}
	b = b[1:]
	n += 1

	if len(b) >= 1 {
		accessSelector = new(tDlmsAccessSelector)
		err = binary.Read(bytes.NewBuffer(b[0:1]), binary.BigEndian, accessSelector)
		if nil != err {
			errorLog.Println("%s: binary.Read() failed, err: %v", err)
			return err, 0, 0, nil, 0, nil, nil
		}
		b = b[1:]
		n += 1
	}

	//TODO: Fucking cosem green book full of garbage is not precise on how and when access selector parameters are to be encoded/decoded
	// We skip this to avoid reading too much into next item in case of decoding GetRequestWithList.
	//if len(b) >= 1 {
	if false {
		var nn int
		var data *tAsn1Choice
		err, data, nn = decode_Data(b)
		if nil != err {
			return err, 0, 0, nil, 0, nil, nil
		}
		accessParameters = (*tDlmsData)(data)
		b = b[nn:]
		n += nn
	}
	return nil, n, classId, instanceId, attributeId, accessSelector, accessParameters
}

func encode_getResponse(dataAccessResult tDlmsDataAccessResult, data *tDlmsData) (err error, pdu []byte) {
	var FNAME = "encode_getResponse()"

	var w bytes.Buffer

	var buf bytes.Buffer
	err = binary.Write(&buf, binary.BigEndian, dataAccessResult)
	if nil != err {
		errorLog.Printf(fmt.Sprintf("%s: binary.Write() failed, err: %s\n", FNAME, err))
		return err, nil
	}
	b := buf.Bytes()
	_, err = w.Write(b)
	if nil != err {
		errorLog.Printf("%s: w.Wite() failed, err: %v\n", FNAME, err)
		return err, nil
	}

	if nil != data {
		err, b = encode_Data((*tAsn1Choice)(data))
		if nil != err {
			return err, nil
		}
		_, err = w.Write(b)
		if nil != err {
			errorLog.Printf("%s: w.Wite() failed, err: %v\n", FNAME, err)
			return err, nil
		}
	}

	return nil, w.Bytes()
}

func decode_getResponse(pdu []byte) (err error, n int, dataAccessResult tDlmsDataAccessResult, data *tDlmsData) {
	var FNAME = "decode_getResponse()"
	var serr string
	var nn = 0

	b := pdu[0:]
	n = 0

	if len(b) < 1 {
		serr = fmt.Sprintf("%s: short pdu", FNAME)
		errorLog.Println(serr)
		return errors.New(serr), 0, 0, nil
	}
	dataAccessResult = tDlmsDataAccessResult(b[0])
	b = b[1:]
	n += 1

	var cdata *tAsn1Choice
	if dataAccessResult_success == dataAccessResult {
		err, cdata, nn = decode_Data(b)
		if nil != err {
			return err, n + nn, 0, nil
		}
		n += nn
	}

	return nil, n, dataAccessResult, (*tDlmsData)(cdata)
}

func encode_GetRequestNormal(invokeIdAndPriority tDlmsInvokeIdAndPriority, classId tDlmsClassId, instanceId *tDlmsOid, attributeId tDlmsAttributeId, accessSelector *tDlmsAccessSelector, accessParameters *tDlmsData) (err error, pdu []byte) {
	var FNAME = "encode_GetRequestNormal()"

	var w bytes.Buffer

	_, err = w.Write([]byte{0xc0, 0x01})
	if nil != err {
		errorLog.Printf("%s: w.Wite() failed, err: %v\n", FNAME, err)
		return err, nil
	}

	_, err = w.Write([]byte{byte(invokeIdAndPriority)})
	if nil != err {
		errorLog.Printf("%s: w.Wite() failed, err: %v\n", FNAME, err)
		return err, nil
	}

	err, pdu = encode_getRequest(classId, instanceId, attributeId, accessSelector, accessParameters)
	if nil != err {
		errorLog.Printf("%s: encode_getRequest() failed, err: %v\n", FNAME, err)
		return err, nil
	}

	_, err = w.Write(pdu)
	if nil != err {
		errorLog.Printf("%s: w.Wite() failed, err: %v\n", FNAME, err)
		return err, nil
	}

	return nil, w.Bytes()
}

func decode_GetRequestNormal(pdu []byte) (err error, invokeIdAndPriority tDlmsInvokeIdAndPriority, classId tDlmsClassId, instanceId *tDlmsOid, attributeId tDlmsAttributeId, accessSelector *tDlmsAccessSelector, accessParameters *tDlmsData) {
	var FNAME = "decode_GetRequestNormal"
	var serr string
	b := pdu[0:]

	if len(b) < 2 {
		serr = fmt.Sprintf("%s: short pdu", FNAME)
		errorLog.Println(serr)
		return errors.New(serr), 0, 0, nil, 0, nil, nil
	}
	if !bytes.Equal(b[0:2], []byte{0xC0, 0x01}) {
		errorLog.Printf("%s: pdu is not GetRequestNormal: 0x%02X 0x%02X\n", FNAME, b[0], b[1])
		return errors.New("pdu is not GetRequestNormal"), 0, 0, nil, 0, nil, nil
	}
	b = b[2:]

	if len(b) < 1 {
		serr = fmt.Sprintf("%s: short pdu", FNAME)
		errorLog.Println(serr)
		return errors.New(serr), 0, 0, nil, 0, nil, nil
	}
	invokeIdAndPriority = tDlmsInvokeIdAndPriority(b[0])
	b = b[1:]

	err, _, classId, instanceId, attributeId, accessSelector, accessParameters = decode_getRequest(b)
	if nil != err {
		return err, 0, 0, nil, 0, nil, nil
	}
	return nil, invokeIdAndPriority, classId, instanceId, attributeId, accessSelector, accessParameters
}

func encode_GetResponseNormal(invokeIdAndPriority tDlmsInvokeIdAndPriority, dataAccessResult tDlmsDataAccessResult, data *tDlmsData) (err error, pdu []byte) {
	var FNAME = "encode_GetResponseNormal()"

	var w bytes.Buffer
	var b []byte

	_, err = w.Write([]byte{0xC4, 0x01})
	if nil != err {
		errorLog.Printf("%s: w.Wite() failed, err: %v\n", FNAME, err)
		return err, nil
	}

	_, err = w.Write([]byte{byte(invokeIdAndPriority)})
	if nil != err {
		errorLog.Printf("%s: w.Wite() failed, err: %v\n", FNAME, err)
		return err, nil
	}

	err, b = encode_getResponse(dataAccessResult, data)
	if nil != err {
		return err, nil
	}
	_, err = w.Write(b)
	if nil != err {
		errorLog.Printf("%s: w.Wite() failed, err: %v\n", FNAME, err)
		return err, nil
	}

	return nil, w.Bytes()
}

func decode_GetResponseNormal(pdu []byte) (err error, invokeIdAndPriority tDlmsInvokeIdAndPriority, dataAccessResult tDlmsDataAccessResult, data *tDlmsData) {
	var FNAME = "decode_GetResponsenormal()"
	var serr string
	b := pdu[0:]

	if len(b) < 2 {
		return errors.New("short pdu"), 0, 0, nil
	}
	if !bytes.Equal(b[0:2], []byte{0xC4, 0x01}) {
		errorLog.Printf("%s: pdu is not GetResponsenormal: 0x%02X 0x%02X\n", FNAME, b[0], b[1])
		return errors.New("pdu is not GetResponsenormal"), 0, 0, nil
	}
	b = b[2:]

	if len(b) < 1 {
		serr = fmt.Sprintf("%s: short pdu", FNAME)
		errorLog.Println(serr)
		return errors.New(serr), 0, 0, nil
	}
	invokeIdAndPriority = tDlmsInvokeIdAndPriority(b[0])
	b = b[1:]

	err, _, dataAccessResult, data = decode_getResponse(b)
	if nil != err {
		return err, 0, 0, nil
	}

	return nil, invokeIdAndPriority, dataAccessResult, data
}

func encode_GetRequestWithList(invokeIdAndPriority tDlmsInvokeIdAndPriority, classIds []tDlmsClassId, instanceIds []*tDlmsOid, attributeIds []tDlmsAttributeId, accessSelectors []*tDlmsAccessSelector, accessParameters []*tDlmsData) (err error, pdu []byte) {
	var FNAME = "encode_GetRequestWithList()"

	var w bytes.Buffer

	_, err = w.Write([]byte{0xc0, 0x03})
	if nil != err {
		errorLog.Printf("%s: w.Wite() failed, err: %v\n", FNAME, err)
		return err, nil
	}

	_, err = w.Write([]byte{byte(invokeIdAndPriority)})
	if nil != err {
		errorLog.Printf("%s: w.Wite() failed, err: %v\n", FNAME, err)
		return err, nil
	}

	count := len(classIds) // count of get requests

	_, err = w.Write([]byte{byte(count)})
	if nil != err {
		errorLog.Printf("%s: w.Wite() failed, err: %v\n", FNAME, err)
		return err, nil
	}

	for i := 0; i < count; i += 1 {

		err, pdu = encode_getRequest(classIds[i], instanceIds[i], attributeIds[i], accessSelectors[i], accessParameters[i])
		if nil != err {
			errorLog.Printf("%s: encode_getRequest() failed, err: %v\n", FNAME, err)
			return err, nil
		}

		_, err = w.Write(pdu)
		if nil != err {
			errorLog.Printf("%s: w.Wite() failed, err: %v\n", FNAME, err)
			return err, nil
		}
	}

	return nil, w.Bytes()
}

func decode_GetRequestWithList(pdu []byte) (err error, invokeIdAndPriority tDlmsInvokeIdAndPriority, classIds []tDlmsClassId, instanceIds []*tDlmsOid, attributeIds []tDlmsAttributeId, accessSelectors []*tDlmsAccessSelector, accessParameters []*tDlmsData) {
	var FNAME = "decode_GetRequestWithList()"
	var serr string
	b := pdu[0:]

	if len(b) < 2 {
		return errors.New("short pdu"), 0, nil, nil, nil, nil, nil
	}
	if !bytes.Equal(b[0:2], []byte{0xC0, 0x03}) {
		errorLog.Printf("%s: pdu is not GetRequestWithList: 0x%02X 0x%02X\n", FNAME, b[0], b[1])
		return errors.New("pdu is not GetRequestWithList"), 0, nil, nil, nil, nil, nil
	}
	b = b[2:]

	if len(b) < 1 {
		serr = fmt.Sprintf("%s: short pdu", FNAME)
		errorLog.Println(serr)
		return errors.New(serr), 0, nil, nil, nil, nil, nil
	}
	invokeIdAndPriority = tDlmsInvokeIdAndPriority(b[0])
	b = b[1:]

	if len(b) < 1 {
		serr = fmt.Sprintf("%s: short pdu", FNAME)
		errorLog.Println(serr)
		return errors.New(serr), 0, nil, nil, nil, nil, nil
	}
	count := int(b[0])
	b = b[1:]

	classIds = make([]tDlmsClassId, count)
	instanceIds = make([]*tDlmsOid, count)
	attributeIds = make([]tDlmsAttributeId, count)
	accessSelectors = make([]*tDlmsAccessSelector, count)
	accessParameters = make([]*tDlmsData, count)

	for i := 0; i < count; i += 1 {
		err, n, classId, instanceId, attributeId, accessSelector, accessParameter := decode_getRequest(b)
		if nil != err {
			return err, 0, nil, nil, nil, nil, nil
		}
		classIds[i] = classId
		instanceIds[i] = instanceId
		attributeIds[i] = attributeId
		accessSelectors[i] = accessSelector
		accessParameters[i] = accessParameter
		b = b[n:]
	}
	return nil, invokeIdAndPriority, classIds, instanceIds, attributeIds, accessSelectors, accessParameters

}

func encode_GetResponseWithList(invokeIdAndPriority tDlmsInvokeIdAndPriority, dataAccessResults []tDlmsDataAccessResult, datas []*tDlmsData) (err error, pdu []byte) {
	var FNAME = "encode_GetResponseWithList()"

	var w bytes.Buffer

	_, err = w.Write([]byte{0xC4, 0x03})
	if nil != err {
		errorLog.Printf("%s: w.Wite() failed, err: %v\n", FNAME, err)
		return err, nil
	}

	_, err = w.Write([]byte{byte(invokeIdAndPriority)})
	if nil != err {
		errorLog.Printf("%s: w.Wite() failed, err: %v\n", FNAME, err)
		return err, nil
	}

	count := len(dataAccessResults)

	_, err = w.Write([]byte{byte(count)})
	if nil != err {
		errorLog.Printf("%s: w.Wite() failed, err: %v\n", FNAME, err)
		return err, nil
	}

	for i := 0; i < count; i += 1 {

		err, b := encode_getResponse(dataAccessResults[i], datas[i])
		if nil != err {
			return err, nil
		}
		_, err = w.Write(b)
		if nil != err {
			errorLog.Printf("%s: w.Wite() failed, err: %v\n", FNAME, err)
			return err, nil
		}
	}

	return nil, w.Bytes()
}

func decode_GetResponseWithList(pdu []byte) (err error, invokeIdAndPriority tDlmsInvokeIdAndPriority, dataAccessResults []tDlmsDataAccessResult, datas []*tDlmsData) {
	var FNAME = "decode_GetResponseWithList()"
	var serr string
	b := pdu[0:]

	if len(b) < 2 {
		serr = fmt.Sprintf("%s: short pdu", FNAME)
		errorLog.Println(serr)
		return errors.New(serr), 0, nil, nil
	}
	if !bytes.Equal(b[0:2], []byte{0xC4, 0x03}) {
		errorLog.Printf("%s: pdu is not GetResponseWithList: 0x%02X 0x%02X\n", FNAME, b[0], b[1])
		return errors.New("pdu is not GetResponseWithList"), 0, nil, nil
	}
	b = b[2:]

	if len(b) < 1 {
		serr = fmt.Sprintf("%s: short pdu", FNAME)
		errorLog.Println(serr)
		return errors.New(serr), 0, nil, nil
	}
	invokeIdAndPriority = tDlmsInvokeIdAndPriority(b[0])
	b = b[1:]

	if len(b) < 1 {
		serr = fmt.Sprintf("%s: short pdu", FNAME)
		errorLog.Println(serr)
		return errors.New(serr), 0, nil, nil
	}
	count := int(b[0])
	b = b[1:]

	dataAccessResults = make([]tDlmsDataAccessResult, count)
	datas = make([]*tDlmsData, count)

	var dataAccessResult tDlmsDataAccessResult
	var data *tDlmsData
	var n int
	for i := 0; i < count; i += 1 {
		err, n, dataAccessResult, data = decode_getResponse(b)
		if nil != err {
			return err, 0, nil, nil
		}
		b = b[n:]
		dataAccessResults[i] = dataAccessResult
		datas[i] = data
	}

	return nil, invokeIdAndPriority, dataAccessResults, datas
}

func encode_GetResponsewithDataBlock(invokeIdAndPriority tDlmsInvokeIdAndPriority, lastBlock bool, blockNumber uint32, dataAccessResult tDlmsDataAccessResult, rawData []byte) (err error, pdu []byte) {
	var FNAME = "encode_GetResponsewithDataBlock()"

	var w bytes.Buffer

	_, err = w.Write([]byte{0xC4, 0x02})
	if nil != err {
		errorLog.Printf("%s: w.Wite() failed, err: %v\n", FNAME, err)
		return err, nil
	}

	_, err = w.Write([]byte{byte(invokeIdAndPriority)})
	if nil != err {
		errorLog.Printf("%s: w.Wite() failed, err: %v\n", FNAME, err)
		return err, nil
	}

	var bb byte
	if lastBlock {
		bb = 1
	} else {
		bb = 0
	}
	_, err = w.Write([]byte{bb})
	if nil != err {
		errorLog.Printf("%s: w.Wite() failed, err: %v\n", FNAME, err)
		return err, nil
	}

	var buf bytes.Buffer

	err = binary.Write(&buf, binary.BigEndian, blockNumber)
	if nil != err {
		errorLog.Printf(fmt.Sprintf("%s: binary.Write() failed, err: %s\n", FNAME, err))
		return err, nil
	}
	b := buf.Bytes()
	_, err = w.Write(b)
	if nil != err {
		errorLog.Printf("%s: w.Wite() failed, err: %v\n", FNAME, err)
		return err, nil
	}

	_, err = w.Write([]byte{byte(dataAccessResult)})
	if nil != err {
		errorLog.Printf("%s: w.Wite() failed, err: %v\n", FNAME, err)
		return err, nil
	}

	// raw data tag
	_, err = w.Write([]byte{0x1E})
	if nil != err {
		errorLog.Printf("%s: w.Wite() failed, err: %v\n", FNAME, err)
		return err, nil
	}

	_, err = w.Write(rawData)
	if nil != err {
		errorLog.Printf("%s: w.Wite() failed, err: %v\n", FNAME, err)
		return err, nil
	}

	return nil, w.Bytes()

}

func decode_GetResponsewithDataBlock(pdu []byte) (err error, invokeIdAndPriority tDlmsInvokeIdAndPriority, lastBlock bool, blockNumber uint32, dataAccessResult tDlmsDataAccessResult, rawData []byte) {
	var FNAME = "decode_GetResponsewithDataBlock()"
	var serr string
	b := pdu[0:]

	if len(b) < 2 {
		serr = fmt.Sprintf("%s: short pdu", FNAME)
		errorLog.Println(serr)
		return errors.New(serr), 0, false, 0, 0, nil
	}
	if !bytes.Equal(b[0:2], []byte{0xC4, 0x02}) {
		serr = fmt.Sprintf("%s: pdu is not GetResponsewithDataBlock: 0x%02X 0x%02X ", FNAME, b[0], b[1])
		errorLog.Println(serr)
		return errors.New(serr), 0, false, 0, 0, nil
	}
	b = b[2:]

	if len(b) < 1 {
		serr = fmt.Sprintf("%s: short pdu", FNAME)
		errorLog.Println(serr)
		return errors.New(serr), 0, false, 0, 0, nil
	}
	invokeIdAndPriority = tDlmsInvokeIdAndPriority(b[0])
	b = b[1:]

	if len(b) < 1 {
		serr = fmt.Sprintf("%s: short pdu", FNAME)
		errorLog.Println(serr)
		return errors.New(serr), 0, false, 0, 0, nil
	}
	if 0 == b[0] {
		lastBlock = false
	} else {
		lastBlock = true
	}
	b = b[1:]

	if len(b) < 4 {
		serr = fmt.Sprintf("%s: short pdu", FNAME)
		errorLog.Println(serr)
		return errors.New(serr), 0, false, 0, 0, nil
	}
	err = binary.Read(bytes.NewBuffer(b[0:4]), binary.BigEndian, &blockNumber)
	if nil != err {
		errorLog.Println("%s: binary.Read() failed, err: %v", err)
		return err, 0, false, 0, 0, nil
	}
	b = b[4:]

	if len(b) < 1 {
		serr = fmt.Sprintf("%s: short pdu", FNAME)
		errorLog.Println(serr)
		return errors.New(serr), 0, false, 0, 0, nil
	}
	dataAccessResult = tDlmsDataAccessResult(b[0])
	b = b[1:]

	if len(b) < 1 {
		serr = fmt.Sprintf("%s: short pdu", FNAME)
		errorLog.Println(serr)
		return errors.New(serr), 0, false, 0, 0, nil
	}
	tag := b[0]
	b = b[1:]

	if 0x1E != tag {
		serr = fmt.Sprintf("%s: wrong raw data tag: 0X%02X", FNAME)
		errorLog.Println(serr)
		return errors.New(serr), 0, false, 0, 0, nil
	}

	rawData = b

	return nil, invokeIdAndPriority, lastBlock, blockNumber, dataAccessResult, rawData
}

func encode_GetRequestForNextDataBlock(invokeIdAndPriority tDlmsInvokeIdAndPriority, blockNumber uint32) (err error, pdu []byte) {
	var FNAME = "encode_GetRequestForNextDataBlock()"

	var w bytes.Buffer

	_, err = w.Write([]byte{0xc0, 0x02})
	if nil != err {
		errorLog.Printf("%s: w.Wite() failed, err: %v\n", FNAME, err)
		return err, nil
	}

	_, err = w.Write([]byte{byte(invokeIdAndPriority)})
	if nil != err {
		errorLog.Printf("%s: w.Wite() failed, err: %v\n", FNAME, err)
		return err, nil
	}

	var buf bytes.Buffer
	err = binary.Write(&buf, binary.BigEndian, blockNumber)
	if nil != err {
		errorLog.Printf(fmt.Sprintf("%s: binary.Write() failed, err: %s\n", err))
		return err, nil
	}
	b := buf.Bytes()
	_, err = w.Write(b)
	if nil != err {
		errorLog.Printf("%s: w.Wite() failed, err: %v\n", FNAME, err)
		return err, nil
	}

	return nil, w.Bytes()
}

//@@@@@@@@@@@@@@@@@@@@@@@
func decode_GetRequestForNextDataBlock(pdu []byte) (err error, invokeIdAndPriority tDlmsInvokeIdAndPriority, blockNumber uint32) {
	var FNAME = "decode_GetRequestForNextDataBlock()"
	var serr string
	b := pdu[0:]

	if len(b) < 2 {
		serr = fmt.Sprintf("%s: short pdu", FNAME)
		errorLog.Println(serr)
		return errors.New(serr), 0, 0
	}
	if !bytes.Equal(b[0:2], []byte{0xc0, 0x02}) {
		serr = fmt.Sprintf("%s: pdu is not GetRequestForNextDataBlock: 0x%02X 0x%02X ", FNAME, b[0], b[1])
		errorLog.Println(serr)
		return errors.New(serr), 0, 0
	}
	b = b[2:]

	if len(b) < 1 {
		serr = fmt.Sprintf("%s: short pdu", FNAME)
		errorLog.Println(serr)
		return errors.New(serr), 0, 0
	}
	invokeIdAndPriority = tDlmsInvokeIdAndPriority(b[0])
	b = b[1:]

	if len(b) < 4 {
		serr = fmt.Sprintf("%s: short pdu", FNAME)
		errorLog.Println(serr)
		return errors.New(serr), 0, 0
	}
	err = binary.Read(bytes.NewBuffer(b[0:4]), binary.BigEndian, &blockNumber)
	if nil != err {
		errorLog.Println("%s: binary.Read() failed, err: %v", err)
		return err, 0, 0
	}
	b = b[4:]

	return nil, invokeIdAndPriority, blockNumber
}

const (
	COSEM_lowest_level_security_mechanism_name           = uint(0)
	COSEM_low_level_security_mechanism_name              = uint(1)
	COSEM_high_level_security_mechanism_name             = uint(2)
	COSEM_high_level_security_mechanism_name_using_MD5   = uint(3)
	COSEM_high_level_security_mechanism_name_using_SHA_1 = uint(4)
	COSEM_High_Level_Security_Mechanism_Name_Using_GMAC  = uint(5)
)

const (
	Logical_Name_Referencing_No_Ciphering   = uint(1)
	Short_Name_Referencing_No_Ciphering     = uint(2)
	Logical_Name_Referencing_With_Ciphering = uint(3)
	Short_Name_Referencing_With_Ciphering   = uint(4)
)

const (
	ACSE_Requirements_authentication = byte(0x80) // bit 0
)

const (
	Transport_HLDC = int(1)
	Transport_UDP  = int(2)
	Transport_TCP  = int(3)
)

type DlmsChannelMessage struct {
	err  error
	data interface{}
}

type DlmsChannel chan *DlmsChannelMessage

type tWrapperHeader struct {
	ProtocolVersion uint16
	SrcWport        uint16
	DstWport        uint16
	DataLength      uint16
}

func (header *tWrapperHeader) String() string {
	return fmt.Sprintf("tWrapperHeader {protocolVersion: %d, srcWport: %s, dstWport: %d, dataLength: %d}")
}

type DlmsConn struct {
	closed        bool
	rwc           io.ReadWriteCloser
	transportType int
	ch            DlmsChannel // channel used to serialize inbound requests
}

type DlmsTransportSendRequest struct {
	ch                DlmsChannel // reply channel
	applicationClient uint16
	logicalDevice     uint16
	pdu               []byte
}

type DlmsTransportReceiveRequest struct {
	ch DlmsChannel // reply channel
}

var ErrorDlmsTimeout = errors.New("ErrorDlmsTimeout")

func makeWpdu(applicationClient uint16, logicalDevice uint16, pdu []byte) (err error, wpdu []byte) {
	var (
		FNAME  string = "makeWpdu()"
		buf    bytes.Buffer
		header tWrapperHeader
	)

	header.ProtocolVersion = 0x00001
	header.SrcWport = applicationClient
	header.DstWport = logicalDevice
	header.DataLength = uint16(len(pdu))

	err = binary.Write(&buf, binary.BigEndian, &header)
	if nil != err {
		errorLog.Printf("%s:  binary.Write() failed, err: %v\n", FNAME, err)
		return err, nil
	}
	_, err = buf.Write(pdu)
	if nil != err {
		errorLog.Printf("%s:  binary.Write() failed, err: %v\n", FNAME, err)
		return err, nil
	}
	return nil, buf.Bytes()

}

func ipTransportSend(ch DlmsChannel, rwc io.ReadWriteCloser, applicationClient uint16, logicalDevice uint16, pdu []byte) {
	go func() {
		var (
			FNAME string = "ipTransportSend()"
		)

		err, wpdu := makeWpdu(applicationClient, logicalDevice, pdu)
		if nil != err {
			ch <- &DlmsChannelMessage{err, nil}
			return
		}
		debugLog.Printf("%s: sending: %02X\n", FNAME, wpdu)
		_, err = rwc.Write(wpdu)
		if nil != err {
			errorLog.Printf("%s: io.Write() failed, err: %v\n", FNAME, err)
			ch <- &DlmsChannelMessage{err, nil}
			return
		}
		debugLog.Printf("%s: sending: ok", FNAME)
		ch <- &DlmsChannelMessage{nil, nil}
	}()
}

// Never call this method directly or else you risk race condtitions on io.Writer() in case of paralell call.
// Use instead proxy variant 'transportSend()' which queues this method call on sync channel.

func (dconn *DlmsConn) doTransportSend(ch DlmsChannel, applicationClient uint16, logicalDevice uint16, pdu []byte) {
	go func() {
		var (
			FNAME string = "doTransportSend()"
		)

		debugLog.Printf("%s: trnascport type: %d, applicationClient: %d, logicalDevice: %d\n", FNAME, dconn.transportType, applicationClient, logicalDevice)

		if (Transport_TCP == dconn.transportType) || (Transport_UDP == dconn.transportType) {
			ipTransportSend(ch, dconn.rwc, applicationClient, logicalDevice, pdu)
		} else {
			panic(fmt.Sprintf("%s: unsupported transport type: %d", FNAME, dconn.transportType))
		}
	}()
}

func (dconn *DlmsConn) transportSend(ch DlmsChannel, applicationClient uint16, logicalDevice uint16, pdu []byte) {
	go func() {
		msg := new(DlmsChannelMessage)

		data := new(DlmsTransportSendRequest)
		data.ch = ch
		data.applicationClient = applicationClient
		data.logicalDevice = logicalDevice
		data.pdu = pdu

		msg.data = data

		dconn.ch <- msg

	}()
}

func readLength(r io.Reader, length int) (err error, data []byte) {
	var (
		buf bytes.Buffer
		n   int
	)

	p := make([]byte, length)
	for {
		n, err = r.Read(p[0 : length-buf.Len()])
		if n > 0 {
			buf.Write(p[0:n])
			if length == buf.Len() {
				return nil, buf.Bytes()
			} else if length < buf.Len() {
				panic("assertion failed")
			} else {
				continue
			}
		} else if 0 == n {
			if nil != err {
				errorLog.Printf("%s: io.Read() failed, err: %v", err)
				return err, data
			} else {
				panic("assertion failed")
			}
		} else {
			panic("assertion failed")
		}
	}
	panic("assertion failed")
}

func ipTransportReceive(ch DlmsChannel, rwc io.ReadWriteCloser) {
	go func() {
		var (
			FNAME     string = "ipTransportReceive()"
			serr      string
			err       error
			headerPdu []byte
			header    tWrapperHeader
		)

		err, headerPdu = readLength(rwc, int(unsafe.Sizeof(header)))
		if nil != err {
			ch <- &DlmsChannelMessage{err, nil}
			return
		}
		debugLog.Printf("%s: receiving pdu ...\n", FNAME)
		err = binary.Read(bytes.NewBuffer(headerPdu), binary.BigEndian, &header)
		if nil != err {
			errorLog.Printf("%s: binary.Read() failed, err: %v\n", FNAME, err)
			ch <- &DlmsChannelMessage{err, nil}
			return
		}
		debugLog.Printf("%s: header: ok\n", FNAME)
		if header.DataLength <= 0 {
			serr = fmt.Sprintf("%s: wrong pdu length: %d", FNAME, header.DataLength)
			errorLog.Println(serr)
			ch <- &DlmsChannelMessage{errors.New(serr), nil}
			return
		}
		err, pdu := readLength(rwc, int(header.DataLength))
		if nil != err {
			ch <- &DlmsChannelMessage{err, nil}
			return
		}
		debugLog.Printf("%s: pdu: %02X\n", FNAME, pdu)
		ch <- &DlmsChannelMessage{nil, pdu}
		return
	}()

}

// Never call this method directly or else you risk race condtitions on io.Writer() in case of paralell call.
// Use instead proxy variant 'transportReceive()' which queues this method call on sync channel.

func (dconn *DlmsConn) doTransportReceive(ch DlmsChannel) {
	go func() {
		var (
			FNAME string = "transportRecive()"
			serr  string
		)

		debugLog.Printf("%s: trnascport type: %d\n", FNAME, dconn.transportType)

		if (Transport_TCP == dconn.transportType) || (Transport_UDP == dconn.transportType) {

			ipTransportReceive(ch, dconn.rwc)

		} else {
			serr = fmt.Sprintf("%s: unsupported transport type: %d", FNAME, dconn.transportType)
			errorLog.Println(serr)
			ch <- &DlmsChannelMessage{errors.New(serr), nil}
			return
		}
	}()
}

func (dconn *DlmsConn) transportReceive(ch DlmsChannel) {
	go func() {
		data := new(DlmsTransportReceiveRequest)
		data.ch = ch
		msg := new(DlmsChannelMessage)
		msg.data = data
		dconn.ch <- msg
	}()
}

func (dconn *DlmsConn) handleTransportRequests() {
	go func() {
		var (
			FNAME string = "DlmsConn.handleTransportRequests()"
			serr  string
		)

		debugLog.Printf("%s: start\n", FNAME)
		for msg := range dconn.ch {
			debugLog.Printf("%s: message\n", FNAME)
			switch v := msg.data.(type) {
			case *DlmsTransportSendRequest:
				debugLog.Printf("%s: send request\n", FNAME)
				if dconn.closed {
					serr = fmt.Sprintf("%s: tansport send request ignored, transport connection closed", FNAME)
					errorLog.Println(serr)
					v.ch <- &DlmsChannelMessage{errors.New(serr), nil}
				}
				dconn.doTransportSend(v.ch, v.applicationClient, v.logicalDevice, v.pdu)
			case *DlmsTransportReceiveRequest:
				debugLog.Printf("%s: receive request\n", FNAME)
				if dconn.closed {
					serr = fmt.Sprintf("%s: transport receive request ignored, transport connection closed", FNAME)
					errorLog.Println(serr)
					v.ch <- &DlmsChannelMessage{errors.New(serr), nil}
				}
				dconn.doTransportReceive(v.ch)
			default:
				panic(fmt.Sprintf("unknown request type: %T", v))
			}
		}
		debugLog.Printf("%s: finish\n", FNAME)
		dconn.rwc.Close()
	}()
}

func (dconn *DlmsConn) AppConnectWithPassword(ch DlmsChannel, msecTimeout int64, applicationClient uint16, logicalDevice uint16, password string) {
	go func() {
		var (
			serr string
			err  error
			aarq AARQapdu
			pdu  []byte
		)

		_ch := make(DlmsChannel)

		go func() {
			__ch := make(DlmsChannel)

			aarq.applicationContextName = tAsn1ObjectIdentifier([]uint{2, 16, 756, 5, 8, 1, Logical_Name_Referencing_No_Ciphering})
			aarq.senderAcseRequirements = &tAsn1BitString{
				buf:        []byte{ACSE_Requirements_authentication},
				bitsUnused: 7,
			}
			mechanismName := (tAsn1ObjectIdentifier)([]uint{2, 16, 756, 5, 8, 2, COSEM_low_level_security_mechanism_name})
			aarq.mechanismName = &mechanismName
			aarq.callingAuthenticationValue = new(tAsn1Choice)
			_password := tAsn1GraphicString([]byte(password))
			aarq.callingAuthenticationValue.setVal(int(C_Authentication_value_PR_charstring), &_password)

			//TODO A-XDR encoding of userInformation
			userInformation := tAsn1OctetString([]byte{0x01, 0x00, 0x00, 0x00, 0x06, 0x5F, 0x1F, 0x04, 0x00, 0x00, 0x7E, 0x1F, 0x04, 0xB0})

			aarq.userInformation = &userInformation

			err, pdu = encode_AARQapdu(&aarq)
			if nil != err {
				_ch <- &DlmsChannelMessage{err, nil}
				return
			}

			dconn.transportSend(__ch, applicationClient, logicalDevice, pdu)
			msg := <-__ch
			if nil != msg.err {
				_ch <- &DlmsChannelMessage{msg.err, nil}
				return
			}
			dconn.transportReceive(__ch)
			msg = <-__ch
			if nil != msg.err {
				_ch <- &DlmsChannelMessage{msg.err, nil}
				return
			}
			err, aare := decode_AAREapdu((msg.data).([]byte))
			if nil != err {
				_ch <- &DlmsChannelMessage{msg.err, nil}
				return
			}
			if C_Association_result_accepted != int(aare.result) {
				serr = fmt.Sprintf("%s: app connect failed, aare.result %d, aare.resultSourceDiagnostic: %d", aare.result, aare.resultSourceDiagnostic)
				debugLog.Println(serr)
				_ch <- &DlmsChannelMessage{errors.New(serr), nil}
				return
			} else {
				_ch <- &DlmsChannelMessage{nil, nil}
			}

		}()

		select {
		case msg := <-_ch:
			if nil == msg.err {
				aconn := NewAppConn(dconn, applicationClient, logicalDevice)
				ch <- &DlmsChannelMessage{msg.err, aconn}
			} else {
				ch <- &DlmsChannelMessage{msg.err, nil}
			}
		case <-time.After(time.Millisecond * time.Duration(msecTimeout)):
			ch <- &DlmsChannelMessage{ErrorDlmsTimeout, nil}
		}
	}()

}

func TcpConnect(ch DlmsChannel, msecTimeout int64, ipAddr string, port int) {
	go func() {
		var (
			FNAME string = "connectTCP()"
			conn  net.Conn
			err   error
		)

		dconn := new(DlmsConn)
		dconn.closed = false
		dconn.ch = make(DlmsChannel)
		dconn.transportType = Transport_TCP

		_ch := make(DlmsChannel)
		go func() {

			debugLog.Printf("%s: connecting tcp transport: %s:%d\n", FNAME, ipAddr, port)
			conn, err = net.Dial("tcp", fmt.Sprintf("%s:%d", ipAddr, port))
			if nil != err {
				errorLog.Printf("%s: net.Dial() failed, err: %v", FNAME, err)
				_ch <- &DlmsChannelMessage{err, nil}
				return
			}
			dconn.rwc = conn
			_ch <- &DlmsChannelMessage{nil, nil}
		}()

		select {
		case msg := <-_ch:
			if nil == msg.err {
				debugLog.Printf("%s: tcp transport connected: %s:%d\n", FNAME, ipAddr, port)
				dconn.handleTransportRequests()
				ch <- &DlmsChannelMessage{nil, dconn}
			} else {
				debugLog.Printf("%s: tcp transport connection failed: %s:%d, err: %v\n", FNAME, ipAddr, port, msg.err)
				ch <- &DlmsChannelMessage{msg.err, msg.data}
			}
		case <-time.After(time.Millisecond * time.Duration(msecTimeout)):
			errorLog.Printf("%s: tcp transport connection time out: %s:%d\n", FNAME, ipAddr, port)
			ch <- &DlmsChannelMessage{ErrorDlmsTimeout, nil}
		}
	}()

}

func (dconn *DlmsConn) Close() {
	if dconn.closed {
		return
	}
	dconn.closed = true
	close(dconn.ch)
}
