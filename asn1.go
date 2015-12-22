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
// #include "AARE-apdu.h"
// #include "Data.h"
//
//
import "C"

import (
	"bytes"
	"fmt"
	"time"
	"unsafe"
)

// asn1 simple types

type tAsn1BitString struct {
	buf        []byte
	bitsUnused int
}

type tAsn1IA5String string
type tAsn1Integer int
type tAsn1ObjectIdentifier []uint
type tAsn1OctetString []byte
type tAsn1PrintableString string
type tAsn1T61String string
type tAsn1UTCTime time.Time
type tAsn1GraphicString []byte
type tAsn1Any []byte
type tAsn1Null int
type tAsn1Boolean bool

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

const C_RC_OK = int(C.RC_OK)
const C_RC_WMORE = int(C.RC_WMORE)
const C_RC_FAIL = int(C.RC_FAIL)

const C_Authentication_value_PR_NOTHING = int(C.Authentication_value_PR_NOTHING)
const C_Authentication_value_PR_charstring = int(C.Authentication_value_PR_charstring)
const C_Authentication_value_PR_bitstring = int(C.Authentication_value_PR_bitstring)
const C_Authentication_value_PR_external = int(C.Authentication_value_PR_external)
const C_Authentication_value_PR_other = int(C.Authentication_value_PR_other)

const C_Association_result_accepted = int(C.Association_result_accepted)
const C_Association_result_rejected_permanent = int(C.Association_result_rejected_permanent)
const C_Association_result_rejected_transient = int(C.Association_result_rejected_transient)

const C_Associate_source_diagnostic_PR_NOTHING = int(C.Associate_source_diagnostic_PR_NOTHING)
const C_Associate_source_diagnostic_PR_acse_service_user = int(C.Associate_source_diagnostic_PR_acse_service_user)
const CAssociate_source_diagnostic_PR_acse_service_provider = int(C.Associate_source_diagnostic_PR_acse_service_provider)

const C_acse_service_user_null = int(C.acse_service_user_null)
const C_acse_service_user_no_reason_given = int(C.acse_service_user_no_reason_given)
const C_acse_service_user_application_context_name_not_supported = int(C.acse_service_user_application_context_name_not_supported)
const C_acse_service_user_authentication_mechanism_name_not_recognised = int(C.acse_service_user_authentication_mechanism_name_not_recognised)
const C_acse_service_user_authentication_mechanism_name_required = int(C.acse_service_user_authentication_mechanism_name_required)
const C_acse_service_user_authentication_failure = int(C.acse_service_user_authentication_failure)
const C_acse_service_user_authentication_required = int(C.acse_service_user_authentication_required)

const C_acse_service_provider_null = int(C.acse_service_provider_null)
const C_acse_service_provider_no_reason_given = int(C.acse_service_provider_no_reason_given)
const C_acse_service_provider_no_common_acse_version = int(C.acse_service_provider_no_common_acse_version)

const C_Data_PR_NOTHING = int(C.Data_PR_NOTHING)
const C_Data_PR_null_data = int(C.Data_PR_null_data)
const C_Data_PR_array = int(C.Data_PR_array)
const C_Data_PR_structure = int(C.Data_PR_structure)
const C_Data_PR_boolean = int(C.Data_PR_boolean)
const C_Data_PR_bit_string = int(C.Data_PR_bit_string)
const C_Data_PR_double_long = int(C.Data_PR_double_long)
const C_Data_PR_double_long_unsigned = int(C.Data_PR_double_long_unsigned)
const C_Data_PR_floating_point = int(C.Data_PR_floating_point)
const C_Data_PR_octet_string = int(C.Data_PR_octet_string)
const C_Data_PR_visible_string = int(C.Data_PR_visible_string)
const C_Data_PR_bcd = int(C.Data_PR_bcd)
const C_Data_PR_integer = int(C.Data_PR_integer)
const C_Data_PR_long = int(C.Data_PR_long)
const C_Data_PR_unsigned = int(C.Data_PR_unsigned)
const C_Data_PR_long_unsigned = int(C.Data_PR_long_unsigned)
const C_Data_PR_compact_array = int(C.Data_PR_compact_array)
const C_Data_PR_enum = int(C.Data_PR_enum)
const C_Data_PR_float32 = int(C.Data_PR_float32)
const C_Data_PR_float64 = int(C.Data_PR_float64)
const C_Data_PR_date_time = int(C.Data_PR_date_time)
const C_Data_PR_date = int(C.Data_PR_date)
const C_Data_PR_time = int(C.Data_PR_time)
const C_Data_PR_dont_care = int(C.Data_PR_dont_care)

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
	cb := (*[1 << 30]C.uint8_t)(unsafe.Pointer(C.calloc(C.size_t(length), 1)))[:length:length]
	return cb
}

// Make slice backed by newly allocated C array and copies there  'length' of bytes pointed by 'p'

func cslicep(p uintptr, length uintptr) []C.uint8_t {
	n := int(length)
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
	b := (*[1 << 30]byte)(unsafe.Pointer(p))[:n:n]
	cb := (*[1 << 30]C.uint8_t)(unsafe.Pointer(C.calloc(C.size_t(n), 1)))[:n:n]

	for i := 0; i < n; i++ {
		cb[i] = (C.uint8_t)(b[i])
	}

	return cb
}

// Make slice backed by newly allocated C array and copies 'b' onto it.

func cslice(b []byte) []C.uint8_t {

	cb := cslicel(len(b))

	for i := 0; i < len(b); i++ {
		cb[i] = (C.uint8_t)(b[i])
	}

	return cb
}

func goslice(cb *C.uint8_t, n C.int) []byte {
	return C.GoBytes(unsafe.Pointer(cb), n)
}

func cAsn1Integer(i *tAsn1Integer) *C.long {
	if nil == i {
		return nil
	}
	ci := C.hlp__calloc_long(1)
	*ci = C.long(*i)
	return ci
}

func goAsn1Integer(ci *C.long) *tAsn1Integer {
	if nil == ci {
		return nil
	}
	i := new(tAsn1Integer)
	*i = tAsn1Integer(*ci)
	return i
}

func cAsn1BitString(bitString *tAsn1BitString) *C.BIT_STRING_t {

	if nil == bitString {
		return (*C.BIT_STRING_t)(unsafe.Pointer(nil))
	}

	cb := cslice(bitString.buf)
	bitsUnused := bitString.bitsUnused
	return (*C.T_protocol_version_t)(C.hlp__fill_BIT_STRING_t((*C.BIT_STRING_t)(unsafe.Pointer(nil)), &cb[0], C.int(len(cb)), C.int(bitsUnused)))
}

func goAsn1BitString(cBitString *C.BIT_STRING_t) *tAsn1BitString {
	if nil == cBitString {
		return nil
	}

	bitString := new(tAsn1BitString)

	bitString.buf = goslice(cBitString.buf, cBitString.size)
	bitString.bitsUnused = int(cBitString.bits_unused)
	return bitString
}

func cAsn1OctetString(octetString *tAsn1OctetString) *C.OCTET_STRING_t {
	if nil == octetString {
		return (*C.OCTET_STRING_t)(unsafe.Pointer(nil))
	}
	cb := cslice(*octetString)
	return C.hlp__fill_OCTET_STRING_t((*C.OCTET_STRING_t)(unsafe.Pointer(nil)), &cb[0], C.int(len(cb)))
}

func goAsn1OctetString(cOctetString *C.OCTET_STRING_t) *tAsn1OctetString {
	if nil == cOctetString {
		return nil
	}

	octetString := new(tAsn1OctetString)
	*octetString = tAsn1OctetString(goslice(cOctetString.buf, cOctetString.size))
	return octetString
}

func cAsn1GraphicString(graphicString *tAsn1GraphicString) *C.GraphicString_t {
	if nil == graphicString {
		return (*C.GraphicString_t)(unsafe.Pointer(nil))
	}
	cb := cslice(*graphicString)
	return C.hlp__fill_OCTET_STRING_t((*C.GraphicString_t)(unsafe.Pointer(nil)), &cb[0], C.int(len(cb)))
}

func goAsn1GraphicString(cGraphicString *C.GraphicString_t) *tAsn1GraphicString {
	if nil == cGraphicString {
		return nil
	}
	graphicString := new(tAsn1GraphicString)
	*graphicString = tAsn1GraphicString(goslice(cGraphicString.buf, cGraphicString.size))
	return graphicString
}

func cAsn1ObjectIdentifier(oid *tAsn1ObjectIdentifier) *C.OBJECT_IDENTIFIER_t {
	if nil == oid {
		return (*C.OBJECT_IDENTIFIER_t)(unsafe.Pointer(nil))
	}
	cOid := C.hlp__calloc_OBJECT_IDENTIFIER_t()

	length := len(*oid)
	cb := (*[1 << 26]C.uint32_t)(unsafe.Pointer(C.calloc(C.size_t(length), 4)))[:length:length]
	for i := 0; i < length; i++ {
		cb[i] = C.uint32_t((*oid)[i])
	}
	ret := C.OBJECT_IDENTIFIER_set_arcs(cOid, unsafe.Pointer(&cb[0]), 4, C.uint(length))
	if -1 == ret {
		panic("cAsn1ObjectIdentifier(): cannot encode oid")
	}
	C.free(unsafe.Pointer(&cb[0]))
	return cOid
}

func goAsn1ObjectIdentifier(cOid *C.OBJECT_IDENTIFIER_t) *tAsn1ObjectIdentifier {
	if nil == cOid {
		return nil
	}

	length := C.int(20)
	cb := (*[1 << 26]C.uint32_t)(unsafe.Pointer(C.calloc(C.size_t(length), 4)))[:length:length]
	ret := C.OBJECT_IDENTIFIER_get_arcs(cOid, unsafe.Pointer(&cb[0]), 4, C.uint(length))
	if -1 == ret {
		panic("goAsn1ObjectIdentifier(): cannot decode oid")
	}
	n := int(ret)
	b := make([]uint, n)
	for i := 0; i < n; i++ {
		b[i] = uint(cb[i])
	}
	C.free(unsafe.Pointer(&cb[0]))
	return (*tAsn1ObjectIdentifier)(&b)

}

func cAsn1Any(any *tAsn1Any) *C.ANY_t {
	if nil == any {
		return (*C.ANY_t)(unsafe.Pointer(nil))
	}
	cb := cslice(*any)
	return C.hlp__fill_ANY_t((*C.ANY_t)(unsafe.Pointer(nil)), &cb[0], C.int(len(cb)))
}

func goAsn1Any(cAny *C.ANY_t) *tAsn1Any {
	if nil == cAny {
		return nil
	}

	any := new(tAsn1Any)
	*any = goslice(cAny.buf, cAny.size)
	return any
}

func cAsn1Null() *C.NULL_t {
	cNull := C.hlp__calloc_NULL_t()
	*cNull = C.NULL_t(0)
	return cNull
}

func goAsn1Null() *tAsn1Null {
	null := new(tAsn1Null)
	(*null) = 0
	return null
}

func cAsn1Boolean(b *tAsn1Boolean) *C.BOOLEAN_t {
	if nil == b {
		return (*C.BOOLEAN_t)(unsafe.Pointer(nil))
	}
	cb := C.hlp__calloc_BOOLEAN_t()
	if *b {
		*cb = C.BOOLEAN_t(1)
	} else {
		*cb = C.BOOLEAN_t(0)
	}
	return cb
}

func goAsn1Boolean(cb *C.BOOLEAN_t) *tAsn1Boolean {
	if nil == cb {
		return nil
	}
	b := new(tAsn1Boolean)
	if C.BOOLEAN_t(0) == (*cb) {
		*b = false
	} else {
		*b = true
	}
	return b
}

func encode_AARQapdu(_pdu *AARQapdu) []byte {
	var ret C.asn_enc_rval_t
	var pdu *C.AARQ_apdu_t
	var buf bytes.Buffer
	var cb []C.uint8_t

	pdu = C.hlp__calloc_AARQ_apdu_t()

	// protocol_version
	pdu.protocol_version = cAsn1BitString(_pdu.protocolVersion)

	// application_context_name
	pdu.application_context_name = *cAsn1ObjectIdentifier(&_pdu.applicationContextName)

	// called_AP_title
	if nil != _pdu.calledAPtitle {
		pdu.called_AP_title = cAsn1OctetString(_pdu.calledAPtitle)
	}

	// called_AE_qualifier
	if nil != _pdu.calledAEqualifier {
		pdu.called_AE_qualifier = cAsn1OctetString(_pdu.calledAEqualifier)
	}

	// called_AP_invocation_id
	if nil != _pdu.calledAPinvocationId {
		pdu.called_AP_invocation_id = (*C.AP_invocation_identifier_t)(cAsn1Integer(_pdu.calledAPinvocationId))
	}

	// called_AE_invocation_id
	if nil != _pdu.calledAEinvocationId {
		pdu.called_AE_invocation_id = (*C.AE_invocation_identifier_t)(cAsn1Integer(_pdu.calledAEinvocationId))
	}

	// calling_AP_title
	if nil != _pdu.callingAPtitle {
		pdu.calling_AP_title = cAsn1OctetString(_pdu.callingAPtitle)
	}

	// calling_AE_qualifier
	if nil != _pdu.callingAEqualifier {
		pdu.calling_AE_qualifier = cAsn1OctetString(_pdu.callingAEqualifier)
	}

	// calling_AP_invocation_id
	if nil != _pdu.callingAPinvocationId {
		pdu.calling_AP_invocation_id = (*C.AP_invocation_identifier_t)(cAsn1Integer(_pdu.callingAPinvocationId))
	}

	// calling_AE_invocation_id
	if nil != _pdu.callingAEinvocationId {
		pdu.calling_AE_invocation_id = (*C.AE_invocation_identifier_t)(cAsn1Integer(_pdu.callingAEinvocationId))
	}

	// sender_acse_requirements
	if nil != _pdu.senderAcseRequirements {
		pdu.sender_acse_requirements = cAsn1BitString(_pdu.senderAcseRequirements)

	}

	// mechanism_name
	if nil != _pdu.mechanismName {
		pdu.mechanism_name = cAsn1ObjectIdentifier(_pdu.mechanismName)
	}

	//struct Authentication_value	*calling_authentication_value	/* OPTIONAL */;
	if nil != _pdu.callingAuthenticationValue {
		_av := &_pdu.callingAuthenticationValue
		av := &pdu.calling_authentication_value

		(*av) = C.hlp__calloc_struct_Authentication_value()
		/*
			Authentication_value_PR_NOTHING,
			Authentication_value_PR_charstring,
			Authentication_value_PR_bitstring,
			Authentication_value_PR_external,
			Authentication_value_PR_other
		*/

		switch C.enum_Authentication_value_PR((*_av).getTag()) {

		case C.Authentication_value_PR_NOTHING:
			(*av).present = C.Authentication_value_PR_NOTHING

		case C.Authentication_value_PR_charstring:
			charstring := ((*_av).getVal()).(*tAsn1GraphicString)
			(*av).present = C.Authentication_value_PR_charstring
			*(*C.GraphicString_t)(unsafe.Pointer(&(*av).choice[0])) = *cAsn1GraphicString(charstring)

		case C.Authentication_value_PR_bitstring:
			bistring := ((*_av).getVal()).(*tAsn1BitString)
			(*av).present = C.Authentication_value_PR_bitstring
			*(*C.BIT_STRING_t)(unsafe.Pointer(&(*av).choice[0])) = *cAsn1BitString(bistring)

		case C.Authentication_value_PR_external:
			external := ((*_av).getVal()).(*tAsn1OctetString)
			(*av).present = C.Authentication_value_PR_external
			*(*C.OCTET_STRING_t)(unsafe.Pointer(&(*av).choice[0])) = *cAsn1OctetString(external)
		case C.Authentication_value_PR_other:
			var avo C.struct_Authentication_value_other
			other := ((*_av).getVal()).(*tAuthenticationValueOther)

			avo.other_mechanism_name = *cAsn1ObjectIdentifier(&other.otherMechanismName)
			avo.other_mechanism_value = *cAsn1Any(&other.otherMechanismValue)

			(*av).present = C.Authentication_value_PR_other
			cb = cslicep((uintptr)(unsafe.Pointer(&avo)), unsafe.Sizeof(avo))
			C.hlp__fill_OCTET_STRING_t((*C.OCTET_STRING_t)(unsafe.Pointer(&(*av).choice[0])), &cb[0], C.int(len(cb)))
		default:
			panic(fmt.Sprintf("encode_AARQapdu() failed, unknown callingAuthenticationValue tag %v", _pdu.callingAuthenticationValue.getTag()))
		}
	}

	//implementation-information [29] IMPLICIT Implementation-data OPTIONAL,
	if nil != _pdu.implementationInformation {
		pdu.implementation_information = cAsn1GraphicString(_pdu.implementationInformation)
	}

	//user-information [30] EXPLICIT Association-information OPTIONAL
	if nil != _pdu.userInformation {
		pdu.user_information = cAsn1OctetString(_pdu.userInformation)
	}

	ret, errno := C.der_encode(&C.asn_DEF_AARQ_apdu, unsafe.Pointer(pdu), (*C.asn_app_consume_bytes_f)(C.consumeBytesWrap), unsafe.Pointer(&buf))
	if -1 == ret.encoded {
		s := C.GoString(ret.failed_type.name)
		panic(fmt.Sprintf("C.der_encode() failed, failed type name: %v, errno: %v", s, errno))
	}
	C.hlp__free_AARQ_apdu_t(pdu)
	return buf.Bytes()
}

func decode_AAREapdu(inb []byte) (pdu *AAREapdu) {

	buf := bytes.NewBuffer(inb)
	var _pdu *C.AARE_apdu_t

	pdu = new(AAREapdu)

	cb := cslice(buf.Bytes())
	ret, errno := C.ber_decode((*C.struct_asn_codec_ctx_s)(unsafe.Pointer(nil)), &C.asn_DEF_AARE_apdu, (*unsafe.Pointer)(unsafe.Pointer(&_pdu)), unsafe.Pointer(&cb[0]), C.size_t(len(cb)))
	C.free(unsafe.Pointer(&cb[0]))
	if C.RC_OK != ret.code {
		panic(fmt.Sprintf("C.ber_decode() failed, code: %v, consumed: , errno %v", ret.code, ret.consumed, errno))
	}

	//-- [APPLICATION 1] == [ 61H ] = [ 97 ]
	//protocol-version [0] IMPLICIT T-protocol-version DEFAULT {version1},
	//protocolVersion *tAsn1BitString
	pdu.protocolVersion = goAsn1BitString(_pdu.protocol_version)

	//application-context-name [1] Application-context-name,
	//applicationContextName tAsn1ObjectIdentifier
	pdu.applicationContextName = *goAsn1ObjectIdentifier(&_pdu.application_context_name)

	//result [2] Association-result,
	//result tAsn1Integer
	pdu.result = *goAsn1Integer((*C.long)(&_pdu.result))

	//result-source-diagnostic [3] Associate-source-diagnostic,
	//resultSourceDiagnostic tAsn1Choice
	b := _pdu.result_source_diagnostic.choice
	switch _pdu.result_source_diagnostic.present {
	case C.Associate_source_diagnostic_PR_NOTHING:
		pdu.resultSourceDiagnostic.setVal(int(C.Associate_source_diagnostic_PR_NOTHING), nil)
	case C.Associate_source_diagnostic_PR_acse_service_user:
		pdu.resultSourceDiagnostic.setVal(int(C.Associate_source_diagnostic_PR_acse_service_user), int(*(*C.long)(unsafe.Pointer(&b[0]))))
	case C.Associate_source_diagnostic_PR_acse_service_provider:
		pdu.resultSourceDiagnostic.setVal(int(C.Associate_source_diagnostic_PR_acse_service_provider), int(*(*C.long)(unsafe.Pointer(&b[0]))))
	default:
		panic(fmt.Sprintf("decode_AAREapdu(): unknown choice tag: %v", int(_pdu.result_source_diagnostic.present)))
	}

	//responding-AP-title [4] AP-title OPTIONAL,
	//respondingAPtitle *tAsn1OctetString
	pdu.respondingAPtitle = goAsn1OctetString(_pdu.responding_AP_title)

	//responding-AE-qualifier [5] AE-qualifier OPTIONAL,
	//respondingAEqualifier *tAsn1OctetString
	pdu.respondingAEqualifier = goAsn1OctetString(_pdu.responding_AE_qualifier)

	//responding-AP-invocation-id [6] AP-invocation-identifier OPTIONAL,
	//respondingAPinvocationId *tAsn1Integer
	pdu.respondingAPinvocationId = goAsn1Integer((*C.long)(_pdu.responding_AP_invocation_id))

	//responding-AE-invocation-id [7] AE-invocation-identifier OPTIONAL,
	//respondingAEinvocationId *tAsn1Integer
	pdu.respondingAEinvocationId = goAsn1Integer((*C.long)(_pdu.responding_AE_invocation_id))

	//-- The following field shall not be present if only the kernel is used.
	//responder-acse-requirements [8] IMPLICIT ACSE-requirements OPTIONAL,
	//responderAcseRequirements *tAsn1BitString
	pdu.responderAcseRequirements = goAsn1BitString(_pdu.responder_acse_requirements)

	//-- The following field shall only be present if the authentication functional unit is selected.
	//mechanism-name [9] IMPLICIT Mechanism-name OPTIONAL,
	//mechanismName *tAsn1ObjectIdentifier
	pdu.mechanismName = goAsn1ObjectIdentifier(_pdu.mechanism_name)

	//-- The following field shall only be present if the authentication functional unit is selected.
	//responding-authentication-value [10] EXPLICIT Authentication-value OPTIONAL,
	//respondingAuthenticationValue *tAsn1Choice
	if nil != _pdu.responding_authentication_value {
		pdu.respondingAuthenticationValue = new(tAsn1Choice)
		b := _pdu.responding_authentication_value.choice
		switch _pdu.responding_authentication_value.present {
		case C.Authentication_value_PR_NOTHING:
			pdu.respondingAuthenticationValue.setVal(int(C.Authentication_value_PR_NOTHING), nil)
		case C.Authentication_value_PR_charstring:
			pdu.respondingAuthenticationValue.setVal(int(C.Authentication_value_PR_charstring), goAsn1GraphicString((*C.GraphicString_t)(unsafe.Pointer(&b[0]))))
		case C.Authentication_value_PR_bitstring:
			pdu.respondingAuthenticationValue.setVal(int(C.Authentication_value_PR_bitstring), goAsn1BitString((*C.BIT_STRING_t)(unsafe.Pointer(&b[0]))))
		case C.Authentication_value_PR_external:
			pdu.respondingAuthenticationValue.setVal(int(C.Authentication_value_PR_external), goAsn1OctetString((*C.OCTET_STRING_t)(unsafe.Pointer(&b[0]))))
		case C.Authentication_value_PR_other:
			var _other *C.Authentication_value_other_t = (*C.Authentication_value_other_t)(unsafe.Pointer(&b[0]))
			var other tAuthenticationValueOther
			other.otherMechanismName = *goAsn1ObjectIdentifier(&_other.other_mechanism_name)
			other.otherMechanismValue = *goAsn1Any(&_other.other_mechanism_value)
			pdu.respondingAuthenticationValue.setVal(int(C.Authentication_value_PR_other), &other)
		default:
			panic(fmt.Sprintf("decode_AAREapdu(): unknown choice tag: %v", int(_pdu.responding_authentication_value.present)))
		}
	}

	//implementation-information [29] IMPLICIT Implementation-data OPTIONAL,
	//implementationInformation *tAsn1GraphicString
	pdu.implementationInformation = goAsn1GraphicString(_pdu.implementation_information)

	//user-information [30] EXPLICIT Association-information OPTIONAL
	//userInformation *tAsn1OctetString
	pdu.userInformation = goAsn1OctetString(_pdu.user_information)

	C.hlp__free_AARE_apdu_t(_pdu)

	return pdu
}

func encode_Data(_data *tAsn1Choice) []byte {
	data := C.hlp__calloc_Data_t()

	switch C.Data_PR((*_data).getTag()) {
	case C.Data_PR_NOTHING:
		data.present = C.Data_PR_NOTHING
	case C.Data_PR_null_data:
		data.present = C.Data_PR_null_data
		*(*C.NULL_t)(unsafe.Pointer(&(*data).choice[0])) = *cAsn1Null()
	case C.Data_PR_array:
		panic(fmt.Sprintf("encode_Data(): array not implemnted"))
	case C.Data_PR_structure:
		panic(fmt.Sprintf("encode_Data(): structure not implemnted"))
	case C.Data_PR_boolean:
		data.present = C.Data_PR_boolean
		cb := cAsn1Boolean((_data.getVal()).(*tAsn1Boolean))
		*(*C.BOOLEAN_t)(unsafe.Pointer(&(*data).choice[0])) = *cb
		C.free(unsafe.Pointer(cb))
	case C.Data_PR_bit_string:
	case C.Data_PR_double_long:
	case C.Data_PR_double_long_unsigned:
	case C.Data_PR_floating_point:
	case C.Data_PR_octet_string:
	case C.Data_PR_visible_string:
	case C.Data_PR_bcd:
	case C.Data_PR_integer:
	case C.Data_PR_long:
	case C.Data_PR_unsigned:
	case C.Data_PR_long_unsigned:
	case C.Data_PR_compact_array:
	case C.Data_PR_enum:
	case C.Data_PR_float32:
	case C.Data_PR_float64:
	case C.Data_PR_date_time:
	case C.Data_PR_date:
	case C.Data_PR_time:
	case C.Data_PR_dont_care:
	default:
		panic(fmt.Sprintf("encode_Data(): unknown choice tag: %v", int((*_data).getTag())))
	}

	return nil

}
