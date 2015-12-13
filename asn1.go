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

// asn1 simple types

type tAsn1BitString []byte
type tAsn1IA5String string
type tAsn1Integer int
type tAsn1ObjectIdentifier []byte
type tAsn1OctetString bytes.Buffer
type tAsn1PrintableString string
type tAsn1T61String string
type tAsn1UTCTime time.Time
type tAsn1GraphicString string

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
	callingAuthenticationValue *tAuthenticationValue
}

//export consumeBytes
func consumeBytes(_buf unsafe.Pointer, _bufLen C.int, ctx unsafe.Pointer) C.int {

	buf := (*bytes.Buffer)(ctx)
	bytes := C.GoBytes(_buf, _bufLen)
	(*buf).Write(bytes)
	return 0
}

func cslice(length int) []byte {
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
	slice := (*[1 << 30]byte)(unsafe.Pointer(C.calloc(C.size_t(length), 1)))[:length:length]
	return slice
}

func encode_AARQapdu(_pdu *AARQapdu) bytes.Buffer {
	var ret C.asn_enc_rval_t
	var pdu *C.AARQ_apdu_t
	var buf bytes.Buffer

	pdu = C.hlp__calloc_AARQ_apdu_t()

	b := cslice(1)
	b[0] = 0x80
	pdu.protocol_version = C.hlp__fill_T_protocol_version_t((*C.T_protocol_version_t)(unsafe.Pointer(nil)), (*C.uint8_t)(&b[0]), C.int(len(b)), C.int(0))

	ret, errno := C.der_encode(&C.asn_DEF_AARQ_apdu, unsafe.Pointer(pdu), (*C.asn_app_consume_bytes_f)(C.consumeBytesWrap), unsafe.Pointer(&buf))
	if -1 == ret.encoded {
		s := C.GoString(ret.failed_type.name)
		panic(fmt.Sprintf("encode_AARQapdu() failed, faile type name: %v, errno: %v", s, errno))
	}
	return buf
}
