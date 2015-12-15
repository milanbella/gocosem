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
	bits       []byte
	unusedBits int
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

/*
typedef enum Association_result {
	Association_result_accepted	= 0,
	Association_result_rejected_permanent	= 1,
	Association_result_rejected_transient	= 2
} e_Association_result;
*/

const C_Association_result_accepted = int(C.Association_result_accepted)
const C_Association_result_rejected_permanent = int(C.Association_result_rejected_permanent)
const C_Association_result_rejected_transient = int(C.Association_result_rejected_transient)

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

/*
AARE-apdu ::= [APPLICATION 1] IMPLICIT SEQUENCE
{
	-- [APPLICATION 1] == [ 61H ] = [ 97 ]
	protocol-version [0] IMPLICIT T-protocol-version DEFAULT {version1},
	application-context-name [1] Application-context-name,
	result [2] Association-result,
	result-source-diagnostic [3] Associate-source-diagnostic,
	responding-AP-title [4] AP-title OPTIONAL,
	responding-AE-qualifier [5] AE-qualifier OPTIONAL,
	responding-AP-invocation-id [6] AP-invocation-identifier OPTIONAL,
	responding-AE-invocation-id [7] AE-invocation-identifier OPTIONAL,
	-- The following field shall not be present if only the kernel is used.
	responder-acse-requirements [8] IMPLICIT ACSE-requirements OPTIONAL,
	-- The following field shall only be present if the authentication functional unit is selected.
	mechanism-name [9] IMPLICIT Mechanism-name OPTIONAL,
	-- The following field shall only be present if the authentication functional unit is selected.
	responding-authentication-value [10] EXPLICIT Authentication-value OPTIONAL,
	implementation-information [29] IMPLICIT Implementation-data OPTIONAL,
	user-information [30] EXPLICIT Association-information OPTIONAL
}
*/

const C_Associate_source_diagnostic_PR_NOTHING = int(C.Associate_source_diagnostic_PR_NOTHING)
const C_Associate_source_diagnostic_PR_acse_service_user = int(C.Associate_source_diagnostic_PR_acse_service_user)
const CAssociate_source_diagnostic_PR_acse_service_provider = int(C.Associate_source_diagnostic_PR_acse_service_provider)

/*
typedef enum acse_service_user {
	acse_service_user_null	= 0,
	acse_service_user_no_reason_given	= 1,
	acse_service_user_application_context_name_not_supported	= 2,
	acse_service_user_authentication_mechanism_name_not_recognised	= 11,
	acse_service_user_authentication_mechanism_name_required	= 12,
	acse_service_user_authentication_failure	= 13,
	acse_service_user_authentication_required	= 14
} e_acse_service_user;
*/
const C_acse_service_user_null = int(C.acse_service_user_null)
const C_acse_service_user_no_reason_given = int(C.acse_service_user_no_reason_given)
const C_acse_service_user_application_context_name_not_supported = int(C.acse_service_user_application_context_name_not_supported)
const C_acse_service_user_authentication_mechanism_name_not_recognised = int(C.acse_service_user_authentication_mechanism_name_not_recognised)
const C_acse_service_user_authentication_mechanism_name_required = int(C.acse_service_user_authentication_mechanism_name_required)
const C_acse_service_user_authentication_failure = int(C.acse_service_user_authentication_failure)
const C_acse_service_user_authentication_required = int(C.acse_service_user_authentication_required)

/*
typedef enum acse_service_provider {
	acse_service_provider_null	= 0,
	acse_service_provider_no_reason_given	= 1,
	acse_service_provider_no_common_acse_version	= 2
} e_acse_service_provider;
*/
const C_acse_service_provider_null = int(C.acse_service_provider_null)
const C_acse_service_provider_no_reason_given = int(C.acse_service_provider_no_reason_given)
const C_acse_service_provider_no_common_acse_version = int(C.acse_service_provider_no_common_acse_version)

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

func cAsn1BitString(bitString *tAsn1BitString) *C.BIT_STRING_t {

	if nil == bitString {
		return (*C.BIT_STRING_t)(unsafe.Pointer(nil))
	}

	cb := cslice(bitString.bits)
	unusedBits := bitString.unusedBits
	return (*C.T_protocol_version_t)(C.hlp__fill_BIT_STRING_t((*C.BIT_STRING_t)(unsafe.Pointer(nil)), &cb[0], C.int(len(cb)), C.int(unusedBits)))
}

func cAsn1OctetString(octetString *tAsn1OctetString) *C.OCTET_STRING_t {
	if nil == octetString {
		return (*C.OCTET_STRING_t)(unsafe.Pointer(nil))
	}
	cb := cslice(*octetString)
	return C.hlp__fill_OCTET_STRING_t((*C.OCTET_STRING_t)(unsafe.Pointer(nil)), &cb[0], C.int(len(cb)))
}

func cAsn1GraphicString(graphicString *tAsn1GraphicString) *C.GraphicString_t {
	if nil == graphicString {
		return (*C.GraphicString_t)(unsafe.Pointer(nil))
	}
	cb := cslice(*graphicString)
	return C.hlp__fill_OCTET_STRING_t((*C.GraphicString_t)(unsafe.Pointer(nil)), &cb[0], C.int(len(cb)))
}

func cAsn1ObjectIdentifier(oid *tAsn1ObjectIdentifier) *C.OBJECT_IDENTIFIER_t {
	if nil == oid {
		return (*C.OBJECT_IDENTIFIER_t)(unsafe.Pointer(nil))
	}
	coid := C.hlp__calloc_OBJECT_IDENTIFIER_t()

	//int OBJECT_IDENTIFIER_get_arcs(const OBJECT_IDENTIFIER_t *_oid,
	//	void *_arcs,			/* e.g., unsigned int arcs[N] */
	//	unsigned int _arc_type_size,	/* e.g., sizeof(arcs[0]) */
	//	unsigned int _arc_slots		/* e.g., N */);
	//	C.OBJECT_IDENTIFIER_get_arcs(coid, C.uint(len(*oid))

	length := len(*oid)
	cb := (*[1 << 26]C.uint32_t)(unsafe.Pointer(C.calloc(C.size_t(length), 4)))[:length:length]
	for i := 0; i < length; i++ {
		cb[i] = C.uint32_t((*oid)[i])
	}
	ret := C.OBJECT_IDENTIFIER_set_arcs(coid, unsafe.Pointer(&cb[0]), 4, C.uint(length))
	if -1 == ret {
		panic("cAsn1ObjectIdentifier(): cannot encode oid")
	}
	C.free(unsafe.Pointer(&cb[0]))
	return coid
}

func cAsn1Any(any *tAsn1Any) *C.ANY_t {
	if nil == any {
		return (*C.ANY_t)(unsafe.Pointer(nil))
	}
	cb := cslice(*any)
	return C.hlp__fill_ANY_t((*C.ANY_t)(unsafe.Pointer(nil)), &cb[0], C.int(len(cb)))
}

func encode_AARQapdu(_pdu *AARQapdu) bytes.Buffer {
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
		pdu.called_AP_invocation_id = (*C.AP_invocation_identifier_t)(unsafe.Pointer(_pdu.calledAPinvocationId))
	}

	// called_AE_invocation_id
	if nil != _pdu.calledAEinvocationId {
		pdu.called_AE_invocation_id = (*C.AE_invocation_identifier_t)(unsafe.Pointer(_pdu.calledAEinvocationId))
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
		pdu.calling_AP_invocation_id = (*C.AP_invocation_identifier_t)(unsafe.Pointer(_pdu.callingAPinvocationId))
	}

	// calling_AE_invocation_id
	if nil != _pdu.callingAEinvocationId {
		pdu.calling_AE_invocation_id = (*C.AE_invocation_identifier_t)(unsafe.Pointer(_pdu.callingAEinvocationId))
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

	if nil != _pdu.implementationInformation {
		pdu.implementation_information = cAsn1GraphicString(_pdu.implementationInformation)
	}

	if nil != _pdu.userInformation {
		pdu.user_information = cAsn1OctetString(_pdu.userInformation)
	}

	ret, errno := C.der_encode(&C.asn_DEF_AARQ_apdu, unsafe.Pointer(pdu), (*C.asn_app_consume_bytes_f)(C.consumeBytesWrap), unsafe.Pointer(&buf))
	if -1 == ret.encoded {
		s := C.GoString(ret.failed_type.name)
		panic(fmt.Sprintf("C.der_encode() failed, failed type name: %v, errno: %v", s, errno))
	}
	C.hlp__free_AARQ_apdu_t(pdu)
	return buf
}

func decode_AAREapdu(b bytes.Buffer) *AAREapdu {
	var _pdu *C.AARE_apdu_t

	cb := cslice(b.Bytes())
	ret, errno := C.ber_decode((*C.struct_asn_codec_ctx_s)(unsafe.Pointer(nil)), &C.asn_DEF_AARE_apdu, (*unsafe.Pointer)(unsafe.Pointer(&_pdu)), unsafe.Pointer(&cb[0]), C.size_t(len(cb)))
	C.free(unsafe.Pointer(&cb[0]))
	if C.RC_OK != ret.code {
		panic(fmt.Sprintf("C.ber_decode() failed, code: %v, consumed: , errno %v", ret.code, ret.consumed, errno))
	}
	return nil

}
