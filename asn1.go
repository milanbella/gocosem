// -build ignore

package gocosem

// #cgo LDFLAGS: -L${SRCDIR}/c/asn1 -L/usr/local/lib -lCosemPdu
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
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"time"
	"unsafe"
)

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
const C_Data_PR_long64 = int(C.Data_PR_long64)
const C_Data_PR_long64_unsigned = int(C.Data_PR_long64_unsigned)
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
	b := C.GoBytes(_buf, _bufLen)
	(*buf).Write(b)
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
	cb := (*[1 << 30]C.uint8_t)(unsafe.Pointer(C.hlp__calloc(C.size_t(length), 1)))[:length:length]
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
	cb := (*[1 << 30]C.uint8_t)(unsafe.Pointer(C.hlp__calloc(C.size_t(n), 1)))[:n:n]

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

func goslice(cb *C.uint8_t, n C.size_t) []byte {
	return C.GoBytes(unsafe.Pointer(cb), C.int(n))
}

func cAsn1Integer(ci *C.long, i *tAsn1Integer) *C.long {
	if nil == i {
		return nil
	}
	if nil == ci {
		ci = C.hlp__calloc_long(1)
	}
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

func cAsn1Integer8(ci *C.Integer8_t, i *tAsn1Integer8) *C.Integer8_t {
	if nil == i {
		return nil
	}
	if nil == ci {
		ci = C.hlp__calloc_Integer8_t(1)
	}
	*ci = C.Integer8_t(*i)
	return ci
}

func goAsn1Integer8(ci *C.Integer8_t) *tAsn1Integer8 {
	if nil == ci {
		return nil
	}
	i := new(tAsn1Integer8)
	*i = tAsn1Integer8(*ci)
	return i
}

func cAsn1Integer16(ci *C.Integer16_t, i *tAsn1Integer16) *C.Integer16_t {
	if nil == i {
		return nil
	}
	if nil == ci {
		ci = C.hlp__calloc_Integer16_t(1)
	}
	*ci = C.Integer16_t(*i)
	return ci
}

func goAsn1Integer16(ci *C.Integer16_t) *tAsn1Integer16 {
	if nil == ci {
		return nil
	}
	i := new(tAsn1Integer16)
	*i = tAsn1Integer16(*ci)
	return i
}

func cAsn1Integer32(ci *C.Integer32_t, i *tAsn1Integer32) *C.Integer32_t {
	if nil == i {
		return nil
	}
	if nil == ci {
		ci = C.hlp__calloc_Integer32_t(1)
	}
	*ci = C.Integer32_t(*i)
	return ci
}

func goAsn1Integer32(ci *C.Integer32_t) *tAsn1Integer32 {
	if nil == ci {
		return nil
	}
	i := new(tAsn1Integer32)
	*i = tAsn1Integer32(*ci)
	return i
}

func cAsn1Long64(ci *C.OCTET_STRING_t, i *tAsn1Long64) *C.OCTET_STRING_t {
	if nil == i {
		return nil
	}

	var buf bytes.Buffer
	err := binary.Write(&buf, binary.BigEndian, i)
	if err != nil {
		panic(fmt.Sprintf("binary.Write() failed: %v", err))
	}
	ib := buf.Bytes()

	cib := cslice(ib)
	return C.hlp__fill_OCTET_STRING_t(ci, (*C.uint8_t)(&cib[0]), C.int(len(cib)))
}

func goAsn1Long64(ci *C.OCTET_STRING_t) *tAsn1Long64 {
	if nil == ci {
		return nil
	}

	i := new(tAsn1Long64)

	if 8 != ci.size {
		panic(fmt.Sprintf("goAsn1Long64(): size of long64 is not 8"))
	}
	b := goslice(ci.buf, ci.size)

	buf := bytes.NewBuffer(b)
	err := binary.Read(buf, binary.BigEndian, i)
	if err != nil {
		panic(fmt.Sprintf("binary.Read() failed: %v", err))
	}

	return i
}

func cAsn1Unsigned8(ci *C.Unsigned8_t, i *tAsn1Unsigned8) *C.Unsigned8_t {
	if nil == i {
		return nil
	}
	if nil == ci {
		ci = C.hlp__calloc_Unsigned8_t(1)
	}
	*ci = C.Unsigned8_t(*i)
	return ci
}

func goAsn1Unsigned8(ci *C.Unsigned8_t) *tAsn1Unsigned8 {
	if nil == ci {
		return nil
	}
	i := new(tAsn1Unsigned8)
	*i = tAsn1Unsigned8(*ci)
	return i
}

func cAsn1Unsigned16(ci *C.Unsigned16_t, i *tAsn1Unsigned16) *C.Unsigned16_t {
	if nil == i {
		return nil
	}
	if nil == ci {
		ci = C.hlp__calloc_Unsigned16_t(1)
	}
	*ci = C.Unsigned16_t(*i)
	return ci
}

func goAsn1Unsigned16(ci *C.Unsigned16_t) *tAsn1Unsigned16 {
	if nil == ci {
		return nil
	}
	i := new(tAsn1Unsigned16)
	*i = tAsn1Unsigned16(*ci)
	return i
}

func cAsn1Unsigned32(ci *C.Unsigned32_t, i *tAsn1Unsigned32) *C.Unsigned32_t {
	if nil == i {
		return nil
	}
	if nil == ci {
		ci = C.hlp__calloc_Unsigned32_t(1)
	}
	*ci = C.Unsigned32_t(*i)
	return ci
}

func goAsn1Unsigned32(ci *C.Unsigned32_t) *tAsn1Unsigned32 {
	if nil == ci {
		return nil
	}
	i := new(tAsn1Unsigned32)
	*i = tAsn1Unsigned32(*ci)
	return i
}

func cAsn1UnsignedLong64(ci *C.OCTET_STRING_t, i *tAsn1UnsignedLong64) *C.OCTET_STRING_t {
	if nil == i {
		return nil
	}

	var buf bytes.Buffer
	err := binary.Write(&buf, binary.BigEndian, i)
	if err != nil {
		panic(fmt.Sprintf("binary.Write() failed: %v", err))
	}
	ib := buf.Bytes()

	cib := cslice(ib)
	return C.hlp__fill_OCTET_STRING_t(ci, (*C.uint8_t)(&cib[0]), C.int(len(cib)))
}

func goAsn1UnsignedLong64(ci *C.OCTET_STRING_t) *tAsn1UnsignedLong64 {
	if nil == ci {
		return nil
	}

	i := new(tAsn1UnsignedLong64)

	if 8 != ci.size {
		panic(fmt.Sprintf("goAsn1UnsignedLong64(): size of long64 is not 8"))
	}
	b := goslice(ci.buf, ci.size)

	buf := bytes.NewBuffer(b)
	err := binary.Read(buf, binary.BigEndian, i)
	if err != nil {
		panic(fmt.Sprintf("binary.Read() failed: %v", err))
	}

	return i
}

func cAsn1Float(cf *C.OCTET_STRING_t, f *tAsn1Float) *C.OCTET_STRING_t {
	if nil == f {
		return nil
	}

	var buf bytes.Buffer
	err := binary.Write(&buf, binary.BigEndian, f)
	if err != nil {
		panic(fmt.Sprintf("binary.Write() failed: %v", err))
	}
	fb := buf.Bytes()

	cfb := cslice(fb)
	return C.hlp__fill_OCTET_STRING_t(cf, (*C.uint8_t)(&cfb[0]), C.int(len(cfb)))
}

func goAsn1Float(cf *C.OCTET_STRING_t) *tAsn1Float {
	if nil == cf {
		return nil
	}
	f := new(tAsn1Float)
	b := goslice(cf.buf, cf.size)

	buf := bytes.NewBuffer(b)
	err := binary.Read(buf, binary.BigEndian, f)
	if err != nil {
		panic(fmt.Sprintf("binary.Read() failed: %v", err))
	}

	return f
}

func cAsn1Float32(cf *C.OCTET_STRING_t, f *tAsn1Float32) *C.OCTET_STRING_t {
	if nil == f {
		return nil
	}

	var buf bytes.Buffer
	err := binary.Write(&buf, binary.BigEndian, f)
	if err != nil {
		panic(fmt.Sprintf("binary.Write() failed: %v", err))
	}
	fb := buf.Bytes()

	cfb := cslice(fb)
	return C.hlp__fill_OCTET_STRING_t(cf, (*C.uint8_t)(&cfb[0]), C.int(len(cfb)))
}

func goAsn1Float32(cf *C.OCTET_STRING_t) *tAsn1Float32 {
	if nil == cf {
		return nil
	}

	f := new(tAsn1Float32)
	if 4 != cf.size {
		panic(fmt.Sprintf("goAsn1Float32(): size of float is not 4"))
	}
	b := goslice(cf.buf, cf.size)

	buf := bytes.NewBuffer(b)
	err := binary.Read(buf, binary.BigEndian, f)
	if err != nil {
		panic(fmt.Sprintf("binary.Read() failed: %v", err))
	}

	return f
}

func cAsn1Float64(cf *C.OCTET_STRING_t, f *tAsn1Float64) *C.OCTET_STRING_t {
	if nil == f {
		return nil
	}

	var buf bytes.Buffer
	err := binary.Write(&buf, binary.BigEndian, f)
	if err != nil {
		panic(fmt.Sprintf("binary.Write() failed: %v", err))
	}
	fb := buf.Bytes()

	cfb := cslice(fb)
	return C.hlp__fill_OCTET_STRING_t(cf, (*C.uint8_t)(&cfb[0]), C.int(len(cfb)))
}

func goAsn1Float64(cf *C.OCTET_STRING_t) *tAsn1Float64 {
	if nil == cf {
		return nil
	}
	f := new(tAsn1Float64)
	if 8 != cf.size {
		panic(fmt.Sprintf("goAsn1Float64(): size of float is not 8"))
	}
	b := goslice(cf.buf, cf.size)

	buf := bytes.NewBuffer(b)
	err := binary.Read(buf, binary.BigEndian, f)
	if err != nil {
		panic(fmt.Sprintf("binary.Read() failed: %v", err))
	}

	return f
}

func cAsn1BitString(cbitString *C.BIT_STRING_t, bitString *tAsn1BitString) *C.BIT_STRING_t {

	if nil == bitString {
		return (*C.BIT_STRING_t)(unsafe.Pointer(nil))
	}

	cb := cslice(bitString.buf)
	bitsUnused := bitString.bitsUnused
	return C.hlp__fill_BIT_STRING_t(cbitString, &cb[0], C.int(len(cb)), C.int(bitsUnused))
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

func cAsn1OctetString(coctetString *C.OCTET_STRING_t, octetString *tAsn1OctetString) *C.OCTET_STRING_t {
	if nil == octetString {
		return (*C.OCTET_STRING_t)(unsafe.Pointer(nil))
	}
	cb := cslice(*octetString)
	return C.hlp__fill_OCTET_STRING_t(coctetString, &cb[0], C.int(len(cb)))
}

func goAsn1OctetString(cOctetString *C.OCTET_STRING_t) *tAsn1OctetString {
	if nil == cOctetString {
		return nil
	}

	octetString := new(tAsn1OctetString)
	*octetString = tAsn1OctetString(goslice(cOctetString.buf, cOctetString.size))
	return octetString
}

func cAsn1GraphicString(cgs *C.GraphicString_t, gs *tAsn1GraphicString) *C.GraphicString_t {
	if nil == gs {
		return (*C.GraphicString_t)(unsafe.Pointer(nil))
	}
	cb := cslice(*gs)
	return C.hlp__fill_OCTET_STRING_t(cgs, &cb[0], C.int(len(cb)))
}

func goAsn1GraphicString(cGraphicString *C.GraphicString_t) *tAsn1GraphicString {
	if nil == cGraphicString {
		return nil
	}
	graphicString := new(tAsn1GraphicString)
	*graphicString = tAsn1GraphicString(goslice(cGraphicString.buf, cGraphicString.size))
	return graphicString
}

func cAsn1VisibleString(cvs *C.VisibleString_t, vs *tAsn1VisibleString) *C.VisibleString_t {
	if nil == vs {
		return (*C.VisibleString_t)(unsafe.Pointer(nil))
	}
	cb := cslice(*vs)
	return C.hlp__fill_OCTET_STRING_t(cvs, &cb[0], C.int(len(cb)))
}

func goAsn1VisibleString(cvs *C.VisibleString_t) *tAsn1VisibleString {
	if nil == cvs {
		return nil
	}
	vs := new(tAsn1VisibleString)
	*vs = tAsn1VisibleString(goslice(cvs.buf, cvs.size))
	return vs
}

func cAsn1ObjectIdentifier(coid *C.OBJECT_IDENTIFIER_t, oid *tAsn1ObjectIdentifier) *C.OBJECT_IDENTIFIER_t {
	if nil == oid {
		return (*C.OBJECT_IDENTIFIER_t)(unsafe.Pointer(nil))
	}
	if nil == coid {
		coid = C.hlp__calloc_OBJECT_IDENTIFIER_t()
	}

	length := len(*oid)
	cb := (*[1 << 26]C.uint32_t)(unsafe.Pointer(C.hlp__calloc(C.size_t(length), 4)))[:length:length]
	for i := 0; i < length; i++ {
		cb[i] = C.uint32_t((*oid)[i])
	}
	ret := C.OBJECT_IDENTIFIER_set_arcs(coid, (*C.asn_oid_arc_t)(unsafe.Pointer(&cb[0])), C.size_t(length))
	if -1 == ret {
		panic("cAsn1ObjectIdentifier(): cannot encode oid")
	}
	C.hlp__free(unsafe.Pointer(&cb[0]))
	return coid
}

func goAsn1ObjectIdentifier(cOid *C.OBJECT_IDENTIFIER_t) *tAsn1ObjectIdentifier {
	if nil == cOid {
		return nil
	}

	length := C.int(20)
	cb := (*[1 << 26]C.uint32_t)(unsafe.Pointer(C.hlp__calloc(C.size_t(length), 4)))[:length:length]
	ret := C.OBJECT_IDENTIFIER_get_arcs(cOid, (*C.asn_oid_arc_t)(unsafe.Pointer(&cb[0])), C.size_t(length))
	if -1 == ret {
		panic("goAsn1ObjectIdentifier(): cannot decode oid")
	}
	n := int(ret)
	b := make([]uint32, n)
	for i := 0; i < n; i++ {
		b[i] = uint32(cb[i])
	}
	C.hlp__free(unsafe.Pointer(&cb[0]))
	return (*tAsn1ObjectIdentifier)(&b)

}

func cAsn1Any(cany *C.ANY_t, any *tAsn1Any) *C.ANY_t {
	if nil == any {
		return (*C.ANY_t)(unsafe.Pointer(nil))
	}
	cb := cslice(*any)
	return C.hlp__fill_ANY_t(cany, &cb[0], C.int(len(cb)))
}

func goAsn1Any(cAny *C.ANY_t) *tAsn1Any {
	if nil == cAny {
		return nil
	}

	any := new(tAsn1Any)
	*any = goslice(cAny.buf, C.size_t(cAny.size))
	return any
}

func cAsn1Null(cnull *C.NULL_t) *C.NULL_t {
	if nil == cnull {
		cnull = C.hlp__calloc_NULL_t()
	}
	*cnull = C.NULL_t(0)
	return cnull
}

func goAsn1Null() *tAsn1Null {
	null := new(tAsn1Null)
	(*null) = 0
	return null
}

func cAsn1Boolean(cb *C.BOOLEAN_t, b *tAsn1Boolean) *C.BOOLEAN_t {
	if nil == b {
		return (*C.BOOLEAN_t)(unsafe.Pointer(nil))
	}
	if nil == cb {
		cb = C.hlp__calloc_BOOLEAN_t()
	}
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

func cAsn1DateTime(cd *C.OCTET_STRING_t, d *tAsn1DateTime) *C.OCTET_STRING_t {
	if nil == d {
		return nil
	}
	cb := cslice(*d)
	return C.hlp__fill_OCTET_STRING_t(cd, (*C.uint8_t)(&cb[0]), C.int(len(cb)))
}

func goAsn1DateTime(d *C.OCTET_STRING_t) *tAsn1DateTime {
	if nil == d {
		return nil
	}
	cb := goslice(d.buf, d.size)
	ad := make([]byte, int(d.size))
	copy(ad, cb)
	return (*tAsn1DateTime)(&ad)
}

func cAsn1Date(cd *C.OCTET_STRING_t, d *tAsn1Date) *C.OCTET_STRING_t {
	if nil == d {
		return nil
	}
	cb := cslice(*d)
	return C.hlp__fill_OCTET_STRING_t(cd, (*C.uint8_t)(&cb[0]), C.int(len(cb)))
}

func goAsn1Date(d *C.OCTET_STRING_t) *tAsn1Date {
	if nil == d {
		return nil
	}
	cb := goslice(d.buf, d.size)
	ad := make([]byte, int(d.size))
	copy(ad, cb)
	return (*tAsn1Date)(&ad)
}

func cAsn1Time(cd *C.OCTET_STRING_t, d *tAsn1Time) *C.OCTET_STRING_t {
	if nil == d {
		return nil
	}
	cb := cslice(*d)
	return C.hlp__fill_OCTET_STRING_t(cd, (*C.uint8_t)(&cb[0]), C.int(len(cb)))
}

func goAsn1Time(d *C.OCTET_STRING_t) *tAsn1Time {
	if nil == d {
		return nil
	}
	cb := goslice(d.buf, d.size)
	ad := make([]byte, int(d.size))
	copy(ad, cb)
	return (*tAsn1Time)(&ad)
}

func encode_AARQapdu(_pdu *AARQapdu) (err error, result []byte) {
	var ret C.asn_enc_rval_t
	var pdu *C.AARQ_apdu_t
	var buf bytes.Buffer
	var cb []C.uint8_t

	pdu = C.hlp__calloc_AARQ_apdu_t()

	// protocol_version
	pdu.protocol_version = cAsn1BitString(pdu.protocol_version, _pdu.protocolVersion)

	// application_context_name
	cAsn1ObjectIdentifier(&pdu.application_context_name, &_pdu.applicationContextName)

	// called_AP_title
	if nil != _pdu.calledAPtitle {
		pdu.called_AP_title = cAsn1OctetString(pdu.called_AP_title, _pdu.calledAPtitle)
	}

	// called_AE_qualifier
	if nil != _pdu.calledAEqualifier {
		pdu.called_AE_qualifier = cAsn1OctetString(pdu.called_AE_qualifier, _pdu.calledAEqualifier)
	}

	// called_AP_invocation_id
	if nil != _pdu.calledAPinvocationId {
		pdu.called_AP_invocation_id = (*C.AP_invocation_identifier_t)(cAsn1Integer((*C.long)(pdu.called_AP_invocation_id), _pdu.calledAPinvocationId))
	}

	// called_AE_invocation_id
	if nil != _pdu.calledAEinvocationId {
		pdu.called_AE_invocation_id = (*C.AE_invocation_identifier_t)(cAsn1Integer((*C.long)(pdu.called_AE_invocation_id), _pdu.calledAEinvocationId))
	}

	// calling_AP_title
	if nil != _pdu.callingAPtitle {
		pdu.calling_AP_title = cAsn1OctetString(pdu.calling_AP_title, _pdu.callingAPtitle)
	}

	// calling_AE_qualifier
	if nil != _pdu.callingAEqualifier {
		pdu.calling_AE_qualifier = cAsn1OctetString(pdu.calling_AE_qualifier, _pdu.callingAEqualifier)
	}

	// calling_AP_invocation_id
	if nil != _pdu.callingAPinvocationId {
		pdu.calling_AP_invocation_id = (*C.AP_invocation_identifier_t)(cAsn1Integer((*C.long)(pdu.calling_AP_invocation_id), _pdu.callingAPinvocationId))
	}

	// calling_AE_invocation_id
	if nil != _pdu.callingAEinvocationId {
		pdu.calling_AE_invocation_id = (*C.AE_invocation_identifier_t)(cAsn1Integer((*C.long)(pdu.calling_AE_invocation_id), _pdu.callingAEinvocationId))
	}

	// sender_acse_requirements
	if nil != _pdu.senderAcseRequirements {
		pdu.sender_acse_requirements = cAsn1BitString(pdu.sender_acse_requirements, _pdu.senderAcseRequirements)

	}

	// mechanism_name
	if nil != _pdu.mechanismName {
		pdu.mechanism_name = cAsn1ObjectIdentifier(pdu.mechanism_name, _pdu.mechanismName)
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

			cCharstring := cAsn1GraphicString(nil, charstring)
			*(*C.GraphicString_t)(unsafe.Pointer(&(*av).choice[0])) = *cCharstring
			C.hlp__free(unsafe.Pointer(cCharstring))

		case C.Authentication_value_PR_bitstring:

			bistring := ((*_av).getVal()).(*tAsn1BitString)
			(*av).present = C.Authentication_value_PR_bitstring

			cBistring := cAsn1BitString(nil, bistring)
			*(*C.BIT_STRING_t)(unsafe.Pointer(&(*av).choice[0])) = *cBistring
			C.hlp__free(unsafe.Pointer(cBistring))

		case C.Authentication_value_PR_external:

			external := ((*_av).getVal()).(*tAsn1OctetString)
			(*av).present = C.Authentication_value_PR_external

			cExternal := cAsn1OctetString(nil, external)
			*(*C.OCTET_STRING_t)(unsafe.Pointer(&(*av).choice[0])) = *cExternal
			C.hlp__free(unsafe.Pointer(cExternal))

		case C.Authentication_value_PR_other:

			var avo C.struct_Authentication_value_other
			other := ((*_av).getVal()).(*tAuthenticationValueOther)

			cAsn1ObjectIdentifier(&avo.other_mechanism_name, &other.otherMechanismName)
			cAsn1Any(&avo.other_mechanism_value, &other.otherMechanismValue)

			(*av).present = C.Authentication_value_PR_other
			cb = cslicep((uintptr)(unsafe.Pointer(&avo)), unsafe.Sizeof(avo))
			C.hlp__fill_OCTET_STRING_t((*C.OCTET_STRING_t)(unsafe.Pointer(&(*av).choice[0])), &cb[0], C.int(len(cb)))

		default:
			err = fmt.Errorf("encode_AARQapdu() failed, unknown callingAuthenticationValue tag %v", _pdu.callingAuthenticationValue.getTag())
			errorLog("%s", err)
			return err, nil
		}
	}

	//implementation-information [29] IMPLICIT Implementation-data OPTIONAL,
	if nil != _pdu.implementationInformation {
		pdu.implementation_information = cAsn1GraphicString(pdu.implementation_information, _pdu.implementationInformation)
	}

	//user-information [30] EXPLICIT Association-information OPTIONAL
	if nil != _pdu.userInformation {
		pdu.user_information = cAsn1OctetString(pdu.user_information, _pdu.userInformation)
	}

	ret, errno := C.der_encode(&C.asn_DEF_AARQ_apdu, unsafe.Pointer(pdu), (*C.asn_app_consume_bytes_f)(C.consumeBytesWrap), unsafe.Pointer(&buf))
	if -1 == ret.encoded {
		C.hlp__free_AARQ_apdu_t(pdu)
		s := C.GoString(ret.failed_type.name)
		err = fmt.Errorf("C.der_encode() failed, failed type name: %v, errno: %v", s, errno)
		errorLog("%s", err)
		return err, nil
	}
	C.hlp__free_AARQ_apdu_t(pdu)
	return nil, buf.Bytes()
}

func decode_AAREapdu(inb []byte) (err error, pdu *AAREapdu) {

	var cpdu *C.AARE_apdu_t

	pdu = new(AAREapdu)

	cb := cslice(inb)
	ret, errno := C.ber_decode((*C.struct_asn_codec_ctx_s)(unsafe.Pointer(nil)), &C.asn_DEF_AARE_apdu, (*unsafe.Pointer)(unsafe.Pointer(&cpdu)), unsafe.Pointer(&cb[0]), C.size_t(len(cb)))
	C.hlp__free(unsafe.Pointer(&cb[0]))
	if C.RC_OK != ret.code {
		err = fmt.Errorf("C.ber_decode() failed, code: %v, consumed: %v, errno %v", ret.code, ret.consumed, errno)
		errorLog("%s", err)
		return err, nil
	}

	//-- [APPLICATION 1] == [ 61H ] = [ 97 ]
	//protocol-version [0] IMPLICIT T-protocol-version DEFAULT {version1},
	//protocolVersion *tAsn1BitString
	pdu.protocolVersion = goAsn1BitString(cpdu.protocol_version)

	//application-context-name [1] Application-context-name,
	//applicationContextName tAsn1ObjectIdentifier
	pdu.applicationContextName = *goAsn1ObjectIdentifier(&cpdu.application_context_name)

	//result [2] Association-result,
	//result tAsn1Integer
	pdu.result = *goAsn1Integer((*C.long)(&cpdu.result))

	//result-source-diagnostic [3] Associate-source-diagnostic,
	//resultSourceDiagnostic tAsn1Choice
	b := cpdu.result_source_diagnostic.choice
	switch cpdu.result_source_diagnostic.present {
	case C.Associate_source_diagnostic_PR_NOTHING:
		pdu.resultSourceDiagnostic.setVal(int(C.Associate_source_diagnostic_PR_NOTHING), nil)
	case C.Associate_source_diagnostic_PR_acse_service_user:
		pdu.resultSourceDiagnostic.setVal(int(C.Associate_source_diagnostic_PR_acse_service_user), int(*(*C.long)(unsafe.Pointer(&b[0]))))
	case C.Associate_source_diagnostic_PR_acse_service_provider:
		pdu.resultSourceDiagnostic.setVal(int(C.Associate_source_diagnostic_PR_acse_service_provider), int(*(*C.long)(unsafe.Pointer(&b[0]))))
	default:
		err = fmt.Errorf("decode_AAREapdu(): unknown choice tag: %v", int(cpdu.result_source_diagnostic.present))
		errorLog("%s", err)
		return err, nil
	}

	//responding-AP-title [4] AP-title OPTIONAL,
	//respondingAPtitle *tAsn1OctetString
	pdu.respondingAPtitle = goAsn1OctetString(cpdu.responding_AP_title)

	//responding-AE-qualifier [5] AE-qualifier OPTIONAL,
	//respondingAEqualifier *tAsn1OctetString
	pdu.respondingAEqualifier = goAsn1OctetString(cpdu.responding_AE_qualifier)

	//responding-AP-invocation-id [6] AP-invocation-identifier OPTIONAL,
	//respondingAPinvocationId *tAsn1Integer
	pdu.respondingAPinvocationId = goAsn1Integer((*C.long)(cpdu.responding_AP_invocation_id))

	//responding-AE-invocation-id [7] AE-invocation-identifier OPTIONAL,
	//respondingAEinvocationId *tAsn1Integer
	pdu.respondingAEinvocationId = goAsn1Integer((*C.long)(cpdu.responding_AE_invocation_id))

	//-- The following field shall not be present if only the kernel is used.
	//responder-acse-requirements [8] IMPLICIT ACSE-requirements OPTIONAL,
	//responderAcseRequirements *tAsn1BitString
	pdu.responderAcseRequirements = goAsn1BitString(cpdu.responder_acse_requirements)

	//-- The following field shall only be present if the authentication functional unit is selected.
	//mechanism-name [9] IMPLICIT Mechanism-name OPTIONAL,
	//mechanismName *tAsn1ObjectIdentifier
	pdu.mechanismName = goAsn1ObjectIdentifier(cpdu.mechanism_name)

	//-- The following field shall only be present if the authentication functional unit is selected.
	//responding-authentication-value [10] EXPLICIT Authentication-value OPTIONAL,
	//respondingAuthenticationValue *tAsn1Choice
	if nil != cpdu.responding_authentication_value {
		pdu.respondingAuthenticationValue = new(tAsn1Choice)
		b := cpdu.responding_authentication_value.choice
		switch cpdu.responding_authentication_value.present {
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
			err = fmt.Errorf("decode_AAREapdu(): unknown choice tag: %v", int(cpdu.responding_authentication_value.present))
			errorLog("%s", err)
			return err, nil
		}
	}

	//implementation-information [29] IMPLICIT Implementation-data OPTIONAL,
	//implementationInformation *tAsn1GraphicString
	pdu.implementationInformation = goAsn1GraphicString(cpdu.implementation_information)

	//user-information [30] EXPLICIT Association-information OPTIONAL
	//userInformation *tAsn1OctetString
	pdu.userInformation = goAsn1OctetString(cpdu.user_information)

	C.hlp__free_AARE_apdu_t(cpdu)

	return nil, pdu
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
}

func encode_uint32_base128(w io.Writer, val uint32) (err error) {

	b := make([]uint8, 1)

	if val <= uint32(0x7f) {
		b[0] = uint8(val & 0x7f)
		n, err := w.Write(b)
		if nil != err {
			errorLog("io.Write(): %v", err)
			return err
		}
	} else if val <= uint32(0x3fff) {
		b[0] = uint8((val&0x3f80)>>7) | 0x10
		n, err := w.Write(b)
		if nil != err {
			errorLog("io.Write(): %v", err)
			return err
		}
		b[0] = uint8(val & 0x7f)
		n, err = w.Write(b)
		if nil != err {
			errorLog("io.Write(): %v", err)
			return err
		}
	} else if val <= uint32(0x1fffff) {
		b[0] = uint8((val&0x1fc000)>>14) | 0x10
		n, err := w.Write(b)
		if nil != err {
			errorLog("io.Write(): %v", err)
			return err
		}
		b[0] = uint8((val&0x3f80)>>7) | 0x10
		n, err = w.Write(b)
		if nil != err {
			errorLog("io.Write(): %v", err)
			return err
		}
		b[0] = uint8(val & 0x7f)
		n, err = w.Write(b)
		if nil != err {
			errorLog("io.Write(): %v", err)
			return err
		}
	} else if val <= uint32(0x0fffffff) {
		b[0] = uint8((val&0xfe00000)>>21) | 0x10
		n, err := w.Write(b)
		if nil != err {
			errorLog("io.Write(): %v", err)
			return err
		}
		b[0] = uint8((val&0x1fc000)>>14) | 0x10
		n, err = w.Write(b)
		if nil != err {
			errorLog("io.Write(): %v", err)
			return err
		}
		b[0] = uint8((val&0x3f80)>>7) | 0x10
		n, err = w.Write(b)
		if nil != err {
			errorLog("io.Write(): %v", err)
			return err
		}
		b[0] = uint8(val & 0x7f)
		n, err = w.Write(b)
		if nil != err {
			errorLog("io.Write(): %v", err)
			return err
		}
	} else if val <= uint32(0xffffffff) {
		b[0] = uint8(val&0x00000000>>28) | 0x10
		n, err := w.Write(b)
		if nil != err {
			errorLog("io.Write(): %v", err)
			return err
		}
		b[0] = uint8(val&0xfe00000>>21) | 0x10
		n, err = w.Write(b)
		if nil != err {
			errorLog("io.Write(): %v", err)
			return err
		}
		b[0] = uint8(val&0x1fc000>>14) | 0x10
		n, err = w.Write(b)
		if nil != err {
			errorLog("io.Write(): %v", err)
			return err
		}
		b[0] = uint8(val&0x3f80>>7) | 0x10
		n, err = w.Write(b)
		if nil != err {
			errorLog("io.Write(): %v", err)
			return err
		}
		b[0] = uint8(val & 0x7f)
		n, err = w.Write(b)
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

func decode_uint32_base128(r io.Reader) (err error, _val uint32) {

	var val uint32
	b := make([]byte, 1)

	n, err := r.Read(b)
	if nil != err {
		errorLog("io.Read(): %v", err)
		return err, 0
	}
	val = uint32(0x7f & b[0])
	if 1 == b[0]&0x80 {
		n, err := r.Read(b)
		if nil != err {
			errorLog("io.Read(): %v", err)
			return err, 0
		}
		val = (val << 7) | uint32(0x7f&b[0])
		if 1 == uint32(b[0]&0x80) {
			n, err := r.Read(b)
			if nil != err {
				errorLog("io.Read(): %v", err)
				return err, 0
			}
			val = (val << 7) | uint32(0x7f&b[0])
			if 1 == b[0]&0x80 {
				n, err := r.Read(b)
				if nil != err {
					errorLog("io.Read(): %v", err)
					return err, 0
				}
				val = (val << 7) | uint32(0x7f&b[0])
				if 1 == b[0]&0x80 {
					n, err := r.Read(b)
					if nil != err {
						errorLog("io.Read(): %v", err)
						return err, 0
					}
					if b[0] > 0x0f {
						err = fmt.Errorf("value of tag exceeds limit: %v", math.MaxUint32)
						errorLog("%s", err)
						return err, 0
					}
					val = (val << 4) | uint32(0x7f&b[0])
				}
			}
		}
	}
	return nil, val
}

func der_encode_Integer(i tAsn1Integer) []uint8 {

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

	content := make([]uint8, n)
	for j := 0; j < n; j++ {
		content[j] = uint8(_i >> uint((n-1-j)*8))
	}

	return content

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

func der_encode_BitString(bs *tAsn1BitString) []uint8 {
	var content []uint8

	if nil == bs || nil == bs.buf {
		content = make([]uint8, 0)
		return content
	}

	if 0 == len(bs.buf) {
		content = make([]uint8, 1)
		content[0] = 0
		return content
	}

	content = make([]uint8, len(bs.buf)+1)
	if bs.bitsUnused > 7 {
		panic("wrong count of unused bits")
	}
	content[0] = uint8(bs.bitsUnused)
	for i := 0; i < len(bs.buf); i++ {
		content[i+1] = bs.buf[i]
	}
	return content
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

func der_encode_ObjectIdentifier(oi *tAsn1ObjectIdentifier) []uint8 {
	var content []uint8

	if nil == oi {
		content = make([]uint8, 0)
		return content
	}

	_oi := ([]uint32)(*oi)

	if 0 == len(_oi) {
		content = make([]uint8, 0)
		return content
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
	err := encode_uint32_base128(&buf, 40*_oi[0]+_oi[1])
	if nil != err {
		panic(err)
	}
	for i := 2; i < len(_oi); i++ {
		err := encode_uint32_base128(&buf, 40*_oi[0]+_oi[1])
		if nil != err {
			panic(err)
		}
	}

	content = buf.Bytes()

	return content
}

func der_decode_ObjectIdentifier(content []uint8) (err error, oi *tAsn1ObjectIdentifier) {

	if len(content) < 1 {
		return nil, nil
	}

	buf := bytes.NewBuffer(content)
	COMPONENTS_BUFFER_SIZE := 100
	components := make([]uint32, COMPONENTS_BUFFER_SIZE)
	i := 0
	for i := 0; i < len(components); i++ {
		err, component := decode_uint32_base128(buf)
		if err == io.EOF {
			break
		}
		if nil != err {
			panic(err)
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
			panic("COMPONENTS_BUFFER_SIZE small")
		}
	}

	return nil, (*tAsn1ObjectIdentifier)(&components)
}

func der_encode_chunk(w io.Writer, ch *t_der_chunk) (err error) {
	b := make([]byte, 1)

	// class

	b[0] = ch.asn1_class << 6

	// encoding

	b[0] |= ch.encoding << 5

	// tag

	if ch.asn1_tag < 31 {
		b[0] |= uint8(ch.asn1_tag)
		n, err := w.Write(b)
		if nil != err {
			errorLog("io.Write(): %v", err)
			return err
		}
	} else {
		b[0] |= 0x1f
		n, err := w.Write(b)
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
		n, err := w.Write(b)
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
			panic()
		}
		b[0] = uint8(m) | 0x10
		n, err := w.Write(b)
		if nil != err {
			errorLog("io.Write(): %v", err)
			return err
		}
		for i := 0; i < m; i++ {
			b[0] = uint8((length >> uint(i)) & 0xff)
			n, err := w.Write(b)
			if nil != err {
				errorLog("io.Write(): %v", err)
				return err
			}
		}
	}

	return err
}

func der_decode_chunk(r io.Reader) (err error, _ch *t_der_chunk) {
	var n, m int
	var ch t_der_chunk

	b := make([]byte, 1)
	n, err = r.Read(b)
	if nil != err {
		errorLog("io.Read(): %v", err)
		return err, nil
	}

	// class

	ch.asn1_class = (b[0] & 0xc0) >> 6

	// encoding

	ch.encoding = (b[0] & 0x20) >> 5

	// tag

	if 0x1f == b[0] {
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

	n, err = r.Read(b)
	if nil != err {
		errorLog("io.Read(): %v", err)
		return err, nil
	}
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
		n, err := r.Read(b)
		if nil != err {
			errorLog("io.Read(): %v", err)
			return err, nil
		}
		length = uint64(b[0])
		for i := 0; i < m-1; m-- {
			n, err := r.Read(b)
			if nil != err {
				errorLog("io.Read(): %v", err)
				return err, nil
			}
			length = (length << 8) | uint64(b[0])
		}
	}

	if length > 1024 { // guard against allocating too much
		err = fmt.Errorf("content too log: %v bytes", length)
		errorLog("%s", err)
		return err, nil
	}
	ch.content = make([]byte, length)
	n, err = r.Read(b)
	if nil != err {
		errorLog("io.Read(): %v", err)
		return err, nil
	}

	return nil, &ch

}

func encode_AARQapdu_1(w io.Writer, aarq *AARQapdu) (err error) {

	// protocol-version [0] IMPLICIT T-protocol-version DEFAULT {version1}

	var protocolVersion *t_der_chunk
	if nil != aarq.protocolVersion {
		ch := new(t_der_chunk)
		ch.asn1_class = ASN1_CLASS_CONTEXT_SPECIFIC
		ch.encoding = BER_ENCODING_PRIMITIVE
		ch.asn1_tag = 0
		ch.content = der_encode_BitString(aarq.protocolVersion)
		protocolVersion = ch
	}

	// application-context-name [1] Application-context-name,

	var applicationContextName *t_der_chunk
	if nil != aarq.applicationContextName {
		ch := new(t_der_chunk)
		ch.asn1_class = ASN1_CLASS_CONTEXT_SPECIFIC
		ch.encoding = BER_ENCODING_CONSTRUCTED
		ch.asn1_tag = 1

		ch1 := new(t_der_chunk)
		ch1.asn1_class = ASN1_CLASS_UNIVERSAL
		ch1.encoding = BER_ENCODING_PRIMITIVE
		ch1.asn1_tag = 6
		ch1.content = der_encode_ObjectIdentifier(&aarq.applicationContextName)
		var buf bytes.Buffer
		err = der_encode_chunk(&buf, ch1)
		if nil != err {
			return err
		}

		ch.content = buf.Bytes()
		applicationContextName = ch
	}

	// called-AP-title [2] AP-title OPTIONAL,

	var calledAPtitle *t_der_chunk
	if nil != aarq.calledAPtitle {
		ch := new(t_der_chunk)
		ch.asn1_class = ASN1_CLASS_CONTEXT_SPECIFIC
		ch.encoding = BER_ENCODING_CONSTRUCTED
		ch.asn1_tag = 2

		ch1 := new(t_der_chunk)
		ch1.asn1_class = ASN1_CLASS_UNIVERSAL
		ch1.encoding = BER_ENCODING_PRIMITIVE
		ch1.asn1_tag = 4
		octetString := ([]uint8)(*aarq.calledAPtitle)
		ch1.content = make([]uint8, len(octetString))
		copy(ch1.content, octetString)
		var buf bytes.Buffer
		err = der_encode_chunk(&buf, ch1)
		if nil != err {
			return err
		}

		ch.content = buf.Bytes()
		calledAPtitle = ch
	}

	// called-AE-qualifier [3] AE-qualifier OPTIONAL,

	var calledAEqualifier *t_der_chunk
	if nil != aarq.calledAEqualifier {
		ch := new(t_der_chunk)
		ch.asn1_class = ASN1_CLASS_CONTEXT_SPECIFIC
		ch.encoding = BER_ENCODING_CONSTRUCTED
		ch.asn1_tag = 3

		ch1 := new(t_der_chunk)
		ch1.asn1_class = ASN1_CLASS_UNIVERSAL
		ch1.encoding = BER_ENCODING_PRIMITIVE
		ch1.asn1_tag = 4
		octetString := ([]uint8)(*aarq.calledAEqualifier)
		ch1.content = make([]uint8, len(octetString))
		copy(ch1.content, octetString)
		var buf bytes.Buffer
		err = der_encode_chunk(&buf, ch1)
		if nil != err {
			return err
		}

		ch.content = buf.Bytes()
		calledAEqualifier = ch
	}

	// called-AP-invocation-id [4] AP-invocation-identifier OPTIONAL,

	var calledAPinvocationId *t_der_chunk
	if nil != aarq.calledAPinvocationId {
		ch := new(t_der_chunk)
		ch.asn1_class = ASN1_CLASS_CONTEXT_SPECIFIC
		ch.encoding = BER_ENCODING_CONSTRUCTED
		ch.asn1_tag = 4

		ch1 := new(t_der_chunk)
		ch1.asn1_class = ASN1_CLASS_UNIVERSAL
		ch1.encoding = BER_ENCODING_PRIMITIVE
		ch1.asn1_tag = 2
		ch1.content = der_encode_Integer(*aarq.calledAPinvocationId)
		var buf bytes.Buffer
		err = der_encode_chunk(&buf, ch1)
		if nil != err {
			return err
		}

		ch.content = buf.Bytes()
		calledAPinvocationId = ch
	}

	// called-AE-invocation-id [5] AE-invocation-identifier OPTIONAL,

	var calledAEinvocationId *t_der_chunk
	if nil != aarq.calledAEinvocationId {
		ch := new(t_der_chunk)
		ch.asn1_class = ASN1_CLASS_CONTEXT_SPECIFIC
		ch.encoding = BER_ENCODING_CONSTRUCTED
		ch.asn1_tag = 5

		ch1 := new(t_der_chunk)
		ch1.asn1_class = ASN1_CLASS_UNIVERSAL
		ch1.encoding = BER_ENCODING_PRIMITIVE
		ch1.asn1_tag = 2
		ch1.content = der_encode_Integer(*aarq.calledAEinvocationId)
		var buf bytes.Buffer
		err = der_encode_chunk(&buf, ch1)
		if nil != err {
			return err
		}

		ch.content = buf.Bytes()
		calledAEinvocationId = ch
	}

	// calling-AP-title [6] AP-title OPTIONAL,

	var callingAPtitle *t_der_chunk
	if nil != aarq.callingAPtitle {
		ch := new(t_der_chunk)
		ch.asn1_class = ASN1_CLASS_CONTEXT_SPECIFIC
		ch.encoding = BER_ENCODING_CONSTRUCTED
		ch.asn1_tag = 6

		ch1 := new(t_der_chunk)
		ch1.asn1_class = ASN1_CLASS_UNIVERSAL
		ch1.encoding = BER_ENCODING_PRIMITIVE
		ch1.asn1_tag = 4
		octetString := ([]uint8)(*aarq.callingAPtitle)
		ch1.content = make([]uint8, len(octetString))
		copy(ch1.content, octetString)
		var buf bytes.Buffer
		err = der_encode_chunk(&buf, ch1)
		if nil != err {
			return err
		}

		ch.content = buf.Bytes()
		callingAPtitle = ch
	}

	// calling-AE-qualifier [7] AE-qualifier OPTIONAL,

	var callingAEqualifier *t_der_chunk
	if nil != aarq.callingAEqualifier {
		ch := new(t_der_chunk)
		ch.asn1_class = ASN1_CLASS_CONTEXT_SPECIFIC
		ch.encoding = BER_ENCODING_CONSTRUCTED
		ch.asn1_tag = 7

		ch1 := new(t_der_chunk)
		ch1.asn1_class = ASN1_CLASS_UNIVERSAL
		ch1.encoding = BER_ENCODING_PRIMITIVE
		ch1.asn1_tag = 4
		octetString := ([]uint8)(*aarq.callingAEqualifier)
		ch1.content = make([]uint8, len(octetString))
		copy(ch1.content, octetString)
		var buf bytes.Buffer
		err = der_encode_chunk(&buf, ch1)
		if nil != err {
			return err
		}

		ch.content = buf.Bytes()
		callingAEqualifier = ch
	}

	// calling-AP-invocation-id [8] AP-invocation-identifier OPTIONAL,

	var callingAPinvocationId *t_der_chunk
	if nil != aarq.callingAPinvocationId {
		ch := new(t_der_chunk)
		ch.asn1_class = ASN1_CLASS_CONTEXT_SPECIFIC
		ch.encoding = BER_ENCODING_CONSTRUCTED
		ch.asn1_tag = 8

		ch1 := new(t_der_chunk)
		ch1.asn1_class = ASN1_CLASS_UNIVERSAL
		ch1.encoding = BER_ENCODING_PRIMITIVE
		ch1.asn1_tag = 2
		ch1.content = der_encode_Integer(*aarq.callingAPinvocationId)
		var buf bytes.Buffer
		err = der_encode_chunk(&buf, ch1)
		if nil != err {
			return err
		}

		ch.content = buf.Bytes()
		callingAPinvocationId = ch
	}

	// calling-AE-invocation-id [9] AE-invocation-identifier OPTIONAL,

	var callingAEinvocationId *t_der_chunk
	if nil != aarq.callingAEinvocationId {
		ch := new(t_der_chunk)
		ch.asn1_class = ASN1_CLASS_CONTEXT_SPECIFIC
		ch.encoding = BER_ENCODING_CONSTRUCTED
		ch.asn1_tag = 9

		ch1 := new(t_der_chunk)
		ch1.asn1_class = ASN1_CLASS_UNIVERSAL
		ch1.encoding = BER_ENCODING_PRIMITIVE
		ch1.asn1_tag = 2
		ch1.content = der_encode_Integer(*aarq.callingAEinvocationId)
		var buf bytes.Buffer
		err = der_encode_chunk(&buf, ch1)
		if nil != err {
			return err
		}

		ch.content = buf.Bytes()
		callingAEinvocationId = ch
	}

	// sender-acse-requirements [10] IMPLICIT ACSE-requirements OPTIONAL,

	var senderAcseRequirements *t_der_chunk
	if nil != aarq.senderAcseRequirements {
		ch := new(t_der_chunk)
		ch.asn1_class = ASN1_CLASS_CONTEXT_SPECIFIC
		ch.encoding = BER_ENCODING_PRIMITIVE
		ch.asn1_tag = 10
		ch.content = der_encode_BitString(aarq.senderAcseRequirements)

		senderAcseRequirements = ch
	}

	// mechanism-name [11] IMPLICIT Mechanism-name OPTIONAL,
	var mechanismName *t_der_chunk
	if nil != aarq.mechanismName {
		ch := new(t_der_chunk)
		ch.asn1_class = ASN1_CLASS_CONTEXT_SPECIFIC
		ch.encoding = BER_ENCODING_PRIMITIVE
		ch.asn1_tag = 11
		ch.content = der_encode_ObjectIdentifier(aarq.mechanismName)

		mechanismName = ch
	}

	// calling-authentication-value [12] EXPLICIT Authentication-value OPTIONAL,

	/*
	   type tAsn1Choice struct {
	   	tag int
	   	val interface{}
	   }
	*/
	var callingAuthenticationValue *t_der_chunk
	if nil != aarq.callingAuthenticationValue {
		ch := new(t_der_chunk)
		ch.asn1_class = ASN1_CLASS_CONTEXT_SPECIFIC
		ch.encoding = BER_ENCODING_CONSTRUCTED
		ch.asn1_tag = 12

		ch1 := new(t_der_chunk)

		if 0 == aarq.callingAuthenticationValue.tag {
			// charstring [0] IMPLICIT GraphicString,
			ch1.asn1_class = ASN1_CLASS_CONTEXT_SPECIFIC
			ch1.encoding = BER_ENCODING_PRIMITIVE
			ch1.asn1_tag = 0
			octetString := aarq.callingAuthenticationValue.val.([]uint8)
			ch1.content = make([]uint8, len(octetString))
			copy(ch1.content, octetString)
		} else if 1 == aarq.callingAuthenticationValue.tag {
			// bitstring [1] IMPLICIT BIT STRING,
			ch1.asn1_class = ASN1_CLASS_CONTEXT_SPECIFIC
			ch1.encoding = BER_ENCODING_PRIMITIVE
			ch1.asn1_tag = 1
			bitString := aarq.callingAuthenticationValue.val.(tAsn1BitString)
			ch1.content = der_encode_BitString(&bitString)
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

			ch2 := new(t_der_chunk)
			ch2.asn1_class = ASN1_CLASS_UNIVERSAL
			ch2.encoding = BER_ENCODING_PRIMITIVE
			ch2.asn1_tag = 6
			ch2.content = der_encode_ObjectIdentifier(&authenticationValueOther.otherMechanismName)

			ch3 := new(t_der_chunk)
			ch3.asn1_class = ASN1_CLASS_UNIVERSAL
			ch3.encoding = BER_ENCODING_PRIMITIVE
			ch3.asn1_tag = 4
			octetString := ([]uint8)(authenticationValueOther.otherMechanismValue)
			ch3.content = make([]uint8, len(octetString))
			copy(ch3.content, octetString)

			var buf bytes.Buffer
			err = der_encode_chunk(&buf, ch2)
			if nil != err {
				return err
			}
			err = der_encode_chunk(&buf, ch3)
			if nil != err {
				return err
			}

			ch1.content = buf.Bytes()
		}

		var buf bytes.Buffer
		err = der_encode_chunk(&buf, ch1)
		if nil != err {
			return err
		}
		ch.content = buf.Bytes()

		callingAuthenticationValue = ch
	}

	// implementation-information [29] IMPLICIT Implementation-data OPTIONAL,

	var implementationInformation *t_der_chunk
	if nil != aarq.implementationInformation {
		ch := new(t_der_chunk)
		ch.asn1_class = ASN1_CLASS_CONTEXT_SPECIFIC
		ch.encoding = BER_ENCODING_PRIMITIVE
		ch.asn1_tag = 29

		octetString := ([]uint8)(*aarq.implementationInformation)
		ch.content = make([]uint8, len(octetString))
		copy(ch.content, octetString)

		implementationInformation = ch
	}

	// user-information [30] EXPLICIT Association-information OPTIONAL

	var userInformation *t_der_chunk
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
		copy(ch.content, octetString)

		var buf bytes.Buffer
		err = der_encode_chunk(&buf, ch1)
		if nil != err {
			return err
		}

		ch.content = buf.Bytes()
		userInformation = ch
	}

	// AARQ-apdu ::= [APPLICATION 0] IMPLICIT SEQUENCE

	ch := new(t_der_chunk)
	ch.asn1_class = ASN1_CLASS_APPLICATION
	ch.encoding = BER_ENCODING_CONSTRUCTED
	ch.asn1_tag = 0

	var buf bytes.Buffer
	if nil != protocolVersion {
		err = der_encode_chunk(&buf, protocolVersion)
		if nil != err {
			return err
		}
	}
	if nil != applicationContextName {
		err = der_encode_chunk(&buf, applicationContextName)
		if nil != err {
			return err
		}
	}
	if nil != calledAPtitle {
		err = der_encode_chunk(&buf, calledAPtitle)
		if nil != err {
			return err
		}
	}
	if nil != calledAEqualifier {
		err = der_encode_chunk(&buf, calledAEqualifier)
		if nil != err {
			return err
		}
	}
	if nil != calledAPinvocationId {
		err = der_encode_chunk(&buf, calledAPinvocationId)
		if nil != err {
			return err
		}
	}
	if nil != calledAEinvocationId {
		err = der_encode_chunk(&buf, calledAEinvocationId)
		if nil != err {
			return err
		}
	}
	if nil != callingAPtitle {
		err = der_encode_chunk(&buf, callingAPtitle)
		if nil != err {
			return err
		}
	}
	if nil != callingAEqualifier {
		err = der_encode_chunk(&buf, callingAEqualifier)
		if nil != err {
			return err
		}
	}
	if nil != callingAPinvocationId {
		err = der_encode_chunk(&buf, callingAPinvocationId)
		if nil != err {
			return err
		}
	}
	if nil != callingAEinvocationId {
		err = der_encode_chunk(&buf, callingAEinvocationId)
		if nil != err {
			return err
		}
	}
	if nil != senderAcseRequirements {
		err = der_encode_chunk(&buf, senderAcseRequirements)
		if nil != err {
			return err
		}
	}
	if nil != mechanismName {
		err = der_encode_chunk(&buf, mechanismName)
		if nil != err {
			return err
		}
	}
	if nil != callingAuthenticationValue {
		err = der_encode_chunk(&buf, callingAuthenticationValue)
		if nil != err {
			return err
		}
	}
	if nil != implementationInformation {
		err = der_encode_chunk(&buf, implementationInformation)
		if nil != err {
			return err
		}
	}
	if nil != userInformation {
		err = der_encode_chunk(&buf, userInformation)
		if nil != err {
			return err
		}
	}

	ch.content = buf.Bytes()
	err = der_encode_chunk(w, ch)
	if nil != err {
		return err
	}

	return nil

}

func decode_AAREapdu_protocolVersion(ch *t_der_chunk, aarq *AARQapdu) (err error, found bool) {
	// protocol-version [0] IMPLICIT T-protocol-version DEFAULT {version1},
	if 0 == ch.asn1_tag {
		found = true
		err, aarq.protocolVersion = der_decode_BitString(ch.content)
		if nil != err {
			return err, found
		}
	}
	err, ch1 := der_decode_chunk(bytes.NewReader(ch.content))
	if nil != err {
		return err, found
	}

	err, _found := decode_AAREapdu_applicationContextName(ch, aarq)
	if nil != err {
		return err, found
	}
	if !_found {
		err = fmt.Errorf("decoding error")
		return err, found
	}
	return err, found
}

func decode_AAREapdu_applicationContextName(ch *t_der_chunk, aarq *AARQapdu) (err error, found bool) {
	//application-context-name [1] Application-context-name,
	if 1 == ch.asn1_tag {
		found = true
		err, ch1 := der_decode_chunk(bytes.NewReader(ch.content))
		if nil != err {
			return err, found
		}
		if 6 != ch1.asn1_tag {
			err = fmt.Errorf("decoding error")
			return err, found
		}
		err, applicationContextName := der_decode_ObjectIdentifier(ch1.content)
		if nil != err {
			return err, found
		}
		aarq.applicationContextName = *applicationContextName
	}

	err, _found := decode_AAREapdu_calledAPtitle(ch, aarq)
	if nil != err || _found {
		return err, found
	}
	err, _found = decode_AAREapdu_calledAEqualifier(ch, aarq)
	if nil != err || _found {
		return err, found
	}
	err, _found = decode_AAREapdu_calledAPinvocationId(ch, aarq)
	if nil != err || _found {
		return err, found
	}
	err, _found = decode_AAREapdu_calledAEinvocationId(ch, aarq)
	if nil != err || _found {
		return err, found
	}
	err, _found = decode_AAREapdu_callingAPtitle(ch, aarq)
	if nil != err || _found {
		return err, found
	}
	err, _found = decode_AAREapdu_callingAEqualifier(ch, aarq)
	if nil != err || _found {
		return err, found
	}
	err, _found = decode_AAREapdu_callingAPinvocationId(ch, aarq)
	if nil != err || _found {
		return err, found
	}
	err, _found = decode_AAREapdu_callingAEinvocationId(ch, aarq)
	if nil != err || _found {
		return err, found
	}
	err, _found = decode_AAREapdu_senderAcseRequirements(ch, aarq)
	if nil != err || _found {
		return err, found
	}
	err, _found = decode_AAREapdu_mechanismName(ch, aarq)
	if nil != err || _found {
		return err, found
	}
	err, _found = decode_AAREapdu_callingAuthenticationValue(ch, aarq)
	if nil != err || _found {
		return err, found
	}
	err, _found = decode_AAREapdu_implementationInformation(ch, aarq)
	if nil != err || _found {
		return err, found
	}
	err, _found = decode_AAREapdu_userInformation(ch, aarq)
	if nil != err || _found {
		return err, found
	}
	return err, found
}

func decode_AAREapdu_calledAPtitle(ch *t_der_chunk, aarq *AARQapdu) (err error, found bool) {
	//called-AP-title [2] AP-title OPTIONAL,
	if 2 == ch.asn1_tag {
		found = true
		err, ch1 := der_decode_chunk(bytes.NewReader(ch.content))
		if nil != err {
			return err, found
		}
		if 4 != ch1.asn1_tag {
			err = fmt.Errorf("decoding error")
			return err, found
		}
		octetString := make([]uint8, len(ch1.content))
		copy(octetString, ch1.content)
		aarq.calledAPtitle = (*tAsn1OctetString)(&octetString)

	}

	err, _found := decode_AAREapdu_calledAEqualifier(ch, aarq)
	if nil != err || _found {
		return err, found
	}
	err, _found = decode_AAREapdu_calledAPinvocationId(ch, aarq)
	if nil != err || _found {
		return err, found
	}
	err, _found = decode_AAREapdu_calledAEinvocationId(ch, aarq)
	if nil != err || _found {
		return err, found
	}
	err, _found = decode_AAREapdu_callingAPtitle(ch, aarq)
	if nil != err || _found {
		return err, found
	}
	err, _found = decode_AAREapdu_callingAEqualifier(ch, aarq)
	if nil != err || _found {
		return err, found
	}
	err, _found = decode_AAREapdu_callingAPinvocationId(ch, aarq)
	if nil != err || _found {
		return err, found
	}
	err, _found = decode_AAREapdu_callingAEinvocationId(ch, aarq)
	if nil != err || _found {
		return err, found
	}
	err, _found = decode_AAREapdu_senderAcseRequirements(ch, aarq)
	if nil != err || _found {
		return err, found
	}
	err, _found = decode_AAREapdu_mechanismName(ch, aarq)
	if nil != err || _found {
		return err, found
	}
	err, _found = decode_AAREapdu_callingAuthenticationValue(ch, aarq)
	if nil != err || _found {
		return err, found
	}
	err, _found = decode_AAREapdu_implementationInformation(ch, aarq)
	if nil != err || _found {
		return err, found
	}
	err, _found = decode_AAREapdu_userInformation(ch, aarq)
	if nil != err || _found {
		return err, found
	}
	return err, found
}

func decode_AAREapdu_calledAEqualifier(ch *t_der_chunk, aarq *AARQapdu) (err error, found bool) {
	//called-AE-qualifier [3] AE-qualifier OPTIONAL,
	if 3 == ch.asn1_tag {
		found = true
		err, ch1 := der_decode_chunk(bytes.NewReader(ch.content))
		if nil != err {
			return err, found
		}
		if 4 != ch1.asn1_tag {
			err = fmt.Errorf("decoding error")
			return err, found
		}
		octetString := make([]uint8, len(ch1.content))
		copy(octetString, ch1.content)
		aarq.calledAEqualifier = (*tAsn1OctetString)(&octetString)
	}

	err, _found := decode_AAREapdu_calledAPinvocationId(ch, aarq)
	if nil != err || _found {
		return err, found
	}
	err, _found = decode_AAREapdu_calledAEinvocationId(ch, aarq)
	if nil != err || _found {
		return err, found
	}
	err, _found = decode_AAREapdu_callingAPtitle(ch, aarq)
	if nil != err || _found {
		return err, found
	}
	err, _found = decode_AAREapdu_callingAEqualifier(ch, aarq)
	if nil != err || _found {
		return err, found
	}
	err, _found = decode_AAREapdu_callingAPinvocationId(ch, aarq)
	if nil != err || _found {
		return err, found
	}
	err, _found = decode_AAREapdu_callingAEinvocationId(ch, aarq)
	if nil != err || _found {
		return err, found
	}
	err, _found = decode_AAREapdu_senderAcseRequirements(ch, aarq)
	if nil != err || _found {
		return err, found
	}
	err, _found = decode_AAREapdu_mechanismName(ch, aarq)
	if nil != err || _found {
		return err, found
	}
	err, _found = decode_AAREapdu_callingAuthenticationValue(ch, aarq)
	if nil != err || _found {
		return err, found
	}
	err, _found = decode_AAREapdu_implementationInformation(ch, aarq)
	if nil != err || _found {
		return err, found
	}
	err, _found = decode_AAREapdu_userInformation(ch, aarq)
	if nil != err || _found {
		return err, found
	}
	return err, found
}

func decode_AAREapdu_calledAPinvocationId(ch *t_der_chunk, aarq *AARQapdu) (err error, found bool) {
	//called-AP-invocation-id [4] AP-invocation-identifier OPTIONAL,
	if 4 == ch.asn1_tag {
		found = true
		err, ch1 := der_decode_chunk(bytes.NewReader(ch.content))
		if nil != err {
			return err, found
		}
		if 2 != ch1.asn1_tag {
			err = fmt.Errorf("decoding error")
			return err, found
		}
		err, calledAPinvocationId := der_decode_Integer(ch1.content)
		if nil != err {
			return err, found
		}
		aarq.calledAPinvocationId = &calledAPinvocationId
	}

	err, _found := decode_AAREapdu_calledAEinvocationId(ch, aarq)
	if nil != err || _found {
		return err, found
	}
	err, _found = decode_AAREapdu_callingAPtitle(ch, aarq)
	if nil != err || _found {
		return err, found
	}
	err, _found = decode_AAREapdu_callingAEqualifier(ch, aarq)
	if nil != err || _found {
		return err, found
	}
	err, _found = decode_AAREapdu_callingAPinvocationId(ch, aarq)
	if nil != err || _found {
		return err, found
	}
	err, _found = decode_AAREapdu_callingAEinvocationId(ch, aarq)
	if nil != err || _found {
		return err, found
	}
	err, _found = decode_AAREapdu_senderAcseRequirements(ch, aarq)
	if nil != err || _found {
		return err, found
	}
	err, _found = decode_AAREapdu_mechanismName(ch, aarq)
	if nil != err || _found {
		return err, found
	}
	err, _found = decode_AAREapdu_callingAuthenticationValue(ch, aarq)
	if nil != err || _found {
		return err, found
	}
	err, _found = decode_AAREapdu_implementationInformation(ch, aarq)
	if nil != err || _found {
		return err, found
	}
	err, _found = decode_AAREapdu_userInformation(ch, aarq)
	if nil != err || _found {
		return err, found
	}
	return err, found
}

func decode_AAREapdu_calledAEinvocationId(ch *t_der_chunk, aarq *AARQapdu) (err error, found bool) {
	//called-AE-invocation-id [5] AE-invocation-identifier OPTIONAL,
	if 5 == ch.asn1_tag {
		found = true
		err, ch1 := der_decode_chunk(bytes.NewReader(ch.content))
		if nil != err {
			return err, found
		}
		if 2 != ch1.asn1_tag {
			err = fmt.Errorf("decoding error")
			return err, found
		}
		err, calledAEinvocationId := der_decode_Integer(ch1.content)
		if nil != err {
			return err, found
		}
		aarq.calledAEinvocationId = &calledAEinvocationId
	}

	err, _found := decode_AAREapdu_callingAPtitle(ch, aarq)
	if nil != err || _found {
		return err, found
	}
	err, _found = decode_AAREapdu_callingAEqualifier(ch, aarq)
	if nil != err || _found {
		return err, found
	}
	err, _found = decode_AAREapdu_callingAPinvocationId(ch, aarq)
	if nil != err || _found {
		return err, found
	}
	err, _found = decode_AAREapdu_callingAEinvocationId(ch, aarq)
	if nil != err || _found {
		return err, found
	}
	err, _found = decode_AAREapdu_senderAcseRequirements(ch, aarq)
	if nil != err || _found {
		return err, found
	}
	err, _found = decode_AAREapdu_mechanismName(ch, aarq)
	if nil != err || _found {
		return err, found
	}
	err, _found = decode_AAREapdu_callingAuthenticationValue(ch, aarq)
	if nil != err || _found {
		return err, found
	}
	err, _found = decode_AAREapdu_implementationInformation(ch, aarq)
	if nil != err || _found {
		return err, found
	}
	err, _found = decode_AAREapdu_userInformation(ch, aarq)
	if nil != err || _found {
		return err, found
	}
	return err, found
}

func decode_AAREapdu_callingAPtitle(ch *t_der_chunk, aarq *AARQapdu) (err error, found bool) {
	//calling-AP-title [6] AP-title OPTIONAL,
	if 6 == ch.asn1_tag {
		found = true
		err, ch1 := der_decode_chunk(bytes.NewReader(ch.content))
		if nil != err {
			return err, found
		}
		if 4 != ch1.asn1_tag {
			err = fmt.Errorf("decoding error")
			return err, found
		}
		octetString := make([]uint8, len(ch1.content))
		copy(octetString, ch1.content)
		aarq.callingAPtitle = (*tAsn1OctetString)(&octetString)
	}

	err, _found := decode_AAREapdu_callingAEqualifier(ch, aarq)
	if nil != err || _found {
		return err, found
	}
	err, _found = decode_AAREapdu_callingAPinvocationId(ch, aarq)
	if nil != err || _found {
		return err, found
	}
	err, _found = decode_AAREapdu_callingAEinvocationId(ch, aarq)
	if nil != err || _found {
		return err, found
	}
	err, _found = decode_AAREapdu_senderAcseRequirements(ch, aarq)
	if nil != err || _found {
		return err, found
	}
	err, _found = decode_AAREapdu_mechanismName(ch, aarq)
	if nil != err || _found {
		return err, found
	}
	err, _found = decode_AAREapdu_callingAuthenticationValue(ch, aarq)
	if nil != err || _found {
		return err, found
	}
	err, _found = decode_AAREapdu_implementationInformation(ch, aarq)
	if nil != err || _found {
		return err, found
	}
	err, _found = decode_AAREapdu_userInformation(ch, aarq)
	if nil != err || _found {
		return err, found
	}
	return err, found
}

func decode_AAREapdu_callingAEqualifier(ch *t_der_chunk, aarq *AARQapdu) (err error, found bool) {
	//calling-AE-qualifier [7] AE-qualifier OPTIONAL,
	if 7 == ch.asn1_tag {
		found = true
		err, ch1 := der_decode_chunk(bytes.NewReader(ch.content))
		if nil != err {
			return err, found
		}
		if 4 != ch1.asn1_tag {
			err = fmt.Errorf("decoding error")
			return err, found
		}
		octetString := make([]uint8, len(ch1.content))
		copy(octetString, ch1.content)
		aarq.callingAEqualifier = (*tAsn1OctetString)(&octetString)
	}

	err, _found := decode_AAREapdu_callingAPinvocationId(ch, aarq)
	if nil != err || _found {
		return err, found
	}
	err, _found = decode_AAREapdu_callingAEinvocationId(ch, aarq)
	if nil != err || _found {
		return err, found
	}
	err, _found = decode_AAREapdu_senderAcseRequirements(ch, aarq)
	if nil != err || _found {
		return err, found
	}
	err, _found = decode_AAREapdu_mechanismName(ch, aarq)
	if nil != err || _found {
		return err, found
	}
	err, _found = decode_AAREapdu_callingAuthenticationValue(ch, aarq)
	if nil != err || _found {
		return err, found
	}
	err, _found = decode_AAREapdu_implementationInformation(ch, aarq)
	if nil != err || _found {
		return err, found
	}
	err, _found = decode_AAREapdu_userInformation(ch, aarq)
	if nil != err || _found {
		return err, found
	}
	return err, found
}

func decode_AAREapdu_callingAPinvocationId(ch *t_der_chunk, aarq *AARQapdu) (err error, found bool) {
	//calling-AP-invocation-id [8] AP-invocation-identifier OPTIONAL,
	if 8 == ch.asn1_tag {
		found = true
		err, ch1 := der_decode_chunk(bytes.NewReader(ch.content))
		if nil != err {
			return err, found
		}
		if 2 != ch1.asn1_tag {
			err = fmt.Errorf("decoding error")
			return err, found
		}
		err, callingAPinvocationId := der_decode_Integer(ch1.content)
		if nil != err {
			return err, found
		}
		aarq.callingAPinvocationId = &callingAPinvocationId
	}

	err, _found := decode_AAREapdu_callingAEinvocationId(ch, aarq)
	if nil != err || _found {
		return err, found
	}
	err, _found = decode_AAREapdu_senderAcseRequirements(ch, aarq)
	if nil != err || _found {
		return err, found
	}
	err, _found = decode_AAREapdu_mechanismName(ch, aarq)
	if nil != err || _found {
		return err, found
	}
	err, _found = decode_AAREapdu_callingAuthenticationValue(ch, aarq)
	if nil != err || _found {
		return err, found
	}
	err, _found = decode_AAREapdu_implementationInformation(ch, aarq)
	if nil != err || _found {
		return err, found
	}
	err, _found = decode_AAREapdu_userInformation(ch, aarq)
	if nil != err || _found {
		return err, found
	}
	return err, found
}

func decode_AAREapdu_callingAEinvocationId(ch *t_der_chunk, aarq *AARQapdu) (err error, found bool) {
	//calling-AE-invocation-id [9] AE-invocation-identifier OPTIONAL,
	if 9 == ch.asn1_tag {
		found = true
		err, ch1 := der_decode_chunk(bytes.NewReader(ch.content))
		if nil != err {
			return err, found
		}
		if 2 != ch1.asn1_tag {
			err = fmt.Errorf("decoding error")
			return err, found
		}
		err, callingAEinvocationId := der_decode_Integer(ch1.content)
		if nil != err {
			return err, found
		}
		aarq.callingAEinvocationId = &callingAEinvocationId
	}

	err, _found := decode_AAREapdu_senderAcseRequirements(ch, aarq)
	if nil != err || _found {
		return err, found
	}
	err, _found = decode_AAREapdu_mechanismName(ch, aarq)
	if nil != err || _found {
		return err, found
	}
	err, _found = decode_AAREapdu_callingAuthenticationValue(ch, aarq)
	if nil != err || _found {
		return err, found
	}
	err, _found = decode_AAREapdu_implementationInformation(ch, aarq)
	if nil != err || _found {
		return err, found
	}
	err, _found = decode_AAREapdu_userInformation(ch, aarq)
	if nil != err || _found {
		return err, found
	}
	return err, found
}

func decode_AAREapdu_senderAcseRequirements(ch *t_der_chunk, aarq *AARQapdu) (err error, found bool) {
	//sender-acse-requirements [10] IMPLICIT ACSE-requirements OPTIONAL,
	if 10 == ch.asn1_tag {
		found = true
		err, aarq.senderAcseRequirements = der_decode_BitString(ch.content)
		if nil != err {
			return err, found
		}
	}

	err, _found := decode_AAREapdu_mechanismName(ch, aarq)
	if nil != err || _found {
		return err, found
	}
	err, _found = decode_AAREapdu_callingAuthenticationValue(ch, aarq)
	if nil != err || _found {
		return err, found
	}
	err, _found = decode_AAREapdu_implementationInformation(ch, aarq)
	if nil != err || _found {
		return err, found
	}
	err, _found = decode_AAREapdu_userInformation(ch, aarq)
	if nil != err || _found {
		return err, found
	}
	return err, found
}
func decode_AAREapdu_mechanismName(ch *t_der_chunk, aarq *AARQapdu) (err error, found bool) {
	//mechanism-name [11] IMPLICIT Mechanism-name OPTIONAL,
	//mechanismName *tAsn1ObjectIdentifier
	if 11 == ch.asn1_tag {
		found = true
		err, aarq.mechanismName = der_decode_ObjectIdentifier(ch.content)
		if nil != err {
			return err, found
		}
	}

	err, _found := decode_AAREapdu_callingAuthenticationValue(ch, aarq)
	if nil != err || _found {
		return err, found
	}
	err, _found = decode_AAREapdu_implementationInformation(ch, aarq)
	if nil != err || _found {
		return err, found
	}
	err, _found = decode_AAREapdu_userInformation(ch, aarq)
	if nil != err || _found {
		return err, found
	}
	return err, found
}

func decode_AAREapdu_callingAuthenticationValue(ch *t_der_chunk, aarq *AARQapdu) (err error, found bool) {
	//calling-authentication-value [12] EXPLICIT Authentication-value OPTIONAL,
	if 12 == ch.asn1_tag {
		found = true
		err, ch1 := der_decode_chunk(bytes.NewReader(ch.content))
		if nil != err {
			return err, found
		}
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
		err, _found := decode_AAREapdu_callingAuthenticationValue_charstring(ch1, aarq)
		if nil != err {
			return err, found
		}
		if !_found {
			err, _found := decode_AAREapdu_callingAuthenticationValue_bitstring(ch1, aarq)
			if nil != err {
				return err, found
			}
		}
		if !_found {
			err, _found := decode_AAREapdu_callingAuthenticationValue_external(ch1, aarq)
			if nil != err {
				return err, found
			}
		}
		if !_found {
			err, _found := decode_AAREapdu_callingAuthenticationValue_other(ch1, aarq)
			if nil != err {
				return err, found
			}
		}
		if !_found {
			err = fmt.Errorf("decoding error")
			return err, found
		}

	}

	err, _found := decode_AAREapdu_implementationInformation(ch, aarq)
	if nil != err || _found {
		return err, found
	}
	err, _found = decode_AAREapdu_userInformation(ch, aarq)
	if nil != err || _found {
		return err, found
	}
	return err, found
}

func decode_AAREapdu_callingAuthenticationValue_charstring(ch *t_der_chunk, aarq *AARQapdu) (err error, found bool) {
	// charstring [0] IMPLICIT GraphicString,
	if 0 == ch.asn1_tag {
		found = true
		callingAuthenticationValue := new(tAsn1Choice)
		callingAuthenticationValue.tag = 0
		octetString := make([]uint8, len(ch.content))
		copy(octetString, ch.content)
		callingAuthenticationValue.val = octetString
		aarq.callingAuthenticationValue = callingAuthenticationValue
		return err, found
	}
	return err, found
}

func decode_AAREapdu_callingAuthenticationValue_bitstring(ch *t_der_chunk, aarq *AARQapdu) (err error, found bool) {
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

func decode_AAREapdu_callingAuthenticationValue_external(ch *t_der_chunk, aarq *AARQapdu) (err error, found bool) {
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

func decode_AAREapdu_callingAuthenticationValue_other(ch *t_der_chunk, aarq *AARQapdu) (err error, found bool) {
	/*
		other [3] IMPLICIT SEQUENCE
		{
			other-mechanism-name Mechanism-name,
			other-mechanism-value ANY DEFINED BY other-mechanism-name
		}
	*/
	if 3 == ch.asn1_tag {
		found = true
		callingAuthenticationValue := new(tAsn1Choice)
		callingAuthenticationValue.tag = 1

		var authenticationValueOther tAsn1CosemAuthenticationValueOther

		buf := bytes.NewReader(ch.content)
		err, ch1 := der_decode_chunk(buf)
		if nil != err {
			err = fmt.Errorf("decoding error")
			return err, found
		}
		if 6 != ch1.asn1_tag {
			err = fmt.Errorf("decoding error")
			return err, found
		}
		err, objectIdentifier := der_decode_ObjectIdentifier(ch1.content)
		if nil != err {
			err = fmt.Errorf("decoding error")
			return err, found
		}
		authenticationValueOther.otherMechanismName = *objectIdentifier

		err, ch1 = der_decode_chunk(buf)
		if nil != err {
			err = fmt.Errorf("decoding error")
			return err, found
		}
		if 4 != ch1.asn1_tag {
			err = fmt.Errorf("decoding error")
			return err, found
		}
		authenticationValueOther.otherMechanismValue = make([]uint8, len(ch1.content))
		copy(authenticationValueOther.otherMechanismValue, ch1.content)

		callingAuthenticationValue.val = authenticationValueOther
		aarq.callingAuthenticationValue = callingAuthenticationValue
	}
	return err, found
}

func decode_AAREapdu_implementationInformation(ch *t_der_chunk, aarq *AARQapdu) (err error, found bool) {
	//implementation-information [29] IMPLICIT Implementation-data OPTIONAL,
	if 29 == ch.asn1_tag {
		found = true
		octetString := make([]uint8, len(ch.content))
		copy(octetString, ch.content)
		aarq.implementationInformation = (*tAsn1GraphicString)(&octetString)
	}

	err, _found := decode_AAREapdu_userInformation(ch, aarq)
	if nil != err || _found {
		return err, found
	}
	return err, found
}

func decode_AAREapdu_userInformation(ch *t_der_chunk, aarq *AARQapdu) (err error, found bool) {
	//user-information [30] EXPLICIT Association-information OPTIONAL
	if 30 == ch.asn1_tag {
		found = true
		err, ch1 := der_decode_chunk(bytes.NewReader(ch.content))
		if nil != err {
			return err, found
		}
		if 4 != ch1.asn1_tag {
			err = fmt.Errorf("decoding error")
			return err, found
		}
		octetString := make([]uint8, len(ch1.content))
		copy(octetString, ch1.content)
		aarq.userInformation = (*tAsn1OctetString)(&octetString)
	}
	return err, found
}

func decode_AARQapdu_1(r io.Reader) (err error, aarq *AARQapdu) {
	//func der_decode_chunk(r io.Reader) (err error, _ch *t_der_chunk) {

	err, ch := der_decode_chunk(r)
	if nil != err {
		return err, nil
	}

	// AARQ-apdu ::= [APPLICATION 0] IMPLICIT SEQUENCE
	if 0 != ch.asn1_tag {
		err = fmt.Errorf("decoding error")
		return err, nil
	}
	buf := bytes.NewReader(ch.content)

	// protocol-version [0] IMPLICIT T-protocol-version DEFAULT {version1},

	err, ch = der_decode_chunk(buf)
	if nil != err {
		return err, nil
	}
	if 0 == ch.asn1_tag {
		err = fmt.Errorf("decoding error")
		return err, nil
	}

	return nil, nil
}
