package gocosem

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
)

type tDlmsInvokeIdAndPriority uint8
type tDlmsClassId uint16
type tDlmsOid [6]uint8
type tDlmsAttributeId uint8
type tDlmsAccessSelector uint8
type tDlmsData tAsn1Choice

type tDlmsDataAccessResult uint8

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

var logger *log.Logger = getLogger()

func getRequest(classId tDlmsClassId, instanceId *tDlmsOid, attributeId tDlmsAttributeId, accessSelector *tDlmsAccessSelector, accessParameters *tDlmsData) (pdu []byte, err error) {
	var FNAME = "getRequest()"

	var w bytes.Buffer

	var buf bytes.Buffer
	err = binary.Write(&buf, binary.BigEndian, classId)
	if nil != err {
		logger.Printf(fmt.Sprintf("%s: binary.Write() failed, err: %s", err))
		return nil, err
	}
	b := buf.Bytes()
	_, err = w.Write(b)
	if nil != err {
		logger.Printf("%s: w.Wite() failed, err: %v", FNAME, err)
		return nil, err
	}

	_, err = w.Write((*instanceId)[0:6])
	if nil != err {
		logger.Printf("%s: w.Wite() failed, err: %v", FNAME, err)
		return nil, err
	}

	_, err = w.Write([]byte{byte(attributeId)})
	if nil != err {
		logger.Printf("%s: w.Wite() failed, err: %v", FNAME, err)
		return nil, err
	}

	if 0 != attributeId {
		var as []byte
		var ap []byte
		if nil == accessSelector {
			as = []byte{0}
		} else {
			as = []byte{byte(*accessSelector)}
		}
		if nil != accessParameters {
			ap = encode_Data((*tAsn1Choice)(accessParameters))
		} else {
			ap = make([]byte, 0)
		}

		_, err = w.Write(as)
		if nil != err {
			logger.Printf("%s: w.Wite() failed, err: %v", FNAME, err)
			return nil, err
		}

		_, err = w.Write(ap)
		if nil != err {
			logger.Printf("%s: w.Wite() failed, err: %v", FNAME, err)
			return nil, err
		}
	}

	return w.Bytes(), nil
}

func encode_GetRequestNormal(invokeIdAndPriority tDlmsInvokeIdAndPriority, classId tDlmsClassId, instanceId *tDlmsOid, attributeId tDlmsAttributeId, accessSelector *tDlmsAccessSelector, accessParameters *tDlmsData) (pdu []byte, err error) {
	var FNAME = "encode_GetRequestNormal()"

	var w bytes.Buffer

	_, err = w.Write([]byte{0xc0, 0x01})
	if nil != err {
		logger.Printf("%s: w.Wite() failed, err: %v", FNAME, err)
		return nil, err
	}

	_, err = w.Write([]byte{byte(invokeIdAndPriority)})
	if nil != err {
		logger.Printf("%s: w.Wite() failed, err: %v", FNAME, err)
		return nil, err
	}

	pdu, err = getRequest(classId, instanceId, attributeId, accessSelector, accessParameters)
	if nil != err {
		logger.Printf("%s: getRequest() failed, err: %v", FNAME, err)
		return nil, err
	}

	_, err = w.Write(pdu)
	if nil != err {
		logger.Printf("%s: w.Wite() failed, err: %v", FNAME, err)
		return nil, err
	}

	return w.Bytes(), nil
}

func decode_GetResponsenormal(pdu []byte) (err error, invokeIdAndPriority tDlmsInvokeIdAndPriority, dataAccessResult tDlmsDataAccessResult, data *tDlmsData) {
	var FNAME = "decode_GetResponsenormal"
	b := pdu[0:]

	if len(b) < 2 {
		return errors.New("short pdu"), 0, 0, nil
	}
	if !bytes.Equal(b[0:2], []byte{0xC4, 0x01}) {
		logger.Printf("%s: pdu is not GetResponsenormal: 0x%02X 0x%02X ", FNAME, b[0], b[1])
		return errors.New("pdu is not GetResponsenormal"), 0, 0, nil
	}
	b = b[2:]

	if len(b) < 1 {
		return errors.New("short pdu"), 0, 0, nil
	}
	invokeIdAndPriority = tDlmsInvokeIdAndPriority(b[0])
	b = b[1:]

	if len(b) < 1 {
		return errors.New("short pdu"), 0, 0, nil
	}
	dataAccessResult = tDlmsDataAccessResult(b[0])
	b = b[1:]

	if dataAccessResult_success == dataAccessResult {
		data = (*tDlmsData)(decode_Data(b))
	}

	return nil, invokeIdAndPriority, dataAccessResult, data
}

func encode_GetRequestWithList(invokeIdAndPriority tDlmsInvokeIdAndPriority, classIds []tDlmsClassId, instanceIds []*tDlmsOid, attributeIds []tDlmsAttributeId, accessSelectors []*tDlmsAccessSelector, accessParameters []*tDlmsData) (pdu []byte, err error) {
	var FNAME = "encode_GetRequestWithList"

	var w bytes.Buffer

	_, err = w.Write([]byte{0xc0, 0x03})
	if nil != err {
		logger.Printf("%s: w.Wite() failed, err: %v", FNAME, err)
		return nil, err
	}

	_, err = w.Write([]byte{byte(invokeIdAndPriority)})
	if nil != err {
		logger.Printf("%s: w.Wite() failed, err: %v", FNAME, err)
		return nil, err
	}

	count := len(classIds) // count of get requests

	_, err = w.Write([]byte{byte(count)})
	if nil != err {
		logger.Printf("%s: w.Wite() failed, err: %v", FNAME, err)
		return nil, err
	}

	for i := 0; i < count; i += 1 {

		pdu, err = getRequest(classIds[i], instanceIds[i], attributeIds[i], accessSelectors[i], accessParameters[i])
		if nil != err {
			logger.Printf("%s: getRequest() failed, err: %v", FNAME, err)
			return nil, err
		}

		_, err = w.Write(pdu)
		if nil != err {
			logger.Printf("%s: w.Wite() failed, err: %v", FNAME, err)
			return nil, err
		}
	}

	return w.Bytes(), nil
}
