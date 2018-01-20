package gocosem

// #cgo CPPFLAGS: -I${SRCDIR}/c/cosem_crypto
// #cgo LDFLAGS: -L${SRCDIR}/c/cosem_crypto -lcosemcrypto
// #include <stdio.h>
// #include <errno.h>
// #include <stdlib.h>
// #include "tomcrypt.h"
import "C"

import (
	"fmt"
	"unsafe"
)

const GCM_TAG_LEN = 12

// Note: Use 'direction' 0 for encrypt, 'direction' 1 for decrypt.

func aesgcm(key []byte, IV []byte, adata []byte, pdu []byte, direction int) (err error, opdu []byte, tag []byte) {

	ckey := C.CBytes(key)
	defer C.free(ckey)
	ckeyLen := len(key)
	cIV := C.CBytes(IV)
	defer C.free(cIV)
	cIVLen := len(IV)
	cadata := C.CBytes(adata)
	defer C.free(cadata)
	cadataLen := len(adata)
	cpdu := C.CBytes(pdu)
	defer C.free(cpdu)
	cpduLen := len(pdu)
	ctag := C.CBytes(make([]byte, GCM_TAG_LEN))
	defer C.free(ctag)
	ctagLen := C.ulong(GCM_TAG_LEN)

	copdu := C.CBytes(make([]byte, cpduLen))
	defer C.free(copdu)

	var res C.int
	if !(direction == 0) || (direction == 1) {
		err = fmt.Errorf("wrogn direction parameter: use 0 for encrypt, 1 for decrypt")
		errorLog("%s", err)
		return err, nil, nil
	}
	res = C.gcm_memory(0, (*C.uchar)(ckey), C.ulong(ckeyLen), (*C.uchar)(cIV), C.ulong(cIVLen), (*C.uchar)(cadata), C.ulong(cadataLen), (*C.uchar)(cpdu), C.ulong(cpduLen), (*C.uchar)(copdu), (*C.uchar)(ctag), (*C.ulong)(unsafe.Pointer(&ctagLen)), C.int(direction))
	if int(res) != 0 {
		err = fmt.Errorf("aesgcm() failed: %d", int(res))
		errorLog("%s", err)
		return err, nil, nil
	}
	opdu = C.GoBytes(copdu, C.int(cpduLen))
	tag = C.GoBytes(ctag, GCM_TAG_LEN)

	return nil, opdu, tag
}
