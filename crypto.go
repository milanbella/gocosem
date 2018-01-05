package gocosem

// #cgo CPPFLAGS: -I${SRCDIR}/c/cosem_crypto
// #cgo LDFLAGS: -L${SRCDIR}/c/cosem_crypto -lcosemcrypto
// #include <stdio.h>
// #include <errno.h>
// #include "tomcrypt.h"
import "C"

import (
	"fmt"
	"unsafe"
)

const GCM_TAG_LEN = 12

func aesgcm(key []byte, IV []byte, adata []byte, pdu []byte, direction int) (err error, opdu []byte, tag []byte) {

	ckey := C.CBytes(key)
	ckeyLen := len(key)
	cIV := C.CBytes(IV)
	cIVLen := len(IV)
	cadata := C.CBytes(adata)
	cadataLen := len(adata)
	cpdu := C.CBytes(pdu)
	cpduLen := len(pdu)
	ctag := C.CBytes(make([]byte, GCM_TAG_LEN))
	ctagLen := C.ulong(GCM_TAG_LEN)

	cipher := C.CBytes(make([]byte, len(pdu)))

	var res C.int
	if !(direction == 0) || (direction == 1) {
		err = fmt.Errorf("wrogn direction parameter: use 0 for encrypt, 1 for decrypt")
		errorLog("%s", err)
		return err, nil, nil
	}
	res = C.gcm_memory(0, (*C.uchar)(ckey), C.ulong(ckeyLen), (*C.uchar)(cIV), C.ulong(cIVLen), (*C.uchar)(cadata), C.ulong(cadataLen), (*C.uchar)(cpdu), C.ulong(cpduLen), (*C.uchar)(cipher), (*C.uchar)(ctag), (*C.ulong)(unsafe.Pointer(&ctagLen)), C.int(direction))
	if int(res) != 0 {
		err = fmt.Errorf("aesgcm() failed: %d", int(res))
		errorLog("%s", err)
		return err, nil, nil
	}
	opdu = C.GoBytes(cpdu, C.int(cpduLen))
	tag = C.GoBytes(ctag, GCM_TAG_LEN)

	return nil, opdu, tag
}
