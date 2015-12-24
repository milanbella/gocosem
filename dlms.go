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

func getRequest(classId tDlmsClassId, instanceId *tDlmsOid, attributeId tDlmsAttributeId, accessSelector *tDlmsAccessSelector, accessParameters *tDlmsData) (err error, pdu []byte) {
	var FNAME = "getRequest()"

	var w bytes.Buffer

	var buf bytes.Buffer
	err = binary.Write(&buf, binary.BigEndian, classId)
	if nil != err {
		logger.Printf(fmt.Sprintf("%s: binary.Write() failed, err: %s", err))
		return err, nil
	}
	b := buf.Bytes()
	_, err = w.Write(b)
	if nil != err {
		logger.Printf("%s: w.Wite() failed, err: %v", FNAME, err)
		return err, nil
	}

	_, err = w.Write((*instanceId)[0:6])
	if nil != err {
		logger.Printf("%s: w.Wite() failed, err: %v", FNAME, err)
		return err, nil
	}

	_, err = w.Write([]byte{byte(attributeId)})
	if nil != err {
		logger.Printf("%s: w.Wite() failed, err: %v", FNAME, err)
		return err, nil
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
			err, ap = encode_Data((*tAsn1Choice)(accessParameters))
			if nil != err {
				return err, nil
			}
		} else {
			ap = make([]byte, 0)
		}

		_, err = w.Write(as)
		if nil != err {
			logger.Printf("%s: w.Wite() failed, err: %v", FNAME, err)
			return err, nil
		}

		_, err = w.Write(ap)
		if nil != err {
			logger.Printf("%s: w.Wite() failed, err: %v", FNAME, err)
			return err, nil
		}
	}

	return nil, w.Bytes()
}

func getResponse(pdu []byte) (err error, n int, dataAccessResult tDlmsDataAccessResult, data *tDlmsData) {
	var FNAME = "getResponse()"
	var serr string
	var nn = 0

	b := pdu[0:]
	n = 0

	if len(b) < 1 {
		serr = fmt.Sprintf("%s: short pdu", FNAME)
		logger.Printf(serr)
		return errors.New(serr), 0, 0, nil
	}
	dataAccessResult = tDlmsDataAccessResult(b[0])
	b = b[1:]
	n += 1

	var cdata *tAsn1Choice
	if dataAccessResult_success == dataAccessResult {
		err, cdata, nn = decode_Data(b)
		if nil != err {
			return err, n + nn, 0, nil
		}
		n += nn
	}

	return nil, n, dataAccessResult, (*tDlmsData)(cdata)
}

func encode_GetRequestNormal(invokeIdAndPriority tDlmsInvokeIdAndPriority, classId tDlmsClassId, instanceId *tDlmsOid, attributeId tDlmsAttributeId, accessSelector *tDlmsAccessSelector, accessParameters *tDlmsData) (err error, pdu []byte) {
	var FNAME = "encode_GetRequestNormal()"

	var w bytes.Buffer

	_, err = w.Write([]byte{0xc0, 0x01})
	if nil != err {
		logger.Printf("%s: w.Wite() failed, err: %v", FNAME, err)
		return err, nil
	}

	_, err = w.Write([]byte{byte(invokeIdAndPriority)})
	if nil != err {
		logger.Printf("%s: w.Wite() failed, err: %v", FNAME, err)
		return err, nil
	}

	err, pdu = getRequest(classId, instanceId, attributeId, accessSelector, accessParameters)
	if nil != err {
		logger.Printf("%s: getRequest() failed, err: %v", FNAME, err)
		return err, nil
	}

	_, err = w.Write(pdu)
	if nil != err {
		logger.Printf("%s: w.Wite() failed, err: %v", FNAME, err)
		return err, nil
	}

	return nil, w.Bytes()
}

func decode_GetResponseNormal(pdu []byte) (err error, invokeIdAndPriority tDlmsInvokeIdAndPriority, dataAccessResult tDlmsDataAccessResult, data *tDlmsData) {
	var FNAME = "decode_GetResponsenormal()"
	var serr string
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
		serr = fmt.Sprintf("%s: short pdu", FNAME)
		logger.Printf(serr)
		return errors.New(serr), 0, 0, nil
	}
	invokeIdAndPriority = tDlmsInvokeIdAndPriority(b[0])
	b = b[1:]

	err, _, dataAccessResult, data = getResponse(b)
	if nil != err {
		return err, 0, 0, nil
	}

	return nil, invokeIdAndPriority, dataAccessResult, data
}

func encode_GetRequestWithList(invokeIdAndPriority tDlmsInvokeIdAndPriority, classIds []tDlmsClassId, instanceIds []*tDlmsOid, attributeIds []tDlmsAttributeId, accessSelectors []*tDlmsAccessSelector, accessParameters []*tDlmsData) (err error, pdu []byte) {
	var FNAME = "encode_GetRequestWithList()"

	var w bytes.Buffer

	_, err = w.Write([]byte{0xc0, 0x03})
	if nil != err {
		logger.Printf("%s: w.Wite() failed, err: %v", FNAME, err)
		return err, nil
	}

	_, err = w.Write([]byte{byte(invokeIdAndPriority)})
	if nil != err {
		logger.Printf("%s: w.Wite() failed, err: %v", FNAME, err)
		return err, nil
	}

	count := len(classIds) // count of get requests

	_, err = w.Write([]byte{byte(count)})
	if nil != err {
		logger.Printf("%s: w.Wite() failed, err: %v", FNAME, err)
		return err, nil
	}

	for i := 0; i < count; i += 1 {

		err, pdu = getRequest(classIds[i], instanceIds[i], attributeIds[i], accessSelectors[i], accessParameters[i])
		if nil != err {
			logger.Printf("%s: getRequest() failed, err: %v", FNAME, err)
			return err, nil
		}

		_, err = w.Write(pdu)
		if nil != err {
			logger.Printf("%s: w.Wite() failed, err: %v", FNAME, err)
			return err, nil
		}
	}

	return nil, w.Bytes()
}

func decode_GetResponseWithList(pdu []byte) (err error, invokeIdAndPriority tDlmsInvokeIdAndPriority, dataAccessResults []tDlmsDataAccessResult, datas []*tDlmsData) {
	var FNAME = "decode_GetResponseWithList()"
	var serr string
	b := pdu[0:]

	if len(b) < 2 {
		serr = fmt.Sprintf("%s: short pdu", FNAME)
		logger.Printf(serr)
		return errors.New(serr), 0, nil, nil
	}
	if !bytes.Equal(b[0:2], []byte{0xC4, 0x03}) {
		logger.Printf("%s: pdu is not GetResponseWithList: 0x%02X 0x%02X ", FNAME, b[0], b[1])
		return errors.New("pdu is not GetResponseWithList"), 0, nil, nil
	}
	b = b[2:]

	if len(b) < 1 {
		serr = fmt.Sprintf("%s: short pdu", FNAME)
		logger.Printf(serr)
		return errors.New(serr), 0, nil, nil
	}
	invokeIdAndPriority = tDlmsInvokeIdAndPriority(b[0])
	b = b[1:]

	if len(b) < 1 {
		serr = fmt.Sprintf("%s: short pdu", FNAME)
		logger.Printf(serr)
		return errors.New(serr), 0, nil, nil
	}
	count := int(b[0])
	b = b[1:]

	dataAccessResults = make([]tDlmsDataAccessResult, count)
	datas = make([]*tDlmsData, count)

	var dataAccessResult tDlmsDataAccessResult
	var data *tDlmsData
	var n int
	for i := 0; i < count; i += 1 {
		err, n, dataAccessResult, data = getResponse(b)
		if nil != err {
			return err, 0, nil, nil
		}
		b = b[n:]
		dataAccessResults[i] = dataAccessResult
		datas[i] = data
	}

	return nil, invokeIdAndPriority, dataAccessResults, datas
}

func decode_GetResponsewithDataBlock(pdu []byte) (err error, invokeIdAndPriority tDlmsInvokeIdAndPriority, lastBlock bool, blockNumber uint32, dataAccessResult tDlmsDataAccessResult, rawData []byte) {
	var FNAME = "decode_GetResponsewithDataBlock()"
	var serr string
	b := pdu[0:]

	if len(b) < 2 {
		serr = fmt.Sprintf("%s: short pdu", FNAME)
		logger.Printf(serr)
		return errors.New(serr), 0, false, 0, 0, nil
	}
	if !bytes.Equal(b[0:2], []byte{0xC4, 0x02}) {
		serr = fmt.Sprintf("%s: pdu is not GetResponsewithDataBlock: 0x%02X 0x%02X ", FNAME, b[0], b[1])
		logger.Printf(serr)
		return errors.New(serr), 0, false, 0, 0, nil
	}
	b = b[2:]

	if len(b) < 1 {
		serr = fmt.Sprintf("%s: short pdu", FNAME)
		logger.Printf(serr)
		return errors.New(serr), 0, false, 0, 0, nil
	}
	invokeIdAndPriority = tDlmsInvokeIdAndPriority(b[0])
	b = b[1:]

	if len(b) < 1 {
		serr = fmt.Sprintf("%s: short pdu", FNAME)
		logger.Printf(serr)
		return errors.New(serr), 0, false, 0, 0, nil
	}
	if 0 == b[0] {
		lastBlock = false
	} else {
		lastBlock = true
	}
	b = b[1:]

	if len(b) < 4 {
		serr = fmt.Sprintf("%s: short pdu", FNAME)
		logger.Printf(serr)
		return errors.New(serr), 0, false, 0, 0, nil
	}
	err = binary.Read(bytes.NewBuffer(b[0:4]), binary.BigEndian, &blockNumber)
	if nil != err {
		serr = fmt.Sprintf("%s: binary.Read() failed, err: %v", err)
		logger.Printf(serr)
		return errors.New(serr), 0, false, 0, 0, nil
	}
	b = b[4:]

	if len(b) < 1 {
		serr = fmt.Sprintf("%s: short pdu", FNAME)
		logger.Printf(serr)
		return errors.New(serr), 0, false, 0, 0, nil
	}
	dataAccessResult = tDlmsDataAccessResult(b[0])
	b = b[1:]

	if len(b) < 1 {
		serr = fmt.Sprintf("%s: short pdu", FNAME)
		logger.Printf(serr)
		return errors.New(serr), 0, false, 0, 0, nil
	}
	tag := b[0]
	b = b[1:]

	if 0x1E != tag {
		serr = fmt.Sprintf("%s: wrong raw data tag: 0X%02X", FNAME)
		logger.Printf(serr)
		return errors.New(serr), 0, false, 0, 0, nil
	}

	rawData = b

	return nil, invokeIdAndPriority, lastBlock, blockNumber, dataAccessResult, rawData
}

func encode_GetRequestForNextDataBlock(invokeIdAndPriority tDlmsInvokeIdAndPriority, blockNumber uint32) (err error, pdu []byte) {
	var FNAME = "encode_GetRequestForNextDataBlock()"

	var w bytes.Buffer

	_, err = w.Write([]byte{0xc0, 0x02})
	if nil != err {
		logger.Printf("%s: w.Wite() failed, err: %v", FNAME, err)
		return err, nil
	}

	_, err = w.Write([]byte{byte(invokeIdAndPriority)})
	if nil != err {
		logger.Printf("%s: w.Wite() failed, err: %v", FNAME, err)
		return err, nil
	}

	var buf bytes.Buffer
	err = binary.Write(&buf, binary.BigEndian, blockNumber)
	if nil != err {
		logger.Printf(fmt.Sprintf("%s: binary.Write() failed, err: %s", err))
		return err, nil
	}
	b := buf.Bytes()
	_, err = w.Write(b)
	if nil != err {
		logger.Printf("%s: w.Wite() failed, err: %v", FNAME, err)
		return err, nil
	}

	return nil, w.Bytes()
}
