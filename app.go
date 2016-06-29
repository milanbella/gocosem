package gocosem

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"time"
)

var ErrorRequestTimeout = errors.New("request timeout")
var ErrorBlockTimeout = errors.New("block receive timeout")

type DlmsRequest struct {
	ClassId         DlmsClassId
	InstanceId      *DlmsOid
	AttributeId     DlmsAttributeId
	AccessSelector  DlmsAccessSelector
	AccessParameter *DlmsData
	Data            *DlmsData // Data to be sent with SetRequest. Must be nil if GetRequest.
	BlockSize       int       // If > 0 then data sent with SetReuqest are sent in bolocks.

	rawData     []byte // Remaining data to be sent using block transfer.
	blockNumber uint32 // Number of last block sent.
}

type DlmsResponse struct {
	DataAccessResult DlmsDataAccessResult
	Data             *DlmsData
}

type DlmsRequestResponse struct {
	Req *DlmsRequest
	Rep *DlmsResponse

	RequestSubmittedAt time.Time
	ReplyDeliveredAt   time.Time
	highPriority       bool
	rawData            []byte
}

type DlmsAppLevelSendRequest struct {
	ch       chan *DlmsMessage // reply channel
	invokeId uint8
	rips     []*DlmsRequestResponse
	pdu      []byte
}
type DlmsAppLevelReceiveRequest struct {
	ch chan *DlmsMessage // reply channel
}

type AppConn struct {
	dconn             *DlmsConn
	applicationClient uint16
	logicalDevice     uint16
}

type DlmsResultResponse []*DlmsRequestResponse

func (rep DlmsResultResponse) RequestAt(i int) (req *DlmsRequest) {
	return rep[i].Req
}

func (rep DlmsResultResponse) DataAt(i int) *DlmsData {
	return rep[i].Rep.Data
}

func (rep DlmsResultResponse) DataAccessResultAt(i int) DlmsDataAccessResult {
	return rep[i].Rep.DataAccessResult
}

func (rep DlmsResultResponse) DeliveredIn() time.Duration {
	return rep[0].ReplyDeliveredAt.Sub(rep[0].RequestSubmittedAt)
}

func NewAppConn(dconn *DlmsConn, applicationClient uint16, logicalDevice uint16) (aconn *AppConn) {
	aconn = new(AppConn)
	aconn.dconn = dconn
	aconn.applicationClient = applicationClient
	aconn.logicalDevice = logicalDevice

	return aconn
}

func (aconn *AppConn) Close() {
	debugLog("closing")
}

func (aconn *AppConn) processGetResponseNormal(rips []*DlmsRequestResponse, r io.Reader, errr error) error {

	err, dataAccessResult, data := decode_GetResponseNormal(r)

	rips[0].Rep = new(DlmsResponse)
	rips[0].Rep.DataAccessResult = dataAccessResult
	rips[0].Rep.Data = data

	if nil == err {
		return nil
	} else {
		if nil != errr {
			return errr
		} else {
			return err
		}
	}
}

func (aconn *AppConn) processGetResponseNormalBlock(rips []*DlmsRequestResponse, r io.Reader, errr error) error {

	err, data := decode_GetResponseNormalBlock(r)

	rips[0].Rep = new(DlmsResponse)
	rips[0].Rep.DataAccessResult = dataAccessResult_success
	rips[0].Rep.Data = data

	if nil == err {
		return nil
	} else {
		if nil != errr {
			return errr
		} else {
			return err
		}
	}
}

func (aconn *AppConn) processGetResponseWithList(rips []*DlmsRequestResponse, r io.Reader, errr error) error {

	err, dataAccessResults, datas := decode_GetResponseWithList(r)

	if len(dataAccessResults) != len(rips) {
		err = fmt.Errorf("unexpected count of received list entries")
		errorLog("%s", err)

		if len(dataAccessResults) > len(rips) {
			dataAccessResults = dataAccessResults[0:len(rips)]
		}
	}

	for i := 0; i < len(dataAccessResults); i += 1 {
		rip := rips[i]
		rip.Rep = new(DlmsResponse)
		rip.Rep.DataAccessResult = dataAccessResults[i]
		rip.Rep.Data = datas[i]
	}

	if nil == err {
		return nil
	} else {
		if nil != errr {
			return errr
		} else {
			return err
		}
	}

}

func (aconn *AppConn) processBlockResponse(rips []*DlmsRequestResponse, r io.Reader, err error) error {
	if 1 == len(rips) {
		debugLog("blocks received, processing ResponseNormal")
		return aconn.processGetResponseNormalBlock(rips, r, err)
	} else {
		debugLog("blocks received, processing ResponseWithList")
		return aconn.processGetResponseWithList(rips, r, err)
	}
}

func (aconn *AppConn) processSetResponseNormal(rips []*DlmsRequestResponse, r io.Reader, errr error) error {

	err, dataAccessResult := decode_SetResponseNormal(r)

	rips[0].Rep = new(DlmsResponse)
	rips[0].Rep.DataAccessResult = dataAccessResult

	if nil == err {
		return nil
	} else {
		if nil != errr {
			return errr
		} else {
			return err
		}
	}
}

func (aconn *AppConn) processSetResponseWithList(rips []*DlmsRequestResponse, r io.Reader, errr error) error {

	err, dataAccessResults := decode_SetResponseWithList(r)

	if len(dataAccessResults) != len(rips) {
		err = fmt.Errorf("unexpected count of received list entries")
		errorLog("%s", err)

		if len(dataAccessResults) > len(rips) {
			dataAccessResults = dataAccessResults[0:len(rips)]
		}
	}

	for i := 0; i < len(dataAccessResults); i += 1 {
		rip := rips[i]
		rip.Rep = new(DlmsResponse)
		rip.Rep.DataAccessResult = dataAccessResults[i]
	}

	if nil == err {
		return nil
	} else {
		if nil != errr {
			return errr
		} else {
			return err
		}
	}
}

func (aconn *AppConn) processReply(rips []*DlmsRequestResponse, p []byte, r io.Reader) error {

	invokeId := uint8((p[2] & 0xF0) >> 4)
	debugLog("invokeId %d\n", invokeId)

	if (0xC4 == p[0]) && (0x01 == p[1]) {
		debugLog("processing GetResponseNormal")

		return aconn.processGetResponseNormal(rips, r, nil)

	} else if (0xC4 == p[0]) && (0x03 == p[1]) {
		debugLog("processing GetResponseWithList")

		return aconn.processGetResponseWithList(rips, r, nil)

	} else if (0xC4 == p[0]) && (0x02 == p[1]) {
		// data blocks response
		debugLog("processing GetResponsewithDataBlock")

		err, lastBlock, blockNumber, dataAccessResult, rawData := decode_GetResponsewithDataBlock(r)
		if nil != err {
			return err
		}
		if 0 != dataAccessResult {
			err = fmt.Errorf("error occured receiving response block, invokeId: %d, blockNumber: %d, dataAccessResult: %d", invokeId, blockNumber, dataAccessResult)
			errorLog("%s", err)
			return err
		}

		if nil == rips[0].rawData {
			rips[0].rawData = rawData
		} else {
			rips[0].rawData = append(rips[0].rawData, rawData...)
		}
		_pdu := rips[0].rawData

		if lastBlock {
			return aconn.processBlockResponse(rips, bytes.NewBuffer(_pdu), nil)
		} else {
			// requests next data block

			debugLog("requesting next data block after block %d", blockNumber)

			buf := new(bytes.Buffer)
			invokeIdAndPriority := p[2]
			_, err := buf.Write([]byte{0xC0, 0x02, byte(invokeIdAndPriority)})
			if nil != err {
				return err
			}

			err = encode_GetRequestForNextDataBlock(buf, blockNumber)
			if nil != err {
				return err
			}

			err = aconn.dconn.transportSend(aconn.applicationClient, aconn.logicalDevice, buf.Bytes())
			if nil != err {
				return err
			}

			pdu, err := aconn.dconn.transportReceive(aconn.logicalDevice, aconn.applicationClient)
			if nil != err {
				return err
			}

			buf = bytes.NewBuffer(pdu)

			p = make([]byte, 3)
			err = binary.Read(buf, binary.BigEndian, p)
			if nil != err {
				errorLog("io.Read() failed: %v", err)
				return err
			}
			invokeIdRcv := uint8((p[2] & 0xF0) >> 4)
			if invokeIdRcv != invokeId {
				err := fmt.Errorf("invoke ids differs: invokeId sent: %v, invokeId received: %v", invokeId, invokeIdRcv)
				errorLog("%s", err)
				return err
			}

			return aconn.processReply(rips, p, buf)
		}

	} else if (0xC5 == p[0]) && (0x01 == p[1]) {
		debugLog("processing SetResponseNormal")

		return aconn.processSetResponseNormal(rips, r, nil)

	} else if (0xC5 == p[0]) && (0x05 == p[1]) {
		debugLog("processing SetResponseWithList")

		return aconn.processSetResponseWithList(rips, r, nil)

	} else if (0xC5 == p[0]) && (0x02 == p[1]) {
		debugLog("processing SetResponseForDataBlock")

		req := rips[0].Req

		err, blockNumber := decode_SetResponseForDataBlock(r)
		if nil != err {
			return err
		}
		if req.blockNumber != blockNumber {
			err = fmt.Errorf("error occured receiving response block: received unexpected blockNumber: %d, invokeId: %d ", blockNumber, invokeId)
			errorLog("%s", err)
			return err
		}

		// set next block

		n := req.BlockSize
		if n > len(req.rawData) {
			n = len(req.rawData)
		}

		rawData := req.rawData[0:n]
		req.rawData = req.rawData[n:]

		lastBlock := false
		if 0 == len(req.rawData) {
			lastBlock = true
		}

		debugLog("setting next data block (current block is %d)", blockNumber)

		buf := new(bytes.Buffer)
		invokeIdAndPriority := p[2]
		_, err = buf.Write([]byte{0xC1, 0x03, byte(invokeIdAndPriority)})
		if nil != err {
			return err
		}

		err = encode_SetRequestWithDataBlock(buf, lastBlock, blockNumber+1, rawData)
		if nil != err {
			return err
		}
		req.blockNumber += 1

		err = aconn.dconn.transportSend(aconn.applicationClient, aconn.logicalDevice, buf.Bytes())
		if nil != err {
			return err
		}

		pdu, err := aconn.dconn.transportReceive(aconn.logicalDevice, aconn.applicationClient)
		if nil != err {
			return err
		}

		buf = bytes.NewBuffer(pdu)

		p = make([]byte, 3)
		err = binary.Read(buf, binary.BigEndian, p)
		if nil != err {
			errorLog("io.Read() failed: %v", err)
			aconn.Close()
			return err
		}
		invokeIdRcv := uint8((p[2] & 0xF0) >> 4)
		if invokeIdRcv != invokeId {
			err := fmt.Errorf("invoke ids differs: invokeId sent: %v, invokeId received: %v", invokeId, invokeIdRcv)
			errorLog("%s", err)
			return err
		}

		return aconn.processReply(rips, p, buf)

	} else if (0xC5 == p[0]) && (0x03 == p[1]) {
		debugLog("processing SetResponseForLastDataBlock")

		req := rips[0].Req

		err, dataAccessResult, blockNumber := decode_SetResponseForLastDataBlock(r)
		if nil != err {
			return err
		}
		if req.blockNumber != blockNumber {
			err = fmt.Errorf("error occured receiving response block: received unexpected blockNumber: %d, invokeId: %d ", blockNumber, invokeId)
			errorLog("%s", err)
			return err
		}

		rips[0].Rep = new(DlmsResponse)
		rips[0].Rep.DataAccessResult = dataAccessResult

		return nil

	} else if (0xC5 == p[0]) && (0x04 == p[1]) {
		debugLog("processing SetResponseForLastDataBlockWithList")

		req := rips[0].Req

		err, dataAccessResults, blockNumber := decode_SetResponseForLastDataBlockWithList(r)
		if nil != err {
			return err
		}
		if req.blockNumber != blockNumber {
			err = fmt.Errorf("error occured receiving response block: received unexpected blockNumber: %d, invokeId: %d ", blockNumber, invokeId)
			errorLog("%s", err)
			return err
		}

		if len(rips) != len(dataAccessResults) {
			err = fmt.Errorf("error occured receiving response block: received unexpected number of results: %d, expected: %d, invokeId: %d ", len(dataAccessResults), len(rips), invokeId)
			errorLog("%s", err)
			return err
		}

		for i := 0; i < len(rips); i++ {
			rips[i].Rep = new(DlmsResponse)
			rips[i].Rep.DataAccessResult = dataAccessResults[i]
		}

		return nil

	} else {
		err := fmt.Errorf("received pdu discarded due to unknown tag: %02X %02X", p[0], p[1])
		errorLog("%s", err)
		return err
	}
}

func (aconn *AppConn) SendRequest(vals []*DlmsRequest, invokeId uint8) (response DlmsResultResponse, err error) {
	debugLog("enter")
	highPriority := true

	if 0 == len(vals) {
		return nil, nil
	}

	debugLog("invokeId %d\n", invokeId)
	if invokeId > 0x0F {
		err := fmt.Errorf("invokeId exceeds limit")
		errorLog("%s", err)
		return nil, err
	}

	rips := make([]*DlmsRequestResponse, len(vals))
	for i := 0; i < len(vals); i += 1 {
		rip := new(DlmsRequestResponse)
		rip.Req = vals[i]

		rip.RequestSubmittedAt = time.Now()
		rip.highPriority = highPriority
		rips[i] = rip
	}

	// build and forward pdu to transport

	var invokeIdAndPriority tDlmsInvokeIdAndPriority
	if highPriority {
		invokeIdAndPriority = tDlmsInvokeIdAndPriority((invokeId << 4) | 0x01)
	} else {
		invokeIdAndPriority = tDlmsInvokeIdAndPriority(invokeId << 4)
	}

	buf := new(bytes.Buffer) // buffer for encoded application layer data pdu

	// encode application layer data pdu

	if 1 == len(vals) {
		if nil == vals[0].Data {
			_, err = buf.Write([]byte{0xC0, 0x01, byte(invokeIdAndPriority)})
			if nil != err {
				errorLog("buf.Write() failed: %v\n", err)
				return nil, err
			}
			err = encode_GetRequestNormal(buf, vals[0].ClassId, vals[0].InstanceId, vals[0].AttributeId, vals[0].AccessSelector, vals[0].AccessParameter)
			if nil != err {
				return nil, err
			}
		} else {
			if 0 == vals[0].BlockSize {
				_, err = buf.Write([]byte{0xC1, 0x01, byte(invokeIdAndPriority)})
				if nil != err {
					errorLog("buf.Write() failed: %v\n", err)
					return nil, err
				}

				err = encode_SetRequestNormal(buf, vals[0].ClassId, vals[0].InstanceId, vals[0].AttributeId, vals[0].AccessSelector, vals[0].AccessParameter, vals[0].Data)
				if nil != err {
					return nil, err
				}
			} else {
				_, err = buf.Write([]byte{0xC1, 0x02, byte(invokeIdAndPriority)})
				if nil != err {
					errorLog("buf.Write() failed: %v\n", err)
					return nil, err
				}

				var _buf bytes.Buffer
				err = vals[0].Data.Encode(&_buf)
				if nil != err {
					return nil, err
				}
				vals[0].rawData = _buf.Bytes()

				n := vals[0].BlockSize
				if n > len(vals[0].rawData) {
					n = len(vals[0].rawData)
				}

				rawData := vals[0].rawData[0:n]
				vals[0].rawData = vals[0].rawData[n:]

				lastBlock := false
				if 0 == len(vals[0].rawData) {
					lastBlock = true
				}

				err = encode_SetRequestNormalBlock(buf, vals[0].ClassId, vals[0].InstanceId, vals[0].AttributeId, vals[0].AccessSelector, vals[0].AccessParameter, lastBlock, vals[0].blockNumber+1, rawData)
				if nil != err {
					return nil, err
				} else {
					vals[0].blockNumber += 1
				}
			}
		}
	} else if len(vals) > 1 {
		if nil == vals[0].Data {
			_, err = buf.Write([]byte{0xC0, 0x03, byte(invokeIdAndPriority)})
			if nil != err {
				errorLog("buf.Write() failed: %v\n", err)
				return nil, err
			}
			var (
				classIds         []DlmsClassId        = make([]DlmsClassId, len(vals))
				instanceIds      []*DlmsOid           = make([]*DlmsOid, len(vals))
				attributeIds     []DlmsAttributeId    = make([]DlmsAttributeId, len(vals))
				accessSelectors  []DlmsAccessSelector = make([]DlmsAccessSelector, len(vals))
				accessParameters []*DlmsData          = make([]*DlmsData, len(vals))
			)
			for i := 0; i < len(vals); i += 1 {
				classIds[i] = vals[i].ClassId
				instanceIds[i] = vals[i].InstanceId
				attributeIds[i] = vals[i].AttributeId
				accessSelectors[i] = vals[i].AccessSelector
				accessParameters[i] = vals[i].AccessParameter
			}
			err = encode_GetRequestWithList(buf, classIds, instanceIds, attributeIds, accessSelectors, accessParameters)
			if nil != err {
				return nil, err
			}
		} else {
			var (
				classIds         []DlmsClassId        = make([]DlmsClassId, len(vals))
				instanceIds      []*DlmsOid           = make([]*DlmsOid, len(vals))
				attributeIds     []DlmsAttributeId    = make([]DlmsAttributeId, len(vals))
				accessSelectors  []DlmsAccessSelector = make([]DlmsAccessSelector, len(vals))
				accessParameters []*DlmsData          = make([]*DlmsData, len(vals))
				datas            []*DlmsData          = make([]*DlmsData, len(vals))
			)
			for i := 0; i < len(vals); i += 1 {
				classIds[i] = vals[i].ClassId
				instanceIds[i] = vals[i].InstanceId
				attributeIds[i] = vals[i].AttributeId
				accessSelectors[i] = vals[i].AccessSelector
				accessParameters[i] = vals[i].AccessParameter
				datas[i] = vals[i].Data
			}
			if 0 == vals[0].BlockSize {
				_, err = buf.Write([]byte{0xC1, 0x04, byte(invokeIdAndPriority)})
				if nil != err {
					errorLog("buf.Write() failed: %v\n", err)
					return nil, err
				}

				err = encode_SetRequestWithList(buf, classIds, instanceIds, attributeIds, accessSelectors, accessParameters, datas)
				if nil != err {
					return nil, err
				}
			} else {
				_, err = buf.Write([]byte{0xC1, 0x05, byte(invokeIdAndPriority)})
				if nil != err {
					errorLog("buf.Write() failed: %v\n", err)
					return nil, err
				}

				var _buf bytes.Buffer

				count := uint8(len(classIds))
				err = binary.Write(&_buf, binary.BigEndian, count)
				if nil != err {
					panic(fmt.Sprintf("binary.Write() failed: %v", err))
				}
				for i := 0; i < int(count); i++ {
					err = vals[i].Data.Encode(&_buf)
					if nil != err {
						return nil, err
					}
				}
				vals[0].rawData = _buf.Bytes()

				n := vals[0].BlockSize
				if n > len(vals[0].rawData) {
					n = len(vals[0].rawData)
				}

				rawData := vals[0].rawData[0:n]
				vals[0].rawData = vals[0].rawData[n:]

				lastBlock := false
				if 0 == len(vals[0].rawData) {
					lastBlock = true
				}

				err = encode_SetRequestWithListBlock(buf, classIds, instanceIds, attributeIds, accessSelectors, accessParameters, lastBlock, vals[0].blockNumber+1, rawData)
				if nil != err {
					return nil, err
				} else {
					vals[0].blockNumber += 1
				}
			}
		}
	} else {
		panic("assertion failed")
	}

	// send request

	debugLog("send request")

	err = aconn.dconn.transportSend(aconn.applicationClient, aconn.logicalDevice, buf.Bytes())
	if nil != err {
		return nil, err
	}

	// receive reply

	debugLog("receive request")

	pdu, err := aconn.dconn.transportReceive(aconn.logicalDevice, aconn.applicationClient)
	if nil != err {
		return nil, err
	}

	buf = bytes.NewBuffer(pdu)

	p := make([]byte, 3)
	err = binary.Read(buf, binary.BigEndian, p)
	if nil != err {
		errorLog("io.Read() failed: %v", err)
		return nil, err
	}
	invokeIdRcv := uint8((p[2] & 0xF0) >> 4)
	if invokeIdRcv != invokeId {
		err := fmt.Errorf("invoke ids differs: invokeId sent: %v, invokeId received: %v", invokeId, invokeIdRcv)
		errorLog("%s", err)
		return nil, err
	}

	err = aconn.processReply(rips, p, buf)
	if nil == err {
		t := time.Now()
		for i := 0; i < rips.length; i++ {
			rips[i].ReplyDeliveredAt(t)
		}
	}
	return DlmsResultResponse(rips), err
}
