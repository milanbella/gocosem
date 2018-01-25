package gocosem

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
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

const (
	actionResult_success                 = 0
	actionResult_hardwareFault           = 1
	actionResult_temporaryFailure        = 2
	actionResult_readWriteDenied         = 3
	actionResult_objectUndefined         = 4
	actionResult_objectClassInconsistent = 9
	actionResult_objectUnavailable       = 11
	actionResult_typeUnmatched           = 12
	actionResult_scopeOfAccessViolated   = 13
	actionResult_dataBlockUnavailable    = 14
	actionResult_longActionAborted       = 15
	actionResult_noLongActionInProgress  = 16
	actionResult_otherReason             = 250
)

type tDlmsInvokeIdAndPriority uint8

type DlmsClassId uint16
type DlmsOid [6]uint8
type DlmsAttributeId uint8
type DlmsAccessSelector uint8
type DlmsMethodId uint8

type DlmsData struct {
	Err error
	Typ uint8
	Val interface{}
	Len uint16
	Arr []*DlmsData // array
}

type DlmsDataAccessResult uint8
type DlmsActionResult uint8

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
	Deviation   uint16
	ClockStatus uint8
}

func DlmsDateFromBytes(b []byte) (date *DlmsDate) {
	date = new(DlmsDate)
	err := binary.Read(bytes.NewBuffer(b[0:2]), binary.BigEndian, &date.Year)
	if nil != err {
		err = fmt.Errorf("binary.Read() failed: %v", err)
		errorLog("%s", err)
		panic(err)
	}
	date.Month = b[2]
	date.DayOfMonth = b[3]
	date.DayOfWeek = b[4]
	return date
}

func (date *DlmsDate) ToBytes() []byte {
	var buf bytes.Buffer
	err := binary.Write(&buf, binary.BigEndian, date)
	if nil != err {
		panic(fmt.Sprintf("binary.Write() failed: %v", err))
	}
	return buf.Bytes()
}

func (date *DlmsDate) PrintDate() string {
	var (
		year       string
		month      string
		dayOfMonth string
		dayOfWeek  string
	)

	if date.IsYearWildcard() {
		year = "*"
	} else {
		year = fmt.Sprintf("%04d", date.Year)
	}

	if date.IsMonthWildcard() {
		month = "*"
	} else if date.IsDaylightSavingsBegin() {
		month = "db"
	} else if date.IsDaylightSavingsEnd() {
		month = "de"
	} else {
		month = fmt.Sprintf("%02d", date.Month)
	}

	dayOfMonth = fmt.Sprintf("%02d", date.DayOfMonth)

	if date.IsDayOfWeekWildcard() {
		dayOfWeek = fmt.Sprintf("*_wd")
	} else {
		dayOfWeek = fmt.Sprintf("%02d_wd", date.DayOfWeek)
	}

	return fmt.Sprintf("%s-%s-%s (%s)", year, month, dayOfMonth, dayOfWeek)
}

func (date *DlmsDate) SetYearWildcard() {
	date.Year = 0xFFFF
}

func (date *DlmsDate) IsYearWildcard() bool {
	return date.Year == 0xFFFF
}

func (date *DlmsDate) SetMonthWildcard() {
	date.Month = 0xFF
}

func (date *DlmsDate) IsMonthWildcard() bool {
	return date.Month == 0xFF
}

func (date *DlmsDate) SetDaylightSavingsEnd() {
	date.Month = 0xFD
}

func (date *DlmsDate) IsDaylightSavingsEnd() bool {
	return date.Month == 0xFD
}

func (date *DlmsDate) SetDaylightSavingsBegin() {
	date.Month = 0xFE
}

func (date *DlmsDate) IsDaylightSavingsBegin() bool {
	return date.Month == 0xFE
}

func (date *DlmsDate) SetDayOfWeekWildcard() {
	date.DayOfWeek = 0xFF
}

func (date *DlmsDate) IsDayOfWeekWildcard() bool {
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

func (tim *DlmsTime) ToBytes() []byte {
	var buf bytes.Buffer
	err := binary.Write(&buf, binary.BigEndian, tim)
	if nil != err {
		panic(fmt.Sprintf("binary.Write() failed: %v", err))
	}
	return buf.Bytes()
}

func (tim *DlmsTime) PrintTime() string {
	var (
		hour       string
		minute     string
		second     string
		hundredths string
	)

	if tim.IsHourWildcard() {
		hour = fmt.Sprintf("*")
	} else {
		hour = fmt.Sprintf("%02d", tim.Hour)
	}

	if tim.IsMinuteWildcard() {
		minute = fmt.Sprintf("*")
	} else {
		minute = fmt.Sprintf("%02d", tim.Minute)
	}

	if tim.IsSecondWildcard() {
		second = fmt.Sprintf("*")
	} else {
		second = fmt.Sprintf("%02d", tim.Second)
	}

	if tim.IsHundredthsWildcard() {
		hundredths = fmt.Sprintf("*")
	} else {
		hundredths = fmt.Sprintf("%02d", tim.Hundredths)
	}

	return fmt.Sprintf("%s-%s-%s-%s", hour, minute, second, hundredths)
}

func (tim *DlmsTime) SetHourWildcard() {
	tim.Hour = 0xFF
}

func (tim *DlmsTime) IsHourWildcard() bool {
	return tim.Hour == 0xFF
}

func (tim *DlmsTime) SetMinuteWildcard() {
	tim.Minute = 0xFF
}

func (tim *DlmsTime) IsMinuteWildcard() bool {
	return tim.Minute == 0xFF
}

func (tim *DlmsTime) SetSecondWildcard() {
	tim.Second = 0xFF
}

func (tim *DlmsTime) IsSecondWildcard() bool {
	return tim.Second == 0xFF
}

func (tim *DlmsTime) SetHundredthsWildcard() {
	tim.Hundredths = 0xFF
}

func (tim *DlmsTime) IsHundredthsWildcard() bool {
	return tim.Hundredths == 0xFF
}

func DlmsDateTimeFromBytes(b []byte) (dateTime *DlmsDateTime) {
	dateTime = new(DlmsDateTime)
	err := binary.Read(bytes.NewBuffer(b[0:2]), binary.BigEndian, &dateTime.Year)
	if nil != err {
		err = fmt.Errorf("binary.Read() failed: %v", err)
		errorLog("%s", err)
		panic(err)
	}
	dateTime.Month = b[2]
	dateTime.DayOfMonth = b[3]
	dateTime.DayOfWeek = b[4]
	dateTime.Hour = b[5]
	dateTime.Minute = b[6]
	dateTime.Second = b[7]
	dateTime.Hundredths = b[8]
	err = binary.Read(bytes.NewBuffer(b[9:11]), binary.BigEndian, &dateTime.Deviation)
	if nil != err {
		err = fmt.Errorf("binary.Read() failed: %v", err)
		errorLog("%s", err)
		panic(err)
	}
	dateTime.ClockStatus = b[11]

	return dateTime
}

func (dateTime *DlmsDateTime) ToBytes() []byte {
	var buf bytes.Buffer
	err := binary.Write(&buf, binary.BigEndian, dateTime)
	if nil != err {
		panic(fmt.Sprintf("binary.Write() failed: %v", err))
	}
	return buf.Bytes()
}

func (dateTime *DlmsDateTime) PrintDateTime() string {
	var (
		date        string
		time        string
		deviation   string
		clockStatus string
	)

	time = dateTime.PrintTime()
	date = dateTime.PrintDate()

	if 0x8000 == dateTime.Deviation {
		deviation = "*_dv"
	} else {
		deviation = fmt.Sprintf("%04d_dv", int16(dateTime.Deviation))
	}
	clockStatus = fmt.Sprintf("%02X_st", *(*[1]byte)(unsafe.Pointer(&dateTime.ClockStatus)))

	return fmt.Sprintf("%s %s (%s, %s)", date, time, deviation, clockStatus)
}

func (dateTime *DlmsDateTime) SetDeviationWildcard() {
	dateTime.Deviation = uint16(0x8000)
}

func (dateTime *DlmsDateTime) IsDeviationWildcard() bool {
	return uint16(dateTime.Deviation) == uint16(0x8000)
}

func (dateTime *DlmsDateTime) SetClockStatusInvalid() {
	dateTime.ClockStatus |= 0x01
}

func (dateTime *DlmsDateTime) IsClockStatusInvalid() bool {
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
	if length <= 0x80 {
		err = binary.Write(w, binary.BigEndian, uint8(length))
		if nil != err {
			errorLog("binary.Write() failed: %v", err)
			return err
		}
		return nil
	} else if length <= 0xFF {
		err = binary.Write(w, binary.BigEndian, []uint8{0x81, uint8(length)})
		if nil != err {
			errorLog("binary.Write() failed: %v", err)
			return err
		}
		return nil
	} else {
		err = binary.Write(w, binary.BigEndian, []uint8{0x82, uint8(length & 0xFF00 >> 8), uint8(0x00FF & length)})
		if nil != err {
			errorLog("binary.Write() failed: %v", err)
			return err
		}
		return nil
	}
}

func decodeAxdrLength(r io.Reader) (err error, length uint16) {
	var (
		u8  uint8
		u16 uint16
	)
	err = binary.Read(r, binary.BigEndian, &u8)
	if nil != err {
		errorLog("binary.Read() failed: %v", err)
		return err, 0
	}
	if u8 <= 0x80 {
		return nil, uint16(u8)
	} else if u8 == 0x81 {
		err = binary.Read(r, binary.BigEndian, &u8)
		if nil != err {
			errorLog("binary.Read() failed: %v", err)
			return err, 0
		}
		return nil, uint16(u8)
	} else if u8 == 0x82 {
		err = binary.Read(r, binary.BigEndian, &u16)
		if nil != err {
			errorLog("binary.Read() failed: %v", err)
			return err, 0
		}
		return nil, u16
	} else {
		err = fmt.Errorf("incorrect encoding\n")
		errorLog("%s", err)
		return err, 0
	}
}

func (data *DlmsData) Encode(w io.Writer) error {
	switch data.Typ {
	case DATA_TYPE_NULL:
		return data.encodeNULL(w)
	case DATA_TYPE_ARRAY, DATA_TYPE_STRUCTURE:
		err := binary.Write(w, binary.BigEndian, []byte{data.Typ})
		if nil != err {
			data.Err = err
			errorLog("binary.Write() failed: %v\n", err)
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
		err := fmt.Errorf("unknown data tag: %d\n", data.Typ)
		errorLog("%s", err)
		data.Err = err
		return err
	}
	return nil
}

func (data *DlmsData) Decode(r io.Reader) error {
	if err := binary.Read(r, binary.BigEndian, &data.Typ); nil != err {
		errorLog("binary.Read() failed: %v\n", err)
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
		err := fmt.Errorf("unknown data tag: %d\n", data.Typ)
		errorLog("%s", err)
		data.Err = err
		return err
	}
	return nil
}

func (data *DlmsData) Print() string {

	if nil != data.Err {
		return "!! data error !!"
	}

	switch data.Typ {
	case DATA_TYPE_NULL:
		return data.PrintNULL()
	case DATA_TYPE_ARRAY:
		return data.PrintArray()
	case DATA_TYPE_STRUCTURE:
		return data.PrintStructure()
	case DATA_TYPE_BOOLEAN:
		return data.PrintBoolean()
	case DATA_TYPE_BIT_STRING:
		return data.PrintBitString()
	case DATA_TYPE_DOUBLE_LONG:
		return data.PrintDoubleLong()
	case DATA_TYPE_DOUBLE_LONG_UNSIGNED:
		return data.PrintDoubleLongUnsigned()
	case DATA_TYPE_FLOATING_POINT:
		return data.PrintFloatingPoint()
	case DATA_TYPE_OCTET_STRING:
		return data.PrintOctetString()
	case DATA_TYPE_VISIBLE_STRING:
		return data.PrintVisibleString()
	case DATA_TYPE_BCD:
		return data.PrintBcd()
	case DATA_TYPE_INTEGER:
		return data.PrintInteger()
	case DATA_TYPE_LONG:
		return data.PrintLong()
	case DATA_TYPE_UNSIGNED:
		return data.PrintUnsigned()
	case DATA_TYPE_LONG_UNSIGNED:
		return data.PrintLongUnsigned()
	case DATA_TYPE_LONG64:
		return data.PrintLong64()
	case DATA_TYPE_UNSIGNED_LONG64:
		return data.PrintUnsignedLong64()
	case DATA_TYPE_ENUM:
		return data.PrintEnum()
	case DATA_TYPE_REAL32:
		return data.PrintReal32()
	case DATA_TYPE_REAL64:
		return data.PrintReal64()
	case DATA_TYPE_DATETIME:
		return data.PrintDateTime()
	case DATA_TYPE_DATE:
		return data.PrintDate()
	case DATA_TYPE_TIME:
		return data.PrintTime()
	default:
		return "!! data error !!"
	}
}

func (data *DlmsData) GetType() uint8 {
	return data.Typ
}

func (data *DlmsData) SetArray(length int) {
	data.Typ = DATA_TYPE_ARRAY
	data.Arr = make([]*DlmsData, length)
	for i := 0; i < length; i++ {
		data.Arr[i] = new(DlmsData)
	}
}

func (data *DlmsData) PrintArray() string {
	str := "a["
	for i := 0; i < len(data.Arr)-1; i++ {
		str += data.Arr[i].Print() + ", "
	}
	if len(data.Arr) > 0 {
		str += data.Arr[len(data.Arr)-1].Print() + "]"
	} else {
		str += "]"
	}
	return str
}

func (data *DlmsData) SetStructure(length int) {
	data.Typ = DATA_TYPE_STRUCTURE
	data.Arr = make([]*DlmsData, length)
	for i := 0; i < length; i++ {
		data.Arr[i] = new(DlmsData)
	}
}

func (data *DlmsData) PrintStructure() string {
	str := "s["
	for i := 0; i < len(data.Arr)-1; i++ {
		str += data.Arr[i].Print() + ", "
	}
	str += data.Arr[len(data.Arr)-1].Print() + "]"
	return str
}

func (data *DlmsData) SetNULL() {
	data.Typ = DATA_TYPE_NULL
}

func (data *DlmsData) PrintNULL() string {
	return "null"
}

func (data *DlmsData) encodeNULL(w io.Writer) (err error) {
	err = binary.Write(w, binary.BigEndian, []byte{DATA_TYPE_NULL})
	if nil != err {
		errorLog("binary.Write() failed: %v\n", err)
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

func (data *DlmsData) PrintBoolean() string {
	return fmt.Sprintf("%t (boolean)", data.GetBoolean())
}

func (data *DlmsData) encodeBoolean(w io.Writer) (err error) {
	if data.Val.(bool) {
		err = binary.Write(w, binary.BigEndian, []byte{DATA_TYPE_BOOLEAN, 1})
	} else {
		err = binary.Write(w, binary.BigEndian, []byte{DATA_TYPE_BOOLEAN, 0})
	}
	if nil != err {
		errorLog("binary.Write() failed: %v\n", err)
		data.Err = err
		return err
	}
	return nil
}

func (data *DlmsData) decodeBoolean(r io.Reader) (err error) {
	var b uint8
	err = binary.Read(r, binary.BigEndian, &b)
	if nil != err {
		errorLog("binary.Read() failed: %v\n", err)
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

func (data *DlmsData) PrintBitString() string {
	b, _ := data.GetBitString()
	return fmt.Sprintf("%X (BitString)", b)
}

func (data *DlmsData) encodeBitString(w io.Writer) (err error) {
	err = binary.Write(w, binary.BigEndian, []byte{DATA_TYPE_BIT_STRING})
	if nil != err {
		errorLog("binary.Write() failed: %v\n", err)
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
		errorLog("binary.Write() failed: %v", err)
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
		errorLog("binary.Read() failed: %v\n", err)
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

func (data *DlmsData) PrintDoubleLong() string {
	return fmt.Sprintf("%d (DoubleLong)", data.GetDoubleLong())
}

func (data *DlmsData) encodeDoubleLong(w io.Writer) (err error) {
	err = binary.Write(w, binary.BigEndian, []byte{DATA_TYPE_DOUBLE_LONG})
	if nil != err {
		errorLog("binary.Write() failed: %v\n", err)
		data.Err = err
		return err
	}
	err = binary.Write(w, binary.BigEndian, data.Val.(int32))
	if nil != err {
		errorLog("binary.Write() failed: %v\n", err)
		data.Err = err
		return err
	}
	return nil
}

func (data *DlmsData) decodeDoubleLong(r io.Reader) (err error) {
	var i int32
	err = binary.Read(r, binary.BigEndian, &i)
	if nil != err {
		errorLog("binary.Read() failed: %v\n", err)
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

func (data *DlmsData) PrintDoubleLongUnsigned() string {
	return fmt.Sprintf("%d (DoubleLongUnsigned)", data.GetDoubleLongUnsigned())
}

func (data *DlmsData) encodeDoubleLongUnsigned(w io.Writer) (err error) {
	err = binary.Write(w, binary.BigEndian, []byte{DATA_TYPE_DOUBLE_LONG_UNSIGNED})
	if nil != err {
		errorLog("binary.Write() failed: %v\n", err)
		data.Err = err
		return err
	}
	err = binary.Write(w, binary.BigEndian, data.Val.(uint32))
	if nil != err {
		errorLog("binary.Write() failed: %v\n", err)
		data.Err = err
		return err
	}
	return nil
}

func (data *DlmsData) decodeDoubleLongUnsigned(r io.Reader) (err error) {
	var i uint32
	err = binary.Read(r, binary.BigEndian, &i)
	if nil != err {
		errorLog("binary.Read() failed: %v\n", err)
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

func (data *DlmsData) PrintFloatingPoint() string {
	return fmt.Sprintf("%f (FloatingPoint)", data.GetFloatingPoint())
}

func (data *DlmsData) encodeFloatingPoint(w io.Writer) (err error) {
	err = binary.Write(w, binary.BigEndian, []byte{DATA_TYPE_FLOATING_POINT})
	if nil != err {
		errorLog("binary.Write() failed: %v\n", err)
		data.Err = err
		return err
	}
	err = binary.Write(w, binary.BigEndian, data.Val.(float32))
	if nil != err {
		errorLog("binary.Write() failed: %v\n", err)
		data.Err = err
		return err
	}
	return nil
}

func (data *DlmsData) decodeFloatingPoint(r io.Reader) (err error) {
	var f float32
	err = binary.Read(r, binary.BigEndian, &f)
	if nil != err {
		errorLog("binary.Read() failed: %v\n", err)
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

func (data *DlmsData) PrintOctetString() string {
	return fmt.Sprintf("%02X (OctetString)", data.GetOctetString())
}

func (data *DlmsData) encodeOctetString(w io.Writer) (err error) {
	err = binary.Write(w, binary.BigEndian, []byte{DATA_TYPE_OCTET_STRING})
	if nil != err {
		errorLog("binary.Write() failed: %v\n", err)
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
		errorLog("binary.Write() failed: %v\n", err)
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
		errorLog("binary.Read() failed: %v\n", err)
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

func (data *DlmsData) PrintVisibleString() string {
	return fmt.Sprintf("%02X (VisibleString)", data.GetVisibleString())
}

func (data *DlmsData) encodeVisibleString(w io.Writer) (err error) {
	err = binary.Write(w, binary.BigEndian, []byte{DATA_TYPE_VISIBLE_STRING})
	if nil != err {
		errorLog("binary.Write() failed: %v\n", err)
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
		errorLog("binary.Write() failed: %v\n", err)
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
		errorLog("binary.Read() failed: %v\n", err)
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

func (data *DlmsData) PrintBcd() string {
	return fmt.Sprintf("%d (Bcd)", data.GetBcd())
}

func (data *DlmsData) encodeBcd(w io.Writer) (err error) {
	err = binary.Write(w, binary.BigEndian, []byte{DATA_TYPE_BCD})
	if nil != err {
		errorLog("binary.Write() failed: %v\n", err)
		data.Err = err
		return err
	}
	err = binary.Write(w, binary.BigEndian, data.Val.(int8))
	if nil != err {
		errorLog("binary.Write() failed: %v\n", err)
		data.Err = err
		return err
	}
	return nil
}

func (data *DlmsData) decodeBcd(r io.Reader) (err error) {
	var bcd int8
	err = binary.Read(r, binary.BigEndian, &bcd)
	if nil != err {
		errorLog("binary.Read() failed: %v\n", err)
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

func (data *DlmsData) PrintInteger() string {
	return fmt.Sprintf("%d (Integer)", data.GetInteger())
}

func (data *DlmsData) encodeInteger(w io.Writer) (err error) {
	err = binary.Write(w, binary.BigEndian, []byte{DATA_TYPE_INTEGER})
	if nil != err {
		errorLog("binary.Write() failed: %v\n", err)
		data.Err = err
		return err
	}
	err = binary.Write(w, binary.BigEndian, data.Val.(int8))
	if nil != err {
		errorLog("binary.Write() failed: %v\n", err)
		data.Err = err
		return err
	}
	return nil
}

func (data *DlmsData) decodeInteger(r io.Reader) (err error) {
	var i int8
	err = binary.Read(r, binary.BigEndian, &i)
	if nil != err {
		errorLog("binary.Read() failed: %v\n", err)
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

func (data *DlmsData) PrintLong() string {
	return fmt.Sprintf("%d (Long)", data.GetLong())
}

func (data *DlmsData) encodeLong(w io.Writer) (err error) {
	err = binary.Write(w, binary.BigEndian, []byte{DATA_TYPE_LONG})
	if nil != err {
		errorLog("binary.Write() failed: %v\n", err)
		data.Err = err
		return err
	}
	err = binary.Write(w, binary.BigEndian, data.Val.(int16))
	if nil != err {
		errorLog("binary.Write() failed: %v\n", err)
		data.Err = err
		return err
	}
	return nil
}

func (data *DlmsData) decodeLong(r io.Reader) (err error) {
	var i int16
	err = binary.Read(r, binary.BigEndian, &i)
	if nil != err {
		errorLog("binary.Read() failed: %v\n", err)
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

func (data *DlmsData) PrintUnsigned() string {
	return fmt.Sprintf("%d (Unsigned)", data.GetUnsigned())
}

func (data *DlmsData) encodeUnsigned(w io.Writer) (err error) {
	err = binary.Write(w, binary.BigEndian, []byte{DATA_TYPE_UNSIGNED})
	if nil != err {
		errorLog("binary.Write() failed: %v\n", err)
		data.Err = err
		return err
	}
	err = binary.Write(w, binary.BigEndian, data.Val.(uint8))
	if nil != err {
		errorLog("binary.Write() failed: %v\n", err)
		data.Err = err
		return err
	}
	return nil
}

func (data *DlmsData) decodeUnsigned(r io.Reader) (err error) {
	var i uint8
	err = binary.Read(r, binary.BigEndian, &i)
	if nil != err {
		errorLog("binary.Read() failed: %v", err)
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

func (data *DlmsData) PrintLongUnsigned() string {
	return fmt.Sprintf("%d (LongUnsigned)", data.GetLongUnsigned())
}

func (data *DlmsData) encodeLongUnsigned(w io.Writer) (err error) {
	err = binary.Write(w, binary.BigEndian, []byte{DATA_TYPE_LONG_UNSIGNED})
	if nil != err {
		errorLog("binary.Write() failed: %v\n", err)
		data.Err = err
		return err
	}
	err = binary.Write(w, binary.BigEndian, data.Val.(uint16))
	if nil != err {
		errorLog("binary.Write() failed: %v\n", err)
		data.Err = err
		return err
	}
	return nil
}

func (data *DlmsData) decodeLongUnsigned(r io.Reader) (err error) {
	var i uint16
	err = binary.Read(r, binary.BigEndian, &i)
	if nil != err {
		errorLog("binary.Read() failed: %v", err)
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

func (data *DlmsData) PrintLong64() string {
	return fmt.Sprintf("%d (Long64)", data.GetLong64())
}

func (data *DlmsData) encodeLong64(w io.Writer) (err error) {
	err = binary.Write(w, binary.BigEndian, []byte{DATA_TYPE_LONG64})
	if nil != err {
		errorLog("binary.Write() failed: %v\n", err)
		data.Err = err
		return err
	}
	err = binary.Write(w, binary.BigEndian, data.Val.(int64))
	if nil != err {
		errorLog("binary.Write() failed: %v\n", err)
		data.Err = err
		return err
	}
	return nil
}

func (data *DlmsData) decodeLong64(r io.Reader) (err error) {
	var i int64
	err = binary.Read(r, binary.BigEndian, &i)
	if nil != err {
		errorLog("binary.Read() failed: %v\n", err)
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

func (data *DlmsData) PrintUnsignedLong64() string {
	return fmt.Sprintf("%d (UnsignedLong64)", data.GetUnsignedLong64())
}

func (data *DlmsData) encodeUnsignedLong64(w io.Writer) (err error) {
	err = binary.Write(w, binary.BigEndian, []byte{DATA_TYPE_UNSIGNED_LONG64})
	if nil != err {
		errorLog("binary.Write() failed: %v\n", err)
		data.Err = err
		return err
	}
	err = binary.Write(w, binary.BigEndian, data.Val.(uint64))
	if nil != err {
		errorLog("binary.Write() failed: %v\n", err)
		data.Err = err
		return err
	}
	return nil
}

func (data *DlmsData) decodeUnsignedLong64(r io.Reader) (err error) {
	var i uint64
	err = binary.Read(r, binary.BigEndian, &i)
	if nil != err {
		errorLog("binary.Read() failed: %v\n", err)
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

func (data *DlmsData) PrintEnum() string {
	return fmt.Sprintf("%d (Enum)", data.GetEnum())
}

func (data *DlmsData) encodeEnum(w io.Writer) (err error) {
	err = binary.Write(w, binary.BigEndian, []byte{DATA_TYPE_ENUM})
	if nil != err {
		errorLog("binary.Write() failed: %v\n", err)
		data.Err = err
		return err
	}
	err = binary.Write(w, binary.BigEndian, data.Val.(uint8))
	if nil != err {
		errorLog("binary.Write() failed: %v\n", err)
		data.Err = err
		return err
	}
	return nil
}

func (data *DlmsData) decodeEnum(r io.Reader) (err error) {
	var i uint8
	err = binary.Read(r, binary.BigEndian, &i)
	if nil != err {
		errorLog("binary.Read() failed: %v\n", err)
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

func (data *DlmsData) PrintReal32() string {
	return fmt.Sprintf("%f (Real32)", data.GetReal32())
}

func (data *DlmsData) encodeReal32(w io.Writer) (err error) {
	err = binary.Write(w, binary.BigEndian, []byte{DATA_TYPE_REAL32})
	if nil != err {
		errorLog("binary.Write() failed: %v\n", err)
		data.Err = err
		return err
	}
	err = binary.Write(w, binary.BigEndian, data.Val.(float32))
	if nil != err {
		errorLog("binary.Write() failed: %v\n", err)
		data.Err = err
		return err
	}
	return nil
}

func (data *DlmsData) decodeReal32(r io.Reader) (err error) {
	var f float32
	err = binary.Read(r, binary.BigEndian, &f)
	if nil != err {
		errorLog("binary.Read() failed: %v\n", err)
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

func (data *DlmsData) PrintReal64() string {
	return fmt.Sprintf("%f (Real64)", data.GetReal64())
}

func (data *DlmsData) encodeReal64(w io.Writer) (err error) {
	err = binary.Write(w, binary.BigEndian, []byte{DATA_TYPE_REAL64})
	if nil != err {
		errorLog("binary.Write() failed: %v\n", err)
		data.Err = err
		return err
	}
	err = binary.Write(w, binary.BigEndian, data.Val.(float64))
	if nil != err {
		errorLog("binary.Write() failed: %v\n", err)
		data.Err = err
		return err
	}
	return nil
}

func (data *DlmsData) decodeReal64(r io.Reader) (err error) {
	var f float64
	err = binary.Read(r, binary.BigEndian, &f)
	if nil != err {
		errorLog("binary.Read() failed: %v\n", err)
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

func (data *DlmsData) PrintDateTime() string {
	dateTime := DlmsDateTimeFromBytes(data.GetDateTime())
	return dateTime.PrintDateTime()
}

func (data *DlmsData) encodeDateTime(w io.Writer) (err error) {
	err = binary.Write(w, binary.BigEndian, []byte{DATA_TYPE_DATETIME})
	if nil != err {
		errorLog("binary.Write() failed: %v\n", err)
		data.Err = err
		return err
	}
	err = binary.Write(w, binary.BigEndian, data.Val.([]byte))
	if nil != err {
		errorLog("binary.Write() failed: %v\n", err)
		data.Err = err
		return err
	}
	return nil
}

func (data *DlmsData) decodeDateTime(r io.Reader) (err error) {
	b := make([]byte, 12)
	err = binary.Read(r, binary.BigEndian, b)
	if nil != err {
		errorLog("binary.Read() failed: %v\n", err)
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

func (data *DlmsData) PrintDate() string {
	date := DlmsDateFromBytes(data.GetDate())
	return date.PrintDate()
}

func (data *DlmsData) encodeDate(w io.Writer) (err error) {
	err = binary.Write(w, binary.BigEndian, []byte{DATA_TYPE_DATE})
	if nil != err {
		errorLog("binary.Write() failed: %v\n", err)
		data.Err = err
		return err
	}
	err = binary.Write(w, binary.BigEndian, data.Val.([]byte))
	if nil != err {
		errorLog("binary.Write() failed: %v\n", err)
		data.Err = err
		return err
	}
	return nil
}

func (data *DlmsData) decodeDate(r io.Reader) (err error) {
	b := make([]byte, 5)
	err = binary.Read(r, binary.BigEndian, b)
	if nil != err {
		errorLog("binary.Read() failed: %v\n", err)
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

func (data *DlmsData) PrintTime() string {
	time := DlmsTimeFromBytes(data.GetTime())
	return time.PrintTime()
}

func (data *DlmsData) encodeTime(w io.Writer) (err error) {
	err = binary.Write(w, binary.BigEndian, []byte{DATA_TYPE_TIME})
	if nil != err {
		errorLog("binary.Write() failed: %v\n", err)
		data.Err = err
		return err
	}
	err = binary.Write(w, binary.BigEndian, data.Val.([]byte))
	if nil != err {
		errorLog("binary.Write() failed: %v\n", err)
		data.Err = err
		return err
	}
	return nil
}

func (data *DlmsData) decodeTime(r io.Reader) (err error) {
	b := make([]byte, 4)
	err = binary.Read(r, binary.BigEndian, b)
	if nil != err {
		errorLog("binary.Read() failed: %v\n", err)
		return err
	}
	data.Typ = DATA_TYPE_TIME
	data.Val = b
	return nil
}

type DlmsInitiateRequest struct {
	dedicatedKey              *[]byte // optional
	responseAllowed           bool
	proposedQualityOfService  *int8 // optional
	proposedDlmsVersionNumber uint8
	proposedConformance       tAsn1BitString // bit string encoded in BER
	clientMaxReceivePduSize   uint16
}

/*
InitiateResponse ::= SEQUENCE
{
	negotiated-quality-of-service [0] IMPLICIT Integer8 OPTIONAL,
	negotiated-dlms-version-number Unsigned8,
	negotiated-conformance Conformance, -- Shall be encoded in BER
	server-max-receive-pdu-size Unsigned16,
	vaa-name ObjectName
}
*/

type DlmsInitiateResponse struct {
	negotiatedQualityOfService  *int8 // optional
	negotiatedDlmsVersionNumber uint8
	negotiatedConformance       tAsn1BitString
	serverMaxReceivePduSize     uint16
	vaaName                     int16
}

func (req *DlmsInitiateRequest) encode(w io.Writer) (err error) {

	err = binary.Write(w, binary.BigEndian, uint8(1)) // initiateRequest [1] IMPLICIT InitiateRequest
	if nil != err {
		errorLog(fmt.Sprintf("binary.Write() failed, err: %s\n", err))
		return err
	}

	// dedicated-key OCTET STRING OPTIONAL
	if req.dedicatedKey != nil {
		err = binary.Write(w, binary.BigEndian, uint8(1)) // used
		if nil != err {
			errorLog(fmt.Sprintf("binary.Write() failed, err: %s\n", err))
			return err
		}
		err = encodeAxdrLength(w, uint16(len(*req.dedicatedKey)))
		if nil != err {
			errorLog("encodeAxdrLength() failed, err: %v\n", err)
			return err
		}

		_, err = w.Write(*req.dedicatedKey)
		if nil != err {
			errorLog("w.Wite() failed, err: %v\n", err)
			return err
		}
	} else {
		err = binary.Write(w, binary.BigEndian, uint8(0)) // unused
		if nil != err {
			errorLog(fmt.Sprintf("binary.Write() failed, err: %s\n", err))
			return err
		}
	}

	// response-allowed BOOLEAN DEFAULT TRUE
	if req.responseAllowed == true {
		err = binary.Write(w, binary.BigEndian, uint8(0)) // unused, value is default value true
		if nil != err {
			errorLog(fmt.Sprintf("binary.Write() failed, err: %s\n", err))
			return err
		}
	} else {
		err = binary.Write(w, binary.BigEndian, uint8(1)) // used
		if nil != err {
			errorLog(fmt.Sprintf("binary.Write() failed, err: %s\n", err))
			return err
		}
		err = binary.Write(w, binary.BigEndian, req.responseAllowed) // value is false
		if nil != err {
			errorLog(fmt.Sprintf("binary.Write() failed, err: %s\n", err))
			return err
		}
	}

	// proposed-quality-of-service [0] IMPLICIT Integer8 OPTIONAL,
	if nil != req.proposedQualityOfService {
		err = binary.Write(w, binary.BigEndian, uint8(1)) // used
		if nil != err {
			errorLog(fmt.Sprintf("binary.Write() failed, err: %s\n", err))
			return err
		}
		err = binary.Write(w, binary.BigEndian, int8(*req.proposedQualityOfService)) // used
		if nil != err {
			errorLog(fmt.Sprintf("binary.Write() failed, err: %s\n", err))
			return err
		}
	} else {
		err = binary.Write(w, binary.BigEndian, uint8(0)) // unused
		if nil != err {
			errorLog(fmt.Sprintf("binary.Write() failed, err: %s\n", err))
			return err
		}
	}

	// proposed-dlms-version-number Unsigned8,
	err = binary.Write(w, binary.BigEndian, uint8(req.proposedDlmsVersionNumber))
	if nil != err {
		errorLog(fmt.Sprintf("binary.Write() failed, err: %s\n", err))
		return err
	}

	// proposed-conformance Conformance, -- Shall be encoded in BER
	// Conformance ::= [APPLICATION 31] IMPLICIT BIT STRING
	ch := new(t_der_chunk)
	ch.asn1_class = ASN1_CLASS_APPLICATION
	ch.encoding = BER_ENCODING_PRIMITIVE
	ch.asn1_tag = 31
	err, ch.content = der_encode_BitString(&req.proposedConformance)
	if nil != err {
		return err
	}
	err = der_encode_chunk(w, ch)
	if nil != err {
		return err
	}

	// client-max-receive-pdu-size Unsigned16
	err = binary.Write(w, binary.BigEndian, uint16(req.clientMaxReceivePduSize))
	if nil != err {
		errorLog(fmt.Sprintf("binary.Write() failed, err: %s\n", err))
		return err
	}

	return err
}

func (req *DlmsInitiateRequest) decode(r io.Reader) (err error) {

	// initiateRequest [1] IMPLICIT InitiateRequest
	var tag uint8
	err = binary.Read(r, binary.BigEndian, &tag)
	if nil != err {
		errorLog("binary.Read() failed, err: %v", err)
		return err
	}
	if 1 != tag {
		err = fmt.Errorf("wrong tag: %d", tag)
		return err
	}

	// dedicated-key OCTET STRING OPTIONAL
	var used uint8
	err = binary.Read(r, binary.BigEndian, &used)
	if nil != err {
		errorLog("binary.Read() failed, err: %v", err)
		return err
	}
	if used != 0 {
		err, length := decodeAxdrLength(r)
		if nil != err {
			errorLog("decodeAxdrLength() failed, err: %v\n", err)
			return err
		}
		dedicatedKey := make([]byte, length)
		err = binary.Read(r, binary.BigEndian, dedicatedKey)
		if nil != err {
			errorLog("binary.Read() failed, err: %v", err)
			return err
		}
		req.dedicatedKey = &dedicatedKey
	}

	// response-allowed BOOLEAN DEFAULT TRUE
	err = binary.Read(r, binary.BigEndian, &used)
	if nil != err {
		errorLog("binary.Read() failed, err: %v", err)
		return err
	}
	if 0 == used {
		// unused => set default value
		req.responseAllowed = true
	} else {
		var val uint8
		err = binary.Read(r, binary.BigEndian, &val)
		if nil != err {
			errorLog("binary.Read() failed, err: %v", err)
			return err
		}
		if val > 0 {
			req.responseAllowed = true
		} else {
			req.responseAllowed = false
		}
	}

	// proposed-quality-of-service [0] IMPLICIT Integer8 OPTIONAL,
	err = binary.Read(r, binary.BigEndian, &used)
	if nil != err {
		errorLog("binary.Read() failed, err: %v", err)
		return err
	}
	if 0 != used {
		req.proposedQualityOfService = new(int8)
		err = binary.Read(r, binary.BigEndian, req.proposedQualityOfService)
		if nil != err {
			errorLog("binary.Read() failed, err: %v", err)
			return err
		}
	}

	// proposed-dlms-version-number Unsigned8,
	err = binary.Read(r, binary.BigEndian, &req.proposedDlmsVersionNumber)
	if nil != err {
		errorLog("binary.Read() failed, err: %v", err)
		return err
	}

	// proposed-conformance Conformance, -- Shall be encoded in BER
	// Conformance ::= [APPLICATION 31] IMPLICIT BIT STRING
	err, ch := der_decode_chunk(r)
	if nil != err {
		return err
	}
	if 31 != ch.asn1_tag {
		err = fmt.Errorf("wrong tag: %d", ch.asn1_tag)
		return err
	}
	err, bitString := der_decode_BitString(ch.content)
	if nil != err {
		return err
	}
	req.proposedConformance = *bitString

	// client-max-receive-pdu-size Unsigned16
	err = binary.Read(r, binary.BigEndian, &req.clientMaxReceivePduSize)
	if nil != err {
		errorLog("binary.Read() failed, err: %v", err)
		return err
	}

	return err
}

func (rep *DlmsInitiateResponse) encode(w io.Writer) (err error) {
	err = binary.Write(w, binary.BigEndian, uint8(8)) // initiateResponse [8] IMPLICIT InitiateResponse,
	if nil != err {
		errorLog(fmt.Sprintf("binary.Write() failed, err: %s\n", err))
		return err
	}

	// negotiated-quality-of-service [0] IMPLICIT Integer8 OPTIONAL,
	if rep.negotiatedQualityOfService != nil {
		err = binary.Write(w, binary.BigEndian, uint8(1)) // used
		if nil != err {
			errorLog(fmt.Sprintf("binary.Write() failed, err: %s\n", err))
			return err
		}

		err = binary.Write(w, binary.BigEndian, *rep.negotiatedQualityOfService)
		if nil != err {
			errorLog(fmt.Sprintf("binary.Write() failed, err: %s\n", err))
			return err
		}
	} else {
		err = binary.Write(w, binary.BigEndian, uint8(0)) // unused
		if nil != err {
			errorLog(fmt.Sprintf("binary.Write() failed, err: %s\n", err))
			return err
		}
	}

	// negotiated-dlms-version-number Unsigned8,
	err = binary.Write(w, binary.BigEndian, rep.negotiatedDlmsVersionNumber)
	if nil != err {
		errorLog(fmt.Sprintf("binary.Write() failed, err: %s\n", err))
		return err
	}

	// negotiated-conformance Conformance, -- Shall be encoded in BER
	// Conformance ::= [APPLICATION 31] IMPLICIT BIT STRING
	ch := new(t_der_chunk)
	ch.asn1_class = ASN1_CLASS_APPLICATION
	ch.encoding = BER_ENCODING_PRIMITIVE
	ch.asn1_tag = 31
	err, ch.content = der_encode_BitString(&rep.negotiatedConformance)
	if nil != err {
		return err
	}
	err = der_encode_chunk(w, ch)
	if nil != err {
		return err
	}

	// server-max-receive-pdu-size Unsigned16,
	err = binary.Write(w, binary.BigEndian, rep.serverMaxReceivePduSize)
	if nil != err {
		errorLog(fmt.Sprintf("binary.Write() failed, err: %s\n", err))
		return err
	}

	// vaa-name ObjectName
	err = binary.Write(w, binary.BigEndian, rep.vaaName)
	if nil != err {
		errorLog(fmt.Sprintf("binary.Write() failed, err: %s\n", err))
		return err
	}

	return err
}

func (rep *DlmsInitiateResponse) decode(r io.Reader) (err error) {
	// initiateResponse [8] IMPLICIT InitiateResponse,
	var tag uint8
	err = binary.Read(r, binary.BigEndian, &tag)
	if nil != err {
		errorLog("binary.Read() failed, err: %v", err)
		return err
	}
	if 8 != tag {
		err = fmt.Errorf("wrong tag: %d", tag)
		return err
	}

	// negotiated-quality-of-service [0] IMPLICIT Integer8 OPTIONAL,
	var used uint8
	err = binary.Read(r, binary.BigEndian, &used)
	if nil != err {
		errorLog("binary.Read() failed, err: %v", err)
		return err
	}
	if used != 0 {
		err = binary.Read(r, binary.BigEndian, rep.negotiatedQualityOfService)
		if nil != err {
			errorLog("binary.Read() failed, err: %v", err)
			return err
		}
	}

	// negotiated-dlms-version-number Unsigned8,
	err = binary.Read(r, binary.BigEndian, &rep.negotiatedDlmsVersionNumber)
	if nil != err {
		errorLog("binary.Read() failed, err: %v", err)
		return err
	}

	// negotiated-conformance Conformance, -- Shall be encoded in BER
	// Conformance ::= [APPLICATION 31] IMPLICIT BIT STRING
	err, ch := der_decode_chunk(r)
	if nil != err {
		return err
	}
	if 31 != ch.asn1_tag {
		err = fmt.Errorf("wrong tag: %d", ch.asn1_tag)
		return err
	}
	err, bitString := der_decode_BitString(ch.content)
	if nil != err {
		return err
	}
	rep.negotiatedConformance = *bitString

	// server-max-receive-pdu-size Unsigned16,
	err = binary.Read(r, binary.BigEndian, &rep.serverMaxReceivePduSize)
	if nil != err {
		errorLog("binary.Read() failed, err: %v", err)
		return err
	}

	// vaa-name ObjectName
	err = binary.Read(r, binary.BigEndian, &rep.vaaName)
	if nil != err {
		errorLog("binary.Read() failed, err: %v", err)
		return err
	}

	return err
}

func encode_getRequest(w io.Writer, classId DlmsClassId, instanceId *DlmsOid, attributeId DlmsAttributeId, accessSelector DlmsAccessSelector, accessParameters *DlmsData) (err error) {

	err = binary.Write(w, binary.BigEndian, classId)
	if nil != err {
		errorLog(fmt.Sprintf("binary.Write() failed, err: %s\n", err))
		return err
	}

	err = binary.Write(w, binary.BigEndian, instanceId)
	if nil != err {
		errorLog(fmt.Sprintf("binary.Write() failed, err: %s\n", err))
		return err
	}

	err = binary.Write(w, binary.BigEndian, attributeId)
	if nil != err {
		errorLog(fmt.Sprintf("binary.Write() failed, err: %s\n", err))
		return err
	}

	if 0 != attributeId {
		if (0 == accessSelector) || (nil == accessParameters) {
			// access selection is false
			err = binary.Write(w, binary.BigEndian, uint8(0))
			if nil != err {
				errorLog(fmt.Sprintf("binary.Write() failed, err: %s\n", err))
				return err
			}
		} else {
			// access selection is true
			err = binary.Write(w, binary.BigEndian, uint8(1))
			if nil != err {
				errorLog(fmt.Sprintf("binary.Write() failed, err: %s\n", err))
				return err
			}

			err = binary.Write(w, binary.BigEndian, accessSelector)
			if nil != err {
				errorLog(fmt.Sprintf("binary.Write() failed, err: %s\n", err))
				return err
			}
			if nil != accessParameters {
				err = accessParameters.Encode(w)
				if nil != err {
					return err
				}
			}
		}
	}
	return nil
}

func decode_getRequest(r io.Reader) (err error, classId DlmsClassId, instanceId *DlmsOid, attributeId DlmsAttributeId, accessSelector DlmsAccessSelector, accessParameters *DlmsData) {

	var _classId DlmsClassId
	err = binary.Read(r, binary.BigEndian, &_classId)
	if nil != err {
		errorLog("binary.Read() failed, err: %v", err)
		return err, 0, nil, 0, 0, nil
	}

	_instanceId := new(DlmsOid)
	err = binary.Read(r, binary.BigEndian, _instanceId)
	if nil != err {
		errorLog("binary.Read() failed, err: %v", err)
		return err, _classId, nil, 0, 0, nil
	}

	var _attributeId DlmsAttributeId
	err = binary.Read(r, binary.BigEndian, &_attributeId)
	if nil != err {
		errorLog("binary.Read() failed, err: %v", err)
		return err, _classId, _instanceId, 0, 0, nil
	}

	var accessSelection uint8
	err = binary.Read(r, binary.BigEndian, &accessSelection)
	if nil != err {
		errorLog("binary.Read() failed, err: %v", err)
		return err, _classId, _instanceId, _attributeId, 0, nil
	}

	var _accessSelector DlmsAccessSelector = 0
	var _accessParameters *DlmsData = nil

	if 0 > accessSelection {
		// access selection is true

		err = binary.Read(r, binary.BigEndian, &_accessSelector)
		if nil != err {
			errorLog("binary.Read() failed, err: %v", err)
			return err, _classId, _instanceId, _attributeId, 0, nil
		}

		if _accessSelector > 0 {
			data := new(DlmsData)
			err = data.Decode(r)
			if nil != err {
				return err, _classId, _instanceId, _attributeId, _accessSelector, nil
			}
			_accessParameters = data
		}
	}
	return nil, _classId, _instanceId, _attributeId, _accessSelector, _accessParameters
}

func encode_getResponse(w io.Writer, dataAccessResult DlmsDataAccessResult, data *DlmsData) (err error) {

	if dataAccessResult_success == dataAccessResult {
		err = binary.Write(w, binary.BigEndian, uint8(0)) // choice [0] - data
		if nil != err {
			errorLog(fmt.Sprintf("binary.Write() failed, err: %s\n", err))
			return err
		}

		err = data.Encode(w)
		if nil != err {
			return err
		}
	} else {
		err = binary.Write(w, binary.BigEndian, uint8(1)) // choice [1] - dataAccessResult
		if nil != err {
			errorLog(fmt.Sprintf("binary.Write() failed, err: %s\n", err))
			return err
		}

		err = binary.Write(w, binary.BigEndian, dataAccessResult)
		if nil != err {
			errorLog(fmt.Sprintf("binary.Write() failed, err: %s\n", err))
			return err
		}
	}

	return nil
}

func encode_getResponseBlock(w io.Writer, data *DlmsData) (err error) {

	if nil != data {
		err = data.Encode(w)
		if nil != err {
			return err
		}
	}

	return nil
}

func decode_getResponse(r io.Reader) (err error, dataAccessResult DlmsDataAccessResult, data *DlmsData) {

	err = binary.Read(r, binary.BigEndian, &dataAccessResult)
	if nil != err {
		errorLog("binary.Read() failed, err: %v", err)
		return err, 0, nil
	}

	data = nil
	if dataAccessResult_success == dataAccessResult {
		data = new(DlmsData)
		err = data.Decode(r)
		if nil != err {
			return err, dataAccessResult, data
		}
	} else {
		err = binary.Read(r, binary.BigEndian, &dataAccessResult)
		if nil != err {
			errorLog("binary.Read() failed, err: %v", err)
			return err, 0, nil
		}
	}

	return nil, dataAccessResult, data
}

func decode_getResponseBlock(r io.Reader) (err error, data *DlmsData) {
	data = new(DlmsData)
	err = data.Decode(r)
	if nil != err {
		return err, data
	}

	return nil, data
}

func encode_GetRequestNormal(w io.Writer, classId DlmsClassId, instanceId *DlmsOid, attributeId DlmsAttributeId, accessSelector DlmsAccessSelector, accessParameters *DlmsData) (err error) {
	err = encode_getRequest(w, classId, instanceId, attributeId, accessSelector, accessParameters)
	if nil != err {
		return err
	}

	return nil
}

func decode_GetRequestNormal(r io.Reader) (err error, classId DlmsClassId, instanceId *DlmsOid, attributeId DlmsAttributeId, accessSelector DlmsAccessSelector, accessParameters *DlmsData) {
	err, classId, instanceId, attributeId, accessSelector, accessParameters = decode_getRequest(r)
	if nil != err {
		return err, 0, nil, 0, 0, nil
	}
	return nil, classId, instanceId, attributeId, accessSelector, accessParameters
}

func encode_GetResponseNormal(w io.Writer, dataAccessResult DlmsDataAccessResult, data *DlmsData) (err error) {
	err = encode_getResponse(w, dataAccessResult, data)
	if nil != err {
		return err
	}

	return nil
}

func encode_GetResponseNormalBlock(w io.Writer, data *DlmsData) (err error) {
	err = encode_getResponseBlock(w, data)
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

func decode_GetResponseNormalBlock(r io.Reader) (err error, data *DlmsData) {
	err, data = decode_getResponseBlock(r)
	if nil != err {
		return err, data
	}

	return nil, data
}

func encode_GetRequestWithList(w io.Writer, classIds []DlmsClassId, instanceIds []*DlmsOid, attributeIds []DlmsAttributeId, accessSelectors []DlmsAccessSelector, accessParameters []*DlmsData) (err error) {
	count := uint8(len(classIds)) // count of get requests

	err = binary.Write(w, binary.BigEndian, count)
	if nil != err {
		errorLog(fmt.Sprintf("binary.Write() failed, err: %s\n", err))
		return err
	}

	for i := uint8(0); i < count; i += 1 {

		err = encode_getRequest(w, classIds[i], instanceIds[i], attributeIds[i], accessSelectors[i], accessParameters[i])
		if nil != err {
			errorLog("encode_getRequest() failed, err: %v\n", err)
			return err
		}

	}

	return nil
}

func decode_GetRequestWithList(r io.Reader) (err error, classIds []DlmsClassId, instanceIds []*DlmsOid, attributeIds []DlmsAttributeId, accessSelectors []DlmsAccessSelector, accessParameters []*DlmsData) {
	var count uint8
	err = binary.Read(r, binary.BigEndian, &count)
	if nil != err {
		errorLog("binary.Read() failed, err: %v", err)
		return err, nil, nil, nil, nil, nil
	}

	classIds = make([]DlmsClassId, count)
	instanceIds = make([]*DlmsOid, count)
	attributeIds = make([]DlmsAttributeId, count)
	accessSelectors = make([]DlmsAccessSelector, count)
	accessParameters = make([]*DlmsData, count)

	for i := uint8(0); i < count; i += 1 {
		err, classId, instanceId, attributeId, accessSelector, accessParameter := decode_getRequest(r)
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
	count := uint8(len(dataAccessResults))

	err = binary.Write(w, binary.BigEndian, count)
	if nil != err {
		errorLog(fmt.Sprintf("binary.Write() failed, err: %s\n", err))
		return err
	}

	for i := uint8(0); i < count; i += 1 {

		err := encode_getResponse(w, dataAccessResults[i], datas[i])
		if nil != err {
			return err
		}
	}

	return nil
}

func decode_GetResponseWithList(r io.Reader) (err error, dataAccessResults []DlmsDataAccessResult, datas []*DlmsData) {

	var count uint8
	err = binary.Read(r, binary.BigEndian, &count)
	if nil != err {
		errorLog("binary.Read() failed, err: %v", err)
		return err, nil, nil
	}

	dataAccessResults = make([]DlmsDataAccessResult, count)
	datas = make([]*DlmsData, count)

	var dataAccessResult DlmsDataAccessResult
	var data *DlmsData
	for i := uint8(0); i < count; i += 1 {
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
	var bb byte
	if lastBlock {
		bb = 1
	} else {
		bb = 0
	}
	err = binary.Write(w, binary.BigEndian, bb)
	if nil != err {
		errorLog(fmt.Sprintf("binary.Write() failed, err: %s\n", err))
		return err
	}

	err = binary.Write(w, binary.BigEndian, blockNumber)
	if nil != err {
		errorLog(fmt.Sprintf("binary.Write() failed, err: %s\n", err))
		return err
	}

	if 0 == dataAccessResult {
		err = binary.Write(w, binary.BigEndian, uint8(0)) // CHOICE raw-data [0] IMPLICIT OCTET STRING,
		if nil != err {
			errorLog(fmt.Sprintf("binary.Write() failed, err: %s\n", err))
			return err
		}

		if nil != rawData {
			err = encodeAxdrLength(w, uint16(len(rawData)))
			if nil != err {
				errorLog("encodeAxdrLength() failed, err: %v\n", err)
				return err
			}

			_, err = w.Write(rawData)
			if nil != err {
				errorLog("w.Wite() failed, err: %v\n", err)
				return err
			}
		}

	} else {
		err = binary.Write(w, binary.BigEndian, uint8(1)) // CHOICE data-access-result [1] IMPLICIT Data-Access-Result
		if nil != err {
			errorLog(fmt.Sprintf("binary.Write() failed, err: %s\n", err))
			return err
		}
		err = binary.Write(w, binary.BigEndian, dataAccessResult)
		if nil != err {
			errorLog(fmt.Sprintf("binary.Write() failed, err: %s\n", err))
			return err
		}
	}

	return nil
}

func decode_GetResponsewithDataBlock(r io.Reader) (err error, lastBlock bool, blockNumber uint32, dataAccessResult DlmsDataAccessResult, rawData []byte) {
	var __lastBlock uint8
	err = binary.Read(r, binary.BigEndian, &__lastBlock)
	if nil != err {
		errorLog("binary.Read() failed, err: %v", err)
		return err, false, 0, 0, nil
	}
	var _lastBlock bool
	if 0 == __lastBlock {
		_lastBlock = false
	} else {
		_lastBlock = true
	}

	var _blockNumber uint32
	err = binary.Read(r, binary.BigEndian, &_blockNumber)
	if nil != err {
		errorLog("binary.Read() failed, err: %v", err)
		return err, _lastBlock, 0, 0, nil
	}

	var _dataAccessResult DlmsDataAccessResult
	err = binary.Read(r, binary.BigEndian, &_dataAccessResult) // read the tag
	if nil != err {
		errorLog("binary.Read() failed, err: %v", err)
		return err, _lastBlock, _blockNumber, 0, nil
	}

	if 0 == _dataAccessResult {

		err, length := decodeAxdrLength(r)
		if nil != err {
			errorLog("decodeAxdrLength() failed, err: %v\n", err)
			return err, _lastBlock, _blockNumber, _dataAccessResult, nil
		}

		rawData = make([]byte, length)
		err = binary.Read(r, binary.BigEndian, rawData)
		if nil != err {
			errorLog("binary.Read() failed, err: %v", err)
			return err, _lastBlock, _blockNumber, _dataAccessResult, nil
		}
	} else {
		err = binary.Read(r, binary.BigEndian, &_dataAccessResult)
		if nil != err {
			errorLog("binary.Read() failed, err: %v", err)
			return err, _lastBlock, _blockNumber, 0, nil
		}
	}

	return nil, _lastBlock, _blockNumber, _dataAccessResult, rawData
}

func encode_GetRequestForNextDataBlock(w io.Writer, blockNumber uint32) (err error) {
	err = binary.Write(w, binary.BigEndian, blockNumber)
	if nil != err {
		errorLog(fmt.Sprintf("binary.Write() failed, err: %s\n", err))
		return err
	}

	return nil
}

func decode_GetRequestForNextDataBlock(r io.Reader) (err error, blockNumber uint32) {
	var _blockNumber uint32
	err = binary.Read(r, binary.BigEndian, &_blockNumber)
	if nil != err {
		errorLog("binary.Read() failed, err: %v", err)
		return err, 0
	}

	return nil, _blockNumber
}

func encode_setRequestData(w io.Writer, data *DlmsData) (err error) {
	if nil != data {
		err := data.Encode(w)
		if nil != err {
			return err
		}
	}
	return nil
}

func decode_setRequestData(r io.Reader) (err error, data *DlmsData) {
	data = new(DlmsData)
	err = data.Decode(r)
	if nil != err {
		return err, nil
	}
	return nil, data
}

func encode_setRequestBlock(w io.Writer, lastBlock bool, blockNumber uint32, rawData []byte) (err error) {
	var _lastBlock uint8

	if lastBlock {
		_lastBlock = 1
	} else {
		_lastBlock = 0
	}

	err = binary.Write(w, binary.BigEndian, _lastBlock)
	if nil != err {
		errorLog(fmt.Sprintf("binary.Write() failed, err: %s\n", err))
		return err
	}
	err = binary.Write(w, binary.BigEndian, blockNumber)
	if nil != err {
		errorLog(fmt.Sprintf("binary.Write() failed, err: %s\n", err))
		return err
	}

	err = encodeAxdrLength(w, uint16(len(rawData)))
	if nil != err {
		errorLog("encodeAxdrLength() failed, err: %v\n", err)
		return err
	}
	_, err = w.Write(rawData)
	if nil != err {
		errorLog("w.Wite() failed, err: %v\n", err)
		return err
	}

	return nil
}

func decode_setRequestBlock(r io.Reader) (err error, lastBlock bool, blockNumber uint32, rawData []byte) {
	var _lastBlock bool
	var u8 uint8
	err = binary.Read(r, binary.BigEndian, &u8)
	if nil != err {
		errorLog("binary.Read() failed, err: %v", err)
		return err, false, 0, nil
	}
	if 0 == u8 {
		_lastBlock = false
	} else {
		_lastBlock = true
	}

	var _blockNumber uint32
	err = binary.Read(r, binary.BigEndian, &_blockNumber)
	if nil != err {
		errorLog("binary.Read() failed, err: %v", err)
		return err, _lastBlock, 0, nil
	}

	err, length := decodeAxdrLength(r)
	if nil != err {
		errorLog("decodeAxdrLength() failed, err: %v\n", err)
		return err, _lastBlock, _blockNumber, nil
	}

	_rawData := make([]byte, length)
	err = binary.Read(r, binary.BigEndian, _rawData)
	if nil != err {
		errorLog("binary.Read() failed, err: %v", err)
		return err, _lastBlock, _blockNumber, nil
	}

	return err, _lastBlock, _blockNumber, _rawData
}

func encode_setRequest(w io.Writer, classId DlmsClassId, instanceId *DlmsOid, attributeId DlmsAttributeId, accessSelector DlmsAccessSelector, accessParameters *DlmsData) (err error) {
	err = binary.Write(w, binary.BigEndian, classId)
	if nil != err {
		errorLog(fmt.Sprintf("binary.Write() failed, err: %s\n", err))
		return err
	}

	err = binary.Write(w, binary.BigEndian, instanceId)
	if nil != err {
		errorLog(fmt.Sprintf("binary.Write() failed, err: %s\n", err))
		return err
	}

	err = binary.Write(w, binary.BigEndian, attributeId)
	if nil != err {
		errorLog(fmt.Sprintf("binary.Write() failed, err: %s\n", err))
		return err
	}

	if 0 != attributeId {
		if (0 == accessSelector) || (nil == accessParameters) {
			// access selection is false
			err = binary.Write(w, binary.BigEndian, uint8(0))
			if nil != err {
				errorLog(fmt.Sprintf("binary.Write() failed, err: %s\n", err))
				return err
			}
		} else {
			// access selection is true
			err = binary.Write(w, binary.BigEndian, uint8(1))
			if nil != err {
				errorLog(fmt.Sprintf("binary.Write() failed, err: %s\n", err))
				return err
			}

			err = binary.Write(w, binary.BigEndian, accessSelector)
			if nil != err {
				errorLog(fmt.Sprintf("binary.Write() failed, err: %s\n", err))
				return err
			}
			if nil != accessParameters {
				err = accessParameters.Encode(w)
				if nil != err {
					return err
				}
			}
		}
	}
	return nil
}

func decode_setRequest(r io.Reader) (err error, classId DlmsClassId, instanceId *DlmsOid, attributeId DlmsAttributeId, accessSelector DlmsAccessSelector, accessParameters *DlmsData) {
	var _classId DlmsClassId
	err = binary.Read(r, binary.BigEndian, &_classId)
	if nil != err {
		errorLog("binary.Read() failed, err: %v", err)
		return err, 0, nil, 0, 0, nil
	}

	_instanceId := new(DlmsOid)
	err = binary.Read(r, binary.BigEndian, _instanceId)
	if nil != err {
		errorLog("binary.Read() failed, err: %v", err)
		return err, _classId, nil, 0, 0, nil
	}

	var _attributeId DlmsAttributeId
	err = binary.Read(r, binary.BigEndian, &_attributeId)
	if nil != err {
		errorLog("binary.Read() failed, err: %v", err)
		return err, _classId, _instanceId, 0, 0, nil
	}

	var accessSelection uint8
	err = binary.Read(r, binary.BigEndian, &accessSelection)
	if nil != err {
		errorLog("binary.Read() failed, err: %v", err)
		return err, _classId, _instanceId, _attributeId, 0, nil
	}

	var _accessSelector DlmsAccessSelector = 0
	var _accessParameters *DlmsData = nil

	if 0 > accessSelection {
		// access selection is true

		err = binary.Read(r, binary.BigEndian, &_accessSelector)
		if nil != err {
			errorLog("binary.Read() failed, err: %v", err)
			return err, _classId, _instanceId, _attributeId, 0, nil
		}

		if _accessSelector > 0 {
			data := new(DlmsData)
			err = data.Decode(r)
			if nil != err {
				return err, _classId, _instanceId, _attributeId, _accessSelector, nil
			}
			_accessParameters = data
			return err, _classId, _instanceId, _attributeId, _accessSelector, _accessParameters
		} else {
			return err, _classId, _instanceId, _attributeId, _accessSelector, nil
		}
	} else {
		return err, _classId, _instanceId, _attributeId, 0, nil
	}
}

func encode_SetRequestNormal(w io.Writer, classId DlmsClassId, instanceId *DlmsOid, attributeId DlmsAttributeId, accessSelector DlmsAccessSelector, accessParameters *DlmsData, data *DlmsData) (err error) {

	err = encode_setRequest(w, classId, instanceId, attributeId, accessSelector, accessParameters)
	if nil != err {
		return err
	}
	err = encode_setRequestData(w, data)
	if nil != err {
		return err
	}

	return nil
}

func decode_SetRequestNormal(r io.Reader) (err error, classId DlmsClassId, instanceId *DlmsOid, attributeId DlmsAttributeId, accessSelector DlmsAccessSelector, accessParameters *DlmsData, data *DlmsData) {

	err, classId, instanceId, attributeId, accessSelector, accessParameters = decode_setRequest(r)
	if nil != err {
		return err, 0, nil, 0, 0, nil, nil
	}
	err, data = decode_setRequestData(r)
	if nil != err {
		return err, classId, instanceId, attributeId, accessSelector, accessParameters, nil
	}
	return nil, classId, instanceId, attributeId, accessSelector, accessParameters, data
}

func encode_SetRequestNormalBlock(w io.Writer, classId DlmsClassId, instanceId *DlmsOid, attributeId DlmsAttributeId, accessSelector DlmsAccessSelector, accessParameters *DlmsData, lastBlock bool, blockNumber uint32, rawData []byte) (err error) {

	err = encode_setRequest(w, classId, instanceId, attributeId, accessSelector, accessParameters)
	if nil != err {
		return err
	}
	err = encode_setRequestBlock(w, lastBlock, blockNumber, rawData)
	if nil != err {
		return err
	}

	return nil
}

func decode_SetRequestNormalBlock(r io.Reader) (err error, classId DlmsClassId, instanceId *DlmsOid, attributeId DlmsAttributeId, accessSelector DlmsAccessSelector, accessParameters *DlmsData, lastBlock bool, blockNumber uint32, rawData []byte) {

	err, classId, instanceId, attributeId, accessSelector, accessParameters = decode_setRequest(r)
	if nil != err {
		return err, 0, nil, 0, 0, nil, false, 0, nil
	}
	err, lastBlock, blockNumber, rawData = decode_setRequestBlock(r)
	if nil != err {
		return err, classId, instanceId, attributeId, accessSelector, accessParameters, false, 0, nil
	}
	return nil, classId, instanceId, attributeId, accessSelector, accessParameters, lastBlock, blockNumber, rawData
}

func encode_SetRequestWithList(w io.Writer, classIds []DlmsClassId, instanceIds []*DlmsOid, attributeIds []DlmsAttributeId, accessSelectors []DlmsAccessSelector, accessParameters []*DlmsData, datas []*DlmsData) (err error) {
	count := uint8(len(classIds)) // count of get requests

	err = binary.Write(w, binary.BigEndian, count)
	if nil != err {
		errorLog(fmt.Sprintf("binary.Write() failed, err: %s\n", err))
		return err
	}
	for i := uint8(0); i < count; i += 1 {
		err = encode_setRequest(w, classIds[i], instanceIds[i], attributeIds[i], accessSelectors[i], accessParameters[i])
		if nil != err {
			return err
		}
	}

	err = binary.Write(w, binary.BigEndian, count)
	if nil != err {
		errorLog(fmt.Sprintf("binary.Write() failed, err: %s\n", err))
		return err
	}
	for i := uint8(0); i < count; i += 1 {
		err = encode_setRequestData(w, datas[i])
		if nil != err {
			return err
		}
	}

	return nil
}

func decode_SetRequestWithList(r io.Reader) (err error, classIds []DlmsClassId, instanceIds []*DlmsOid, attributeIds []DlmsAttributeId, accessSelectors []DlmsAccessSelector, accessParameters []*DlmsData, datas []*DlmsData) {
	var count uint8

	err = binary.Read(r, binary.BigEndian, &count)
	if nil != err {
		errorLog("binary.Read() failed, err: %v", err)
		return err, nil, nil, nil, nil, nil, nil
	}
	classIds = make([]DlmsClassId, count)
	instanceIds = make([]*DlmsOid, count)
	attributeIds = make([]DlmsAttributeId, count)
	accessSelectors = make([]DlmsAccessSelector, count)
	accessParameters = make([]*DlmsData, count)

	for i := uint8(0); i < count; i += 1 {
		err, classId, instanceId, attributeId, accessSelector, accessParameter := decode_setRequest(r)
		if nil != err {
			return err, classIds[0:i], instanceIds[0:i], attributeIds[0:i], accessSelectors[0:i], accessParameters[0:i], nil
		}
		classIds[i] = classId
		instanceIds[i] = instanceId
		attributeIds[i] = attributeId
		accessSelectors[i] = accessSelector
		accessParameters[i] = accessParameter
	}

	err = binary.Read(r, binary.BigEndian, &count)
	if nil != err {
		errorLog("binary.Read() failed, err: %v", err)
		return err, classIds, instanceIds, attributeIds, accessSelectors, accessParameters, nil
	}
	if int(count) < len(classIds) {
		err = fmt.Errorf("missing data")
		return err, classIds, instanceIds, attributeIds, accessSelectors, accessParameters, nil
	}
	datas = make([]*DlmsData, count)
	for i := uint8(0); i < count; i += 1 {
		err, data := decode_setRequestData(r)
		if nil != err {
			return err, classIds, instanceIds, attributeIds, accessSelectors, accessParameters, datas[0:i]
		}
		datas[i] = data
	}

	return nil, classIds, instanceIds, attributeIds, accessSelectors, accessParameters, datas
}

func encode_SetRequestWithListBlock(w io.Writer, classIds []DlmsClassId, instanceIds []*DlmsOid, attributeIds []DlmsAttributeId, accessSelectors []DlmsAccessSelector, accessParameters []*DlmsData, lastBlock bool, blockNumber uint32, rawData []byte) (err error) {
	count := uint8(len(classIds)) // count of get requests

	err = binary.Write(w, binary.BigEndian, count)
	if nil != err {
		errorLog(fmt.Sprintf("binary.Write() failed, err: %s\n", err))
		return err
	}
	for i := uint8(0); i < count; i += 1 {
		err = encode_setRequest(w, classIds[i], instanceIds[i], attributeIds[i], accessSelectors[i], accessParameters[i])
		if nil != err {
			return err
		}
	}

	err = encode_setRequestBlock(w, lastBlock, blockNumber, rawData)
	if nil != err {
		return err
	}

	return nil
}

func decode_SetRequestWithListBlock(r io.Reader) (err error, classIds []DlmsClassId, instanceIds []*DlmsOid, attributeIds []DlmsAttributeId, accessSelectors []DlmsAccessSelector, accessParameters []*DlmsData, lastBlock bool, blockNumber uint32, rawData []byte) {
	var count uint8

	err = binary.Read(r, binary.BigEndian, &count)
	if nil != err {
		errorLog("binary.Read() failed, err: %v", err)
		return err, nil, nil, nil, nil, nil, false, 0, nil
	}
	classIds = make([]DlmsClassId, count)
	instanceIds = make([]*DlmsOid, count)
	attributeIds = make([]DlmsAttributeId, count)
	accessSelectors = make([]DlmsAccessSelector, count)
	accessParameters = make([]*DlmsData, count)

	for i := uint8(0); i < count; i += 1 {
		err, classId, instanceId, attributeId, accessSelector, accessParameter := decode_setRequest(r)
		if nil != err {
			return err, classIds[0:i], instanceIds[0:i], attributeIds[0:i], accessSelectors[0:i], accessParameters[0:i], false, 0, nil
		}
		classIds[i] = classId
		instanceIds[i] = instanceId
		attributeIds[i] = attributeId
		accessSelectors[i] = accessSelector
		accessParameters[i] = accessParameter
	}

	err, lastBlock, blockNumber, rawData = decode_setRequestBlock(r)
	if nil != err {
		return err, classIds, instanceIds, attributeIds, accessSelectors, accessParameters, false, 0, nil
	}

	return nil, classIds, instanceIds, attributeIds, accessSelectors, accessParameters, lastBlock, blockNumber, rawData

}

func encode_SetRequestWithDataBlock(w io.Writer, lastBlock bool, blockNumber uint32, rawData []byte) (err error) {
	var _lastBlock uint8 = 0
	if lastBlock {
		_lastBlock = 1
	}

	err = binary.Write(w, binary.BigEndian, _lastBlock)
	if nil != err {
		errorLog("binary.Write() failed, err: %v", err)
		return err
	}

	err = binary.Write(w, binary.BigEndian, blockNumber)
	if nil != err {
		errorLog("binary.Write() failed, err: %v", err)
		return err
	}

	err = encodeAxdrLength(w, uint16(len(rawData)))
	if nil != err {
		errorLog("encodeAxdrLength() failed, err: %v\n", err)
		return err
	}
	_, err = w.Write(rawData)
	if nil != err {
		errorLog("w.Wite() failed, err: %v\n", err)
		return err
	}

	return nil
}

func decode_SetRequestWithDataBlock(r io.Reader) (err error, lastBlock bool, blockNumber uint32, rawData []byte) {
	var _lastBlock uint8
	err = binary.Read(r, binary.BigEndian, &_lastBlock)
	if nil != err {
		errorLog("binary.Read() failed, err: %v", err)
		return err, false, 0, nil
	}
	if _lastBlock > 0 {
		lastBlock = true
	} else {
		lastBlock = false
	}

	err = binary.Read(r, binary.BigEndian, &blockNumber)
	if nil != err {
		errorLog("binary.Read() failed, err: %v", err)
		return err, lastBlock, 0, nil
	}

	err, length := decodeAxdrLength(r)
	if nil != err {
		errorLog("decodeAxdrLength() failed, err: %v\n", err)
		return err, lastBlock, blockNumber, nil
	}

	rawData = make([]byte, length)
	err = binary.Read(r, binary.BigEndian, rawData)
	if nil != err {
		errorLog("binary.Read() failed, err: %v", err)
		return err, lastBlock, blockNumber, nil
	}

	return err, lastBlock, blockNumber, rawData
}

func encode_SetResponseNormal(w io.Writer, dataAccessResult DlmsDataAccessResult) (err error) {
	err = binary.Write(w, binary.BigEndian, dataAccessResult)
	if nil != err {
		errorLog("binary.Write() failed, err: %v", err)
		return err
	}

	return nil
}

func decode_SetResponseNormal(r io.Reader) (err error, dataAccessResult DlmsDataAccessResult) {
	err = binary.Read(r, binary.BigEndian, &dataAccessResult)
	if nil != err {
		errorLog("binary.Read() failed, err: %v", err)
		return err, 0
	}

	return nil, dataAccessResult
}

func encode_SetResponseWithList(w io.Writer, dataAccessResults []DlmsDataAccessResult) (err error) {
	count := uint8(len(dataAccessResults))
	err = binary.Write(w, binary.BigEndian, count)
	if nil != err {
		errorLog("binary.Write() failed, err: %v", err)
		return err
	}

	for i := uint8(0); i < count; i++ {
		err = binary.Write(w, binary.BigEndian, dataAccessResults[i])
		if nil != err {
			errorLog("binary.Write() failed, err: %v", err)
			return err
		}
	}

	return nil
}

func decode_SetResponseWithList(r io.Reader) (err error, dataAccessResults []DlmsDataAccessResult) {
	var count uint8
	err = binary.Read(r, binary.BigEndian, &count)
	if nil != err {
		errorLog("binary.Read() failed, err: %v", err)
		return err, nil
	}

	dataAccessResults = make([]DlmsDataAccessResult, count)
	for i := uint8(0); i < count; i++ {
		err = binary.Read(r, binary.BigEndian, &dataAccessResults[i])
		if nil != err {
			errorLog("binary.Read() failed, err: %v", err)
			return err, dataAccessResults[0:i]
		}
	}

	return nil, dataAccessResults
}

func encode_SetResponseForDataBlock(w io.Writer, blockNumber uint32) (err error) {
	err = binary.Write(w, binary.BigEndian, blockNumber)
	if nil != err {
		errorLog("binary.Write() failed, err: %v", err)
		return err
	}

	return nil
}

func decode_SetResponseForDataBlock(r io.Reader) (err error, blockNumber uint32) {
	err = binary.Read(r, binary.BigEndian, &blockNumber)
	if nil != err {
		errorLog("binary.Read() failed, err: %v", err)
		return err, 0
	}

	return nil, blockNumber
}

func encode_SetResponseForLastDataBlock(w io.Writer, dataAccessResult DlmsDataAccessResult, blockNumber uint32) (err error) {
	err = binary.Write(w, binary.BigEndian, dataAccessResult)
	if nil != err {
		errorLog("binary.Write() failed, err: %v", err)
		return err
	}

	err = binary.Write(w, binary.BigEndian, blockNumber)
	if nil != err {
		errorLog("binary.Write() failed, err: %v", err)
		return err
	}

	return nil
}

func decode_SetResponseForLastDataBlock(r io.Reader) (err error, dataAccessResult DlmsDataAccessResult, blockNumber uint32) {
	err = binary.Read(r, binary.BigEndian, &dataAccessResult)
	if nil != err {
		errorLog("binary.Read() failed, err: %v", err)
		return err, 0, 0
	}

	err = binary.Read(r, binary.BigEndian, &blockNumber)
	if nil != err {
		errorLog("binary.Read() failed, err: %v", err)
		return err, 0, 0
	}

	return nil, dataAccessResult, blockNumber
}

func encode_SetResponseForLastDataBlockWithList(w io.Writer, dataAccessResults []DlmsDataAccessResult, blockNumber uint32) (err error) {
	count := uint8(len(dataAccessResults))
	err = binary.Write(w, binary.BigEndian, count)
	if nil != err {
		errorLog("binary.Write() failed, err: %v", err)
		return err
	}

	for i := uint8(0); i < count; i++ {
		err = binary.Write(w, binary.BigEndian, dataAccessResults[i])
		if nil != err {
			errorLog("binary.Write() failed, err: %v", err)
			return err
		}
	}

	err = binary.Write(w, binary.BigEndian, blockNumber)
	if nil != err {
		errorLog("binary.Write() failed, err: %v", err)
		return err
	}

	return nil
}

func decode_SetResponseForLastDataBlockWithList(r io.Reader) (err error, dataAccessResults []DlmsDataAccessResult, blockNumber uint32) {
	var count uint8
	err = binary.Read(r, binary.BigEndian, &count)
	if nil != err {
		errorLog("binary.Read() failed, err: %v", err)
		return err, nil, 0
	}

	dataAccessResults = make([]DlmsDataAccessResult, count)
	for i := uint8(0); i < count; i++ {
		err = binary.Read(r, binary.BigEndian, &dataAccessResults[i])
		if nil != err {
			errorLog("binary.Read() failed, err: %v", err)
			return err, nil, 0
		}
	}

	err = binary.Read(r, binary.BigEndian, &blockNumber)
	if nil != err {
		errorLog("binary.Read() failed, err: %v", err)
		return err, nil, 0
	}

	return nil, dataAccessResults, blockNumber
}

func encode_actionRequest(w io.Writer, classId DlmsClassId, instanceId *DlmsOid, methodId DlmsMethodId) (err error) {

	err = binary.Write(w, binary.BigEndian, classId)
	if nil != err {
		errorLog(fmt.Sprintf("binary.Write() failed, err: %s\n", err))
		return err
	}

	err = binary.Write(w, binary.BigEndian, instanceId)
	if nil != err {
		errorLog(fmt.Sprintf("binary.Write() failed, err: %s\n", err))
		return err
	}

	err = binary.Write(w, binary.BigEndian, methodId)
	if nil != err {
		errorLog(fmt.Sprintf("binary.Write() failed, err: %s\n", err))
		return err
	}

	return nil
}

func decode_actionRequest(r io.Reader) (err error, classId DlmsClassId, instanceId *DlmsOid, methodId DlmsMethodId) {
	var _classId DlmsClassId
	err = binary.Read(r, binary.BigEndian, &_classId)
	if nil != err {
		errorLog("binary.Read() failed, err: %v", err)
		return err, 0, nil, 0
	}

	_instanceId := new(DlmsOid)
	err = binary.Read(r, binary.BigEndian, _instanceId)
	if nil != err {
		errorLog("binary.Read() failed, err: %v", err)
		return err, _classId, nil, 0
	}

	var _methodId DlmsMethodId
	err = binary.Read(r, binary.BigEndian, &_methodId)
	if nil != err {
		errorLog("binary.Read() failed, err: %v", err)
		return err, _classId, _instanceId, 0
	}

	return err, _classId, _instanceId, _methodId

}

func encode_ActionRequestNormal(w io.Writer, classId DlmsClassId, instanceId *DlmsOid, methodId DlmsMethodId, methodParameters *DlmsData) (err error) {

	err = encode_actionRequest(w, classId, instanceId, methodId)
	if nil != err {
		return err
	}

	var isUsed uint8

	if nil == methodParameters {

		isUsed = 0
		err = binary.Write(w, binary.BigEndian, isUsed)
		if nil != err {
			errorLog(fmt.Sprintf("binary.Write() failed, err: %s\n", err))
			return err
		}

	} else {

		isUsed = 1
		err = binary.Write(w, binary.BigEndian, isUsed)
		if nil != err {
			errorLog(fmt.Sprintf("binary.Write() failed, err: %s\n", err))
			return err
		}
		err := methodParameters.Encode(w)
		if nil != err {
			return err
		}
	}
	return nil
}

func decode_ActionRequestNormal(r io.Reader) (err error, classId DlmsClassId, instanceId *DlmsOid, methodId DlmsMethodId, methodParameters *DlmsData) {

	err, classId, instanceId, methodId = decode_actionRequest(r)
	if nil != err {
		return err, classId, instanceId, methodId, nil
	}

	var isUsed uint8
	err = binary.Read(r, binary.BigEndian, &isUsed)
	if nil != err {
		errorLog("binary.Read() failed, err: %v", err)
		return err, classId, instanceId, methodId, nil
	}

	methodParameters = nil

	if isUsed > 0 {
		data := new(DlmsData)
		err = data.Decode(r)
		if nil != err {
			return err, classId, instanceId, methodId, nil
		}
		methodParameters = data
		return err, classId, instanceId, methodId, methodParameters
	} else {
		return err, classId, instanceId, methodId, nil
	}
}

func encode_ActionRequestWithFirstPblock(w io.Writer, classId DlmsClassId, instanceId *DlmsOid, methodId DlmsMethodId, lastBlock bool, blockNumber uint32, rawData []byte) (err error) {

	err = encode_actionRequest(w, classId, instanceId, methodId)
	if nil != err {
		return err
	}

	_lastBlock := uint8(0)
	if lastBlock {
		_lastBlock = 1
	}

	err = binary.Write(w, binary.BigEndian, _lastBlock)
	if nil != err {
		errorLog("binary.Write() failed, err: %v", err)
		return err
	}

	err = binary.Write(w, binary.BigEndian, blockNumber)
	if nil != err {
		errorLog("binary.Write() failed, err: %v", err)
		return err
	}

	err = encodeAxdrLength(w, uint16(len(rawData)))
	if nil != err {
		errorLog("encodeAxdrLength() failed, err: %v\n", err)
		return err
	}
	_, err = w.Write(rawData)
	if nil != err {
		errorLog("w.Wite() failed, err: %v\n", err)
		return err
	}

	return nil
}

func decode_ActionRequestWithFirstPblock(r io.Reader) (err error, classId DlmsClassId, instanceId *DlmsOid, methodId DlmsMethodId, lastBlock bool, blockNumber uint32, rawData []byte) {

	err, classId, instanceId, methodId = decode_actionRequest(r)
	if nil != err {
		return err, classId, instanceId, methodId, false, 0, nil
	}

	var _lastBlock uint8
	err = binary.Read(r, binary.BigEndian, &_lastBlock)
	if nil != err {
		errorLog("binary.Read() failed, err: %v", err)
		return err, classId, instanceId, methodId, false, 0, nil
	}
	if _lastBlock > 0 {
		lastBlock = true
	} else {
		lastBlock = false
	}

	err = binary.Read(r, binary.BigEndian, &blockNumber)
	if nil != err {
		errorLog("binary.Read() failed, err: %v", err)
		return err, classId, instanceId, methodId, lastBlock, 0, nil
	}

	err, length := decodeAxdrLength(r)
	if nil != err {
		errorLog("decodeAxdrLength() failed, err: %v\n", err)
		return err, classId, instanceId, methodId, lastBlock, blockNumber, nil
	}

	rawData = make([]byte, length)
	err = binary.Read(r, binary.BigEndian, rawData)
	if nil != err {
		errorLog("binary.Read() failed, err: %v", err)
		return err, classId, instanceId, methodId, lastBlock, blockNumber, nil
	}

	return err, classId, instanceId, methodId, lastBlock, blockNumber, rawData
}

func encode_ActionRequestWithPblock(w io.Writer, lastBlock bool, blockNumber uint32, rawData []byte) (err error) {

	_lastBlock := uint8(0)
	if lastBlock {
		_lastBlock = 1
	}

	err = binary.Write(w, binary.BigEndian, _lastBlock)
	if nil != err {
		errorLog("binary.Write() failed, err: %v", err)
		return err
	}

	err = binary.Write(w, binary.BigEndian, blockNumber)
	if nil != err {
		errorLog("binary.Write() failed, err: %v", err)
		return err
	}

	err = encodeAxdrLength(w, uint16(len(rawData)))
	if nil != err {
		errorLog("encodeAxdrLength() failed, err: %v\n", err)
		return err
	}
	_, err = w.Write(rawData)
	if nil != err {
		errorLog("w.Wite() failed, err: %v\n", err)
		return err
	}

	return nil
}

func decode_ActionRequestWithPblock(r io.Reader) (err error, lastBlock bool, blockNumber uint32, rawData []byte) {

	var _lastBlock uint8
	err = binary.Read(r, binary.BigEndian, &_lastBlock)
	if nil != err {
		errorLog("binary.Read() failed, err: %v", err)
		return err, false, 0, nil
	}
	if _lastBlock > 0 {
		lastBlock = true
	} else {
		lastBlock = false
	}

	err = binary.Read(r, binary.BigEndian, &blockNumber)
	if nil != err {
		errorLog("binary.Read() failed, err: %v", err)
		return err, lastBlock, 0, nil
	}

	err, length := decodeAxdrLength(r)
	if nil != err {
		errorLog("decodeAxdrLength() failed, err: %v\n", err)
		return err, lastBlock, blockNumber, nil
	}

	rawData = make([]byte, length)
	err = binary.Read(r, binary.BigEndian, rawData)
	if nil != err {
		errorLog("binary.Read() failed, err: %v", err)
		return err, lastBlock, blockNumber, nil
	}

	return err, lastBlock, blockNumber, rawData
}

func encode_ActionRequestNextPblock(w io.Writer, blockNumber uint32) (err error) {

	err = binary.Write(w, binary.BigEndian, blockNumber)
	if nil != err {
		errorLog("binary.Write() failed, err: %v", err)
		return err
	}
	return nil
}

func decode_ActionRequestNextPblock(r io.Reader) (err error, blockNumber uint32) {
	err = binary.Read(r, binary.BigEndian, &blockNumber)
	if nil != err {
		errorLog("binary.Read() failed, err: %v", err)
		return err, 0
	}

	return err, blockNumber
}

func encode_ActionRequestWithList(w io.Writer, classIds []DlmsClassId, instanceIds []*DlmsOid, methodIds []DlmsMethodId, methodParameters []*DlmsData) (err error) {
	count := uint8(len(classIds)) // count of requests

	err = binary.Write(w, binary.BigEndian, count)
	if nil != err {
		errorLog(fmt.Sprintf("binary.Write() failed, err: %s\n", err))
		return err
	}
	for i := uint8(0); i < count; i += 1 {
		err = encode_actionRequest(w, classIds[i], instanceIds[i], methodIds[i])
		if nil != err {
			return err
		}
	}

	err = binary.Write(w, binary.BigEndian, count)
	if nil != err {
		errorLog(fmt.Sprintf("binary.Write() failed, err: %s\n", err))
		return err
	}
	for i := uint8(0); i < count; i += 1 {
		var isUsed uint8

		if nil == methodParameters[i] {

			isUsed = 0
			err = binary.Write(w, binary.BigEndian, isUsed)
			if nil != err {
				errorLog(fmt.Sprintf("binary.Write() failed, err: %s\n", err))
				return err
			}

		} else {

			isUsed = 1
			err = binary.Write(w, binary.BigEndian, isUsed)
			if nil != err {
				errorLog(fmt.Sprintf("binary.Write() failed, err: %s\n", err))
				return err
			}
			err := methodParameters[i].Encode(w)
			if nil != err {
				return err
			}
		}
	}

	return nil
}

func decode_ActionRequestWithList(r io.Reader) (err error, classIds []DlmsClassId, instanceIds []*DlmsOid, methodIds []DlmsMethodId, methodParameters []*DlmsData) {
	var count uint8

	err = binary.Read(r, binary.BigEndian, &count)
	if nil != err {
		errorLog("binary.Read() failed, err: %v", err)
		return err, nil, nil, nil, nil
	}
	classIds = make([]DlmsClassId, count)
	instanceIds = make([]*DlmsOid, count)
	methodIds = make([]DlmsMethodId, count)
	methodParameters = make([]*DlmsData, count)

	for i := uint8(0); i < count; i += 1 {
		err, classId, instanceId, methodId := decode_actionRequest(r)
		if nil != err {
			return err, classIds[0:i], instanceIds[0:i], methodIds[0:i], nil
		}
		classIds[i] = classId
		instanceIds[i] = instanceId
		methodIds[i] = methodId
	}

	err = binary.Read(r, binary.BigEndian, &count)
	if nil != err {
		errorLog("binary.Read() failed, err: %v", err)
		return err, classIds, instanceIds, methodIds, nil
	}
	if int(count) < len(classIds) {
		err = fmt.Errorf("missing data")
		return err, classIds, instanceIds, methodIds, nil
	}

	methodParameters = make([]*DlmsData, count)

	for i := uint8(0); i < count; i += 1 {
		methodParameters[i] = nil

		var isUsed uint8
		err = binary.Read(r, binary.BigEndian, &isUsed)
		if nil != err {
			errorLog("binary.Read() failed, err: %v", err)
			return err, classIds, instanceIds, methodIds, methodParameters[0:i]
		}

		if isUsed > 0 {
			data := new(DlmsData)
			err = data.Decode(r)
			if nil != err {
				return err, classIds, instanceIds, methodIds, methodParameters
			}
			methodParameters[i] = data
		} else {
			methodParameters[i] = nil
		}
	}

	return nil, classIds, instanceIds, methodIds, methodParameters
}

func encode_ActionRequestWithListAndFirstPblock(w io.Writer, classIds []DlmsClassId, instanceIds []*DlmsOid, methodIds []DlmsMethodId, lastBlock bool, blockNumber uint32, rawData []byte) (err error) {
	count := uint8(len(classIds)) // count of requests

	err = binary.Write(w, binary.BigEndian, count)
	if nil != err {
		errorLog(fmt.Sprintf("binary.Write() failed, err: %s\n", err))
		return err
	}
	for i := uint8(0); i < count; i += 1 {
		err = encode_actionRequest(w, classIds[i], instanceIds[i], methodIds[i])
		if nil != err {
			return err
		}
	}

	_lastBlock := uint8(0)
	if lastBlock {
		_lastBlock = 1
	}
	blockNumber = uint32(1)

	err = binary.Write(w, binary.BigEndian, _lastBlock)
	if nil != err {
		errorLog("binary.Write() failed, err: %v", err)
		return err
	}

	err = binary.Write(w, binary.BigEndian, blockNumber)
	if nil != err {
		errorLog("binary.Write() failed, err: %v", err)
		return err
	}

	err = encodeAxdrLength(w, uint16(len(rawData)))
	if nil != err {
		errorLog("encodeAxdrLength() failed, err: %v\n", err)
		return err
	}
	_, err = w.Write(rawData)
	if nil != err {
		errorLog("w.Wite() failed, err: %v\n", err)
		return err
	}

	return nil
}

func decode_ActionRequestWithListAndFirstPblock(r io.Reader) (err error, classIds []DlmsClassId, instanceIds []*DlmsOid, methodIds []DlmsMethodId, lastBlock bool, blockNumber uint32, rawData []byte) {
	var count uint8

	err = binary.Read(r, binary.BigEndian, &count)
	if nil != err {
		errorLog("binary.Read() failed, err: %v", err)
		return err, nil, nil, nil, false, 0, nil
	}
	classIds = make([]DlmsClassId, count)
	instanceIds = make([]*DlmsOid, count)
	methodIds = make([]DlmsMethodId, count)

	for i := uint8(0); i < count; i += 1 {
		err, classId, instanceId, methodId := decode_actionRequest(r)
		if nil != err {
			return err, classIds[0:i], instanceIds[0:i], methodIds[0:i], false, 0, nil
		}
		classIds[i] = classId
		instanceIds[i] = instanceId
		methodIds[i] = methodId
	}

	var _lastBlock uint8
	err = binary.Read(r, binary.BigEndian, &_lastBlock)
	if nil != err {
		errorLog("binary.Read() failed, err: %v", err)
		return err, classIds, instanceIds, methodIds, false, 0, nil
	}
	if _lastBlock > 0 {
		lastBlock = true
	} else {
		lastBlock = false
	}

	err = binary.Read(r, binary.BigEndian, &blockNumber)
	if nil != err {
		errorLog("binary.Read() failed, err: %v", err)
		return err, classIds, instanceIds, methodIds, lastBlock, 0, nil
	}

	err, length := decodeAxdrLength(r)
	if nil != err {
		errorLog("decodeAxdrLength() failed, err: %v\n", err)
		return err, classIds, instanceIds, methodIds, lastBlock, blockNumber, nil
	}

	rawData = make([]byte, length)
	err = binary.Read(r, binary.BigEndian, rawData)
	if nil != err {
		errorLog("binary.Read() failed, err: %v", err)
		return err, classIds, instanceIds, methodIds, lastBlock, blockNumber, nil
	}

	return nil, classIds, instanceIds, methodIds, lastBlock, blockNumber, rawData

}

func encode_ActionResponseNormal(w io.Writer, actionResult DlmsActionResult, dataAccessResult *DlmsDataAccessResult, data *DlmsData) (err error) {
	//func encode_getResponse(w io.Writer, dataAccessResult DlmsDataAccessResult, data *DlmsData) (err error) {

	err = binary.Write(w, binary.BigEndian, actionResult)
	if nil != err {
		errorLog(fmt.Sprintf("binary.Write() failed, err: %s\n", err))
		return err
	}

	var isUsed uint8
	if nil == dataAccessResult {
		isUsed = 0
		err = binary.Write(w, binary.BigEndian, isUsed)
		if nil != err {
			errorLog(fmt.Sprintf("binary.Write() failed, err: %s\n", err))
			return err
		}
	} else {
		isUsed = 1

		err = binary.Write(w, binary.BigEndian, isUsed)
		if nil != err {
			errorLog(fmt.Sprintf("binary.Write() failed, err: %s\n", err))
			return err
		}

		err = encode_getResponse(w, *dataAccessResult, data)
	}

	return err

}

func decode_ActionResponseNormal(r io.Reader) (err error, actionResult DlmsActionResult, dataAccessResult *DlmsDataAccessResult, data *DlmsData) {

	err = binary.Read(r, binary.BigEndian, &actionResult)
	if nil != err {
		errorLog("binary.Read() failed, err: %v", err)
		return err, 0, nil, nil
	}

	var isUsed uint8
	err = binary.Read(r, binary.BigEndian, &isUsed)
	if nil != err {
		errorLog("binary.Read() failed, err: %v", err)
		return err, actionResult, nil, nil
	}

	if 0 == isUsed {
		return err, actionResult, nil, nil
	} else {
		err, _dataAccessResult, _data := decode_getResponse(r)
		if nil == err {
			dataAccessResult = new(DlmsDataAccessResult)
			*dataAccessResult = _dataAccessResult
			data = _data
			return nil, actionResult, dataAccessResult, data
		} else {
			return err, actionResult, nil, nil
		}
	}

}

func encode_ActionResponseWithPblock(w io.Writer, lastBlock bool, blockNumber uint32, rawData []byte) (err error) {

	_lastBlock := uint8(0)
	if lastBlock {
		_lastBlock = 1
	}

	err = binary.Write(w, binary.BigEndian, _lastBlock)
	if nil != err {
		errorLog("binary.Write() failed, err: %v", err)
		return err
	}

	err = binary.Write(w, binary.BigEndian, blockNumber)
	if nil != err {
		errorLog("binary.Write() failed, err: %v", err)
		return err
	}

	err = encodeAxdrLength(w, uint16(len(rawData)))
	if nil != err {
		errorLog("encodeAxdrLength() failed, err: %v\n", err)
		return err
	}
	_, err = w.Write(rawData)
	if nil != err {
		errorLog("w.Wite() failed, err: %v\n", err)
		return err
	}

	return nil
}

func decode_ActionResponseWithPblock(r io.Reader) (err error, lastBlock bool, blockNumber uint32, rawData []byte) {

	var _lastBlock uint8
	err = binary.Read(r, binary.BigEndian, &_lastBlock)
	if nil != err {
		errorLog("binary.Read() failed, err: %v", err)
		return err, false, 0, nil
	}
	if _lastBlock > 0 {
		lastBlock = true
	} else {
		lastBlock = false
	}

	err = binary.Read(r, binary.BigEndian, &blockNumber)
	if nil != err {
		errorLog("binary.Read() failed, err: %v", err)
		return err, lastBlock, 0, nil
	}

	err, length := decodeAxdrLength(r)
	if nil != err {
		errorLog("decodeAxdrLength() failed, err: %v\n", err)
		return err, lastBlock, blockNumber, nil
	}

	rawData = make([]byte, length)
	err = binary.Read(r, binary.BigEndian, rawData)
	if nil != err {
		errorLog("binary.Read() failed, err: %v", err)
		return err, lastBlock, blockNumber, nil
	}

	return err, lastBlock, blockNumber, rawData
}

func encode_ActionResponseNextPblock(w io.Writer, blockNumber uint32) (err error) {

	err = binary.Write(w, binary.BigEndian, blockNumber)
	if nil != err {
		errorLog("binary.Write() failed, err: %v", err)
		return err
	}
	return nil
}

func decode_ActionResponseNextPblock(r io.Reader) (err error, blockNumber uint32) {
	err = binary.Read(r, binary.BigEndian, &blockNumber)
	if nil != err {
		errorLog("binary.Read() failed, err: %v", err)
		return err, 0
	}

	return err, blockNumber
}

func encode_ActionResponseWithList(w io.Writer, actionResults []DlmsActionResult, dataAccessResults []*DlmsDataAccessResult, data []*DlmsData) (err error) {
	count := uint8(len(actionResults)) // count of requests

	err = binary.Write(w, binary.BigEndian, count)
	if nil != err {
		errorLog(fmt.Sprintf("binary.Write() failed, err: %s\n", err))
		return err
	}

	for i := uint8(0); i < count; i++ {

		err = binary.Write(w, binary.BigEndian, actionResults[i])
		if nil != err {
			errorLog(fmt.Sprintf("binary.Write() failed, err: %s\n", err))
			return err
		}

		var isUsed uint8
		if nil == dataAccessResults[i] {
			isUsed = 0
			err = binary.Write(w, binary.BigEndian, isUsed)
			if nil != err {
				errorLog(fmt.Sprintf("binary.Write() failed, err: %s\n", err))
				return err
			}
		} else {
			isUsed = 1

			err = binary.Write(w, binary.BigEndian, isUsed)
			if nil != err {
				errorLog(fmt.Sprintf("binary.Write() failed, err: %s\n", err))
				return err
			}

			err = encode_getResponse(w, *(dataAccessResults[i]), data[i])
			if nil != err {
				return err
			}
		}
	}

	return nil
}

func decode_ActionResponseWithList(r io.Reader) (err error, actionResults []DlmsActionResult, dataAccessResults []*DlmsDataAccessResult, data []*DlmsData) {
	var count uint8

	err = binary.Read(r, binary.BigEndian, &count)
	if nil != err {
		errorLog("binary.Read() failed, err: %v", err)
		return err, nil, nil, nil
	}
	actionResults = make([]DlmsActionResult, count)
	dataAccessResults = make([]*DlmsDataAccessResult, count)
	data = make([]*DlmsData, count)

	for i := uint8(0); i < count; i++ {

		err = binary.Read(r, binary.BigEndian, &actionResults[i])
		if nil != err {
			errorLog("binary.Read() failed, err: %v", err)
			return err, nil, nil, nil
		}

		var isUsed uint8
		err = binary.Read(r, binary.BigEndian, &isUsed)
		if nil != err {
			errorLog("binary.Read() failed, err: %v", err)
			return err, actionResults[0:i], nil, nil
		}

		if 0 == isUsed {
			continue
		} else {
			err, _dataAccessResult, _data := decode_getResponse(r)
			if nil == err {
				dataAccessResults[i] = new(DlmsDataAccessResult)
				*(dataAccessResults[i]) = _dataAccessResult
				data[i] = _data
				continue
			} else {
				return err, actionResults[0:i], dataAccessResults[0:i], data[0:i]
			}
		}

	}
	return err, actionResults, dataAccessResults, data

}
