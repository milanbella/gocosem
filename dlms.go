package gocosem

import (
	"io"
)

type tDlmsInvokeIdAndPriority uint8
type tDlmsOid [6]uint8
type tDlmsClassId uint16
type tDlmsAttributeId uint8
type tDlmsAccessSelector uint8

func encodeGetRequest(rw io.ReadWriter, invokeIdAndPriority tDlmsInvokeIdAndPriority, classId tDlmsClassId, oid *tDlmsOid, attributeId tDlmsAttributeId) (n int, err error) {

	var _n int

	n = 0
	_n, err = rw.Write([]byte{0xc0, 0x01, byte(invokeIdAndPriority)})
	n += _n
	if nil != err {
		return n, err
	}

	n, err = rw.Write([]byte{byte(classId)})
	n += _n
	if nil != err {
		return n, err
	}

	n, err = rw.Write((*oid)[0:6])
	n += _n
	if nil != err {
		return n, err
	}

	n, err = rw.Write([]byte{byte(classId)})
	n += _n
	if nil != err {
		return n, err
	}

	n, err = rw.Write([]byte{byte(attributeId)})
	n += _n
	if nil != err {
		return n, err
	}

	return n, nil
}
