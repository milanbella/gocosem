// -build ignore

package gocosem

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"time"
)

const DEBUG_ASN1 = false

// asn1 simple types

type tAsn1BitString struct {
	buf        []uint8
	bitsUnused int
}

type tAsn1IA5String string
type tAsn1Integer int32
type tAsn1Integer8 int8
type tAsn1Integer16 int16
type tAsn1Integer32 int32
type tAsn1Long64 int64
type tAsn1Unsigned8 uint8
type tAsn1Unsigned16 uint16
type tAsn1Unsigned32 uint32
type tAsn1UnsignedLong64 uint64
type tAsn1Float float32
type tAsn1Float32 float32
type tAsn1Float64 float64
type tAsn1ObjectIdentifier []uint32
type tAsn1OctetString []uint8
type tAsn1PrintableString string
type tAsn1VisibleString []byte
type tAsn1T61String string
type tAsn1UTCTime time.Time
type tAsn1GraphicString []byte
type tAsn1Any []byte
type tAsn1Null int
type tAsn1Boolean bool
type tAsn1DateTime []byte
type tAsn1Date []byte
type tAsn1Time []byte

type tAsn1Choice struct {
	tag int
	val interface{}
}

type tAsn1CosemAuthenticationValueOther struct {
	otherMechanismName  tAsn1ObjectIdentifier
	otherMechanismValue tAsn1OctetString
}

func (ch *tAsn1Choice) setVal(tag int, val interface{}) {
	ch.tag = tag
	ch.val = val
}

func (ch *tAsn1Choice) getTag() int {
	return ch.tag
}

func (ch *tAsn1Choice) getVal() interface{} {
	return ch.val
}

type tAuthenticationValueOther struct {
	otherMechanismName  tAsn1ObjectIdentifier
	otherMechanismValue tAsn1Any
}

type AARQapdu struct {
	//protocol-version [0] IMPLICIT T-protocol-version DEFAULT {version1},
	protocolVersion *tAsn1BitString

	//application-context-name [1] Application-context-name,
	applicationContextName tAsn1ObjectIdentifier

	//called-AP-title [2] AP-title OPTIONAL,
	calledAPtitle *tAsn1OctetString

	//called-AE-qualifier [3] AE-qualifier OPTIONAL,
	calledAEqualifier *tAsn1OctetString

	//called-AP-invocation-id [4] AP-invocation-identifier OPTIONAL,
	calledAPinvocationId *tAsn1Integer

	//called-AE-invocation-id [5] AE-invocation-identifier OPTIONAL,
	calledAEinvocationId *tAsn1Integer

	//calling-AP-title [6] AP-title OPTIONAL,
	callingAPtitle *tAsn1OctetString

	//calling-AE-qualifier [7] AE-qualifier OPTIONAL,
	callingAEqualifier *tAsn1OctetString

	//calling-AP-invocation-id [8] AP-invocation-identifier OPTIONAL,
	callingAPinvocationId *tAsn1Integer

	//calling-AE-invocation-id [9] AE-invocation-identifier OPTIONAL,
	callingAEinvocationId *tAsn1Integer

	//-- The following field shall not be present if only the kernel is used.
	//sender-acse-requirements [10] IMPLICIT ACSE-requirements OPTIONAL,
	senderAcseRequirements *tAsn1BitString

	//-- The following field shall only be present if the authentication functional unit is selected.
	//mechanism-name [11] IMPLICIT Mechanism-name OPTIONAL,
	mechanismName *tAsn1ObjectIdentifier

	//-- The following field shall only be present if the authentication functional unit is selected.
	//calling-authentication-value [12] EXPLICIT Authentication-value OPTIONAL,
	callingAuthenticationValue *tAsn1Choice

	//implementation-information [29] IMPLICIT Implementation-data OPTIONAL,
	implementationInformation *tAsn1GraphicString

	//user-information [30] EXPLICIT Association-information OPTIONAL
	userInformation *tAsn1OctetString
}

//AARE-apdu ::= [APPLICATION 1] IMPLICIT SEQUENCE
type AAREapdu struct {
	//-- [APPLICATION 1] == [ 61H ] = [ 97 ]
	//protocol-version [0] IMPLICIT T-protocol-version DEFAULT {version1},
	protocolVersion *tAsn1BitString

	//application-context-name [1] Application-context-name,
	applicationContextName tAsn1ObjectIdentifier

	//result [2] Association-result,
	result tAsn1Integer

	//result-source-diagnostic [3] Associate-source-diagnostic,
	resultSourceDiagnostic tAsn1Choice

	//responding-AP-title [4] AP-title OPTIONAL,
	respondingAPtitle *tAsn1OctetString

	//responding-AE-qualifier [5] AE-qualifier OPTIONAL,
	respondingAEqualifier *tAsn1OctetString

	//responding-AP-invocation-id [6] AP-invocation-identifier OPTIONAL,
	respondingAPinvocationId *tAsn1Integer

	//responding-AE-invocation-id [7] AE-invocation-identifier OPTIONAL,
	respondingAEinvocationId *tAsn1Integer

	//-- The following field shall not be present if only the kernel is used.
	//responder-acse-requirements [8] IMPLICIT ACSE-requirements OPTIONAL,
	responderAcseRequirements *tAsn1BitString

	//-- The following field shall only be present if the authentication functional unit is selected.
	//mechanism-name [9] IMPLICIT Mechanism-name OPTIONAL,
	mechanismName *tAsn1ObjectIdentifier

	//-- The following field shall only be present if the authentication functional unit is selected.
	//responding-authentication-value [10] EXPLICIT Authentication-value OPTIONAL,
	respondingAuthenticationValue *tAsn1Choice

	//implementation-information [29] IMPLICIT Implementation-data OPTIONAL,
	implementationInformation *tAsn1GraphicString

	//user-information [30] EXPLICIT Association-information OPTIONAL
	userInformation *tAsn1OctetString
}

const ASN1_CLASS_UNIVERSAL = 0x00
const ASN1_CLASS_APPLICATION = 0x01
const ASN1_CLASS_CONTEXT_SPECIFIC = 0x02
const ASN1_CLASS_PRIVATE = 0x04

const BER_ENCODING_PRIMITIVE = 0x00
const BER_ENCODING_CONSTRUCTED = 0x01

type t_der_encode_contect_fn func(buf bytes.Buffer, val interface{}) (err error)
type t_der_decode_contect_fn func(r io.Reader, ch *t_der_chunk, val interface{}) (err error)

type t_der_chunk struct {
	asn1_class uint8
	encoding   uint8
	asn1_tag   uint32
	content    []byte
	length     uint32
}

func der_print_chunk(ch *t_der_chunk) string {

	var asn1_class string

	if ASN1_CLASS_UNIVERSAL == ch.asn1_class {
		asn1_class = "u"
	} else if ASN1_CLASS_APPLICATION == ch.asn1_class {
		asn1_class = "a"
	} else if ASN1_CLASS_CONTEXT_SPECIFIC == ch.asn1_class {
		asn1_class = "c"
	} else if ASN1_CLASS_PRIVATE == ch.asn1_class {
		asn1_class = "p"
	} else {
		asn1_class = "u"
	}

	var encoding string

	if BER_ENCODING_PRIMITIVE == ch.encoding {
		encoding = "p"
	} else if BER_ENCODING_CONSTRUCTED == ch.encoding {
		encoding = "c"
	} else {
		encoding = "u"
	}

	tag := fmt.Sprintf("%3d", ch.asn1_tag)

	length := fmt.Sprintf("%5d", ch.length)

	content := fmt.Sprintf("%X", ch.content)

	chunk := fmt.Sprintf("%s %s %s %s %s", asn1_class, encoding, tag, length, content)

	return chunk
}

func encode_uint32_base128(w io.Writer, val uint32) (err error) {

	b := make([]uint8, 1)

	if val <= uint32(0x7f) {
		b[0] = uint8(val & 0x7f)
		_, err := w.Write(b)
		if nil != err {
			errorLog("io.Write(): %v", err)
			return err
		}
	} else if val <= uint32(0x3fff) {
		b[0] = uint8((val&0x3f80)>>7) | 0x80
		_, err := w.Write(b)
		if nil != err {
			errorLog("io.Write(): %v", err)
			return err
		}
		b[0] = uint8(val & 0x7f)
		_, err = w.Write(b)
		if nil != err {
			errorLog("io.Write(): %v", err)
			return err
		}
	} else if val <= uint32(0x1fffff) {
		b[0] = uint8((val&0x1fc000)>>14) | 0x80
		_, err := w.Write(b)
		if nil != err {
			errorLog("io.Write(): %v", err)
			return err
		}
		b[0] = uint8((val&0x3f80)>>7) | 0x80
		_, err = w.Write(b)
		if nil != err {
			errorLog("io.Write(): %v", err)
			return err
		}
		b[0] = uint8(val & 0x7f)
		_, err = w.Write(b)
		if nil != err {
			errorLog("io.Write(): %v", err)
			return err
		}
	} else if val <= uint32(0x0fffffff) {
		b[0] = uint8((val&0xfe00000)>>21) | 0x80
		_, err := w.Write(b)
		if nil != err {
			errorLog("io.Write(): %v", err)
			return err
		}
		b[0] = uint8((val&0x1fc000)>>14) | 0x80
		_, err = w.Write(b)
		if nil != err {
			errorLog("io.Write(): %v", err)
			return err
		}
		b[0] = uint8((val&0x3f80)>>7) | 0x80
		_, err = w.Write(b)
		if nil != err {
			errorLog("io.Write(): %v", err)
			return err
		}
		b[0] = uint8(val & 0x7f)
		_, err = w.Write(b)
		if nil != err {
			errorLog("io.Write(): %v", err)
			return err
		}
	} else if val <= uint32(0xffffffff) {
		b[0] = uint8(val&0x00000000>>28) | 0x80
		_, err := w.Write(b)
		if nil != err {
			errorLog("io.Write(): %v", err)
			return err
		}
		b[0] = uint8(val&0xfe00000>>21) | 0x80
		_, err = w.Write(b)
		if nil != err {
			errorLog("io.Write(): %v", err)
			return err
		}
		b[0] = uint8(val&0x1fc000>>14) | 0x80
		_, err = w.Write(b)
		if nil != err {
			errorLog("io.Write(): %v", err)
			return err
		}
		b[0] = uint8(val&0x3f80>>7) | 0x80
		_, err = w.Write(b)
		if nil != err {
			errorLog("io.Write(): %v", err)
			return err
		}
		b[0] = uint8(val & 0x7f)
		_, err = w.Write(b)
		if nil != err {
			errorLog("io.Write(): %v", err)
			return err
		}
	} else {
		errorLog("asserion failed", err)
		return err
	}

	return nil
}

func decode_uint32_base128(r io.Reader) (err error, val uint32) {

	b := make([]byte, 1)

	_, err = r.Read(b)
	if nil != err {
		errorLog("io.Read(): %v", err)
		return err, 0
	}
	val = uint32(0x7f & b[0])
	if b[0]&0x80 > 0 {
		_, err = r.Read(b)
		if nil != err {
			errorLog("io.Read(): %v", err)
			return err, val
		}
		val = (val << 7) | uint32(0x7f&b[0])
		if uint32(b[0]&0x80) > 0 {
			_, err = r.Read(b)
			if nil != err {
				errorLog("io.Read(): %v", err)
				return err, val
			}
			val = (val << 7) | uint32(0x7f&b[0])
			if b[0]&0x80 > 0 {
				_, err = r.Read(b)
				if nil != err {
					errorLog("io.Read(): %v", err)
					return err, val
				}
				val = (val << 7) | uint32(0x7f&b[0])
				if b[0]&0x80 > 0 {
					_, err = r.Read(b)
					if nil != err {
						errorLog("io.Read(): %v", err)
						return err, val
					}
					if b[0] > 0x0f {
						err = fmt.Errorf("value of tag exceeds limit: %v", math.MaxUint32)
						errorLog("%s", err)
						return err, val
					}
					val = (val << 4) | uint32(0x7f&b[0])
				}
			}
		}
	}
	return err, val
}

func der_encode_Integer(i tAsn1Integer) (err error, content []uint8) {

	var _i int64 = int64(i)

	// compute minimum number of bytes needed for encoding integer

	n := 1

	for _i > 127 { // minimum bytes if integer is positive
		n++
		_i >>= 8
	}
	for i < -128 { // minimum bytes if integer is negative
		n++
		i >>= 8
	}

	content = make([]uint8, n)
	for j := 0; j < n; j++ {
		content[j] = uint8(_i >> uint((n-1-j)*8))
	}

	return nil, content

}

func der_decode_Integer(content []uint8) (err error, i tAsn1Integer) {

	if len(content) < 1 {
		err = fmt.Errorf("decoding error: empty integer")
		errorLog("%s", err)
		return err, 0
	}
	if len(content) > 1 {
		if (content[0] == 0 && content[1]&0x80 == 0) || (content[0] == 0xff && content[1]&0x80 == 0x80) {
			err = fmt.Errorf("decoding error: not minimally encoded integer")
			errorLog("%s", err)
			return err, 0
		}
	}

	var _i int64

	if len(content) > 8 {
		err = fmt.Errorf("decoding error: integer is too big")
		errorLog("%s", err)
		return err, 0
	}

	for j := 0; j < len(content); j++ {
		_i <<= 8
		_i |= int64(content[j])
	}

	// shift left and right to sign extend
	// https://en.wikipedia.org/wiki/Sign_extension

	_i <<= 64 - uint8(len(content))*8
	_i >>= 64 - uint8(len(content))*8

	if _i != int64(int32(_i)) {
		err = fmt.Errorf("decoding error: integer is too big")
		errorLog("%s", err)
		return err, 0
	}

	return nil, tAsn1Integer(_i)
}

func der_encode_BitString(bs *tAsn1BitString) (err error, content []uint8) {

	if nil == bs || nil == bs.buf {
		content = make([]uint8, 0)
		return nil, content
	}

	if 0 == len(bs.buf) {
		content = make([]uint8, 1)
		content[0] = 0
		return nil, content
	}

	content = make([]uint8, len(bs.buf)+1)
	if bs.bitsUnused > 7 {
		err = fmt.Errorf("wrong count of unused bits")
		errorLog("%v", err)
		return err, nil
	}
	content[0] = uint8(bs.bitsUnused)
	for i := 0; i < len(bs.buf); i++ {
		content[i+1] = bs.buf[i]
	}
	return nil, content
}

func der_decode_BitString(content []uint8) (err error, bs *tAsn1BitString) {
	if len(content) < 1 {
		return nil, nil
	}

	var _bs tAsn1BitString
	_bs.buf = make([]byte, len(content)-1)

	_bs.bitsUnused = int(content[0])
	for i := 0; i < len(content)-1; i++ {
		_bs.buf[i] = content[i+1]
	}

	return nil, &_bs

}

func der_encode_ObjectIdentifier(oi *tAsn1ObjectIdentifier) (err error, content []uint8) {

	if nil == oi {
		content = make([]uint8, 0)
		return nil, content
	}

	_oi := ([]uint32)(*oi)

	if 0 == len(_oi) {
		content = make([]uint8, 0)
		return nil, content
	}

	if len(_oi) < 2 {
		panic("object identifier must contain at least 2 components")
	}
	if _oi[0] > 2 {
		panic("first component value must be in range 0-2")
	}
	if _oi[0] < 2 {
		if _oi[1] > 39 {
			panic("second component must be in range 0-39")
		}
	}

	var buf bytes.Buffer
	err = encode_uint32_base128(&buf, 40*_oi[0]+_oi[1])
	if nil != err {
		panic(err)
	}
	for i := 2; i < len(_oi); i++ {
		err := encode_uint32_base128(&buf, _oi[i])
		if nil != err {
			return err, nil
		}
	}

	content = buf.Bytes()

	return nil, content
}

func der_decode_ObjectIdentifier(content []uint8) (err error, oi *tAsn1ObjectIdentifier) {

	if len(content) < 1 {
		return nil, nil
	}

	buf := bytes.NewReader(content)
	COMPONENTS_BUFFER_SIZE := 100
	components := make([]uint32, COMPONENTS_BUFFER_SIZE)
	var i int
	var component uint32

	for i = 0; i < len(components); i++ {

		if buf.Len() > 0 {
			err, component = decode_uint32_base128(buf)
			if nil != err {
				return err, nil
			}
		} else {
			break
		}

		if 0 == i {
			if (0 >= component) && (39 <= component) {
				components[0] = 0
				components[1] = component
			} else if (40 >= component) && (79 <= component) {
				components[0] = 1
				components[1] = component - 40
			} else {
				components[0] = 2
				components[1] = component - 80
			}
			i = 1
		} else {
			components[i] = component
		}
	}

	if i == len(components) {
		if buf.Len() > 0 { // there are still some components remaining to be read
			err = fmt.Errorf("COMPONENTS_BUFFER_SIZE small")
			errorLog("%v", err)
			return err, nil
		}
	}
	components = components[0:i]

	return nil, (*tAsn1ObjectIdentifier)(&components)
}

func der_encode_chunk(w io.Writer, ch *t_der_chunk) (err error) {

	if DEBUG_ASN1 {
		fmt.Printf("%s\n", der_print_chunk(ch))
	}

	b := make([]byte, 1)

	// class

	b[0] = ch.asn1_class << 6

	// encoding

	b[0] |= ch.encoding << 5

	// tag

	if ch.asn1_tag < 31 {
		b[0] |= uint8(ch.asn1_tag)
		_, err := w.Write(b)
		if nil != err {
			errorLog("io.Write(): %v", err)
			return err
		}
	} else {
		b[0] |= 0x1f
		_, err := w.Write(b)
		if nil != err {
			errorLog("io.Write(): %v", err)
			return err
		}
		err = encode_uint32_base128(w, ch.asn1_tag)
		if nil != err {
			return err
		}
	}

	length := uint64(len(ch.content))

	// length

	if length <= 0x7f {
		b[0] = uint8(length)
		_, err := w.Write(b)
		if nil != err {
			errorLog("io.Write(): %v", err)
			return err
		}
	} else {
		var m int
		if length <= 0xff {
			m = 1
		} else if length <= 0xffff {
			m = 2
		} else if length <= 0xffffff {
			m = 3
		} else if length <= 0xffffffff {
			m = 4
		} else if length <= 0xffffffffff {
			m = 5
		} else if length <= 0xffffffffffff {
			m = 6
		} else if length <= 0xffffffffffffff {
			m = 7
		} else if length <= 0xffffffffffffffff {
			m = 8
		} else {
			err = fmt.Errorf("asserion failed")
			errorLog("%v", err)
			return err
		}
		b[0] = uint8(m) | 0x10
		_, err := w.Write(b)
		if nil != err {
			errorLog("io.Write(): %v", err)
			return err
		}
		for i := 0; i < m; i++ {
			b[0] = uint8((length >> uint(i)) & 0xff)
			_, err := w.Write(b)
			if nil != err {
				errorLog("io.Write(): %v", err)
				return err
			}
		}
	}

	_, err = w.Write(ch.content)
	if nil != err {
		errorLog("io.Write(): %v", err)
		return err
	}

	return err
}

func der_decode_chunk(r io.Reader) (err error, _ch *t_der_chunk) {
	var m int
	var ch t_der_chunk

	b := make([]byte, 1)

	if DEBUG_ASN1 {
		// print bytes to be parsed

		var _err error
		var pbuf bytes.Buffer
		for {
			_, _err = r.Read(b)
			if io.EOF == _err {
				break
			}
			if nil != _err {
				errorLog("io.Read(): %v", err)
				return _err, nil
			}
			_, _err = pbuf.Write(b)
			if nil != _err {
				errorLog("pbuf.Write(): %v", err)
				return _err, nil
			}
		}
		readerContent := pbuf.Bytes()
		fmt.Printf("%X\n", readerContent)
		r = bytes.NewReader(readerContent)

	}

	_, err = r.Read(b)
	if nil != err {
		errorLog("io.Read(): %v", err)
		return err, nil
	}
	ch.length++

	// class

	ch.asn1_class = (b[0] & 0xc0) >> 6

	// encoding

	ch.encoding = (b[0] & 0x20) >> 5

	// tag

	if 0x1f == b[0]&0x1F {
		err, ch.asn1_tag = decode_uint32_base128(r)
		if nil != err {
			return err, nil
		}
		if ch.asn1_tag < 31 {
			err = fmt.Errorf("wrong tag value")
			errorLog("%s", err)
			return err, nil
		}
	} else {
		ch.asn1_tag = uint32(0x1f & b[0])
		if ch.asn1_tag > 30 {
			err = fmt.Errorf("wrong tag value")
			errorLog("%s", err)
			return err, nil
		}
	}

	// length

	var length uint64

	_, err = r.Read(b)
	if nil != err {
		errorLog("io.Read(): %v", err)
		return err, nil
	}
	ch.length++
	if 0x80 == b[0] {
		err = fmt.Errorf("wrog encoding: length 0x80")
		errorLog("%s", err)
		return err, nil
	} else if 0x00 == b[0]&0x80 {
		length = uint64(b[0] & 0x7f)
	} else {
		m = int(0x7f & b[0])
		if m > 8 {
			err = fmt.Errorf("value of length exceeds limit")
			errorLog("%s", err)
			return err, nil
		}
		_, err := r.Read(b)
		if nil != err {
			errorLog("io.Read(): %v", err)
			return err, nil
		}
		ch.length++
		length = uint64(b[0])
		for i := 0; i < m-1; m-- {
			_, err := r.Read(b)
			if nil != err {
				errorLog("io.Read(): %v", err)
				return err, nil
			}
			ch.length++
			length = (length << 8) | uint64(b[0])
		}
	}

	if length > 1024 { // guard against allocating too much
		err = fmt.Errorf("content too long: %v bytes", length)
		errorLog("%s", err)
		return err, nil
	}
	ch.content = make([]byte, length)
	_, err = r.Read(ch.content)
	if nil != err {
		errorLog("io.Read(): %v", err)
		return err, nil
	}
	ch.length += uint32(len(ch.content))

	if DEBUG_ASN1 {
		fmt.Printf("%s\n", der_print_chunk(&ch))
	}

	return nil, &ch

}

func encode_AARQapdu(w io.Writer, aarq *AARQapdu) (err error) {
	var ch, ch1, ch2, ch3, chAARQ *t_der_chunk
	var buf, bufAARQ *bytes.Buffer

	if nil == aarq {
		return nil
	}

	// AARQ-apdu ::= [APPLICATION 0] IMPLICIT SEQUENCE

	chAARQ = new(t_der_chunk)
	chAARQ.asn1_class = ASN1_CLASS_APPLICATION
	chAARQ.encoding = BER_ENCODING_CONSTRUCTED
	chAARQ.asn1_tag = 0

	bufAARQ = new(bytes.Buffer)

	// protocol-version [0] IMPLICIT T-protocol-version DEFAULT {version1}

	if nil != aarq.protocolVersion {
		ch = new(t_der_chunk)
		ch.asn1_class = ASN1_CLASS_CONTEXT_SPECIFIC
		ch.encoding = BER_ENCODING_PRIMITIVE
		ch.asn1_tag = 0
		err, ch.content = der_encode_BitString(aarq.protocolVersion)
		if nil != err {
			return err
		}

		err = der_encode_chunk(bufAARQ, ch)
		if nil != err {
			return err
		}
	}

	// application-context-name [1] Application-context-name,

	if nil != aarq.applicationContextName {
		ch = new(t_der_chunk)
		ch.asn1_class = ASN1_CLASS_CONTEXT_SPECIFIC
		ch.encoding = BER_ENCODING_CONSTRUCTED
		ch.asn1_tag = 1

		ch1 = new(t_der_chunk)
		ch1.asn1_class = ASN1_CLASS_UNIVERSAL
		ch1.encoding = BER_ENCODING_PRIMITIVE
		ch1.asn1_tag = 6
		err, ch1.content = der_encode_ObjectIdentifier(&aarq.applicationContextName)
		if nil != err {
			return err
		}
		buf = new(bytes.Buffer)
		err = der_encode_chunk(buf, ch1)
		if nil != err {
			return err
		}

		ch.content = buf.Bytes()

		err = der_encode_chunk(bufAARQ, ch)
		if nil != err {
			return err
		}
	}

	// called-AP-title [2] AP-title OPTIONAL,

	if nil != aarq.calledAPtitle {
		ch = new(t_der_chunk)
		ch.asn1_class = ASN1_CLASS_CONTEXT_SPECIFIC
		ch.encoding = BER_ENCODING_CONSTRUCTED
		ch.asn1_tag = 2

		ch1 = new(t_der_chunk)
		ch1.asn1_class = ASN1_CLASS_UNIVERSAL
		ch1.encoding = BER_ENCODING_PRIMITIVE
		ch1.asn1_tag = 4
		octetString := ([]uint8)(*aarq.calledAPtitle)
		ch1.content = make([]uint8, len(octetString))
		copy(ch1.content, octetString)
		buf = new(bytes.Buffer)
		err = der_encode_chunk(buf, ch1)
		if nil != err {
			return err
		}

		ch.content = buf.Bytes()

		err = der_encode_chunk(bufAARQ, ch)
		if nil != err {
			return err
		}
	}

	// called-AE-qualifier [3] AE-qualifier OPTIONAL,

	if nil != aarq.calledAEqualifier {
		ch = new(t_der_chunk)
		ch.asn1_class = ASN1_CLASS_CONTEXT_SPECIFIC
		ch.encoding = BER_ENCODING_CONSTRUCTED
		ch.asn1_tag = 3

		ch1 = new(t_der_chunk)
		ch1.asn1_class = ASN1_CLASS_UNIVERSAL
		ch1.encoding = BER_ENCODING_PRIMITIVE
		ch1.asn1_tag = 4
		octetString := ([]uint8)(*aarq.calledAEqualifier)
		ch1.content = make([]uint8, len(octetString))
		copy(ch1.content, octetString)
		buf = new(bytes.Buffer)
		err = der_encode_chunk(buf, ch1)
		if nil != err {
			return err
		}

		ch.content = buf.Bytes()

		err = der_encode_chunk(bufAARQ, ch)
		if nil != err {
			return err
		}
	}

	// called-AP-invocation-id [4] AP-invocation-identifier OPTIONAL,

	if nil != aarq.calledAPinvocationId {
		ch = new(t_der_chunk)
		ch.asn1_class = ASN1_CLASS_CONTEXT_SPECIFIC
		ch.encoding = BER_ENCODING_CONSTRUCTED
		ch.asn1_tag = 4

		ch1 = new(t_der_chunk)
		ch1.asn1_class = ASN1_CLASS_UNIVERSAL
		ch1.encoding = BER_ENCODING_PRIMITIVE
		ch1.asn1_tag = 2
		err, ch1.content = der_encode_Integer(*aarq.calledAPinvocationId)
		if nil != err {
			return err
		}
		buf = new(bytes.Buffer)
		err = der_encode_chunk(buf, ch1)
		if nil != err {
			return err
		}

		ch.content = buf.Bytes()

		err = der_encode_chunk(bufAARQ, ch)
		if nil != err {
			return err
		}
	}

	// called-AE-invocation-id [5] AE-invocation-identifier OPTIONAL,

	if nil != aarq.calledAEinvocationId {
		ch = new(t_der_chunk)
		ch.asn1_class = ASN1_CLASS_CONTEXT_SPECIFIC
		ch.encoding = BER_ENCODING_CONSTRUCTED
		ch.asn1_tag = 5

		ch1 = new(t_der_chunk)
		ch1.asn1_class = ASN1_CLASS_UNIVERSAL
		ch1.encoding = BER_ENCODING_PRIMITIVE
		ch1.asn1_tag = 2
		err, ch1.content = der_encode_Integer(*aarq.calledAEinvocationId)
		if nil != err {
			return err
		}
		buf = new(bytes.Buffer)
		err = der_encode_chunk(buf, ch1)
		if nil != err {
			return err
		}

		ch.content = buf.Bytes()

		err = der_encode_chunk(bufAARQ, ch)
		if nil != err {
			return err
		}
	}

	// calling-AP-title [6] AP-title OPTIONAL,

	if nil != aarq.callingAPtitle {
		ch = new(t_der_chunk)
		ch.asn1_class = ASN1_CLASS_CONTEXT_SPECIFIC
		ch.encoding = BER_ENCODING_CONSTRUCTED
		ch.asn1_tag = 6

		ch1 = new(t_der_chunk)
		ch1.asn1_class = ASN1_CLASS_UNIVERSAL
		ch1.encoding = BER_ENCODING_PRIMITIVE
		ch1.asn1_tag = 4
		octetString := ([]uint8)(*aarq.callingAPtitle)
		ch1.content = make([]uint8, len(octetString))
		copy(ch1.content, octetString)
		buf = new(bytes.Buffer)
		err = der_encode_chunk(buf, ch1)
		if nil != err {
			return err
		}

		ch.content = buf.Bytes()

		err = der_encode_chunk(bufAARQ, ch)
		if nil != err {
			return err
		}
	}

	// calling-AE-qualifier [7] AE-qualifier OPTIONAL,

	if nil != aarq.callingAEqualifier {
		ch = new(t_der_chunk)
		ch.asn1_class = ASN1_CLASS_CONTEXT_SPECIFIC
		ch.encoding = BER_ENCODING_CONSTRUCTED
		ch.asn1_tag = 7

		ch1 = new(t_der_chunk)
		ch1.asn1_class = ASN1_CLASS_UNIVERSAL
		ch1.encoding = BER_ENCODING_PRIMITIVE
		ch1.asn1_tag = 4
		octetString := ([]uint8)(*aarq.callingAEqualifier)
		ch1.content = make([]uint8, len(octetString))
		copy(ch1.content, octetString)
		buf = new(bytes.Buffer)
		err = der_encode_chunk(buf, ch1)
		if nil != err {
			return err
		}

		ch.content = buf.Bytes()

		err = der_encode_chunk(bufAARQ, ch)
		if nil != err {
			return err
		}
	}

	// calling-AP-invocation-id [8] AP-invocation-identifier OPTIONAL,

	if nil != aarq.callingAPinvocationId {
		ch = new(t_der_chunk)
		ch.asn1_class = ASN1_CLASS_CONTEXT_SPECIFIC
		ch.encoding = BER_ENCODING_CONSTRUCTED
		ch.asn1_tag = 8

		ch1 = new(t_der_chunk)
		ch1.asn1_class = ASN1_CLASS_UNIVERSAL
		ch1.encoding = BER_ENCODING_PRIMITIVE
		ch1.asn1_tag = 2
		err, ch1.content = der_encode_Integer(*aarq.callingAPinvocationId)
		if nil != err {
			return err
		}
		buf = new(bytes.Buffer)
		err = der_encode_chunk(buf, ch1)
		if nil != err {
			return err
		}

		ch.content = buf.Bytes()

		err = der_encode_chunk(bufAARQ, ch)
		if nil != err {
			return err
		}
	}

	// calling-AE-invocation-id [9] AE-invocation-identifier OPTIONAL,

	if nil != aarq.callingAEinvocationId {
		ch = new(t_der_chunk)
		ch.asn1_class = ASN1_CLASS_CONTEXT_SPECIFIC
		ch.encoding = BER_ENCODING_CONSTRUCTED
		ch.asn1_tag = 9

		ch1 = new(t_der_chunk)
		ch1.asn1_class = ASN1_CLASS_UNIVERSAL
		ch1.encoding = BER_ENCODING_PRIMITIVE
		ch1.asn1_tag = 2
		err, ch1.content = der_encode_Integer(*aarq.callingAEinvocationId)
		if nil != err {
			return err
		}
		buf = new(bytes.Buffer)
		err = der_encode_chunk(buf, ch1)
		if nil != err {
			return err
		}

		ch.content = buf.Bytes()

		err = der_encode_chunk(bufAARQ, ch)
		if nil != err {
			return err
		}
	}

	// sender-acse-requirements [10] IMPLICIT ACSE-requirements OPTIONAL,

	if nil != aarq.senderAcseRequirements {
		ch = new(t_der_chunk)
		ch.asn1_class = ASN1_CLASS_CONTEXT_SPECIFIC
		ch.encoding = BER_ENCODING_PRIMITIVE
		ch.asn1_tag = 10
		err, ch.content = der_encode_BitString(aarq.senderAcseRequirements)
		if nil != err {
			return err
		}

		err = der_encode_chunk(bufAARQ, ch)
		if nil != err {
			return err
		}
	}

	// mechanism-name [11] IMPLICIT Mechanism-name OPTIONAL,
	if nil != aarq.mechanismName {
		ch = new(t_der_chunk)
		ch.asn1_class = ASN1_CLASS_CONTEXT_SPECIFIC
		ch.encoding = BER_ENCODING_PRIMITIVE
		ch.asn1_tag = 11
		err, ch.content = der_encode_ObjectIdentifier(aarq.mechanismName)
		if nil != err {
			return err
		}

		err = der_encode_chunk(bufAARQ, ch)
		if nil != err {
			return err
		}
	}

	// calling-authentication-value [12] EXPLICIT Authentication-value OPTIONAL,

	/*
	   type tAsn1Choice struct {
	   	tag int
	   	val interface{}
	   }
	*/
	if nil != aarq.callingAuthenticationValue {
		ch = new(t_der_chunk)
		ch.asn1_class = ASN1_CLASS_CONTEXT_SPECIFIC
		ch.encoding = BER_ENCODING_CONSTRUCTED
		ch.asn1_tag = 12

		ch1 = new(t_der_chunk)

		if 0 == aarq.callingAuthenticationValue.tag {
			// charstring [0] IMPLICIT GraphicString,
			ch1.asn1_class = ASN1_CLASS_CONTEXT_SPECIFIC
			ch1.encoding = BER_ENCODING_PRIMITIVE
			ch1.asn1_tag = 0
			octetString := ([]uint8)(aarq.callingAuthenticationValue.val.(tAsn1GraphicString))
			ch1.content = make([]uint8, len(octetString))
			copy(ch1.content, octetString)
		} else if 1 == aarq.callingAuthenticationValue.tag {
			// bitstring [1] IMPLICIT BIT STRING,
			ch1.asn1_class = ASN1_CLASS_CONTEXT_SPECIFIC
			ch1.encoding = BER_ENCODING_PRIMITIVE
			ch1.asn1_tag = 1
			bitString := aarq.callingAuthenticationValue.val.(tAsn1BitString)
			err, ch1.content = der_encode_BitString(&bitString)
			if nil != err {
				return err
			}
		} else if 2 == aarq.callingAuthenticationValue.tag {
			// external [2] IMPLICIT OCTET STRING,
			ch1.asn1_class = ASN1_CLASS_CONTEXT_SPECIFIC
			ch1.encoding = BER_ENCODING_PRIMITIVE
			ch1.asn1_tag = 2
			octetString := aarq.callingAuthenticationValue.val.([]uint8)
			ch1.content = make([]uint8, len(octetString))
			copy(ch1.content, octetString)
		} else if 3 == aarq.callingAuthenticationValue.tag {

			/*
				other [3] IMPLICIT SEQUENCE
				{
					other-mechanism-name Mechanism-name,
					other-mechanism-value ANY DEFINED BY other-mechanism-name
				}
			*/

			ch1.asn1_class = ASN1_CLASS_CONTEXT_SPECIFIC
			ch1.encoding = BER_ENCODING_CONSTRUCTED
			ch1.asn1_tag = 3

			authenticationValueOther := aarq.callingAuthenticationValue.val.(tAsn1CosemAuthenticationValueOther)

			ch2 = new(t_der_chunk)
			ch2.asn1_class = ASN1_CLASS_UNIVERSAL
			ch2.encoding = BER_ENCODING_PRIMITIVE
			ch2.asn1_tag = 6
			err, ch2.content = der_encode_ObjectIdentifier(&authenticationValueOther.otherMechanismName)
			if nil != err {
				return err
			}

			ch3 = new(t_der_chunk)
			ch3.asn1_class = ASN1_CLASS_UNIVERSAL
			ch3.encoding = BER_ENCODING_PRIMITIVE
			ch3.asn1_tag = 4
			octetString := ([]uint8)(authenticationValueOther.otherMechanismValue)
			ch3.content = make([]uint8, len(octetString))
			copy(ch3.content, octetString)

			buf = new(bytes.Buffer)
			err = der_encode_chunk(buf, ch2)
			if nil != err {
				return err
			}
			err = der_encode_chunk(buf, ch3)
			if nil != err {
				return err
			}

			ch1.content = buf.Bytes()
		} else {
			err = fmt.Errorf("no tag")
			errorLog("%v", err)
			return err
		}

		buf = new(bytes.Buffer)
		err = der_encode_chunk(buf, ch1)
		if nil != err {
			return err
		}
		ch.content = buf.Bytes()

		err = der_encode_chunk(bufAARQ, ch)
		if nil != err {
			return err
		}
	}

	// implementation-information [29] IMPLICIT Implementation-data OPTIONAL,

	if nil != aarq.implementationInformation {
		ch := new(t_der_chunk)
		ch.asn1_class = ASN1_CLASS_CONTEXT_SPECIFIC
		ch.encoding = BER_ENCODING_PRIMITIVE
		ch.asn1_tag = 29

		octetString := ([]uint8)(*aarq.implementationInformation)
		ch.content = make([]uint8, len(octetString))
		copy(ch.content, octetString)

		err = der_encode_chunk(bufAARQ, ch)
		if nil != err {
			return err
		}
	}

	// user-information [30] EXPLICIT Association-information OPTIONAL

	if nil != aarq.userInformation {
		ch := new(t_der_chunk)
		ch.asn1_class = ASN1_CLASS_CONTEXT_SPECIFIC
		ch.encoding = BER_ENCODING_CONSTRUCTED
		ch.asn1_tag = 30

		ch1 := new(t_der_chunk)
		ch1.asn1_class = ASN1_CLASS_UNIVERSAL
		ch1.encoding = BER_ENCODING_PRIMITIVE
		ch1.asn1_tag = 4
		octetString := ([]uint8)(*aarq.userInformation)
		ch1.content = make([]uint8, len(octetString))
		copy(ch1.content, octetString)

		buf = new(bytes.Buffer)
		err = der_encode_chunk(buf, ch1)
		if nil != err {
			return err
		}

		ch.content = buf.Bytes()

		err = der_encode_chunk(bufAARQ, ch)
		if nil != err {
			return err
		}
	}

	chAARQ.content = bufAARQ.Bytes()

	err = der_encode_chunk(w, chAARQ)
	if nil != err {
		return err
	}

	return nil

}

func decode_AARQapdu_protocolVersion(ch *t_der_chunk, aarq *AARQapdu) (err error, found bool) {
	// protocol-version [0] IMPLICIT T-protocol-version DEFAULT {version1},
	if 0 == ch.asn1_tag {
		found = true
		err, aarq.protocolVersion = der_decode_BitString(ch.content)
		if nil != err {
			return err, found
		}
	}
	return err, found
}

func decode_AARQapdu_applicationContextName(ch *t_der_chunk, aarq *AARQapdu) (err error, found bool) {
	//application-context-name [1] Application-context-name,
	if 1 == ch.asn1_tag {
		found = true

		content := ch.content[0:]

		r := bytes.NewReader(content)
		err, ch = der_decode_chunk(r)
		if nil != err {
			return err, found
		}
		content = content[ch.length:]

		if 6 != ch.asn1_tag {
			err = fmt.Errorf("decoding error")
			errorLog("%v", err)
			return err, found
		}
		err, applicationContextName := der_decode_ObjectIdentifier(ch.content)
		if nil != err {
			return err, found
		}
		aarq.applicationContextName = *applicationContextName
	}
	return err, found
}

func decode_AARQapdu_calledAPtitle(ch *t_der_chunk, aarq *AARQapdu) (err error, found bool) {
	//called-AP-title [2] AP-title OPTIONAL,
	if 2 == ch.asn1_tag {
		found = true

		content := ch.content[0:]

		r := bytes.NewReader(content)
		err, ch = der_decode_chunk(r)
		if nil != err {
			return err, found
		}
		content = content[ch.length:]

		if 4 != ch.asn1_tag {
			err = fmt.Errorf("decoding error")
			errorLog("%v", err)
			return err, found
		}
		octetString := make([]uint8, len(ch.content))
		copy(octetString, ch.content)
		aarq.calledAPtitle = (*tAsn1OctetString)(&octetString)

	}
	return err, found
}

func decode_AARQapdu_calledAEqualifier(ch *t_der_chunk, aarq *AARQapdu) (err error, found bool) {
	//called-AE-qualifier [3] AE-qualifier OPTIONAL,
	if 3 == ch.asn1_tag {
		found = true

		content := ch.content[0:]

		r := bytes.NewReader(content)
		err, ch = der_decode_chunk(r)
		if nil != err {
			return err, found
		}
		content = content[ch.length:]

		if 4 != ch.asn1_tag {
			err = fmt.Errorf("decoding error")
			errorLog("%v", err)
			return err, found
		}
		octetString := make([]uint8, len(ch.content))
		copy(octetString, ch.content)
		aarq.calledAEqualifier = (*tAsn1OctetString)(&octetString)
	}
	return err, found
}

func decode_AARQapdu_calledAPinvocationId(ch *t_der_chunk, aarq *AARQapdu) (err error, found bool) {
	//called-AP-invocation-id [4] AP-invocation-identifier OPTIONAL,
	if 4 == ch.asn1_tag {
		found = true

		content := ch.content[0:]

		r := bytes.NewReader(content)
		err, ch = der_decode_chunk(r)
		if nil != err {
			return err, found
		}
		content = content[ch.length:]

		if 2 != ch.asn1_tag {
			err = fmt.Errorf("decoding error")
			errorLog("%v", err)
			return err, found
		}
		err, calledAPinvocationId := der_decode_Integer(ch.content)
		if nil != err {
			return err, found
		}
		aarq.calledAPinvocationId = &calledAPinvocationId
	}
	return err, found
}

func decode_AARQapdu_calledAEinvocationId(ch *t_der_chunk, aarq *AARQapdu) (err error, found bool) {
	//called-AE-invocation-id [5] AE-invocation-identifier OPTIONAL,
	if 5 == ch.asn1_tag {
		found = true

		content := ch.content[0:]

		r := bytes.NewReader(content)
		err, ch = der_decode_chunk(r)
		if nil != err {
			return err, found
		}
		content = content[ch.length:]

		if 2 != ch.asn1_tag {
			err = fmt.Errorf("decoding error")
			errorLog("%v", err)
			return err, found
		}
		err, calledAEinvocationId := der_decode_Integer(ch.content)
		if nil != err {
			return err, found
		}
		aarq.calledAEinvocationId = &calledAEinvocationId
	}
	return err, found
}

func decode_AARQapdu_callingAPtitle(ch *t_der_chunk, aarq *AARQapdu) (err error, found bool) {
	//calling-AP-title [6] AP-title OPTIONAL,
	if 6 == ch.asn1_tag {
		found = true

		content := ch.content[0:]

		r := bytes.NewReader(content)
		err, ch = der_decode_chunk(r)
		if nil != err {
			return err, found
		}
		content = content[ch.length:]

		if 4 != ch.asn1_tag {
			err = fmt.Errorf("decoding error")
			errorLog("%v", err)
			return err, found
		}
		octetString := make([]uint8, len(ch.content))
		copy(octetString, ch.content)
		aarq.callingAPtitle = (*tAsn1OctetString)(&octetString)
	}
	return err, found
}

func decode_AARQapdu_callingAEqualifier(ch *t_der_chunk, aarq *AARQapdu) (err error, found bool) {
	//calling-AE-qualifier [7] AE-qualifier OPTIONAL,
	if 7 == ch.asn1_tag {
		found = true

		content := ch.content[0:]

		r := bytes.NewReader(content)
		err, ch = der_decode_chunk(r)
		if nil != err {
			return err, found
		}
		content = content[ch.length:]

		if 4 != ch.asn1_tag {
			err = fmt.Errorf("decoding error")
			errorLog("%v", err)
			return err, found
		}
		octetString := make([]uint8, len(ch.content))
		copy(octetString, ch.content)
		aarq.callingAEqualifier = (*tAsn1OctetString)(&octetString)
	}
	return err, found
}

func decode_AARQapdu_callingAPinvocationId(ch *t_der_chunk, aarq *AARQapdu) (err error, found bool) {
	//calling-AP-invocation-id [8] AP-invocation-identifier OPTIONAL,
	if 8 == ch.asn1_tag {
		found = true

		content := ch.content[0:]

		r := bytes.NewReader(content)
		err, ch = der_decode_chunk(r)
		if nil != err {
			return err, found
		}
		content = content[ch.length:]

		if 2 != ch.asn1_tag {
			err = fmt.Errorf("decoding error")
			errorLog("%v", err)
			return err, found
		}
		err, callingAPinvocationId := der_decode_Integer(ch.content)
		if nil != err {
			return err, found
		}
		aarq.callingAPinvocationId = &callingAPinvocationId
	}
	return err, found
}

func decode_AARQapdu_callingAEinvocationId(ch *t_der_chunk, aarq *AARQapdu) (err error, found bool) {
	//calling-AE-invocation-id [9] AE-invocation-identifier OPTIONAL,
	if 9 == ch.asn1_tag {
		found = true

		content := ch.content[0:]

		r := bytes.NewReader(content)
		err, ch = der_decode_chunk(r)
		if nil != err {
			return err, found
		}
		content = content[ch.length:]

		if 2 != ch.asn1_tag {
			err = fmt.Errorf("decoding error")
			errorLog("%v", err)
			return err, found
		}
		err, callingAEinvocationId := der_decode_Integer(ch.content)
		if nil != err {
			return err, found
		}
		aarq.callingAEinvocationId = &callingAEinvocationId
	}

	return err, found
}

func decode_AARQapdu_senderAcseRequirements(ch *t_der_chunk, aarq *AARQapdu) (err error, found bool) {
	//sender-acse-requirements [10] IMPLICIT ACSE-requirements OPTIONAL,
	if 10 == ch.asn1_tag {
		found = true
		err, aarq.senderAcseRequirements = der_decode_BitString(ch.content)
		if nil != err {
			return err, found
		}
	}
	return err, found
}

func decode_AARQapdu_mechanismName(ch *t_der_chunk, aarq *AARQapdu) (err error, found bool) {
	//mechanism-name [11] IMPLICIT Mechanism-name OPTIONAL,
	//mechanismName *tAsn1ObjectIdentifier
	if 11 == ch.asn1_tag {
		found = true
		err, aarq.mechanismName = der_decode_ObjectIdentifier(ch.content)
		if nil != err {
			return err, found
		}
	}
	return err, found
}

func decode_AARQapdu_callingAuthenticationValue(ch *t_der_chunk, aarq *AARQapdu) (err error, found bool) {
	//calling-authentication-value [12] EXPLICIT Authentication-value OPTIONAL,
	if 12 == ch.asn1_tag {
		found = true

		/*
			Authentication-value ::= CHOICE
			{
				charstring [0] IMPLICIT GraphicString,
				bitstring [1] IMPLICIT BIT STRING,
				external [2] IMPLICIT OCTET STRING,
				other [3] IMPLICIT SEQUENCE
				{
					other-mechanism-name Mechanism-name,
					other-mechanism-value ANY DEFINED BY other-mechanism-name
				}
			}
		*/

		content := ch.content[0:]

		r := bytes.NewReader(content)
		err, ch = der_decode_chunk(r)
		if nil != err {
			return err, found
		}
		content = content[ch.length:]

		err, _found := decode_AARQapdu_callingAuthenticationValue_charstring(ch, aarq)
		if nil != err {
			return err, found
		}
		if !_found {
			err, _found = decode_AARQapdu_callingAuthenticationValue_bitstring(ch, aarq)
			if nil != err {
				return err, found
			}
		}
		if !_found {
			err, _found = decode_AARQapdu_callingAuthenticationValue_external(ch, aarq)
			if nil != err {
				return err, found
			}
		}
		if !_found {
			err, _found = decode_AARQapdu_callingAuthenticationValue_other(ch, aarq)
			if nil != err {
				return err, found
			}
		}
		if !_found {
			err = fmt.Errorf("decoding error")
			errorLog("%v", err)
			return err, found
		}

	}

	return err, found
}

func decode_AARQapdu_callingAuthenticationValue_charstring(ch *t_der_chunk, aarq *AARQapdu) (err error, found bool) {
	// charstring [0] IMPLICIT GraphicString,
	if 0 == ch.asn1_tag {
		found = true
		callingAuthenticationValue := new(tAsn1Choice)
		callingAuthenticationValue.tag = 0
		octetString := make([]uint8, len(ch.content))
		copy(octetString, ch.content)
		callingAuthenticationValue.val = tAsn1GraphicString(octetString)
		aarq.callingAuthenticationValue = callingAuthenticationValue
		return err, found
	}
	return err, found
}

func decode_AARQapdu_callingAuthenticationValue_bitstring(ch *t_der_chunk, aarq *AARQapdu) (err error, found bool) {
	//bitstring [1] IMPLICIT BIT STRING,
	if 1 == ch.asn1_tag {
		found = true
		callingAuthenticationValue := new(tAsn1Choice)
		callingAuthenticationValue.tag = 1
		err, bitString := der_decode_BitString(ch.content)
		if nil != err {
			return err, found
		}
		callingAuthenticationValue.val = bitString
		aarq.callingAuthenticationValue = callingAuthenticationValue
		return err, found
	}
	return err, found
}

func decode_AARQapdu_callingAuthenticationValue_external(ch *t_der_chunk, aarq *AARQapdu) (err error, found bool) {
	// external [2] IMPLICIT OCTET STRING,
	if 2 == ch.asn1_tag {
		found = true
		callingAuthenticationValue := new(tAsn1Choice)
		callingAuthenticationValue.tag = 2
		octetString := make([]uint8, len(ch.content))
		copy(octetString, ch.content)
		callingAuthenticationValue.val = octetString
		aarq.callingAuthenticationValue = callingAuthenticationValue
		return err, found
	}
	return err, found
}

func decode_AARQapdu_callingAuthenticationValue_other(ch *t_der_chunk, aarq *AARQapdu) (err error, found bool) {
	/*
		other [3] IMPLICIT SEQUENCE
		{
			other-mechanism-name Mechanism-name,
			other-mechanism-value ANY DEFINED BY other-mechanism-name
		}
	*/
	if 3 == ch.asn1_tag {
		found = true

		content := ch.content[0:]

		r := bytes.NewReader(content)
		err, ch = der_decode_chunk(r)
		if nil != err {
			return err, found
		}
		content = content[ch.length:]

		callingAuthenticationValue := new(tAsn1Choice)
		callingAuthenticationValue.tag = 3

		var authenticationValueOther tAsn1CosemAuthenticationValueOther

		if 6 != ch.asn1_tag {
			err = fmt.Errorf("decoding error")
			errorLog("%v", err)
			return err, found
		}
		err, objectIdentifier := der_decode_ObjectIdentifier(ch.content)
		if nil != err {
			err = fmt.Errorf("decoding error")
			errorLog("%v", err)
			return err, found
		}
		authenticationValueOther.otherMechanismName = *objectIdentifier

		r = bytes.NewReader(content)
		err, ch = der_decode_chunk(r)
		if nil != err {
			return err, found
		}
		content = content[ch.length:]

		if 4 != ch.asn1_tag {
			err = fmt.Errorf("decoding error")
			errorLog("%v", err)
			return err, found
		}
		authenticationValueOther.otherMechanismValue = make([]uint8, len(ch.content))
		copy(authenticationValueOther.otherMechanismValue, ch.content)

		callingAuthenticationValue.val = authenticationValueOther
		aarq.callingAuthenticationValue = callingAuthenticationValue
	}
	return err, found
}

func decode_AARQapdu_implementationInformation(ch *t_der_chunk, aarq *AARQapdu) (err error, found bool) {
	//implementation-information [29] IMPLICIT Implementation-data OPTIONAL,
	if 29 == ch.asn1_tag {
		found = true
		octetString := make([]uint8, len(ch.content))
		copy(octetString, ch.content)
		aarq.implementationInformation = (*tAsn1GraphicString)(&octetString)
	}

	return err, found
}

func decode_AARQapdu_userInformation(ch *t_der_chunk, aarq *AARQapdu) (err error, found bool) {
	//user-information [30] EXPLICIT Association-information OPTIONAL
	if 30 == ch.asn1_tag {
		found = true

		content := ch.content[0:]

		r := bytes.NewReader(content)
		err, ch = der_decode_chunk(r)
		if nil != err {
			return err, found
		}
		content = content[ch.length:]

		if 4 != ch.asn1_tag {
			err = fmt.Errorf("decoding error")
			errorLog("%v", err)
			return err, found
		}
		octetString := make([]uint8, len(ch.content))
		copy(octetString, ch.content)
		aarq.userInformation = (*tAsn1OctetString)(&octetString)
	}
	return err, found
}

func decode_AARQapdu(r io.Reader) (err error, aarq *AARQapdu) {
	var found bool
	aarq = new(AARQapdu)

	err, ch := der_decode_chunk(r)
	if nil != err {
		return err, aarq
	}
	content := ch.content[0:]

	// AARQ-apdu ::= [APPLICATION 0] IMPLICIT SEQUENCE
	if 0 == ch.asn1_tag {
		found = true
	}
	if found {
		r = bytes.NewReader(content)
		err, ch = der_decode_chunk(r)
		if nil != err {
			return err, aarq
		}
		content = content[ch.length:]
	} else {
		err = fmt.Errorf("decoding error")
		errorLog("%v", err)
		return err, aarq
	}

	// protocol-version [0] IMPLICIT T-protocol-version DEFAULT {version1},
	err, found = decode_AARQapdu_protocolVersion(ch, aarq)
	if nil != err {
		return err, aarq
	}
	if found {
		r = bytes.NewReader(content)
		err, ch = der_decode_chunk(r)
		if nil != err {
			return err, aarq
		}
		content = content[ch.length:]
	}

	//application-context-name [1] Application-context-name,
	err, found = decode_AARQapdu_applicationContextName(ch, aarq)
	if nil != err {
		return err, aarq
	}
	if found {
		r = bytes.NewReader(content)
		err, ch = der_decode_chunk(r)
		if nil != err {
			return err, aarq
		}
		content = content[ch.length:]
	} else {
		err = fmt.Errorf("decoding error")
		errorLog("%v", err)
		return err, aarq
	}

	//called-AP-title [2] AP-title OPTIONAL,
	err, found = decode_AARQapdu_calledAPtitle(ch, aarq)
	if nil != err {
		return err, aarq
	}
	if found {
		r = bytes.NewReader(content)
		err, ch = der_decode_chunk(r)
		if nil != err {
			return err, aarq
		}
		content = content[ch.length:]
	}

	//called-AE-qualifier [3] AE-qualifier OPTIONAL,
	err, found = decode_AARQapdu_calledAEqualifier(ch, aarq)
	if nil != err {
		return err, aarq
	}
	if found {
		r = bytes.NewReader(content)
		err, ch = der_decode_chunk(r)
		if nil != err {
			return err, aarq
		}
		content = content[ch.length:]
	}

	//called-AP-invocation-id [4] AP-invocation-identifier OPTIONAL,
	err, found = decode_AARQapdu_calledAPinvocationId(ch, aarq)
	if nil != err {
		return err, aarq
	}
	if found {
		r = bytes.NewReader(content)
		err, ch = der_decode_chunk(r)
		if nil != err {
			return err, aarq
		}
		content = content[ch.length:]
	}

	//called-AE-invocation-id [5] AE-invocation-identifier OPTIONAL,
	err, found = decode_AARQapdu_calledAEinvocationId(ch, aarq)
	if nil != err {
		return err, aarq
	}
	if found {
		r = bytes.NewReader(content)
		err, ch = der_decode_chunk(r)
		if nil != err {
			return err, aarq
		}
		content = content[ch.length:]
	}

	//calling-AP-title [6] AP-title OPTIONAL,
	err, found = decode_AARQapdu_callingAPtitle(ch, aarq)
	if nil != err {
		return err, aarq
	}
	if found {
		r = bytes.NewReader(content)
		err, ch = der_decode_chunk(r)
		if nil != err {
			return err, aarq
		}
		content = content[ch.length:]
	}

	//calling-AE-qualifier [7] AE-qualifier OPTIONAL,
	err, found = decode_AARQapdu_callingAEqualifier(ch, aarq)
	if nil != err {
		return err, aarq
	}
	if found {
		r = bytes.NewReader(content)
		err, ch = der_decode_chunk(r)
		if nil != err {
			return err, aarq
		}
		content = content[ch.length:]
	}

	//calling-AP-invocation-id [8] AP-invocation-identifier OPTIONAL,
	err, found = decode_AARQapdu_callingAPinvocationId(ch, aarq)
	if nil != err {
		return err, aarq
	}
	if found {
		r = bytes.NewReader(content)
		err, ch = der_decode_chunk(r)
		if nil != err {
			return err, aarq
		}
		content = content[ch.length:]
	}

	//calling-AE-invocation-id [9] AE-invocation-identifier OPTIONAL,
	err, found = decode_AARQapdu_callingAEinvocationId(ch, aarq)
	if nil != err {
		return err, aarq
	}
	if found {
		r = bytes.NewReader(content)
		err, ch = der_decode_chunk(r)
		if nil != err {
			return err, aarq
		}
		content = content[ch.length:]
	}

	//sender-acse-requirements [10] IMPLICIT ACSE-requirements OPTIONAL,
	err, found = decode_AARQapdu_senderAcseRequirements(ch, aarq)
	if nil != err {
		return err, aarq
	}
	if found {
		r = bytes.NewReader(content)
		err, ch = der_decode_chunk(r)
		if nil != err {
			return err, aarq
		}
		content = content[ch.length:]
	}

	//mechanism-name [11] IMPLICIT Mechanism-name OPTIONAL,
	err, found = decode_AARQapdu_mechanismName(ch, aarq)
	if nil != err {
		return err, aarq
	}
	if found {
		r = bytes.NewReader(content)
		err, ch = der_decode_chunk(r)
		if nil != err {
			return err, aarq
		}
		content = content[ch.length:]
	}

	//calling-authentication-value [12] EXPLICIT Authentication-value OPTIONAL,
	err, found = decode_AARQapdu_callingAuthenticationValue(ch, aarq)
	if nil != err {
		return err, aarq
	}
	if found {
		r = bytes.NewReader(content)
		err, ch = der_decode_chunk(r)
		if nil != err {
			return err, aarq
		}
		content = content[ch.length:]
	}

	//implementation-information [29] IMPLICIT Implementation-data OPTIONAL,
	err, found = decode_AARQapdu_implementationInformation(ch, aarq)
	if nil != err {
		return err, aarq
	}
	if found {
		r = bytes.NewReader(content)
		err, ch = der_decode_chunk(r)
		if nil != err {
			return err, aarq
		}
		content = content[ch.length:]
	}

	//user-information [30] EXPLICIT Association-information OPTIONAL
	err, found = decode_AARQapdu_userInformation(ch, aarq)
	if nil != err {
		return err, aarq
	}

	return err, aarq
}

func encode_AAREapdu(w io.Writer, aare *AAREapdu) (err error) {
	var ch, ch1, ch2, ch3, chAARE *t_der_chunk
	var buf, bufAARE *bytes.Buffer

	if nil == aare {
		return nil
	}

	// AARE-apdu ::= [APPLICATION 1] IMPLICIT SEQUENCE

	chAARE = new(t_der_chunk)
	chAARE.asn1_class = ASN1_CLASS_APPLICATION
	chAARE.encoding = BER_ENCODING_CONSTRUCTED
	chAARE.asn1_tag = 1
	bufAARE = new(bytes.Buffer)

	// protocol-version [0] IMPLICIT T-protocol-version DEFAULT {version1}

	if nil != aare.protocolVersion {
		ch = new(t_der_chunk)
		ch.asn1_class = ASN1_CLASS_CONTEXT_SPECIFIC
		ch.encoding = BER_ENCODING_PRIMITIVE
		ch.asn1_tag = 0
		err, ch.content = der_encode_BitString(aare.protocolVersion)
		if nil != err {
			return err
		}

		err = der_encode_chunk(bufAARE, ch)
		if nil != err {
			return err
		}
	}

	// application-context-name [1] Application-context-name,

	if nil != aare.applicationContextName {
		ch = new(t_der_chunk)
		ch.asn1_class = ASN1_CLASS_CONTEXT_SPECIFIC
		ch.encoding = BER_ENCODING_CONSTRUCTED
		ch.asn1_tag = 1

		ch1 = new(t_der_chunk)
		ch1.asn1_class = ASN1_CLASS_UNIVERSAL
		ch1.encoding = BER_ENCODING_PRIMITIVE
		ch1.asn1_tag = 6
		err, ch1.content = der_encode_ObjectIdentifier(&aare.applicationContextName)
		if nil != err {
			return err
		}
		buf = new(bytes.Buffer)
		err = der_encode_chunk(buf, ch1)
		if nil != err {
			return err
		}

		ch.content = buf.Bytes()

		err = der_encode_chunk(bufAARE, ch)
		if nil != err {
			return err
		}
	}

	// result [2] Association-result,

	ch = new(t_der_chunk)
	ch.asn1_class = ASN1_CLASS_CONTEXT_SPECIFIC
	ch.encoding = BER_ENCODING_CONSTRUCTED
	ch.asn1_tag = 2

	ch1 = new(t_der_chunk)
	ch1.asn1_class = ASN1_CLASS_UNIVERSAL
	ch1.encoding = BER_ENCODING_PRIMITIVE
	ch1.asn1_tag = 2
	err, ch1.content = der_encode_Integer(aare.result)
	if nil != err {
		return err
	}
	buf = new(bytes.Buffer)
	err = der_encode_chunk(buf, ch1)
	if nil != err {
		return err
	}

	ch.content = buf.Bytes()

	err = der_encode_chunk(bufAARE, ch)
	if nil != err {
		return err
	}

	// result-source-diagnostic [3] Associate-source-diagnostic

	ch = new(t_der_chunk)
	ch.asn1_class = ASN1_CLASS_CONTEXT_SPECIFIC
	ch.encoding = BER_ENCODING_CONSTRUCTED
	ch.asn1_tag = 3

	/*
		Associate-source-diagnostic ::= CHOICE
		{
			acse-service-user [1] INTEGER
			{
				null (0),
				no-reason-given (1),
				application-context-name-not-supported (2),
				authentication-mechanism-name-not-recognised (11),
				authentication-mechanism-name-required (12),
				authentication-failure (13),
				authentication-required (14)
			},
			acse-service-provider [2] INTEGER
			{
				null (0),
				no-reason-given (1),
				no-common-acse-version (2)
			}
		}
	*/

	if 1 == aare.resultSourceDiagnostic.tag {
		ch1 = new(t_der_chunk)
		ch1.asn1_class = ASN1_CLASS_CONTEXT_SPECIFIC
		ch1.encoding = BER_ENCODING_CONSTRUCTED
		ch1.asn1_tag = 1

		ch2 = new(t_der_chunk)
		ch2.asn1_class = ASN1_CLASS_UNIVERSAL
		ch2.encoding = BER_ENCODING_PRIMITIVE
		ch2.asn1_tag = 2
		err, ch2.content = der_encode_Integer(aare.resultSourceDiagnostic.val.(tAsn1Integer))
		if nil != err {
			return err
		}
		buf = new(bytes.Buffer)
		err = der_encode_chunk(buf, ch2)
		if nil != err {
			return err
		}

		ch1.content = buf.Bytes()
	} else if 2 == aare.resultSourceDiagnostic.tag {
		ch1 = new(t_der_chunk)
		ch1.asn1_class = ASN1_CLASS_CONTEXT_SPECIFIC
		ch1.encoding = BER_ENCODING_CONSTRUCTED
		ch1.asn1_tag = 2

		ch2 = new(t_der_chunk)
		ch2.asn1_class = ASN1_CLASS_UNIVERSAL
		ch2.encoding = BER_ENCODING_PRIMITIVE
		ch2.asn1_tag = 2
		err, ch2.content = der_encode_Integer(aare.resultSourceDiagnostic.val.(tAsn1Integer))
		if nil != err {
			return err
		}
		buf = new(bytes.Buffer)
		err = der_encode_chunk(buf, ch2)
		if nil != err {
			return err
		}

		ch1.content = buf.Bytes()
	} else {
		err = fmt.Errorf("unknown tag")
		errorLog("unknown tag")
		return err
	}

	buf = new(bytes.Buffer)
	err = der_encode_chunk(buf, ch1)
	if nil != err {
		return err
	}
	ch.content = buf.Bytes()

	err = der_encode_chunk(bufAARE, ch)
	if nil != err {
		return err
	}

	// responding-AP-title [4] AP-title OPTIONAL,

	if nil != aare.respondingAPtitle {
		ch = new(t_der_chunk)
		ch.asn1_class = ASN1_CLASS_CONTEXT_SPECIFIC
		ch.encoding = BER_ENCODING_CONSTRUCTED
		ch.asn1_tag = 4

		ch1 = new(t_der_chunk)
		ch1.asn1_class = ASN1_CLASS_UNIVERSAL
		ch1.encoding = BER_ENCODING_PRIMITIVE
		ch1.asn1_tag = 4
		octetString := ([]uint8)(*aare.respondingAPtitle)
		ch1.content = make([]uint8, len(octetString))
		copy(ch1.content, octetString)
		buf = new(bytes.Buffer)
		err = der_encode_chunk(buf, ch1)
		if nil != err {
			return err
		}

		ch.content = buf.Bytes()

		err = der_encode_chunk(bufAARE, ch)
		if nil != err {
			return err
		}
	}

	// responding-AE-qualifier [5] AE-qualifier OPTIONAL,

	if nil != aare.respondingAEqualifier {
		ch = new(t_der_chunk)
		ch.asn1_class = ASN1_CLASS_CONTEXT_SPECIFIC
		ch.encoding = BER_ENCODING_CONSTRUCTED
		ch.asn1_tag = 5

		ch1 = new(t_der_chunk)
		ch1.asn1_class = ASN1_CLASS_UNIVERSAL
		ch1.encoding = BER_ENCODING_PRIMITIVE
		ch1.asn1_tag = 4
		octetString := ([]uint8)(*aare.respondingAPtitle)
		ch1.content = make([]uint8, len(octetString))
		copy(ch1.content, octetString)
		buf = new(bytes.Buffer)
		err = der_encode_chunk(buf, ch1)
		if nil != err {
			return err
		}

		ch.content = buf.Bytes()

		err = der_encode_chunk(bufAARE, ch)
		if nil != err {
			return err
		}
	}

	//responding-AP-invocation-id [6] AP-invocation-identifier OPTIONAL,

	if nil != aare.respondingAPinvocationId {
		ch = new(t_der_chunk)
		ch.asn1_class = ASN1_CLASS_CONTEXT_SPECIFIC
		ch.encoding = BER_ENCODING_CONSTRUCTED
		ch.asn1_tag = 6

		ch1 = new(t_der_chunk)
		ch1.asn1_class = ASN1_CLASS_UNIVERSAL
		ch1.encoding = BER_ENCODING_PRIMITIVE
		ch1.asn1_tag = 2
		err, ch1.content = der_encode_Integer(*aare.respondingAPinvocationId)
		if nil != err {
			return err
		}
		buf = new(bytes.Buffer)
		err = der_encode_chunk(buf, ch1)
		if nil != err {
			return err
		}

		ch.content = buf.Bytes()

		err = der_encode_chunk(bufAARE, ch)
		if nil != err {
			return err
		}
	}

	//responding-AE-invocation-id [7] AE-invocation-identifier OPTIONAL,

	if nil != aare.respondingAEinvocationId {
		ch = new(t_der_chunk)
		ch.asn1_class = ASN1_CLASS_CONTEXT_SPECIFIC
		ch.encoding = BER_ENCODING_CONSTRUCTED
		ch.asn1_tag = 7

		ch1 = new(t_der_chunk)
		ch1.asn1_class = ASN1_CLASS_UNIVERSAL
		ch1.encoding = BER_ENCODING_PRIMITIVE
		ch1.asn1_tag = 2
		err, ch1.content = der_encode_Integer(*aare.respondingAEinvocationId)
		if nil != err {
			return err
		}
		buf = new(bytes.Buffer)
		err = der_encode_chunk(buf, ch1)
		if nil != err {
			return err
		}

		ch.content = buf.Bytes()

		err = der_encode_chunk(bufAARE, ch)
		if nil != err {
			return err
		}
	}

	//responder-acse-requirements [8] IMPLICIT ACSE-requirements OPTIONAL,

	if nil != aare.responderAcseRequirements {
		ch = new(t_der_chunk)
		ch.asn1_class = ASN1_CLASS_CONTEXT_SPECIFIC
		ch.encoding = BER_ENCODING_PRIMITIVE
		ch.asn1_tag = 8
		err, ch.content = der_encode_BitString(aare.responderAcseRequirements)
		if nil != err {
			return err
		}

		err = der_encode_chunk(bufAARE, ch)
		if nil != err {
			return err
		}
	}

	//mechanism-name [9] IMPLICIT Mechanism-name OPTIONAL,

	if nil != aare.mechanismName {
		ch = new(t_der_chunk)
		ch.asn1_class = ASN1_CLASS_CONTEXT_SPECIFIC
		ch.encoding = BER_ENCODING_PRIMITIVE
		ch.asn1_tag = 9
		err, ch.content = der_encode_ObjectIdentifier(aare.mechanismName)
		if nil != err {
			return err
		}

		err = der_encode_chunk(bufAARE, ch)
		if nil != err {
			return err
		}
	}

	//responding-authentication-value [10] EXPLICIT Authentication-value OPTIONAL,

	if nil != aare.respondingAuthenticationValue {
		ch = new(t_der_chunk)
		ch.asn1_class = ASN1_CLASS_CONTEXT_SPECIFIC
		ch.encoding = BER_ENCODING_CONSTRUCTED
		ch.asn1_tag = 10

		/*
			Authentication-value ::= CHOICE
			{
				charstring [0] IMPLICIT GraphicString,
				bitstring [1] IMPLICIT BIT STRING,
				external [2] IMPLICIT OCTET STRING,
				other [3] IMPLICIT SEQUENCE
				{
					other-mechanism-name Mechanism-name,
					other-mechanism-value ANY DEFINED BY other-mechanism-name
				}
			}
		*/

		ch1 = new(t_der_chunk)

		if 0 == aare.respondingAuthenticationValue.tag {
			// charstring [0] IMPLICIT GraphicString,
			ch1.asn1_class = ASN1_CLASS_CONTEXT_SPECIFIC
			ch1.encoding = BER_ENCODING_PRIMITIVE
			ch1.asn1_tag = 0
			octetString := ([]uint8)(aare.respondingAuthenticationValue.val.(tAsn1GraphicString))
			ch1.content = make([]uint8, len(octetString))
			copy(ch1.content, octetString)
		} else if 1 == aare.respondingAuthenticationValue.tag {
			// bitstring [1] IMPLICIT BIT STRING,
			ch1.asn1_class = ASN1_CLASS_CONTEXT_SPECIFIC
			ch1.encoding = BER_ENCODING_PRIMITIVE
			ch1.asn1_tag = 1
			bitString := aare.respondingAuthenticationValue.val.(tAsn1BitString)
			err, ch1.content = der_encode_BitString(&bitString)
			if nil != err {
				return err
			}
		} else if 2 == aare.respondingAuthenticationValue.tag {
			// external [2] IMPLICIT OCTET STRING,
			ch1.asn1_class = ASN1_CLASS_CONTEXT_SPECIFIC
			ch1.encoding = BER_ENCODING_PRIMITIVE
			ch1.asn1_tag = 2
			octetString := aare.respondingAuthenticationValue.val.([]uint8)
			ch1.content = make([]uint8, len(octetString))
			copy(ch1.content, octetString)
		} else if 3 == aare.respondingAuthenticationValue.tag {

			/*
				other [3] IMPLICIT SEQUENCE
				{
					other-mechanism-name Mechanism-name,
					other-mechanism-value ANY DEFINED BY other-mechanism-name
				}
			*/

			ch1.asn1_class = ASN1_CLASS_CONTEXT_SPECIFIC
			ch1.encoding = BER_ENCODING_CONSTRUCTED
			ch1.asn1_tag = 3

			authenticationValueOther := aare.respondingAuthenticationValue.val.(tAsn1CosemAuthenticationValueOther)

			ch2 = new(t_der_chunk)
			ch2.asn1_class = ASN1_CLASS_UNIVERSAL
			ch2.encoding = BER_ENCODING_PRIMITIVE
			ch2.asn1_tag = 6
			err, ch2.content = der_encode_ObjectIdentifier(&authenticationValueOther.otherMechanismName)
			if nil != err {
				return err
			}

			ch3 = new(t_der_chunk)
			ch3.asn1_class = ASN1_CLASS_UNIVERSAL
			ch3.encoding = BER_ENCODING_PRIMITIVE
			ch3.asn1_tag = 4
			octetString := ([]uint8)(authenticationValueOther.otherMechanismValue)
			ch3.content = make([]uint8, len(octetString))
			copy(ch3.content, octetString)

			buf = new(bytes.Buffer)
			err = der_encode_chunk(buf, ch2)
			if nil != err {
				return err
			}
			err = der_encode_chunk(buf, ch3)
			if nil != err {
				return err
			}

			ch1.content = buf.Bytes()
		} else {
			err = fmt.Errorf("no tag")
			errorLog("%v", err)
			return err
		}

		buf = new(bytes.Buffer)
		err = der_encode_chunk(buf, ch1)
		if nil != err {
			return err
		}
		ch.content = buf.Bytes()

		err = der_encode_chunk(bufAARE, ch)
		if nil != err {
			return err
		}
	}

	//implementation-information [29] IMPLICIT Implementation-data OPTIONAL,

	if nil != aare.implementationInformation {
		ch := new(t_der_chunk)
		ch.asn1_class = ASN1_CLASS_CONTEXT_SPECIFIC
		ch.encoding = BER_ENCODING_PRIMITIVE
		ch.asn1_tag = 29

		octetString := ([]uint8)(*aare.implementationInformation)
		ch.content = make([]uint8, len(octetString))
		copy(ch.content, octetString)

		err = der_encode_chunk(bufAARE, ch)
		if nil != err {
			return err
		}
	}

	//user-information [30] EXPLICIT Association-information OPTIONAL

	if nil != aare.userInformation {
		ch := new(t_der_chunk)
		ch.asn1_class = ASN1_CLASS_CONTEXT_SPECIFIC
		ch.encoding = BER_ENCODING_CONSTRUCTED
		ch.asn1_tag = 30

		ch1 := new(t_der_chunk)
		ch1.asn1_class = ASN1_CLASS_UNIVERSAL
		ch1.encoding = BER_ENCODING_PRIMITIVE
		ch1.asn1_tag = 4
		octetString := ([]uint8)(*aare.userInformation)
		ch1.content = make([]uint8, len(octetString))
		copy(ch1.content, octetString)
		var buf bytes.Buffer
		err = der_encode_chunk(&buf, ch1)
		if nil != err {
			return err
		}

		ch.content = buf.Bytes()

		err = der_encode_chunk(bufAARE, ch)
		if nil != err {
			return err
		}
	}

	chAARE.content = bufAARE.Bytes()
	err = der_encode_chunk(w, chAARE)
	if nil != err {
		return err
	}

	return err
}

func decode_AAREapdu_protocolVersion(ch *t_der_chunk, aare *AAREapdu) (err error, found bool) {
	//protocol-version [0] IMPLICIT T-protocol-version DEFAULT {version1},
	if 0 == ch.asn1_tag {
		found = true
		err, aare.protocolVersion = der_decode_BitString(ch.content)
		if nil != err {
			return err, found
		}
	}
	return err, found
}

func decode_AAREapdu_applicationContextName(ch *t_der_chunk, aare *AAREapdu) (err error, found bool) {
	//application-context-name [1] Application-context-name,
	if 1 == ch.asn1_tag {
		found = true

		content := ch.content[0:]

		r := bytes.NewReader(content)
		err, ch = der_decode_chunk(r)
		if nil != err {
			return err, found
		}
		content = content[ch.length:]

		if 6 != ch.asn1_tag {
			err = fmt.Errorf("decoding error")
			errorLog("%v", err)
			return err, found
		}
		err, applicationContextName := der_decode_ObjectIdentifier(ch.content)
		if nil != err {
			return err, found
		}
		aare.applicationContextName = *applicationContextName
	}
	return err, found
}

func decode_AAREapdu_result(ch *t_der_chunk, aare *AAREapdu) (err error, found bool) {
	//result [2] Association-result,

	if 2 == ch.asn1_tag {
		found = true

		content := ch.content[0:]

		r := bytes.NewReader(content)
		err, ch = der_decode_chunk(r)
		if nil != err {
			return err, found
		}
		content = content[ch.length:]

		if 2 != ch.asn1_tag {
			err = fmt.Errorf("decoding error")
			errorLog("%v", err)
			return err, found
		}
		err, result := der_decode_Integer(ch.content)
		if nil != err {
			return err, found
		}
		aare.result = result
	}
	return err, found
}

func decode_AAREapdu_resultSourceDiagnostic(ch *t_der_chunk, aare *AAREapdu) (err error, found bool) {
	//result-source-diagnostic [3] Associate-source-diagnostic,
	if 3 == ch.asn1_tag {
		found = true

		content := ch.content[0:]

		r := bytes.NewReader(content)
		err, ch = der_decode_chunk(r)
		if nil != err {
			return err, found
		}
		content = content[ch.length:]

		/*
			Associate-source-diagnostic ::= CHOICE
			{
				acse-service-user [1] INTEGER
				{
					null (0),
					no-reason-given (1),
					application-context-name-not-supported (2),
					authentication-mechanism-name-not-recognised (11),
					authentication-mechanism-name-required (12),
					authentication-failure (13),
					authentication-required (14)
				},
				acse-service-provider [2] INTEGER
				{
					null (0),
					no-reason-given (1),
					no-common-acse-version (2)
				}
			}
		*/

		if 1 == ch.asn1_tag {
			// acse-service-user [1] INTEGER
			content := ch.content[0:]

			r := bytes.NewReader(content)
			err, ch = der_decode_chunk(r)
			if nil != err {
				return err, found
			}
			content = content[ch.length:]

			if 2 != ch.asn1_tag {
				err = fmt.Errorf("decoding error")
				errorLog("%v", err)
				return err, found
			}
			err, val := der_decode_Integer(ch.content)
			if nil != err {
				return err, found
			}
			aare.resultSourceDiagnostic.tag = 1
			aare.resultSourceDiagnostic.val = val
		} else if 2 == ch.asn1_tag {
			// acse-service-provider [2] INTEGER
			content := ch.content[0:]

			r := bytes.NewReader(content)
			err, ch = der_decode_chunk(r)
			if nil != err {
				return err, found
			}
			content = content[ch.length:]

			if 2 != ch.asn1_tag {
				err = fmt.Errorf("decoding error")
				errorLog("%v", err)
				return err, found
			}
			err, val := der_decode_Integer(ch.content)
			if nil != err {
				return err, found
			}
			aare.resultSourceDiagnostic.tag = 2
			aare.resultSourceDiagnostic.val = val
		} else {
			err = fmt.Errorf("unknown tag")
			errorLog("%v", err)
			return err, found
		}

	}
	return err, found
}

func decode_AAREapdu_respondingAPtitle(ch *t_der_chunk, aare *AAREapdu) (err error, found bool) {
	//responding-AP-title [4] AP-title OPTIONAL,
	if 4 == ch.asn1_tag {
		found = true

		content := ch.content[0:]

		r := bytes.NewReader(content)
		err, ch = der_decode_chunk(r)
		if nil != err {
			return err, found
		}
		content = content[ch.length:]

		if 4 != ch.asn1_tag {
			err = fmt.Errorf("decoding error")
			errorLog("%v", err)
			return err, found
		}
		octetString := make([]uint8, len(ch.content))
		copy(octetString, ch.content)
		aare.respondingAPtitle = (*tAsn1OctetString)(&octetString)
	}
	return err, found
}

func decode_AAREapdu_respondingAEqualifier(ch *t_der_chunk, aare *AAREapdu) (err error, found bool) {
	//responding-AE-qualifier [5] AE-qualifier OPTIONAL,
	if 5 == ch.asn1_tag {
		found = true

		content := ch.content[0:]

		r := bytes.NewReader(content)
		err, ch = der_decode_chunk(r)
		if nil != err {
			return err, found
		}
		content = content[ch.length:]

		if 4 != ch.asn1_tag {
			err = fmt.Errorf("decoding error")
			errorLog("%v", err)
			return err, found
		}
		octetString := make([]uint8, len(ch.content))
		copy(octetString, ch.content)
		aare.respondingAEqualifier = (*tAsn1OctetString)(&octetString)
	}
	return err, found
}

func decode_AAREapdu_respondingAPinvocationId(ch *t_der_chunk, aare *AAREapdu) (err error, found bool) {
	//responding-AP-invocation-id [6] AP-invocation-identifier OPTIONAL,
	if 6 == ch.asn1_tag {
		found = true

		content := ch.content[0:]

		r := bytes.NewReader(content)
		err, ch = der_decode_chunk(r)
		if nil != err {
			return err, found
		}
		content = content[ch.length:]

		if 2 != ch.asn1_tag {
			err = fmt.Errorf("decoding error")
			errorLog("%v", err)
			return err, found
		}
		var i tAsn1Integer
		err, i = der_decode_Integer(ch.content)
		if nil != err {
			return err, found
		}
		aare.respondingAPinvocationId = &i
	}
	return err, found
}

func decode_AAREapdu_respondingAEinvocationId(ch *t_der_chunk, aare *AAREapdu) (err error, found bool) {
	//responding-AE-invocation-id [7] AE-invocation-identifier OPTIONAL,
	if 7 == ch.asn1_tag {
		found = true

		content := ch.content[0:]

		r := bytes.NewReader(content)
		err, ch = der_decode_chunk(r)
		if nil != err {
			return err, found
		}
		content = content[ch.length:]

		if 2 != ch.asn1_tag {
			err = fmt.Errorf("decoding error")
			errorLog("%v", err)
			return err, found
		}
		var i tAsn1Integer
		err, i = der_decode_Integer(ch.content)
		if nil != err {
			return err, found
		}
		aare.respondingAEinvocationId = &i
	}
	return err, found
}

func decode_AAREapdu_responderAcseRequirements(ch *t_der_chunk, aare *AAREapdu) (err error, found bool) {
	//responder-acse-requirements [8] IMPLICIT ACSE-requirements OPTIONAL,
	if 8 == ch.asn1_tag {
		found = true

		err, aare.responderAcseRequirements = der_decode_BitString(ch.content)
		if nil != err {
			return err, found
		}
	}
	return err, found
}

func decode_AAREapdu_mechanismName(ch *t_der_chunk, aare *AAREapdu) (err error, found bool) {
	//mechanism-name [9] IMPLICIT Mechanism-name OPTIONAL,
	if 9 == ch.asn1_tag {
		found = true

		err, aare.mechanismName = der_decode_ObjectIdentifier(ch.content)
		if nil != err {
			return err, found
		}
	}
	return err, found
}

func decode_AAREapdu_respondingAuthenticationValue(ch *t_der_chunk, aare *AAREapdu) (err error, found bool) {
	//responding-authentication-value [10] EXPLICIT Authentication-value OPTIONAL,
	if 10 == ch.asn1_tag {
		found = true

		/*
			Authentication-value ::= CHOICE
			{
				charstring [0] IMPLICIT GraphicString,
				bitstring [1] IMPLICIT BIT STRING,
				external [2] IMPLICIT OCTET STRING,
				other [3] IMPLICIT SEQUENCE
				{
					other-mechanism-name Mechanism-name,
					other-mechanism-value ANY DEFINED BY other-mechanism-name
				}
			}
		*/

		content := ch.content[0:]

		r := bytes.NewReader(content)
		err, ch = der_decode_chunk(r)
		if nil != err {
			return err, found
		}
		content = content[ch.length:]

		err, _found := decode_AAREapdu_respondingAuthenticationValue_charstring(ch, aare)
		if nil != err {
			return err, found
		}
		if !_found {
			err, _found = decode_AAREapdu_respondingAuthenticationValue_bitstring(ch, aare)
			if nil != err {
				return err, found
			}
		}
		if !_found {
			err, _found = decode_AAREapdu_respondingAuthenticationValue_external(ch, aare)
			if nil != err {
				return err, found
			}
		}
		if !_found {
			err, _found = decode_AAREapdu_respondingAuthenticationValue_other(ch, aare)
			if nil != err {
				return err, found
			}
		}
		if !_found {
			err = fmt.Errorf("decoding error")
			errorLog("%v", err)
			return err, found
		}

	}

	return err, found
}

func decode_AAREapdu_respondingAuthenticationValue_charstring(ch *t_der_chunk, aare *AAREapdu) (err error, found bool) {
	// charstring [0] IMPLICIT GraphicString,
	if 0 == ch.asn1_tag {
		found = true
		respondingAuthenticationValue := new(tAsn1Choice)
		respondingAuthenticationValue.tag = 0
		octetString := make([]uint8, len(ch.content))
		copy(octetString, ch.content)
		respondingAuthenticationValue.val = tAsn1GraphicString(octetString)
		aare.respondingAuthenticationValue = respondingAuthenticationValue
		return err, found
	}
	return err, found
}

func decode_AAREapdu_respondingAuthenticationValue_bitstring(ch *t_der_chunk, aare *AAREapdu) (err error, found bool) {
	//bitstring [1] IMPLICIT BIT STRING,
	if 1 == ch.asn1_tag {
		found = true
		respondingAuthenticationValue := new(tAsn1Choice)
		respondingAuthenticationValue.tag = 1
		err, bitString := der_decode_BitString(ch.content)
		if nil != err {
			return err, found
		}
		respondingAuthenticationValue.val = bitString
		aare.respondingAuthenticationValue = respondingAuthenticationValue
		return err, found
	}
	return err, found
}

func decode_AAREapdu_respondingAuthenticationValue_external(ch *t_der_chunk, aare *AAREapdu) (err error, found bool) {
	// external [2] IMPLICIT OCTET STRING,
	if 2 == ch.asn1_tag {
		found = true
		respondingAuthenticationValue := new(tAsn1Choice)
		respondingAuthenticationValue.tag = 2
		octetString := make([]uint8, len(ch.content))
		copy(octetString, ch.content)
		respondingAuthenticationValue.val = octetString
		aare.respondingAuthenticationValue = respondingAuthenticationValue
		return err, found
	}
	return err, found
}

func decode_AAREapdu_respondingAuthenticationValue_other(ch *t_der_chunk, aare *AAREapdu) (err error, found bool) {
	/*
		other [3] IMPLICIT SEQUENCE
		{
			other-mechanism-name Mechanism-name,
			other-mechanism-value ANY DEFINED BY other-mechanism-name
		}
	*/
	if 3 == ch.asn1_tag {
		found = true

		content := ch.content[0:]

		r := bytes.NewReader(content)
		err, ch = der_decode_chunk(r)
		if nil != err {
			return err, found
		}
		content = content[ch.length:]

		respondingAuthenticationValue := new(tAsn1Choice)
		respondingAuthenticationValue.tag = 3

		var authenticationValueOther tAsn1CosemAuthenticationValueOther

		if 6 != ch.asn1_tag {
			err = fmt.Errorf("decoding error")
			errorLog("%v", err)
			return err, found
		}
		err, objectIdentifier := der_decode_ObjectIdentifier(ch.content)
		if nil != err {
			err = fmt.Errorf("decoding error")
			errorLog("%v", err)
			return err, found
		}
		authenticationValueOther.otherMechanismName = *objectIdentifier

		r = bytes.NewReader(content)
		err, ch = der_decode_chunk(r)
		if nil != err {
			return err, found
		}
		content = content[ch.length:]

		if 4 != ch.asn1_tag {
			err = fmt.Errorf("decoding error")
			errorLog("%v", err)
			return err, found
		}
		authenticationValueOther.otherMechanismValue = make([]uint8, len(ch.content))
		copy(authenticationValueOther.otherMechanismValue, ch.content)

		respondingAuthenticationValue.val = authenticationValueOther
		aare.respondingAuthenticationValue = respondingAuthenticationValue
	}
	return err, found
}

func decode_AAREapdu_implementationInformation(ch *t_der_chunk, aare *AAREapdu) (err error, found bool) {
	//implementation-information [29] IMPLICIT Implementation-data OPTIONAL,
	if 29 == ch.asn1_tag {
		found = true
		octetString := make([]uint8, len(ch.content))
		copy(octetString, ch.content)
		aare.implementationInformation = (*tAsn1GraphicString)(&octetString)
		return err, found
	}
	return err, found
}

func decode_AAREapdu_userInformation(ch *t_der_chunk, aare *AAREapdu) (err error, found bool) {
	//user-information [30] EXPLICIT Association-information OPTIONAL
	if 30 == ch.asn1_tag {
		found = true

		content := ch.content[0:]

		r := bytes.NewReader(content)
		err, ch = der_decode_chunk(r)
		if nil != err {
			return err, found
		}
		content = content[ch.length:]

		if 4 != ch.asn1_tag {
			err = fmt.Errorf("decoding error")
			errorLog("%v", err)
			return err, found
		}
		octetString := make([]uint8, len(ch.content))
		copy(octetString, ch.content)
		aare.userInformation = (*tAsn1OctetString)(&octetString)
	}
	return err, found
}

func decode_AAREapdu(r io.Reader) (err error, aare *AAREapdu) {
	var found bool
	aare = new(AAREapdu)

	err, ch := der_decode_chunk(r)
	if nil != err {
		return err, aare
	}
	content := ch.content[0:]

	//AARE-apdu ::= [APPLICATION 1] IMPLICIT SEQUENCE
	if 1 == ch.asn1_tag {
		found = true
	}
	if found {
		r = bytes.NewReader(content)
		err, ch = der_decode_chunk(r)
		if nil != err {
			return err, aare
		}
		content = content[ch.length:]
	} else {
		err = fmt.Errorf("decoding error")
		errorLog("%v", err)
		return err, aare
	}

	//protocol-version [0] IMPLICIT T-protocol-version DEFAULT {version1},
	err, found = decode_AAREapdu_protocolVersion(ch, aare)
	if nil != err {
		return err, aare
	}
	if found {
		r = bytes.NewReader(content)
		err, ch = der_decode_chunk(r)
		if nil != err {
			return err, aare
		}
		content = content[ch.length:]
	}

	//application-context-name [1] Application-context-name,
	err, found = decode_AAREapdu_applicationContextName(ch, aare)
	if nil != err {
		return err, aare
	}
	if found {
		r = bytes.NewReader(content)
		err, ch = der_decode_chunk(r)
		if nil != err {
			return err, aare
		}
		content = content[ch.length:]
	} else {
		err = fmt.Errorf("decoding error")
		errorLog("%v", err)
		return err, aare
	}

	//result [2] Association-result,
	err, found = decode_AAREapdu_result(ch, aare)
	if nil != err {
		return err, aare
	}
	if found {
		r = bytes.NewReader(content)
		err, ch = der_decode_chunk(r)
		if nil != err {
			return err, aare
		}
		content = content[ch.length:]
	} else {
		err = fmt.Errorf("decoding error")
		errorLog("%v", err)
		return err, aare
	}

	//result-source-diagnostic [3] Associate-source-diagnostic,
	err, found = decode_AAREapdu_resultSourceDiagnostic(ch, aare)
	if nil != err {
		return err, aare
	}
	if found {
		r = bytes.NewReader(content)
		err, ch = der_decode_chunk(r)
		if nil != err {
			return err, aare
		}
		content = content[ch.length:]
	} else {
		err = fmt.Errorf("decoding error")
		errorLog("%v", err)
		return err, aare
	}

	//responding-AP-title [4] AP-title OPTIONAL,
	err, found = decode_AAREapdu_respondingAPtitle(ch, aare)
	if nil != err {
		return err, aare
	}
	if found {
		r = bytes.NewReader(content)
		err, ch = der_decode_chunk(r)
		if nil != err {
			return err, aare
		}
		content = content[ch.length:]
	}

	//responding-AE-qualifier [5] AE-qualifier OPTIONAL,
	err, found = decode_AAREapdu_respondingAEqualifier(ch, aare)
	if nil != err {
		return err, aare
	}
	if found {
		r = bytes.NewReader(content)
		err, ch = der_decode_chunk(r)
		if nil != err {
			return err, aare
		}
		content = content[ch.length:]
	}

	//responding-AP-invocation-id [6] AP-invocation-identifier OPTIONAL,
	err, found = decode_AAREapdu_respondingAPinvocationId(ch, aare)
	if nil != err {
		return err, aare
	}
	if found {
		r = bytes.NewReader(content)
		err, ch = der_decode_chunk(r)
		if nil != err {
			return err, aare
		}
		content = content[ch.length:]
	}

	//responding-AE-invocation-id [7] AE-invocation-identifier OPTIONAL,
	err, found = decode_AAREapdu_respondingAEinvocationId(ch, aare)
	if nil != err {
		return err, aare
	}
	if found {
		r = bytes.NewReader(content)
		err, ch = der_decode_chunk(r)
		if nil != err {
			return err, aare
		}
		content = content[ch.length:]
	}

	//responder-acse-requirements [8] IMPLICIT ACSE-requirements OPTIONAL,
	err, found = decode_AAREapdu_responderAcseRequirements(ch, aare)
	if nil != err {
		return err, aare
	}
	if found {
		r = bytes.NewReader(content)
		err, ch = der_decode_chunk(r)
		if nil != err {
			return err, aare
		}
		content = content[ch.length:]
	}

	//mechanism-name [9] IMPLICIT Mechanism-name OPTIONAL,
	err, found = decode_AAREapdu_mechanismName(ch, aare)
	if nil != err {
		return err, aare
	}
	if found {
		r = bytes.NewReader(content)
		err, ch = der_decode_chunk(r)
		if nil != err {
			return err, aare
		}
		content = content[ch.length:]
	}

	//responding-authentication-value [10] EXPLICIT Authentication-value OPTIONAL,
	err, found = decode_AAREapdu_respondingAuthenticationValue(ch, aare)
	if nil != err {
		return err, aare
	}
	if found {
		r = bytes.NewReader(content)
		err, ch = der_decode_chunk(r)
		if nil != err {
			return err, aare
		}
		content = content[ch.length:]
	}

	//implementation-information [29] IMPLICIT Implementation-data OPTIONAL,
	err, found = decode_AAREapdu_implementationInformation(ch, aare)
	if nil != err {
		return err, aare
	}
	if found {
		r = bytes.NewReader(content)
		err, ch = der_decode_chunk(r)
		if nil != err {
			return err, aare
		}
		content = content[ch.length:]
	}

	//user-information [30] EXPLICIT Association-information OPTIONAL
	err, found = decode_AAREapdu_userInformation(ch, aare)
	if nil != err {
		return err, aare
	}

	return err, aare
}
