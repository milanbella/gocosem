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

type DlmsClassId uint16
type DlmsOid [6]uint8
type DlmsAttributeId uint8
type DlmsAccessSelector uint8

type DlmsData struct {
	Err error
	Typ uint8
	Val interface{}
	Len uint16
	Arr []*DlmsData // array
}

type DlmsDataAccessResult uint8

type DlmsDate struct {
	Year       uint16
	Month      uint8
	DayOfMonth uint8
	DayOfWeek  uint8
}

type DlmsTime struct {
	Hour       uint8
	Minute     uint8
	Second     uint8
	Hundredths uint8
}

type DlmsDateTime struct {
	DlmsDate
	DlmsTime
	Deviation   int16
	ClockStatus uint8
}

var errorLog *log.Logger = getErrorLogger()
var debugLog *log.Logger = getDebugLogger()

func DlmsDateFromBytes(b []byte) (date *DlmsDate) {
	date = new(DlmsDate)
	by := (*[2]byte)(unsafe.Pointer(&date.Year))
	by[0] = b[0]
	by[1] = b[1]
	date.Month = b[2]
	date.DayOfMonth = b[3]
	date.DayOfWeek = b[4]
	return date
}

func (date *DlmsDate) toBytes() []byte {
	b := make([]byte, 5)
	b[0] = byte((date.Year & 0xFF00) >> 8)
	b[1] = byte(date.Year & 0x00FF)
	b[2] = date.Month
	b[3] = date.DayOfMonth
	b[4] = date.DayOfWeek
	return b
}

func (date *DlmsDate) setYearWildcard() {
	date.Year = 0xFFFF
}

func (date *DlmsDate) isYearWildcard() bool {
	return date.Year == 0xFFFF
}

func (date *DlmsDate) setMonthWildcard() {
	date.Month = 0xFF
}

func (date *DlmsDate) isMonthWildcard() bool {
	return date.Month == 0xFF
}

func (date *DlmsDate) setDaylightSavingsEnd() {
	date.Month = 0xFD
}

func (date *DlmsDate) isDaylightSavingsEnd() bool {
	return date.Month == 0xFD
}

func (date *DlmsDate) setDaylightSavingsBegin() {
	date.Month = 0xFE
}

func (date *DlmsDate) isDaylightSavingsBegin() bool {
	return date.Month == 0xFE
}

func (date *DlmsDate) setDayOfWeekWildcard() {
	date.DayOfWeek = 0xFF
}

func (date *DlmsDate) isDayOfWeekWildcard() bool {
	return date.DayOfWeek == 0xFF
}

func DlmsTimeFromBytes(b []byte) (tim *DlmsTime) {
	tim = new(DlmsTime)
	tim.Hour = b[0]
	tim.Minute = b[1]
	tim.Second = b[2]
	tim.Hundredths = b[3]
	return tim
}

func (tim *DlmsTime) toBytes() []byte {
	b := make([]byte, 4)
	b[0] = tim.Hour
	b[1] = tim.Minute
	b[2] = tim.Second
	b[3] = tim.Hundredths
	return b
}

func (tim *DlmsTime) setHourWildcard() {
	tim.Hour = 0xFF
}

func (tim *DlmsTime) isHourWildcard() bool {
	return tim.Hour == 0xFF
}

func (tim *DlmsTime) setMinuteWildcard() {
	tim.Minute = 0xFF
}

func (tim *DlmsTime) isMinuteWildcard() bool {
	return tim.Minute == 0xFF
}

func (tim *DlmsTime) setSecondWildcard() {
	tim.Second = 0xFF
}

func (tim *DlmsTime) isSecondWildcard() bool {
	return tim.Second == 0xFF
}

func (tim *DlmsTime) setHundredthsWildcard() {
	tim.Hundredths = 0xFF
}

func (tim *DlmsTime) isHundredthsWildcard() bool {
	return tim.Hundredths == 0xFF
}

func DlmsDateTimeFromBytes(b []byte) (dateTime *DlmsDateTime) {

	dateTime = new(DlmsDateTime)
	b2 := (*[2]byte)(unsafe.Pointer(&dateTime.Year))
	b2[0] = b[0]
	b2[1] = b[1]
	dateTime.Month = b[2]
	dateTime.DayOfMonth = b[3]
	dateTime.DayOfWeek = b[4]
	dateTime.Hour = b[5]
	dateTime.Minute = b[6]
	dateTime.Second = b[7]
	dateTime.Hundredths = b[8]
	b2 = (*[2]byte)(unsafe.Pointer(&dateTime.Deviation))
	b2[0] = b[9]
	b2[1] = b[10]
	dateTime.ClockStatus = b[11]

	return dateTime
}

func (dateTime *DlmsDateTime) toBytes() []byte {
	b := make([]byte, 12)
	b2 := (*[2]byte)(unsafe.Pointer(&dateTime.Year))
	b[0] = b2[0]
	b[1] = b2[1]
	b[2] = dateTime.Month
	b[3] = dateTime.DayOfMonth
	b[4] = dateTime.DayOfWeek
	b[5] = dateTime.Hour
	b[6] = dateTime.Minute
	b[7] = dateTime.Second
	b[8] = dateTime.Hundredths
	b2 = (*[2]byte)(unsafe.Pointer(&dateTime.Deviation))
	b[9] = b[0]
	b[10] = b[1]
	b[11] = dateTime.ClockStatus
	return b
}

func (dateTime *DlmsDateTime) setDeviationWildcard() {
	b := (*[2]byte)(unsafe.Pointer(&dateTime.Deviation))
	b[0] = 0x80
	b[1] = 0x00
}

func (dateTime *DlmsDateTime) isDeviationWildcard() bool {
	b := (*[2]byte)(unsafe.Pointer(&dateTime.Deviation))
	return (b[0] == 0x80) && (b[1] == 0x00)
}

func (dateTime *DlmsDateTime) setClockStatusInvalid() {
	dateTime.ClockStatus |= 0x01
}

func (dateTime *DlmsDateTime) isClockStatusInvalid() bool {
	return dateTime.ClockStatus&0x01 > 0
}

func (dateTime *DlmsDateTime) setClockStatusDoubtful() {
	dateTime.ClockStatus |= 0x02
}

func (dateTime *DlmsDateTime) isClockStatusDoubtful() bool {
	return dateTime.ClockStatus&0x02 > 0
}

func (dateTime *DlmsDateTime) setClockStatusDifferentClockBase() {
	dateTime.ClockStatus |= 0x04
}

func (dateTime *DlmsDateTime) isClockStatusDifferentClockBase() bool {
	return dateTime.ClockStatus&0x04 > 0
}

func (dateTime *DlmsDateTime) setClockStatusDaylightSavingActive() {
	dateTime.ClockStatus |= 0x80
}

func (dateTime *DlmsDateTime) isClockStatusDaylightSavingActive() bool {
	return dateTime.ClockStatus&0x80 > 0
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

func (data *DlmsData) Encode(w io.Writer) (err error) {
	var (
		FNAME string = "DlmsData.Encode()"
		serr  string
	)
	switch data.Typ {
	case DATA_TYPE_NULL:
		return data.encodeNULL(w)
	case DATA_TYPE_ARRAY, DATA_TYPE_STRUCTURE:
		err = binary.Write(w, binary.BigEndian, []byte{DATA_TYPE_ARRAY})
		if nil != err {
			data.Err = err
			errorLog.Printf("%s: binary.Write() failed: %v\n", FNAME, err)
			return err
		}
		err = encodeAxdrLength(w, uint16(len(data.Arr)))
		if nil != err {
			data.Err = err
			return err
		}
		for i := 0; i < len(data.Arr); i += 1 {
			err = data.Arr[i].Encode(w)
			if nil != err {
				data.Err = err
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
		serr = fmt.Sprintf("%s: unknown data tag: %d: %d\n", FNAME, data.Typ)
		errorLog.Printf(serr)
		err = errors.New(serr)
		data.Err = err
		return err
	}
	return nil
}

func (data *DlmsData) Decode(r io.Reader) (err error) {
	var (
		FNAME string = "DlmsData.Decode()"
		serr  string
	)
	err = binary.Read(r, binary.BigEndian, &data.Typ)
	if nil != err {
		errorLog.Printf("%s: binary.Read() failed: %v\n", FNAME, err)
		data.Err = err
		return err
	}
	switch data.Typ {
	case DATA_TYPE_NULL:
		return data.decodeNULL()
	case DATA_TYPE_ARRAY, DATA_TYPE_STRUCTURE:
		err, length := decodeAxdrLength(r)
		if nil != err {
			data.Err = err
			return err
		}
		data.Arr = make([]*DlmsData, length)
		for i := 0; i < len(data.Arr); i += 1 {
			data.Arr[i] = new(DlmsData)
			err = data.Arr[i].Decode(r)
			if nil != err {
				data.Arr = data.Arr[0:i] // cut off remaining garbage
				data.Err = err
				return err
			}
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
		serr = fmt.Sprintf("%s: unknown data tag: %d\n", FNAME, data.Typ)
		errorLog.Printf(serr)
		err = errors.New(serr)
		data.Err = err
		return err
	}
	return nil
}

func (data *DlmsData) GetType() uint8 {
	return data.Typ
}

func (data *DlmsData) SetNULL() {
	data.Typ = DATA_TYPE_NULL
}

func (data *DlmsData) encodeNULL(w io.Writer) (err error) {
	err = binary.Write(w, binary.BigEndian, []byte{DATA_TYPE_NULL})
	if nil != err {
		errorLog.Printf("binary.Write() failed: %v\n", err)
		data.Err = err
		return err
	}
	return nil
}

func (data *DlmsData) decodeNULL() (err error) {
	return nil
}

func (data *DlmsData) SetBoolean(b bool) {
	data.Typ = DATA_TYPE_BOOLEAN
	data.Val = b
}

func (data *DlmsData) GetBoolean() bool {
	return data.Val.(bool)
}

func (data *DlmsData) encodeBoolean(w io.Writer) (err error) {
	if data.Val.(bool) {
		err = binary.Write(w, binary.BigEndian, []byte{DATA_TYPE_BOOLEAN, 1})
	} else {
		err = binary.Write(w, binary.BigEndian, []byte{DATA_TYPE_BOOLEAN, 0})
	}
	if nil != err {
		errorLog.Printf("binary.Write() failed: %v\n", err)
		data.Err = err
		return err
	}
	return nil
}

func (data *DlmsData) decodeBoolean(r io.Reader) (err error) {
	var b uint8
	err = binary.Read(r, binary.BigEndian, &b)
	if nil != err {
		errorLog.Printf("binary.Read() failed: %v\n", err)
		data.Err = err
		return err
	}
	data.Typ = DATA_TYPE_BOOLEAN
	if b > 0 {
		data.Val = true
	} else {
		data.Val = false
	}
	return nil
}

func (data *DlmsData) SetBitString(b []byte, length uint16) {
	n := length / 8
	if length%8 > 0 {
		n += 1
	}
	if len(b) != int(n) {
		panic("incorrect length")
	}
	data.Typ = DATA_TYPE_BIT_STRING
	data.Val = b
	data.Len = length
}

func (data *DlmsData) GetBitString() (b []byte, length uint16) {
	return data.Val.([]byte), data.Len
}

func (data *DlmsData) encodeBitString(w io.Writer) (err error) {
	err = binary.Write(w, binary.BigEndian, []byte{DATA_TYPE_BIT_STRING})
	if nil != err {
		errorLog.Printf("binary.Write() failed: %v\n", err)
		data.Err = err
		return err
	}
	err = encodeAxdrLength(w, data.Len)
	if nil != err {
		data.Err = err
		return err
	}
	err = binary.Write(w, binary.BigEndian, data.Val.([]byte))
	if nil != err {
		errorLog.Printf("binary.Write() failed: %v", err)
		data.Err = err
		return err
	}
	return nil
}

func (data *DlmsData) decodeBitString(r io.Reader) (err error) {
	var length uint16
	err, length = decodeAxdrLength(r)
	if nil != err {
		data.Err = err
		return err
	}
	n := length / 8
	if length%8 > 0 {
		n += 1
	}
	b := make([]byte, n)
	err = binary.Read(r, binary.BigEndian, b)
	if nil != err {
		errorLog.Printf("binary.Read() failed: %v\n", err)
		data.Err = err
		return err
	}
	data.Typ = DATA_TYPE_BIT_STRING
	data.Val = b
	data.Len = length
	return nil
}

func (data *DlmsData) SetDoubleLong(i int32) {
	data.Typ = DATA_TYPE_DOUBLE_LONG
	data.Val = i
}

func (data *DlmsData) GetDoubleLong() int32 {
	return data.Val.(int32)
}

func (data *DlmsData) encodeDoubleLong(w io.Writer) (err error) {
	err = binary.Write(w, binary.BigEndian, []byte{DATA_TYPE_DOUBLE_LONG})
	if nil != err {
		errorLog.Printf("binary.Write() failed: %v\n", err)
		data.Err = err
		return err
	}
	err = binary.Write(w, binary.BigEndian, data.Val.(int32))
	if nil != err {
		errorLog.Printf("binary.Write() failed: %v\n", err)
		data.Err = err
		return err
	}
	return nil
}

func (data *DlmsData) decodeDoubleLong(r io.Reader) (err error) {
	var i int32
	err = binary.Read(r, binary.BigEndian, &i)
	if nil != err {
		errorLog.Printf("binary.Read() failed: %v\n", err)
		data.Err = err
		return err
	}
	data.Typ = DATA_TYPE_DOUBLE_LONG
	data.Val = i
	return nil
}

func (data *DlmsData) SetDoubleLongUnsigned(i uint32) {
	data.Typ = DATA_TYPE_DOUBLE_LONG_UNSIGNED
	data.Val = i
}

func (data *DlmsData) GetDoubleLongUnsigned() uint32 {
	return data.Val.(uint32)
}

func (data *DlmsData) encodeDoubleLongUnsigned(w io.Writer) (err error) {
	err = binary.Write(w, binary.BigEndian, []byte{DATA_TYPE_DOUBLE_LONG_UNSIGNED})
	if nil != err {
		errorLog.Printf("binary.Write() failed: %v\n", err)
		data.Err = err
		return err
	}
	err = binary.Write(w, binary.BigEndian, data.Val.(uint32))
	if nil != err {
		errorLog.Printf("binary.Write() failed: %v\n", err)
		data.Err = err
		return err
	}
	return nil
}

func (data *DlmsData) decodeDoubleLongUnsigned(r io.Reader) (err error) {
	var i uint32
	err = binary.Read(r, binary.BigEndian, &i)
	if nil != err {
		errorLog.Printf("binary.Read() failed: %v\n", err)
		data.Err = err
		return err
	}
	data.Typ = DATA_TYPE_DOUBLE_LONG_UNSIGNED
	data.Val = i
	return nil
}

func (data *DlmsData) SetFloatingPoint(f float32) {
	data.Typ = DATA_TYPE_FLOATING_POINT
	data.Val = f
}

func (data *DlmsData) GetFloatingPoint() float32 {
	return data.Val.(float32)
}

func (data *DlmsData) encodeFloatingPoint(w io.Writer) (err error) {
	err = binary.Write(w, binary.BigEndian, []byte{DATA_TYPE_FLOATING_POINT})
	if nil != err {
		errorLog.Printf("binary.Write() failed: %v\n", err)
		data.Err = err
		return err
	}
	err = binary.Write(w, binary.BigEndian, data.Val.(float32))
	if nil != err {
		errorLog.Printf("binary.Write() failed: %v\n", err)
		data.Err = err
		return err
	}
	return nil
}

func (data *DlmsData) decodeFloatingPoint(r io.Reader) (err error) {
	var f float32
	err = binary.Read(r, binary.BigEndian, &f)
	if nil != err {
		errorLog.Printf("binary.Read() failed: %v\n", err)
		data.Err = err
		return err
	}
	data.Typ = DATA_TYPE_FLOATING_POINT
	data.Val = f
	return nil
}

func (data *DlmsData) SetOctetString(b []byte) {
	data.Typ = DATA_TYPE_OCTET_STRING
	if len(b) > 0xFFFF {
		panic("octet string too big")
	}
	data.Val = b
}

func (data *DlmsData) GetOctetString() []byte {
	return data.Val.([]byte)
}

func (data *DlmsData) encodeOctetString(w io.Writer) (err error) {
	err = binary.Write(w, binary.BigEndian, []byte{DATA_TYPE_OCTET_STRING})
	if nil != err {
		errorLog.Printf("binary.Write() failed: %v\n", err)
		data.Err = err
		return err
	}
	length := uint16(len(data.Val.([]byte)))
	err = encodeAxdrLength(w, length)
	if nil != err {
		data.Err = err
		return err
	}
	err = binary.Write(w, binary.BigEndian, data.Val.([]byte))
	if nil != err {
		errorLog.Printf("binary.Write() failed: %v\n", err)
		data.Err = err
		return err
	}
	return nil
}

func (data *DlmsData) decodeOctetString(r io.Reader) (err error) {
	var length uint16
	err, length = decodeAxdrLength(r)
	if nil != err {
		data.Err = err
		return err
	}
	b := make([]byte, length)
	err = binary.Read(r, binary.BigEndian, b)
	if nil != err {
		errorLog.Printf("binary.Read() failed: %v\n", err)
		data.Err = err
		return err
	}
	data.Typ = DATA_TYPE_OCTET_STRING
	data.Val = b
	return nil
}

func (data *DlmsData) SetVisibleString(b []byte) {
	if len(b) > 0xFFFF {
		panic("visible string too big")
	}
	data.Typ = DATA_TYPE_VISIBLE_STRING
	data.Val = b
}

func (data *DlmsData) GetVisibleString() []byte {
	return data.Val.([]byte)
}

func (data *DlmsData) encodeVisibleString(w io.Writer) (err error) {
	err = binary.Write(w, binary.BigEndian, []byte{DATA_TYPE_VISIBLE_STRING})
	if nil != err {
		errorLog.Printf("binary.Write() failed: %v\n", err)
		data.Err = err
		return err
	}
	length := uint16(len(data.Val.([]byte)))
	err = encodeAxdrLength(w, length)
	if nil != err {
		data.Err = err
		return err
	}
	err = binary.Write(w, binary.BigEndian, data.Val.([]byte))
	if nil != err {
		errorLog.Printf("binary.Write() failed: %v\n", err)
		data.Err = err
		return err
	}
	return nil
}

func (data *DlmsData) decodeVisibleString(r io.Reader) (err error) {
	var length uint16
	err, length = decodeAxdrLength(r)
	if nil != err {
		data.Err = err
		return err
	}
	b := make([]byte, length)
	err = binary.Read(r, binary.BigEndian, b)
	if nil != err {
		errorLog.Printf("binary.Read() failed: %v\n", err)
		data.Err = err
		return err
	}
	data.Typ = DATA_TYPE_VISIBLE_STRING
	data.Val = b
	return nil
}

func (data *DlmsData) SetBcd(bcd int8) {
	data.Typ = DATA_TYPE_BCD
	data.Val = bcd
}

func (data *DlmsData) GetBcd() int8 {
	return data.Val.(int8)
}

func (data *DlmsData) encodeBcd(w io.Writer) (err error) {
	err = binary.Write(w, binary.BigEndian, []byte{DATA_TYPE_BCD})
	if nil != err {
		errorLog.Printf("binary.Write() failed: %v\n", err)
		data.Err = err
		return err
	}
	err = binary.Write(w, binary.BigEndian, data.Val.(int8))
	if nil != err {
		errorLog.Printf("binary.Write() failed: %v\n", err)
		data.Err = err
		return err
	}
	return nil
}

func (data *DlmsData) decodeBcd(r io.Reader) (err error) {
	var bcd int8
	err = binary.Read(r, binary.BigEndian, &bcd)
	if nil != err {
		errorLog.Printf("binary.Read() failed: %v\n", err)
		data.Err = err
		return err
	}
	data.Typ = DATA_TYPE_BCD
	data.Val = bcd
	return nil
}

func (data *DlmsData) SetInteger(i int8) {
	data.Typ = DATA_TYPE_INTEGER
	data.Val = i
}

func (data *DlmsData) GetInteger() int8 {
	return data.Val.(int8)
}

func (data *DlmsData) encodeInteger(w io.Writer) (err error) {
	err = binary.Write(w, binary.BigEndian, []byte{DATA_TYPE_INTEGER})
	if nil != err {
		errorLog.Printf("binary.Write() failed: %v\n", err)
		data.Err = err
		return err
	}
	err = binary.Write(w, binary.BigEndian, data.Val.(int8))
	if nil != err {
		errorLog.Printf("binary.Write() failed: %v\n", err)
		data.Err = err
		return err
	}
	return nil
}

func (data *DlmsData) decodeInteger(r io.Reader) (err error) {
	var i int8
	err = binary.Read(r, binary.BigEndian, &i)
	if nil != err {
		errorLog.Printf("binary.Read() failed: %v\n", err)
		data.Err = err
		return err
	}
	data.Typ = DATA_TYPE_INTEGER
	data.Val = i
	return nil
}

func (data *DlmsData) SetLong(i int16) {
	data.Typ = DATA_TYPE_LONG
	data.Val = i
}

func (data *DlmsData) GetLong() int16 {
	return data.Val.(int16)
}

func (data *DlmsData) encodeLong(w io.Writer) (err error) {
	err = binary.Write(w, binary.BigEndian, []byte{DATA_TYPE_LONG})
	if nil != err {
		errorLog.Printf("binary.Write() failed: %v\n", err)
		data.Err = err
		return err
	}
	err = binary.Write(w, binary.BigEndian, data.Val.(int16))
	if nil != err {
		errorLog.Printf("binary.Write() failed: %v\n", err)
		data.Err = err
		return err
	}
	return nil
}

func (data *DlmsData) decodeLong(r io.Reader) (err error) {
	var i int16
	err = binary.Read(r, binary.BigEndian, &i)
	if nil != err {
		errorLog.Printf("binary.Read() failed: %v\n", err)
		data.Err = err
		return err
	}
	data.Typ = DATA_TYPE_LONG
	data.Val = i
	return nil
}

func (data *DlmsData) SetUnsigned(i uint8) {
	data.Typ = DATA_TYPE_UNSIGNED
	data.Val = i
}

func (data *DlmsData) GetUnsigned() uint8 {
	return data.Val.(uint8)
}

func (data *DlmsData) encodeUnsigned(w io.Writer) (err error) {
	err = binary.Write(w, binary.BigEndian, []byte{DATA_TYPE_UNSIGNED})
	if nil != err {
		errorLog.Printf("binary.Write() failed: %v\n", err)
		data.Err = err
		return err
	}
	err = binary.Write(w, binary.BigEndian, data.Val.(uint8))
	if nil != err {
		errorLog.Printf("binary.Write() failed: %v\n", err)
		data.Err = err
		return err
	}
	return nil
}

func (data *DlmsData) decodeUnsigned(r io.Reader) (err error) {
	var i uint8
	err = binary.Read(r, binary.BigEndian, &i)
	if nil != err {
		errorLog.Printf("binary.Read() failed: %v", err)
		data.Err = err
		return err
	}
	data.Typ = DATA_TYPE_UNSIGNED
	data.Val = i
	return nil
}

func (data *DlmsData) SetLongUnsigned(i uint16) {
	data.Typ = DATA_TYPE_LONG_UNSIGNED
	data.Val = i
}

func (data *DlmsData) GetLongUnsigned() uint16 {
	return data.Val.(uint16)
}

func (data *DlmsData) encodeLongUnsigned(w io.Writer) (err error) {
	err = binary.Write(w, binary.BigEndian, []byte{DATA_TYPE_LONG_UNSIGNED})
	if nil != err {
		errorLog.Printf("binary.Write() failed: %v\n", err)
		data.Err = err
		return err
	}
	err = binary.Write(w, binary.BigEndian, data.Val.(uint16))
	if nil != err {
		errorLog.Printf("binary.Write() failed: %v\n", err)
		data.Err = err
		return err
	}
	return nil
}

func (data *DlmsData) decodeLongUnsigned(r io.Reader) (err error) {
	var i uint16
	err = binary.Read(r, binary.BigEndian, &i)
	if nil != err {
		errorLog.Printf("binary.Read() failed: %v", err)
		data.Err = err
		return err
	}
	data.Typ = DATA_TYPE_LONG_UNSIGNED
	data.Val = i
	return nil
}

func (data *DlmsData) SetLong64(i int64) {
	data.Typ = DATA_TYPE_LONG64
	data.Val = i
}

func (data *DlmsData) GetLong64() int64 {
	return data.Val.(int64)
}

func (data *DlmsData) encodeLong64(w io.Writer) (err error) {
	err = binary.Write(w, binary.BigEndian, []byte{DATA_TYPE_LONG64})
	if nil != err {
		errorLog.Printf("binary.Write() failed: %v\n", err)
		data.Err = err
		return err
	}
	err = binary.Write(w, binary.BigEndian, data.Val.(int64))
	if nil != err {
		errorLog.Printf("binary.Write() failed: %v\n", err)
		data.Err = err
		return err
	}
	return nil
}

func (data *DlmsData) decodeLong64(r io.Reader) (err error) {
	var i int64
	err = binary.Read(r, binary.BigEndian, &i)
	if nil != err {
		errorLog.Printf("binary.Read() failed: %v\n", err)
		data.Err = err
		return err
	}
	data.Typ = DATA_TYPE_LONG64
	data.Val = i
	return nil
}

func (data *DlmsData) SetUnsignedLong64(i uint64) {
	data.Typ = DATA_TYPE_UNSIGNED_LONG64
	data.Val = i
}

func (data *DlmsData) GetUnsignedLong64() uint64 {
	return data.Val.(uint64)
}

func (data *DlmsData) encodeUnsignedLong64(w io.Writer) (err error) {
	err = binary.Write(w, binary.BigEndian, []byte{DATA_TYPE_UNSIGNED_LONG64})
	if nil != err {
		errorLog.Printf("binary.Write() failed: %v\n", err)
		data.Err = err
		return err
	}
	err = binary.Write(w, binary.BigEndian, data.Val.(uint64))
	if nil != err {
		errorLog.Printf("binary.Write() failed: %v\n", err)
		data.Err = err
		return err
	}
	return nil
}

func (data *DlmsData) decodeUnsignedLong64(r io.Reader) (err error) {
	var i uint64
	err = binary.Read(r, binary.BigEndian, &i)
	if nil != err {
		errorLog.Printf("binary.Read() failed: %v\n", err)
		data.Err = err
		return err
	}
	data.Typ = DATA_TYPE_UNSIGNED_LONG64
	data.Val = i
	return nil
}

func (data *DlmsData) SetEnum(i uint8) {
	data.Typ = DATA_TYPE_ENUM
	data.Val = i
}

func (data *DlmsData) GetEnum() uint8 {
	return data.Val.(uint8)
}

func (data *DlmsData) encodeEnum(w io.Writer) (err error) {
	err = binary.Write(w, binary.BigEndian, []byte{DATA_TYPE_ENUM})
	if nil != err {
		errorLog.Printf("binary.Write() failed: %v\n", err)
		data.Err = err
		return err
	}
	err = binary.Write(w, binary.BigEndian, data.Val.(uint8))
	if nil != err {
		errorLog.Printf("binary.Write() failed: %v\n", err)
		data.Err = err
		return err
	}
	return nil
}

func (data *DlmsData) decodeEnum(r io.Reader) (err error) {
	var i uint8
	err = binary.Read(r, binary.BigEndian, &i)
	if nil != err {
		errorLog.Printf("binary.Read() failed: %v\n", err)
		data.Err = err
		return err
	}
	data.Typ = DATA_TYPE_ENUM
	data.Val = i
	return nil
}

func (data *DlmsData) SetReal32(f float32) {
	data.Typ = DATA_TYPE_REAL32
	data.Val = f
}

func (data *DlmsData) GetReal32() float32 {
	return data.Val.(float32)
}

func (data *DlmsData) encodeReal32(w io.Writer) (err error) {
	err = binary.Write(w, binary.BigEndian, []byte{DATA_TYPE_REAL32})
	if nil != err {
		errorLog.Printf("binary.Write() failed: %v\n", err)
		data.Err = err
		return err
	}
	err = binary.Write(w, binary.BigEndian, data.Val.(float32))
	if nil != err {
		errorLog.Printf("binary.Write() failed: %v\n", err)
		data.Err = err
		return err
	}
	return nil
}

func (data *DlmsData) decodeReal32(r io.Reader) (err error) {
	var f float32
	err = binary.Read(r, binary.BigEndian, &f)
	if nil != err {
		errorLog.Printf("binary.Read() failed: %v\n", err)
		data.Err = err
		return err
	}
	data.Typ = DATA_TYPE_REAL32
	data.Val = f
	return nil
}

func (data *DlmsData) SetReal64(f float64) {
	data.Typ = DATA_TYPE_REAL64
	data.Val = f
}

func (data *DlmsData) GetReal64() float64 {
	return data.Val.(float64)
}

func (data *DlmsData) encodeReal64(w io.Writer) (err error) {
	err = binary.Write(w, binary.BigEndian, []byte{DATA_TYPE_REAL64})
	if nil != err {
		errorLog.Printf("binary.Write() failed: %v\n", err)
		data.Err = err
		return err
	}
	err = binary.Write(w, binary.BigEndian, data.Val.(float64))
	if nil != err {
		errorLog.Printf("binary.Write() failed: %v\n", err)
		data.Err = err
		return err
	}
	return nil
}

func (data *DlmsData) decodeReal64(r io.Reader) (err error) {
	var f float64
	err = binary.Read(r, binary.BigEndian, &f)
	if nil != err {
		errorLog.Printf("binary.Read() failed: %v\n", err)
		data.Err = err
		return err
	}
	data.Typ = DATA_TYPE_REAL64
	data.Val = f
	return nil
}

func (data *DlmsData) SetDateTime(b []byte) {
	data.Typ = DATA_TYPE_DATETIME
	if len(b) != 12 {
		panic("datetime length is not 12")
	}
	data.Val = b
}

func (data *DlmsData) GetDateTime() []byte {
	return data.Val.([]byte)
}

func (data *DlmsData) encodeDateTime(w io.Writer) (err error) {
	err = binary.Write(w, binary.BigEndian, []byte{DATA_TYPE_DATETIME})
	if nil != err {
		errorLog.Printf("binary.Write() failed: %v\n", err)
		data.Err = err
		return err
	}
	err = binary.Write(w, binary.BigEndian, data.Val.([]byte))
	if nil != err {
		errorLog.Printf("binary.Write() failed: %v\n", err)
		data.Err = err
		return err
	}
	return nil
}

func (data *DlmsData) decodeDateTime(r io.Reader) (err error) {
	b := make([]byte, 12)
	err = binary.Read(r, binary.BigEndian, b)
	if nil != err {
		errorLog.Printf("binary.Read() failed: %v\n", err)
		data.Err = err
		return err
	}
	data.Typ = DATA_TYPE_DATETIME
	data.Val = b
	return nil
}

func (data *DlmsData) SetDate(b []byte) {
	data.Typ = DATA_TYPE_DATE
	if len(b) != 5 {
		panic("date length is not 5")
	}
	data.Val = b
}

func (data *DlmsData) GetDate() []byte {
	return data.Val.([]byte)
}

func (data *DlmsData) encodeDate(w io.Writer) (err error) {
	err = binary.Write(w, binary.BigEndian, []byte{DATA_TYPE_DATE})
	if nil != err {
		errorLog.Printf("binary.Write() failed: %v\n", err)
		data.Err = err
		return err
	}
	err = binary.Write(w, binary.BigEndian, data.Val.([]byte))
	if nil != err {
		errorLog.Printf("binary.Write() failed: %v\n", err)
		data.Err = err
		return err
	}
	return nil
}

func (data *DlmsData) decodeDate(r io.Reader) (err error) {
	b := make([]byte, 5)
	err = binary.Read(r, binary.BigEndian, b)
	if nil != err {
		errorLog.Printf("binary.Read() failed: %v\n", err)
		data.Err = err
		return err
	}
	data.Typ = DATA_TYPE_DATE
	data.Val = b
	return nil
}

func (data *DlmsData) SetTime(b []byte) {
	data.Typ = DATA_TYPE_TIME
	if len(b) != 4 {
		panic("time length is not 4")
	}
	data.Val = b
}

func (data *DlmsData) GetTime() []byte {
	return data.Val.([]byte)
}

func (data *DlmsData) encodeTime(w io.Writer) (err error) {
	err = binary.Write(w, binary.BigEndian, []byte{DATA_TYPE_TIME})
	if nil != err {
		errorLog.Printf("binary.Write() failed: %v\n", err)
		data.Err = err
		return err
	}
	err = binary.Write(w, binary.BigEndian, data.Val.([]byte))
	if nil != err {
		errorLog.Printf("binary.Write() failed: %v\n", err)
		data.Err = err
		return err
	}
	return nil
}

func (data *DlmsData) decodeTime(r io.Reader) (err error) {
	b := make([]byte, 4)
	err = binary.Read(r, binary.BigEndian, b)
	if nil != err {
		errorLog.Printf("binary.Read() failed: %v\n", err)
		return err
	}
	data.Typ = DATA_TYPE_TIME
	data.Val = b
	return nil
}

func encode_getRequest(w io.Writer, classId DlmsClassId, instanceId *DlmsOid, attributeId DlmsAttributeId, accessSelector DlmsAccessSelector, accessParameters *DlmsData) (err error) {
	var FNAME string = "encode_getRequest()"

	err = binary.Write(w, binary.BigEndian, classId)
	if nil != err {
		errorLog.Printf(fmt.Sprintf("%s: binary.Write() failed, err: %s\n", FNAME, err))
		return err
	}

	err = binary.Write(w, binary.BigEndian, instanceId)
	if nil != err {
		errorLog.Printf(fmt.Sprintf("%s: binary.Write() failed, err: %s\n", FNAME, err))
		return err
	}

	err = binary.Write(w, binary.BigEndian, attributeId)
	if nil != err {
		errorLog.Printf(fmt.Sprintf("%s: binary.Write() failed, err: %s\n", FNAME, err))
		return err
	}

	if 0 != attributeId {
		err = binary.Write(w, binary.BigEndian, accessSelector)
		if nil != err {
			errorLog.Printf(fmt.Sprintf("%s: binary.Write() failed, err: %s\n", FNAME, err))
			return err
		}
		if nil != accessParameters {
			err = accessParameters.Encode(w)
			if nil != err {
				return err
			}
		}
	}
	return nil
}

func decode_getRequest(r io.Reader) (err error, classId DlmsClassId, instanceId *DlmsOid, attributeId DlmsAttributeId, accessSelector DlmsAccessSelector, accessParameters *DlmsData) {
	var FNAME = "decode_getRequest()"
	var serr string

	var _classId DlmsClassId
	err = binary.Read(r, binary.BigEndian, &_classId)
	if nil != err {
		errorLog.Println("%s: binary.Read() failed, err: %v", err)
		return err, 0, nil, 0, nil, nil
	}

	_instanceId := new(DlmsOid)
	err = binary.Read(r, binary.BigEndian, _instanceId)
	if nil != err {
		errorLog.Println("%s: binary.Read() failed, err: %v", err)
		return err, _classId, nil, 0, nil, nil
	}

	var _attributeId DlmsAttributeId
	err = binary.Read(r, binary.BigEndian, &_attributeId)
	if nil != err {
		errorLog.Println("%s: binary.Read() failed, err: %v", err)
		return err, _classId, _instanceId, 0, nil, nil
	}

	var _accessSelector DlmsAccessSelector
	err = binary.Read(r, binary.BigEndian, &_accessSelector)
	if nil != err {
		errorLog.Println("%s: binary.Read() failed, err: %v", err)
	}

	if accessSelector > 0 {
		data = new(DlmsData)
		err = data.Decode(&r)
		if nil != err {
			return err, _classId, _instanceId, _attributeId, _accessSelector, nil
		}
		accessParameters = data
	}
	return nil, _classId, _instanceId, _attributeId, _accessSelector, accessParameters
}

func encode_getResponse(w io.Writer, dataAccessResult DlmsDataAccessResult, data *DlmsData) (err error) {
	var FNAME = "encode_getResponse()"

	err = binary.Write(&w, binary.BigEndian, dataAccessResult)
	if nil != err {
		errorLog.Printf(fmt.Sprintf("%s: binary.Write() failed, err: %s\n", FNAME, err))
		return err
	}

	if nil != data {
		err = data.Encode(&w)
		if nil != err {
			return err
		}
	}

	return nil
}

func decode_getResponse(r io.Reader) (err error, dataAccessResult DlmsDataAccessResult, data *DlmsData) {
	var FNAME = "decode_getResponse()"
	var serr string

	r := bytes.NewBuffer(pdu)

	err = binary.Read(&r, binary.BigEndian, &dataAccessResult)
	if nil != err {
		errorLog.Println("%s: binary.Read() failed, err: %v", err)
		return err, 0, nil
	}

	data = nil
	if dataAccessResult_success == dataAccessResult {
		data = new(DlmsData)
		err = data.Decode(&r)
		if nil != err {
			return err, dataAccessResult, nil
		}
	}

	return nil, dataAccessResult, data
}

func encode_GetRequestNormal(w io.Writer, classId DlmsClassId, instanceId *DlmsOid, attributeId DlmsAttributeId, accessSelector DlmsAccessSelector, accessParameters *DlmsData) (err error) {
	var FNAME = "encode_GetRequestNormal()"

	err, pdu = encode_getRequest(w, classId, instanceId, attributeId, accessSelector, accessParameters)
	if nil != err {
		errorLog.Printf("%s: encode_getRequest() failed, err: %v\n", FNAME, err)
		return err
	}

	return nil
}

func decode_GetRequestNormal(r io.Reader) (err error, classId DlmsClassId, instanceId *DlmsOid, attributeId DlmsAttributeId, accessSelector DlmsAccessSelector, accessParameters *DlmsData) {
	var FNAME = "decode_GetRequestNormal"

	err, _, classId, instanceId, attributeId, accessSelector, accessParameters = decode_getRequest(r)
	if nil != err {
		return err, 0, nil, 0, 0, nil
	}
	return nil, classId, instanceId, attributeId, accessSelector, accessParameters
}

func encode_GetResponseNormal(w io.Writer, dataAccessResult DlmsDataAccessResult, data *DlmsData) (err error) {
	var FNAME = "encode_GetResponseNormal()"

	err, pdu = encode_getResponse(w, dataAccessResult, data)
	if nil != err {
		return err
	}

	return nil
}

func decode_GetResponseNormal(r io.Reader) (err error, dataAccessResult DlmsDataAccessResult, data *DlmsData) {

	err, dataAccessResult, data = decode_getResponse(r)
	if nil != err {
		return err, dataAccessResult, data
	}

	return nil, dataAccessResult, data
}

func encode_GetRequestWithList(w io.Writer, classIds []DlmsClassId, instanceIds []*DlmsOid, attributeIds []DlmsAttributeId, accessSelectors []DlmsAccessSelector, accessParameters []*DlmsData) (err error) {
	var FNAME = "encode_GetRequestWithList()"

	var w bytes.Buffer

	count := uint8(len(classIds)) // count of get requests

	err = binary.Write(w, binary.BigEndian, count)
	if nil != err {
		errorLog.Printf(fmt.Sprintf("%s: binary.Write() failed, err: %s\n", FNAME, err))
		return err
	}

	for i := 0; i < count; i += 1 {

		err = encode_getRequest(w, classIds[i], instanceIds[i], attributeIds[i], accessSelectors[i], accessParameters[i])
		if nil != err {
			errorLog.Printf("%s: encode_getRequest() failed, err: %v\n", FNAME, err)
			return err
		}

	}

	return nil
}

func decode_GetRequestWithList(r io.Reader) (err error, classIds []DlmsClassId, instanceIds []*DlmsOid, attributeIds []DlmsAttributeId, accessSelectors []DlmsAccessSelector, accessParameters []*DlmsData) {
	var FNAME = "decode_GetRequestWithList()"
	var serr string

	var count uint8
	err = binary.Read(r, binary.BigEndian, &count)
	if nil != err {
		errorLog.Println("%s: binary.Read() failed, err: %v", err)
		return err, nil, nil, nil, nil, nil
	}

	classIds = make([]DlmsClassId, count)
	instanceIds = make([]*DlmsOid, count)
	attributeIds = make([]DlmsAttributeId, count)
	accessSelectors = make([]DlmsAccessSelector, count)
	accessParameters = make([]*DlmsData, count)

	for i := 0; i < count; i += 1 {
		err, n, classId, instanceId, attributeId, accessSelector, accessParameter := decode_getRequest(r)
		if nil != err {
			return err, classIds[0:i], instanceIds[0:i], attributeIds[0:i], accessSelectors[0:i], accessParameters[0:i]
		}
		classIds[i] = classId
		instanceIds[i] = instanceId
		attributeIds[i] = attributeId
		accessSelectors[i] = accessSelector
		accessParameters[i] = accessParameter
	}
	return nil, classIds, instanceIds, attributeIds, accessSelectors, accessParameters
}

func encode_GetResponseWithList(w io.Writer, dataAccessResults []DlmsDataAccessResult, datas []*DlmsData) (err error) {
	var FNAME = "encode_GetResponseWithList()"

	var w bytes.Buffer

	count := uint8(len(dataAccessResults))

	err = binary.Write(w, binary.BigEndian, count)
	if nil != err {
		errorLog.Printf(fmt.Sprintf("%s: binary.Write() failed, err: %s\n", FNAME, err))
		return err
	}

	for i := 0; i < count; i += 1 {

		err := encode_getResponse(w, dataAccessResults[i], datas[i])
		if nil != err {
			return err
		}
	}

	return nil
}

func decode_GetResponseWithList(r io.Reader) (err error, dataAccessResults []DlmsDataAccessResult, datas []*DlmsData) {
	var FNAME = "decode_GetResponseWithList()"
	var serr string

	var count uint8
	err = binary.Read(&r, binary.BigEndian, &count)
	if nil != err {
		errorLog.Println("%s: binary.Read() failed, err: %v", err)
		return err, nil, nil
	}

	dataAccessResults = make([]DlmsDataAccessResult, count)
	datas = make([]*DlmsData, count)

	var dataAccessResult DlmsDataAccessResult
	var data *DlmsData
	for i := 0; i < count; i += 1 {
		err, dataAccessResult, data = decode_getResponse(r)
		if nil != err {
			return err, dataAccessResults[0:i], datas[0:i]
		}
		dataAccessResults[i] = dataAccessResult
		datas[i] = data
	}

	return nil, dataAccessResults, datas
}

func encode_GetResponsewithDataBlock(w io.Writer, lastBlock bool, blockNumber uint32, dataAccessResult DlmsDataAccessResult, rawData []byte) (err error) {
	var FNAME = "encode_GetResponsewithDataBlock()"

	var bb byte
	if lastBlock {
		bb = 1
	} else {
		bb = 0
	}
	err = binary.Write(w, binary.BigEndian, bb)
	if nil != err {
		errorLog.Printf(fmt.Sprintf("%s: binary.Write() failed, err: %s\n", FNAME, err))
		return err
	}

	err = binary.Write(w, binary.BigEndian, blockNumber)
	if nil != err {
		errorLog.Printf(fmt.Sprintf("%s: binary.Write() failed, err: %s\n", FNAME, err))
		return err
	}

	err = binary.Write(w, binary.BigEndian, dataAccessResult)
	if nil != err {
		errorLog.Printf(fmt.Sprintf("%s: binary.Write() failed, err: %s\n", FNAME, err))
		return err
	}

	if nil != rawData {
		err = encodeAxdrLength(&w, uint16(len(rawData)))
		if nil != err {
			errorLog.Printf("%s: encodeAxdrLength() failed, err: %v\n", FNAME, err)
			return err
		}

		_, err = w.Write(rawData)
		if nil != err {
			errorLog.Printf("%s: w.Wite() failed, err: %v\n", FNAME, err)
			return err
		}
	}

	return nil
}

func decode_GetResponsewithDataBlock(r io.Reader) (err error, lastBlock bool, blockNumber uint32, dataAccessResult DlmsDataAccessResult, rawData []byte) {
	var FNAME = "decode_GetResponsewithDataBlock()"
	var serr string

	var _lastBlock bool
	err = binary.Read(r, binary.BigEndian, &_lastBlock)
	if nil != err {
		errorLog.Println("%s: binary.Read() failed, err: %v", err)
		return err, false, 0, 0, nil
	}

	var _blockNumber uint32
	err = binary.Read(r, binary.BigEndian, &_blockNumber)
	if nil != err {
		errorLog.Println("%s: binary.Read() failed, err: %v", err)
		return err, _lastBlock, 0, 0, nil
	}

	var _dataAccessResult DlmsDataAccessResult
	err = binary.Read(r, binary.BigEndian, &_dataAccessResult)
	if nil != err {
		errorLog.Println("%s: binary.Read() failed, err: %v", err)
		return err, _lastBlock, _blockNumber, 0, nil
	}

	err, length := decodeAxdrLength(r)
	if nil != err {
		errorLog.Printf("%s: decodeAxdrLength() failed, err: %v\n", FNAME, err)
		return err, _lastBlock, _blockNumber, _dataAccessResult, nil
	}

	rawData = make([]byte, length)
	err = binary.Read(r, binary.BigEndian, rawData)
	if nil != err {
		errorLog.Println("%s: binary.Read() failed, err: %v", err)
		return err, lastBlock, blockNumber, dataAccessResult, nil
	}

	return nil, _lastBlock, _blockNumber, _dataAccessResult, rawData
}

func encode_GetRequestForNextDataBlock(w io.Writer, blockNumber uint32) (err error) {
	var FNAME = "encode_GetRequestForNextDataBlock()"

	err = binary.Write(w, binary.BigEndian, blockNumber)
	if nil != err {
		errorLog.Printf(fmt.Sprintf("%s: binary.Write() failed, err: %s\n", err))
		return err
	}

	return nil
}

func decode_GetRequestForNextDataBlock(r io.Reader) (err error, blockNumber uint32) {
	var FNAME = "decode_GetRequestForNextDataBlock()"
	var serr string

	var _blockNumber uint32
	err = binary.Read(r, binary.BigEndian, &_blockNumber)
	if nil != err {
		errorLog.Println("%s: binary.Read() failed, err: %v", err)
		return err, 0
	}

	return nil, _blockNumber
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
		err = binary.Read(rwc, binary.BigEndian, &header)
		if nil != err {
			errorLog.Printf("%s: binary.Read() failed, err: %v\n", FNAME, err)
			ch <- &DlmsChannelMessage{err, nil}
			return
		}
		debugLog.Printf("%s: header: ok\n", FNAME)
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
		pdu := make([]byte, header.DataLength)
		err = binary.Read(rwc, binary.BigEndian, pdu)
		if nil != err {
			errorLog.Printf("%s: binary.Read() failed, err: %v\n", FNAME, err)
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
