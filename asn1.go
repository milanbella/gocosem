package gocosem

// #cgo LDFLAGS: -L${SRCDIR}/c/asn1 -lCosemPdu
// #cgo CPPFLAGS: -I. -I${SRCDIR}/c/asn1 -I${SRCDIR}/c/asn1/go
//
// #include <stdlib.h>
// #include <stdio.h>
// #include <errno.h>
//
// #include "asn1_go.h"
// #include "AARQ-apdu.h"
//
//
import "C"

import (
	"bytes"
	"fmt"
	"time"
	"unsafe"
)

/*
	-- [APPLICATION 0] == [ 60H ] = [ 96 ]
	protocol-version [0] IMPLICIT T-protocol-version DEFAULT {version1},
	application-context-name [1] Application-context-name,
	called-AP-title [2] AP-title OPTIONAL,
	called-AE-qualifier [3] AE-qualifier OPTIONAL,
	called-AP-invocation-id [4] AP-invocation-identifier OPTIONAL,
	called-AE-invocation-id [5] AE-invocation-identifier OPTIONAL,
	calling-AP-title [6] AP-title OPTIONAL,
	calling-AE-qualifier [7] AE-qualifier OPTIONAL,
	calling-AP-invocation-id [8] AP-invocation-identifier OPTIONAL,
	calling-AE-invocation-id [9] AE-invocation-identifier OPTIONAL,
	-- The following field shall not be present if only the kernel is used.
	sender-acse-requirements [10] IMPLICIT ACSE-requirements OPTIONAL,
	-- The following field shall only be present if the authentication functional unit is selected.
	mechanism-name [11] IMPLICIT Mechanism-name OPTIONAL,
	-- The following field shall only be present if the authentication functional unit is selected.

	calling-authentication-value [12] EXPLICIT Authentication-value OPTIONAL,
	implementation-information [29] IMPLICIT Implementation-data OPTIONAL,
	user-information [30] EXPLICIT Association-information OPTIONAL
*/

type tAsn1Choice struct {
	tag int
	val interface{}
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

// asn1 simple types

type tAsn1BitString []byte
type tAsn1IA5String string
type tAsn1Integer int
type tAsn1ObjectIdentifier []byte
type tAsn1OctetString []byte
type tAsn1PrintableString string
type tAsn1T61String string
type tAsn1UTCTime time.Time
type tAsn1GraphicString []byte

type tAuthenticationValue struct {
	callingAuthenticationValue_0 *tAsn1GraphicString
	callingAuthenticationValue_1 *tAsn1BitString
	callingAuthenticationValue_2 *tAsn1BitString
	callingAuthenticationValue_3 *struct {
		otherMechanismName tAsn1ObjectIdentifier
		//other-mechanism-value ANY DEFINED BY other-mechanism-name
		otherMechanismValue tAsn1OctetString
	}
}

type AARQapdu struct {
	protocolVersion            tAsn1BitString
	applicationContextName     tAsn1ObjectIdentifier
	calledAPtitle              *tAsn1OctetString
	calledAEqualifier          *tAsn1OctetString
	calledAPinvocationId       *tAsn1Integer
	calledAEinvocationId       *tAsn1Integer
	callingAPtitle             *tAsn1OctetString
	callingAEqualifier         *tAsn1OctetString
	callingAPinvocationId      *tAsn1Integer
	callingAEinvocationId      *tAsn1Integer
	senderAcseRequirements     *tAsn1BitString
	mechanismName              *tAsn1ObjectIdentifier
	callingAuthenticationValue *tAsn1Choice
}

//export consumeBytes
func consumeBytes(_buf unsafe.Pointer, _bufLen C.int, ctx unsafe.Pointer) C.int {

	buf := (*bytes.Buffer)(ctx)
	bytes := C.GoBytes(_buf, _bufLen)
	(*buf).Write(bytes)
	return 0
}

// Make slice backed by newly allocated C array.

func cslicel(length int) []C.uint8_t {
	/*
	   For how to make a slice backed by underlying C array, see:

	   https://github.com/golang/go/wiki/cgo

	   import "C"
	   import "unsafe"
	   ...
	           var theCArray *C.YourType = C.getTheArray()
	           length := C.getTheArrayLength()
	           slice := (*[1 << 30]C.YourType)(unsafe.Pointer(theCArray))[:length:length]


	*/
	slice := (*[1 << 30]C.uint8_t)(unsafe.Pointer(C.calloc(C.size_t(length), 1)))[:length:length]
	return slice
}

// Make slice backed by newly allocated C array and copies 'b' onto it.

func csliceb(b []byte) []C.uint8_t {

	cb := cslicel(len(b))

	for i := 0; i < len(b); i++ {
		cb[i] = (C.uint8_t)(b[i])
	}

	return cb
}

func encode_AARQapdu(_pdu *AARQapdu) bytes.Buffer {
	var ret C.asn_enc_rval_t
	var pdu *C.AARQ_apdu_t
	var buf bytes.Buffer

	pdu = C.hlp__calloc_AARQ_apdu_t()

	// protocol_version
	cb := csliceb(_pdu.protocolVersion)
	pdu.protocol_version = (*C.T_protocol_version_t)(C.hlp__fill_BIT_STRING_t((*C.BIT_STRING_t)(unsafe.Pointer(nil)), &cb[0], C.int(len(cb)), C.int(0)))

	// application_context_name
	cb = csliceb(_pdu.applicationContextName)
	C.hlp__fill_OBJECT_IDENTIFIER_t((*C.OBJECT_IDENTIFIER_t)(unsafe.Pointer(&pdu.application_context_name)), &cb[0], C.int(len(cb)))

	// called_AP_title
	cb = csliceb(*_pdu.calledAPtitle)
	pdu.called_AP_title = C.hlp__fill_OCTET_STRING_t((*C.OCTET_STRING_t)(unsafe.Pointer(nil)), &cb[0], C.int(len(cb)))

	// called_AE_qualifier
	if nil != _pdu.calledAEqualifier {
		cb = csliceb(*_pdu.calledAEqualifier)
		pdu.called_AE_qualifier = C.hlp__fill_OCTET_STRING_t((*C.OCTET_STRING_t)(unsafe.Pointer(nil)), &cb[0], C.int(len(cb)))
	}

	// called_AP_invocation_id
	if nil != _pdu.calledAPinvocationId {
		pdu.called_AP_invocation_id = (*C.AP_invocation_identifier_t)(unsafe.Pointer(_pdu.calledAPinvocationId))
	}

	// called_AE_invocation_id
	if nil != _pdu.calledAEinvocationId {
		pdu.called_AE_invocation_id = (*C.AE_invocation_identifier_t)(unsafe.Pointer(_pdu.calledAEinvocationId))
	}

	// calling_AP_title
	if nil != _pdu.callingAPtitle {
		cb = csliceb(*_pdu.callingAPtitle)
		pdu.calling_AP_title = C.hlp__fill_OCTET_STRING_t((*C.OCTET_STRING_t)(unsafe.Pointer(nil)), &cb[0], C.int(len(cb)))
	}

	// calling_AE_qualifier
	if nil != *_pdu.callingAEqualifier {
		cb = csliceb(*_pdu.callingAEqualifier)
		pdu.calling_AE_qualifier = C.hlp__fill_OCTET_STRING_t((*C.OCTET_STRING_t)(unsafe.Pointer(nil)), &cb[0], C.int(len(cb)))
	}

	// calling_AP_invocation_id
	if nil != _pdu.callingAPinvocationId {
		pdu.calling_AP_invocation_id = (*C.AP_invocation_identifier_t)(unsafe.Pointer(_pdu.callingAPinvocationId))
	}

	// calling_AE_invocation_id
	if nil != _pdu.callingAEinvocationId {
		pdu.calling_AE_invocation_id = (*C.AE_invocation_identifier_t)(unsafe.Pointer(_pdu.callingAEinvocationId))
	}

	// sender_acse_requirements
	if nil != _pdu.senderAcseRequirements {
		cb = csliceb(*_pdu.senderAcseRequirements)
		pdu.sender_acse_requirements = (*C.ACSE_requirements_t)(C.hlp__fill_BIT_STRING_t((*C.BIT_STRING_t)(unsafe.Pointer(nil)), &cb[0], C.int(len(cb)), C.int(0)))
	}

	// mechanism_name
	if nil != _pdu.mechanismName {
		cb = csliceb(*_pdu.mechanismName)
		pdu.mechanism_name = C.hlp__fill_OBJECT_IDENTIFIER_t((*C.OBJECT_IDENTIFIER_t)(unsafe.Pointer(nil)), &cb[0], C.int(len(cb)))
	}

	//struct Authentication_value	*calling_authentication_value	/* OPTIONAL */;
	if nil != _pdu.callingAuthenticationValue {

		pdu.calling_authentication_value = C.hlp__calloc_struct_Authentication_value()

		switch C.enum_Authentication_value_PR(_pdu.callingAuthenticationValue.getTag()) {
		case 1:
			charstring := (_pdu.callingAuthenticationValue.getVal()).(*tAsn1GraphicString)
			cb = csliceb(*charstring)
			graphicString := (*C.GraphicString_t)(unsafe.Pointer(&pdu.calling_authentication_value.choice[0]))
			C.hlp__fill_OCTET_STRING_t(graphicString, &cb[0], C.int(len(cb)))
		default:
			panic(fmt.Sprintf("encode_AARQapdu() failed, unknown callingAuthenticationValue tag %v", _pdu.callingAuthenticationValue.getTag()))
		}

	}

	ret, errno := C.der_encode(&C.asn_DEF_AARQ_apdu, unsafe.Pointer(pdu), (*C.asn_app_consume_bytes_f)(C.consumeBytesWrap), unsafe.Pointer(&buf))
	if -1 == ret.encoded {
		s := C.GoString(ret.failed_type.name)
		panic(fmt.Sprintf("encode_AARQapdu() failed, faile type name: %v, errno: %v", s, errno))
	}
	return buf
}
