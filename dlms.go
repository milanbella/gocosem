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
	DATA_TYPE_NULL                 uint8 = 0
	DATA_TYPE_ARRAY                uint8 = 1
	DATA_TYPE_STRUCTURE            uint8 = 2
	DATA_TYPE_BOOLEAN              uint8 = 3
	DATA_TYPE_BIT_STRING           uint8 = 4
	DATA_TYPE_DOUBLE_LONG          uint8 = 5
	DATA_TYPE_DOUBLE_LONG_UNSIGNED uint8 = 6
	DATA_TYPE_FLOATING_POINT       uint8 = 7
	DATA_TYPE_OCTET_STRING         uint8 = 9
	DATA_TYPE_VISIBLE_STRING       uint8 = 10
	DATA_TYPE_BCD                  uint8 = 13
	DATA_TYPE_INTEGER              uint8 = 15
	DATA_TYPE_LONG                 uint8 = 16
	DATA_TYPE_UNSIGNED             uint8 = 17
	DATA_TYPE_LONG_UNSIGNED        uint8 = 18
	DATA_TYPE_LONG64               uint8 = 20
	DATA_TYPE_UNSIGNED_LONG64      uint8 = 21
	DATA_TYPE_ENUM                 uint8 = 22
	DATA_TYPE_REAL32               uint8 = 23
	DATA_TYPE_REAL64               uint8 = 24
	DATA_TYPE_DATETIME             uint8 = 25
	DATA_TYPE_DATE                 uint8 = 26
	DATA_TYPE_TIME                 uint8 = 27
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

type tDlmsData struct {
	typ uint8
	val interface{}
	len uint16
	sct map[string]*tDlmsData // structure
	arr []*tDlmsData          // array
}

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

func encodeAxdrLength(w io.Writer, length uint16) (err error) {
	var FNAME string = "encodeAxdrLength()"
	if length <= 0x80 {
		err = binary.Write(w, binary.BigEndian, uint8(length))
		if nil != err {
			errorLog.Printf("%s: binary.Write() failed: %v", FNAME, err)
			return err
		}
		return nil
	} else if length <= 0xFF {
		err = binary.Write(w, binary.BigEndian, []uint8{0x81, uint8(length)})
		if nil != err {
			errorLog.Printf("%s: binary.Write() failed: %v", FNAME, err)
			return err
		}
		return nil
	} else {
		err = binary.Write(w, binary.BigEndian, []uint8{0x82, uint8(length & 0xFF00 >> 8), uint8(0x00FF & length)})
		if nil != err {
			errorLog.Printf("%s: binary.Write() failed: %v", FNAME, err)
			return err
		}
		return nil
	}
}

func decodeAxdrLength(r io.Reader) (err error, length uint16) {
	var (
		FNAME string = "decodeAxdrLength()"
		serr  string
		u8    uint8
		u16   uint16
	)
	err = binary.Read(r, binary.BigEndian, &u8)
	if nil != err {
		errorLog.Printf("%s: binary.Read() failed: %v", FNAME, err)
		return err, 0
	}
	if u8 <= 0x80 {
		return nil, uint16(u8)
	} else if u8 == 0x81 {
		err = binary.Read(r, binary.BigEndian, &u8)
		if nil != err {
			errorLog.Printf("%s: binary.Read() failed: %v", FNAME, err)
			return err, 0
		}
		return nil, uint16(u8)
	} else if u8 == 0x82 {
		err = binary.Read(r, binary.BigEndian, &u16)
		if nil != err {
			errorLog.Printf("%s: binary.Read() failed: %v", FNAME, err)
			return err, 0
		}
		return nil, u16
	} else {
		serr = fmt.Sprintf("%s: incorrect encoding\n", FNAME)
		errorLog.Printf(serr)
		return errors.New(serr), 0
	}
}

func (data *tDlmsData) Encode(w io.Writer) (err error) {
	var (
		FNAME string = "DlmsData.Encode()"
		serr  string
	)
	switch data.typ {
	case DATA_TYPE_NULL:
		return data.encodeNULL(w)
	case DATA_TYPE_ARRAY:
		err = binary.Write(w, binary.BigEndian, []byte{DATA_TYPE_ARRAY})
		if nil != err {
			errorLog.Printf("%s: binary.Write() failed: %v\n", FNAME, err)
			return err
		}
		err = encodeAxdrLength(w, uint16(len(data.arr)))
		if nil != err {
			return err
		}
		for i := 0; i < len(data.arr); i += 1 {
			err = data.arr[i].Encode(w)
			if nil != err {
				return err
			}
		}
	case DATA_TYPE_STRUCTURE:
		err = binary.Write(w, binary.BigEndian, []byte{DATA_TYPE_STRUCTURE})
		if nil != err {
			errorLog.Printf("binary.Write() failed: %v\n", err)
			return err
		}
		err = encodeAxdrLength(w, uint16(len(data.sct)))
		if nil != err {
			return err
		}
		for k, _ := range data.sct {
			err = data.sct[k].Encode(w)
			if nil != err {
				return err
			}
		}
	case DATA_TYPE_BOOLEAN:
		return data.encodeBoolean(w)
	case DATA_TYPE_BIT_STRING:
		return data.encodeBitString(w)
	case DATA_TYPE_DOUBLE_LONG:
		return data.encodeDoubleLong(w)
	case DATA_TYPE_DOUBLE_LONG_UNSIGNED:
		return data.encodeDoubleLongUnsigned(w)
	case DATA_TYPE_FLOATING_POINT:
		return data.encodeFloatingPoint(w)
	case DATA_TYPE_OCTET_STRING:
		return data.encodeOctetString(w)
	case DATA_TYPE_VISIBLE_STRING:
		return data.encodeVisibleString(w)
	case DATA_TYPE_BCD:
		return data.encodeBcd(w)
	case DATA_TYPE_INTEGER:
		return data.encodeInteger(w)
	case DATA_TYPE_LONG:
		return data.encodeLong(w)
	case DATA_TYPE_UNSIGNED:
		return data.encodeUnsigned(w)
	case DATA_TYPE_LONG_UNSIGNED:
		return data.encodeLongUnsigned(w)
	case DATA_TYPE_LONG64:
		return data.encodeLong64(w)
	case DATA_TYPE_UNSIGNED_LONG64:
		return data.encodeUnsignedLong64(w)
	case DATA_TYPE_ENUM:
		return data.encodeEnum(w)
	case DATA_TYPE_REAL32:
		return data.encodeReal32(w)
	case DATA_TYPE_REAL64:
		return data.encodeReal64(w)
	case DATA_TYPE_DATETIME:
		return data.encodeDateTime(w)
	case DATA_TYPE_DATE:
		return data.encodeDate(w)
	case DATA_TYPE_TIME:
		return data.encodeTime(w)
	default:
		serr = fmt.Sprintf("%s: unknown data tag: %d: %d\n", FNAME, data.typ)
		errorLog.Printf(serr)
		return errors.New(serr)
	}
	return nil
}

func (data *tDlmsData) Decode(r io.Reader) (err error) {
	var (
		FNAME string = "DlmsData.Decode()"
		serr  string
	)
	err = binary.Read(r, binary.BigEndian, &data.typ)
	if nil != err {
		errorLog.Printf("%s: binary.Read() failed: %v\n", FNAME, err)
		return err
	}
	switch data.typ {
	case DATA_TYPE_NULL:
		return data.decodeNULL()
	case DATA_TYPE_ARRAY:
		err, length := decodeAxdrLength(r)
		if nil != err {
			return err
		}
		data.arr = make([]*tDlmsData, length)
		for i := 0; i < len(data.arr); i += 1 {
			data.arr[i] = new(tDlmsData)
			err = data.arr[i].Decode(r)
			if nil != err {
				return err
			}
		}
	case DATA_TYPE_STRUCTURE:
		err, length := decodeAxdrLength(r)
		if nil != err {
			return err
		}
		data.sct = make(map[string]*tDlmsData, length)
		for i := uint16(0); i < length; i += 1 {
			_data := new(tDlmsData)
			err = _data.Decode(r)
			if nil != err {
				return err
			}
			data.sct[fmt.Sprintf("item%d", i)] = _data
		}
	case DATA_TYPE_BOOLEAN:
		return data.decodeBoolean(r)
	case DATA_TYPE_BIT_STRING:
		return data.decodeBitString(r)
	case DATA_TYPE_DOUBLE_LONG:
		return data.decodeDoubleLong(r)
	case DATA_TYPE_DOUBLE_LONG_UNSIGNED:
		return data.decodeDoubleLongUnsigned(r)
	case DATA_TYPE_FLOATING_POINT:
		return data.decodeFloatingPoint(r)
	case DATA_TYPE_OCTET_STRING:
		return data.decodeOctetString(r)
	case DATA_TYPE_VISIBLE_STRING:
		return data.decodeVisibleString(r)
	case DATA_TYPE_BCD:
		return data.decodeBcd(r)
	case DATA_TYPE_INTEGER:
		return data.decodeInteger(r)
	case DATA_TYPE_LONG:
		return data.decodeLong(r)
	case DATA_TYPE_UNSIGNED:
		return data.decodeUnsigned(r)
	case DATA_TYPE_LONG_UNSIGNED:
		return data.decodeLongUnsigned(r)
	case DATA_TYPE_LONG64:
		return data.decodeLong64(r)
	case DATA_TYPE_UNSIGNED_LONG64:
		return data.decodeUnsignedLong64(r)
	case DATA_TYPE_ENUM:
		return data.decodeEnum(r)
	case DATA_TYPE_REAL32:
		return data.decodeReal32(r)
	case DATA_TYPE_REAL64:
		return data.decodeReal64(r)
	case DATA_TYPE_DATETIME:
		return data.decodeDateTime(r)
	case DATA_TYPE_DATE:
		return data.decodeDate(r)
	case DATA_TYPE_TIME:
		return data.decodeTime(r)
	default:
		serr = fmt.Sprintf("%s: unknown data tag: %d\n", FNAME, data.typ)
		errorLog.Printf(serr)
		return errors.New(serr)
	}
	return nil
}

func (data *tDlmsData) GetType() uint8 {
	return data.typ
}

func (data *tDlmsData) SetNULL() {
	data.typ = DATA_TYPE_NULL
}

func (data *tDlmsData) encodeNULL(w io.Writer) (err error) {
	err = binary.Write(w, binary.BigEndian, []byte{DATA_TYPE_NULL})
	if nil != err {
		errorLog.Printf("binary.Write() failed: %v\n", err)
		return err
	}
	return nil
}

func (data *tDlmsData) decodeNULL() (err error) {
	return nil
}

func (data *tDlmsData) SetBoolean(b bool) {
	data.typ = DATA_TYPE_BOOLEAN
	data.val = b
}

func (data *tDlmsData) GetBool() bool {
	return data.val.(bool)
}

func (data *tDlmsData) encodeBoolean(w io.Writer) (err error) {
	if data.val.(bool) {
		err = binary.Write(w, binary.BigEndian, []byte{DATA_TYPE_BOOLEAN, 1})
	} else {
		err = binary.Write(w, binary.BigEndian, []byte{DATA_TYPE_BOOLEAN, 0})
	}
	if nil != err {
		errorLog.Printf("binary.Write() failed: %v\n", err)
		return err
	}
	return nil
}

func (data *tDlmsData) decodeBoolean(r io.Reader) (err error) {
	var b uint8
	err = binary.Read(r, binary.BigEndian, &b)
	if nil != err {
		errorLog.Printf("binary.Read() failed: %v\n", err)
		return err
	}
	data.typ = DATA_TYPE_BOOLEAN
	if 0 > b {
		data.val = true
	} else {
		data.val = false
	}
	return nil
}

func (data *tDlmsData) SetBitString(b []byte, length uint16) {
	n := length / 8
	if length%8 > 0 {
		n += 1
	}
	if len(b) != int(n) {
		panic("incorrect length")
	}
	data.typ = DATA_TYPE_BIT_STRING
	data.val = b
	data.len = n
}

func (data *tDlmsData) GetBitString() (b []byte, length uint16) {
	return data.val.([]byte), data.len
}

func (data *tDlmsData) encodeBitString(w io.Writer) (err error) {
	err = binary.Write(w, binary.BigEndian, []byte{DATA_TYPE_BIT_STRING})
	if nil != err {
		errorLog.Printf("binary.Write() failed: %v\n", err)
		return err
	}
	err = encodeAxdrLength(w, data.len)
	if nil != err {
		return err
	}
	err = binary.Write(w, binary.BigEndian, data.val.([]byte))
	if nil != err {
		errorLog.Printf("binary.Write() failed: %v", err)
		return err
	}
	return nil
}

func (data *tDlmsData) decodeBitString(r io.Reader) (err error) {
	var length uint16
	err, length = decodeAxdrLength(r)
	if nil != err {
		return err
	}
	b := make([]byte, length)
	err = binary.Read(r, binary.BigEndian, b)
	if nil != err {
		errorLog.Printf("binary.Read() failed: %v\n", err)
		return err
	}
	data.typ = DATA_TYPE_BIT_STRING
	data.val = b
	data.len = length
	return nil
}

func (data *tDlmsData) SetDoubleLong(i int32) {
	data.typ = DATA_TYPE_DOUBLE_LONG
	data.val = i
}

func (data *tDlmsData) GetDoubleLong() int32 {
	return data.val.(int32)
}

func (data *tDlmsData) encodeDoubleLong(w io.Writer) (err error) {
	err = binary.Write(w, binary.BigEndian, []byte{DATA_TYPE_DOUBLE_LONG})
	if nil != err {
		errorLog.Printf("binary.Write() failed: %v\n", err)
		return err
	}
	err = binary.Write(w, binary.BigEndian, data.val.(int32))
	if nil != err {
		errorLog.Printf("binary.Write() failed: %v\n", err)
		return err
	}
	return nil
}

func (data *tDlmsData) decodeDoubleLong(r io.Reader) (err error) {
	var i int32
	err = binary.Read(r, binary.BigEndian, &i)
	if nil != err {
		errorLog.Printf("binary.Read() failed: %v\n", err)
		return err
	}
	data.typ = DATA_TYPE_DOUBLE_LONG
	data.val = i
	return nil
}

func (data *tDlmsData) SetDoubleLongUnsigned(i uint32) {
	data.typ = DATA_TYPE_DOUBLE_LONG_UNSIGNED
	data.val = i
}

func (data *tDlmsData) GetDoubleLongUnsigned() uint32 {
	return data.val.(uint32)
}

func (data *tDlmsData) encodeDoubleLongUnsigned(w io.Writer) (err error) {
	err = binary.Write(w, binary.BigEndian, []byte{DATA_TYPE_DOUBLE_LONG_UNSIGNED})
	if nil != err {
		errorLog.Printf("binary.Write() failed: %v\n", err)
		return err
	}
	err = binary.Write(w, binary.BigEndian, data.val.(uint32))
	if nil != err {
		errorLog.Printf("binary.Write() failed: %v\n", err)
		return err
	}
	return nil
}

func (data *tDlmsData) decodeDoubleLongUnsigned(r io.Reader) (err error) {
	var i uint32
	err = binary.Read(r, binary.BigEndian, &i)
	if nil != err {
		errorLog.Printf("binary.Read() failed: %v\n", err)
		return err
	}
	data.typ = DATA_TYPE_DOUBLE_LONG_UNSIGNED
	data.val = i
	return nil
}

func (data *tDlmsData) SetFloatingPoint(f float32) {
	data.typ = DATA_TYPE_FLOATING_POINT
	data.val = f
}

func (data *tDlmsData) GetFloatingPoint() float32 {
	return data.val.(float32)
}

func (data *tDlmsData) encodeFloatingPoint(w io.Writer) (err error) {
	err = binary.Write(w, binary.BigEndian, []byte{DATA_TYPE_FLOATING_POINT})
	if nil != err {
		errorLog.Printf("binary.Write() failed: %v\n", err)
		return err
	}
	err = binary.Write(w, binary.BigEndian, data.val.(float32))
	if nil != err {
		errorLog.Printf("binary.Write() failed: %v\n", err)
		return err
	}
	return nil
}

func (data *tDlmsData) decodeFloatingPoint(r io.Reader) (err error) {
	var f float32
	err = binary.Read(r, binary.BigEndian, &f)
	if nil != err {
		errorLog.Printf("binary.Read() failed: %v\n", err)
		return err
	}
	data.typ = DATA_TYPE_FLOATING_POINT
	data.val = f
	return nil
}

func (data *tDlmsData) SetOctetString(b []byte) {
	data.typ = DATA_TYPE_OCTET_STRING
	if len(b) > 0xFFFF {
		panic("octet string too big")
	}
	data.val = b
}

func (data *tDlmsData) GetOctetString() []byte {
	return data.val.([]byte)
}

func (data *tDlmsData) encodeOctetString(w io.Writer) (err error) {
	err = binary.Write(w, binary.BigEndian, []byte{DATA_TYPE_OCTET_STRING})
	if nil != err {
		errorLog.Printf("binary.Write() failed: %v\n", err)
		return err
	}
	length := uint16(len(data.val.([]byte)))
	err = encodeAxdrLength(w, length)
	if nil != err {
		return err
	}
	err = binary.Write(w, binary.BigEndian, data.val.([]byte))
	if nil != err {
		errorLog.Printf("binary.Write() failed: %v\n", err)
		return err
	}
	return nil
}

func (data *tDlmsData) decodeOctetString(r io.Reader) (err error) {
	var length uint16
	err, length = decodeAxdrLength(r)
	if nil != err {
		return err
	}
	b := make([]byte, length)
	err = binary.Read(r, binary.BigEndian, b)
	if nil != err {
		errorLog.Printf("binary.Read() failed: %v\n", err)
		return err
	}
	data.typ = DATA_TYPE_OCTET_STRING
	data.val = b
	return nil
}

func (data *tDlmsData) SetVisibleString(b []byte) {
	if len(b) > 0xFFFF {
		panic("visible string too big")
	}
	data.typ = DATA_TYPE_VISIBLE_STRING
	data.val = b
}

func (data *tDlmsData) GetVisibleString() []byte {
	return data.val.([]byte)
}

func (data *tDlmsData) encodeVisibleString(w io.Writer) (err error) {
	err = binary.Write(w, binary.BigEndian, []byte{DATA_TYPE_VISIBLE_STRING})
	if nil != err {
		errorLog.Printf("binary.Write() failed: %v\n", err)
		return err
	}
	length := uint16(len(data.val.([]byte)))
	err = encodeAxdrLength(w, length)
	if nil != err {
		return err
	}
	err = binary.Write(w, binary.BigEndian, data.val.([]byte))
	if nil != err {
		errorLog.Printf("binary.Write() failed: %v\n", err)
		return err
	}
	return nil
}

func (data *tDlmsData) decodeVisibleString(r io.Reader) (err error) {
	var length uint16
	err, length = decodeAxdrLength(r)
	if nil != err {
		return err
	}
	b := make([]byte, length)
	err = binary.Read(r, binary.BigEndian, b)
	if nil != err {
		errorLog.Printf("binary.Read() failed: %v\n", err)
		return err
	}
	data.typ = DATA_TYPE_VISIBLE_STRING
	data.val = b
	return nil
}

func (data *tDlmsData) SetBcd(bcd int8) {
	data.typ = DATA_TYPE_BCD
	data.val = bcd
}

func (data *tDlmsData) GetBcd() int8 {
	return data.val.(int8)
}

func (data *tDlmsData) encodeBcd(w io.Writer) (err error) {
	err = binary.Write(w, binary.BigEndian, []byte{DATA_TYPE_BCD})
	if nil != err {
		errorLog.Printf("binary.Write() failed: %v\n", err)
		return err
	}
	err = binary.Write(w, binary.BigEndian, data.val.(int8))
	if nil != err {
		errorLog.Printf("binary.Write() failed: %v\n", err)
		return err
	}
	return nil
}

func (data *tDlmsData) decodeBcd(r io.Reader) (err error) {
	var bcd int8
	err = binary.Read(r, binary.BigEndian, &bcd)
	if nil != err {
		errorLog.Printf("binary.Read() failed: %v\n", err)
		return err
	}
	data.typ = DATA_TYPE_BCD
	data.val = bcd
	return nil
}

func (data *tDlmsData) SetInteger(i int8) {
	data.typ = DATA_TYPE_INTEGER
	data.val = i
}

func (data *tDlmsData) GetInteger() int8 {
	return data.val.(int8)
}

func (data *tDlmsData) encodeInteger(w io.Writer) (err error) {
	err = binary.Write(w, binary.BigEndian, []byte{DATA_TYPE_INTEGER})
	if nil != err {
		errorLog.Printf("binary.Write() failed: %v\n", err)
		return err
	}
	err = binary.Write(w, binary.BigEndian, data.val.(int8))
	if nil != err {
		errorLog.Printf("binary.Write() failed: %v\n", err)
		return err
	}
	return nil
}

func (data *tDlmsData) decodeInteger(r io.Reader) (err error) {
	var i int8
	err = binary.Read(r, binary.BigEndian, &i)
	if nil != err {
		errorLog.Printf("binary.Read() failed: %v\n", err)
		return err
	}
	data.typ = DATA_TYPE_INTEGER
	data.val = i
	return nil
}

func (data *tDlmsData) SetLong(i int16) {
	data.typ = DATA_TYPE_LONG
	data.val = i
}

func (data *tDlmsData) GetLong() int16 {
	return data.val.(int16)
}

func (data *tDlmsData) encodeLong(w io.Writer) (err error) {
	err = binary.Write(w, binary.BigEndian, []byte{DATA_TYPE_LONG})
	if nil != err {
		errorLog.Printf("binary.Write() failed: %v\n", err)
		return err
	}
	err = binary.Write(w, binary.BigEndian, data.val.(int16))
	if nil != err {
		errorLog.Printf("binary.Write() failed: %v\n", err)
		return err
	}
	return nil
}

func (data *tDlmsData) decodeLong(r io.Reader) (err error) {
	var i int16
	err = binary.Read(r, binary.BigEndian, &i)
	if nil != err {
		errorLog.Printf("binary.Read() failed: %v\n", err)
		return err
	}
	data.typ = DATA_TYPE_LONG
	data.val = i
	return nil
}

func (data *tDlmsData) SetUnsigned(i uint8) {
	data.typ = DATA_TYPE_UNSIGNED
	data.val = i
}

func (data *tDlmsData) GetUnsigned() uint8 {
	return data.val.(uint8)
}

func (data *tDlmsData) encodeUnsigned(w io.Writer) (err error) {
	err = binary.Write(w, binary.BigEndian, []byte{DATA_TYPE_UNSIGNED})
	if nil != err {
		errorLog.Printf("binary.Write() failed: %v\n", err)
		return err
	}
	err = binary.Write(w, binary.BigEndian, data.val.(uint8))
	if nil != err {
		errorLog.Printf("binary.Write() failed: %v\n", err)
		return err
	}
	return nil
}

func (data *tDlmsData) decodeUnsigned(r io.Reader) (err error) {
	var i uint8
	err = binary.Read(r, binary.BigEndian, &i)
	if nil != err {
		errorLog.Printf("binary.Read() failed: %v", err)
		return err
	}
	data.typ = DATA_TYPE_UNSIGNED
	data.val = i
	return nil
}

func (data *tDlmsData) SetLongUnsigned(i uint16) {
	data.typ = DATA_TYPE_LONG_UNSIGNED
	data.val = i
}

func (data *tDlmsData) GetLongUnsigned() uint16 {
	return data.val.(uint16)
}

func (data *tDlmsData) encodeLongUnsigned(w io.Writer) (err error) {
	err = binary.Write(w, binary.BigEndian, []byte{DATA_TYPE_LONG_UNSIGNED})
	if nil != err {
		errorLog.Printf("binary.Write() failed: %v\n", err)
		return err
	}
	err = binary.Write(w, binary.BigEndian, data.val.(uint16))
	if nil != err {
		errorLog.Printf("binary.Write() failed: %v\n", err)
		return err
	}
	return nil
}

func (data *tDlmsData) decodeLongUnsigned(r io.Reader) (err error) {
	var i uint16
	err = binary.Read(r, binary.BigEndian, &i)
	if nil != err {
		errorLog.Printf("binary.Read() failed: %v", err)
		return err
	}
	data.typ = DATA_TYPE_LONG_UNSIGNED
	data.val = i
	return nil
}

func (data *tDlmsData) SetLong64(i int64) {
	data.typ = DATA_TYPE_LONG64
	data.val = i
}

func (data *tDlmsData) GetLong64() int64 {
	return data.val.(int64)
}

func (data *tDlmsData) encodeLong64(w io.Writer) (err error) {
	err = binary.Write(w, binary.BigEndian, []byte{DATA_TYPE_LONG64})
	if nil != err {
		errorLog.Printf("binary.Write() failed: %v\n", err)
		return err
	}
	err = binary.Write(w, binary.BigEndian, data.val.(int64))
	if nil != err {
		errorLog.Printf("binary.Write() failed: %v\n", err)
		return err
	}
	return nil
}

func (data *tDlmsData) decodeLong64(r io.Reader) (err error) {
	var i int64
	err = binary.Read(r, binary.BigEndian, &i)
	if nil != err {
		errorLog.Printf("binary.Read() failed: %v\n", err)
		return err
	}
	data.typ = DATA_TYPE_LONG64
	data.val = i
	return nil
}

func (data *tDlmsData) SetUnsignedLong64(i uint64) {
	data.typ = DATA_TYPE_UNSIGNED_LONG64
	data.val = i
}

func (data *tDlmsData) GetUnsignedLong64() uint64 {
	return data.val.(uint64)
}

func (data *tDlmsData) encodeUnsignedLong64(w io.Writer) (err error) {
	err = binary.Write(w, binary.BigEndian, []byte{DATA_TYPE_UNSIGNED_LONG64})
	if nil != err {
		errorLog.Printf("binary.Write() failed: %v\n", err)
		return err
	}
	err = binary.Write(w, binary.BigEndian, data.val.(uint64))
	if nil != err {
		errorLog.Printf("binary.Write() failed: %v\n", err)
		return err
	}
	return nil
}

func (data *tDlmsData) decodeUnsignedLong64(r io.Reader) (err error) {
	var i uint64
	err = binary.Read(r, binary.BigEndian, &i)
	if nil != err {
		errorLog.Printf("binary.Read() failed: %v\n", err)
		return err
	}
	data.typ = DATA_TYPE_UNSIGNED_LONG64
	data.val = i
	return nil
}

func (data *tDlmsData) SetEnum(i uint8) {
	data.typ = DATA_TYPE_ENUM
	data.val = i
}

func (data *tDlmsData) GetEnum() uint8 {
	return data.val.(uint8)
}

func (data *tDlmsData) encodeEnum(w io.Writer) (err error) {
	err = binary.Write(w, binary.BigEndian, []byte{DATA_TYPE_ENUM})
	if nil != err {
		errorLog.Printf("binary.Write() failed: %v\n", err)
		return err
	}
	err = binary.Write(w, binary.BigEndian, data.val.(uint8))
	if nil != err {
		errorLog.Printf("binary.Write() failed: %v\n", err)
		return err
	}
	return nil
}

func (data *tDlmsData) decodeEnum(r io.Reader) (err error) {
	var i uint8
	err = binary.Read(r, binary.BigEndian, &i)
	if nil != err {
		errorLog.Printf("binary.Read() failed: %v\n", err)
		return err
	}
	data.typ = DATA_TYPE_ENUM
	data.val = i
	return nil
}

func (data *tDlmsData) SetReal32(f float32) {
	data.typ = DATA_TYPE_REAL32
	data.val = f
}

func (data *tDlmsData) GetReal32() float32 {
	return data.val.(float32)
}

func (data *tDlmsData) encodeReal32(w io.Writer) (err error) {
	err = binary.Write(w, binary.BigEndian, []byte{DATA_TYPE_REAL32})
	if nil != err {
		errorLog.Printf("binary.Write() failed: %v\n", err)
		return err
	}
	err = binary.Write(w, binary.BigEndian, data.val.(float32))
	if nil != err {
		errorLog.Printf("binary.Write() failed: %v\n", err)
		return err
	}
	return nil
}

func (data *tDlmsData) decodeReal32(r io.Reader) (err error) {
	var f float32
	err = binary.Read(r, binary.BigEndian, &f)
	if nil != err {
		errorLog.Printf("binary.Read() failed: %v\n", err)
		return err
	}
	data.typ = DATA_TYPE_REAL32
	data.val = f
	return nil
}

func (data *tDlmsData) SetReal64(f float64) {
	data.typ = DATA_TYPE_REAL64
	data.val = f
}

func (data *tDlmsData) GetReal64() float64 {
	return data.val.(float64)
}

func (data *tDlmsData) encodeReal64(w io.Writer) (err error) {
	err = binary.Write(w, binary.BigEndian, []byte{DATA_TYPE_REAL64})
	if nil != err {
		errorLog.Printf("binary.Write() failed: %v\n", err)
		return err
	}
	err = binary.Write(w, binary.BigEndian, data.val.(float64))
	if nil != err {
		errorLog.Printf("binary.Write() failed: %v\n", err)
		return err
	}
	return nil
}

func (data *tDlmsData) decodeReal64(r io.Reader) (err error) {
	var f float64
	err = binary.Read(r, binary.BigEndian, &f)
	if nil != err {
		errorLog.Printf("binary.Read() failed: %v\n", err)
		return err
	}
	data.typ = DATA_TYPE_REAL64
	data.val = f
	return nil
}

func (data *tDlmsData) SetDateTime(b []byte) {
	data.typ = DATA_TYPE_DATETIME
	if len(b) != 12 {
		panic("datetime length is not 12")
	}
	data.val = b
}

func (data *tDlmsData) GetDateTime() []byte {
	return data.val.([]byte)
}

func (data *tDlmsData) encodeDateTime(w io.Writer) (err error) {
	err = binary.Write(w, binary.BigEndian, []byte{DATA_TYPE_DATETIME})
	if nil != err {
		errorLog.Printf("binary.Write() failed: %v\n", err)
		return err
	}
	err = binary.Write(w, binary.BigEndian, data.val.([]byte))
	if nil != err {
		errorLog.Printf("binary.Write() failed: %v\n", err)
		return err
	}
	return nil
}

func (data *tDlmsData) decodeDateTime(r io.Reader) (err error) {
	b := make([]byte, 12)
	err = binary.Read(r, binary.BigEndian, b)
	if nil != err {
		errorLog.Printf("binary.Read() failed: %v\n", err)
		return err
	}
	data.typ = DATA_TYPE_DATETIME
	data.val = b
	return nil
}

func (data *tDlmsData) SetDate(b []byte) {
	data.typ = DATA_TYPE_DATE
	if len(b) != 5 {
		panic("date length is not 5")
	}
	data.val = b
}

func (data *tDlmsData) GetDate() []byte {
	return data.val.([]byte)
}

func (data *tDlmsData) encodeDate(w io.Writer) (err error) {
	err = binary.Write(w, binary.BigEndian, []byte{DATA_TYPE_DATE})
	if nil != err {
		errorLog.Printf("binary.Write() failed: %v\n", err)
		return err
	}
	err = binary.Write(w, binary.BigEndian, data.val.([]byte))
	if nil != err {
		errorLog.Printf("binary.Write() failed: %v\n", err)
		return err
	}
	return nil
}

func (data *tDlmsData) decodeDate(r io.Reader) (err error) {
	b := make([]byte, 5)
	err = binary.Read(r, binary.BigEndian, b)
	if nil != err {
		errorLog.Printf("binary.Read() failed: %v\n", err)
		return err
	}
	data.typ = DATA_TYPE_DATE
	data.val = b
	return nil
}

func (data *tDlmsData) SetTime(b []byte) {
	data.typ = DATA_TYPE_TIME
	if len(b) != 4 {
		panic("time length is not 4")
	}
	data.val = b
}

func (data *tDlmsData) GetTime() []byte {
	return data.val.([]byte)
}

func (data *tDlmsData) encodeTime(w io.Writer) (err error) {
	err = binary.Write(w, binary.BigEndian, []byte{DATA_TYPE_TIME})
	if nil != err {
		errorLog.Printf("binary.Write() failed: %v\n", err)
		return err
	}
	err = binary.Write(w, binary.BigEndian, data.val.([]byte))
	if nil != err {
		errorLog.Printf("binary.Write() failed: %v\n", err)
		return err
	}
	return nil
}

func (data *tDlmsData) decodeTime(r io.Reader) (err error) {
	b := make([]byte, 4)
	err = binary.Read(r, binary.BigEndian, b)
	if nil != err {
		errorLog.Printf("binary.Read() failed: %v\n", err)
		return err
	}
	data.typ = DATA_TYPE_TIME
	data.val = b
	return nil
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
			var buf bytes.Buffer
			err = accessParameters.Encode(&buf)
			if nil != err {
				return err, nil
			}
			ap = buf.Bytes()
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

	//TODO: Cosem green book full of garbage is not precise on how and when access selector parameters are to be encoded/decoded
	// We skip this to avoid reading too much into next item in case of decoding GetRequestWithList.
	//if len(b) >= 1 {
	if false {
		var nn int
		data := new(tDlmsData)
		buf := bytes.NewBuffer(b)
		err = data.Decode(buf)
		if nil != err {
			return err, 0, 0, nil, 0, nil, nil
		}
		accessParameters = data
		nn = len(b) - buf.Len()
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
		var buf bytes.Buffer
		err = data.Encode(&buf)
		if nil != err {
			return err, nil
		}
		_, err = w.Write(buf.Bytes())
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

	if dataAccessResult_success == dataAccessResult {
		buf := bytes.NewBuffer(b)
		data = new(tDlmsData)
		err = data.Decode(buf)
		nn = len(b) - buf.Len()
		if nil != err {
			return err, n + nn, 0, nil
		}
		n += nn
	}

	return nil, n, dataAccessResult, data
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

	if nil != rawData {
		_, err = w.Write([]byte{0x1E}) // raw data tag
		if nil != err {
			errorLog.Printf("%s: w.Wite() failed, err: %v\n", FNAME, err)
			return err, nil
		}

		_, err = w.Write(rawData)
		if nil != err {
			errorLog.Printf("%s: w.Wite() failed, err: %v\n", FNAME, err)
			return err, nil
		}
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
		serr = fmt.Sprintf("%s: wrong raw data tag: 0X%02X", FNAME, tag)
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
	Err  error
	Data interface{}
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
	ch       DlmsChannel // reply channel
	srcWport uint16
	dstWport uint16
	pdu      []byte
}

type DlmsTransportReceiveRequest struct {
	ch       DlmsChannel // reply channel
	srcWport uint16
	dstWport uint16
}

var ErrorDlmsTimeout = errors.New("ErrorDlmsTimeout")

func makeWpdu(srcWport uint16, dstWport uint16, pdu []byte) (err error, wpdu []byte) {
	var (
		FNAME  string = "makeWpdu()"
		buf    bytes.Buffer
		header tWrapperHeader
	)

	header.ProtocolVersion = 0x00001
	header.SrcWport = srcWport
	header.DstWport = dstWport
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

func ipTransportSend(ch DlmsChannel, rwc io.ReadWriteCloser, srcWport uint16, dstWport uint16, pdu []byte) {
	go func() {
		var (
			FNAME string = "ipTransportSend()"
		)

		err, wpdu := makeWpdu(srcWport, dstWport, pdu)
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

func (dconn *DlmsConn) doTransportSend(ch DlmsChannel, srcWport uint16, dstWport uint16, pdu []byte) {
	go func() {
		var (
			FNAME string = "doTransportSend()"
		)

		debugLog.Printf("%s: trnasport type: %d, srcWport: %d, dstWport: %d\n", FNAME, dconn.transportType, srcWport, dstWport)

		if (Transport_TCP == dconn.transportType) || (Transport_UDP == dconn.transportType) {
			ipTransportSend(ch, dconn.rwc, srcWport, dstWport, pdu)
		} else {
			panic(fmt.Sprintf("%s: unsupported transport type: %d", FNAME, dconn.transportType))
		}
	}()
}

func (dconn *DlmsConn) transportSend(ch DlmsChannel, srcWport uint16, dstWport uint16, pdu []byte) {
	go func() {
		msg := new(DlmsChannelMessage)

		data := new(DlmsTransportSendRequest)
		data.ch = ch
		data.srcWport = srcWport
		data.dstWport = dstWport
		data.pdu = pdu

		msg.Data = data

		dconn.ch <- msg

	}()
}

func readLength(r io.Reader, length int) (err error, data []byte) {
	var (
		FNAME string = "readLength()"
		buf   bytes.Buffer
		n     int
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
				errorLog.Printf("%s: io.Read() failed, err: %v", FNAME, err)
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

func ipTransportReceiveForApp(ch DlmsChannel, rwc io.ReadWriteCloser, srcWport uint16, dstWport uint16) {
	ipTransportReceive(ch, rwc, &srcWport, &dstWport)
}

func ipTransportReceiveForAny(ch DlmsChannel, rwc io.ReadWriteCloser) {
	ipTransportReceive(ch, rwc, nil, nil)
}

func ipTransportReceive(ch DlmsChannel, rwc io.ReadWriteCloser, srcWport *uint16, dstWport *uint16) {
	go func() {
		var (
			FNAME     string = "ipTransportReceive()"
			serr      string
			err       error
			headerPdu []byte
			header    tWrapperHeader
		)

		debugLog.Printf("%s: receiving pdu ...\n", FNAME)
		err, headerPdu = readLength(rwc, int(unsafe.Sizeof(header)))
		if nil != err {
			ch <- &DlmsChannelMessage{err, nil}
			return
		}
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
		if (nil != srcWport) && (header.SrcWport != *srcWport) {
			serr = fmt.Sprintf("%s: wrong srcWport: %d, expected: %d", FNAME, header.SrcWport, *srcWport)
			errorLog.Println(serr)
			ch <- &DlmsChannelMessage{errors.New(serr), nil}
			return
		}
		if (nil != dstWport) && (header.DstWport != *dstWport) {
			serr = fmt.Sprintf("%s: wrong dstWport: %d, expected: %d", FNAME, header.DstWport, *dstWport)
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

		// send reply
		m := make(map[string]interface{})
		m["srcWport"] = header.SrcWport
		m["dstWport"] = header.DstWport
		m["pdu"] = pdu
		ch <- &DlmsChannelMessage{nil, m}

		return
	}()

}

// Never call this method directly or else you risk race condtitions on io.Writer() in case of paralell call.
// Use instead proxy variant 'transportReceive()' which queues this method call on sync channel.

func (dconn *DlmsConn) doTransportReceive(ch DlmsChannel, srcWport uint16, dstWport uint16) {
	go func() {
		var (
			FNAME string = "doTransportReceive()"
			serr  string
		)

		debugLog.Printf("%s: trnascport type: %d\n", FNAME, dconn.transportType)

		if (Transport_TCP == dconn.transportType) || (Transport_UDP == dconn.transportType) {

			ipTransportReceiveForApp(ch, dconn.rwc, srcWport, srcWport)

		} else {
			serr = fmt.Sprintf("%s: unsupported transport type: %d", FNAME, dconn.transportType)
			errorLog.Println(serr)
			ch <- &DlmsChannelMessage{errors.New(serr), nil}
			return
		}
	}()
}

func (dconn *DlmsConn) transportReceive(ch DlmsChannel, srcWport uint16, dstWport uint16) {
	go func() {
		data := new(DlmsTransportReceiveRequest)
		data.ch = ch
		data.srcWport = srcWport
		data.dstWport = dstWport
		msg := new(DlmsChannelMessage)
		msg.Data = data
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
			switch v := msg.Data.(type) {
			case *DlmsTransportSendRequest:
				debugLog.Printf("%s: send request\n", FNAME)
				if dconn.closed {
					serr = fmt.Sprintf("%s: tansport send request ignored, transport connection closed", FNAME)
					errorLog.Println(serr)
					v.ch <- &DlmsChannelMessage{errors.New(serr), nil}
				}
				dconn.doTransportSend(v.ch, v.srcWport, v.dstWport, v.pdu)
			case *DlmsTransportReceiveRequest:
				debugLog.Printf("%s: receive request\n", FNAME)
				if dconn.closed {
					serr = fmt.Sprintf("%s: transport receive request ignored, transport connection closed", FNAME)
					errorLog.Println(serr)
					v.ch <- &DlmsChannelMessage{errors.New(serr), nil}
				}
				dconn.doTransportReceive(v.ch, v.srcWport, v.dstWport)
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
			FNAME string = "AppConnectWithPassword"
			serr  string
			err   error
			aarq  AARQapdu
			pdu   []byte
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
			if nil != msg.Err {
				_ch <- &DlmsChannelMessage{msg.Err, nil}
				return
			}
			dconn.transportReceive(__ch, logicalDevice, applicationClient)
			msg = <-__ch
			if nil != msg.Err {
				_ch <- &DlmsChannelMessage{msg.Err, nil}
				return
			}
			m := msg.Data.(map[string]interface{})
			if m["srcWport"] != logicalDevice {
				serr = fmt.Sprintf("%s: incorret srcWport in received pdu: ", FNAME, m["srcWport"])
				errorLog.Println(serr)
				_ch <- &DlmsChannelMessage{errors.New(serr), nil}
				return
			}
			if m["dstWport"] != applicationClient {
				serr = fmt.Sprintf("%s: incorret dstWport in received pdu: ", FNAME, m["dstWport"])
				errorLog.Println(serr)
				_ch <- &DlmsChannelMessage{errors.New(serr), nil}
				return
			}
			err, aare := decode_AAREapdu((m["pdu"]).([]byte))
			if nil != err {
				_ch <- &DlmsChannelMessage{err, nil}
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
			if nil == msg.Err {
				aconn := NewAppConn(dconn, applicationClient, logicalDevice)
				ch <- &DlmsChannelMessage{msg.Err, aconn}
			} else {
				ch <- &DlmsChannelMessage{msg.Err, nil}
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
			if nil == msg.Err {
				debugLog.Printf("%s: tcp transport connected: %s:%d\n", FNAME, ipAddr, port)
				dconn.handleTransportRequests()
				ch <- &DlmsChannelMessage{nil, dconn}
			} else {
				debugLog.Printf("%s: tcp transport connection failed: %s:%d, err: %v\n", FNAME, ipAddr, port, msg.Err)
				ch <- &DlmsChannelMessage{msg.Err, msg.Data}
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
