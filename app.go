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

	invokeId           uint8
	Dead               *string           // If non nil then this request is already dead from whatever reason (e.g. timeot) and MUST NOT be used anymore. Value indicates reason for being dead.
	Ch                 chan *DlmsMessage // Channel to deliver reply.
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
	closed            bool
	ch                chan *DlmsMessage // channel to handle application level requests/replies
	applicationClient uint16
	logicalDevice     uint16
	invokeIdsCh       chan uint8
	finish            chan string
	rips              map[uint8][]*DlmsRequestResponse // Requests in progress. Map key is invokeId. In case of GetRequestNormal value array will comtain just one item. In case of  GetRequestWithList array lengh will be equal to number of values requested.
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
	aconn.closed = false
	aconn.applicationClient = applicationClient
	aconn.logicalDevice = logicalDevice

	aconn.ch = make(chan *DlmsMessage)

	// init invoke ids
	aconn.invokeIdsCh = make(chan uint8, 0x0F+1)
	for i := 0; i <= 0x0F; i += 1 {
		aconn.invokeIdsCh <- uint8(i)
	}
	// -----------

	aconn.rips = make(map[uint8][]*DlmsRequestResponse)

	aconn.finish = make(chan string)

	go aconn.handleAppLevelRequests()

	return aconn
}

func NewAppConnAtChannel(dconn *DlmsConn, applicationClient uint16, logicalDevice uint16, channel uint8) (aconn *AppConn) {
	aconn = new(AppConn)
	aconn.dconn = dconn
	aconn.closed = false
	aconn.applicationClient = applicationClient
	aconn.logicalDevice = logicalDevice

	aconn.ch = make(chan *DlmsMessage)

	// init invoke ids
	if channel > 0x0F {
		panic("channel exceeds limit")
	}
	aconn.invokeIdsCh = make(chan uint8, 1)
	aconn.invokeIdsCh <- channel
	// -----------

	aconn.rips = make(map[uint8][]*DlmsRequestResponse)

	aconn.finish = make(chan string)

	go aconn.handleAppLevelRequests()

	return aconn
}

func (aconn *AppConn) Close() {
	if aconn.closed {
		return
	}
	debugLog("closing app connection")
	aconn.closed = true
	close(aconn.finish)
	for _, rips := range aconn.rips {
		if nil != rips[0].Dead {
			continue
		}
		aconn.killRequest(rips, errors.New("app connection closed"))
	}
}

func (aconn *AppConn) transportSend(invokeId uint8, rips []*DlmsRequestResponse, pdu []byte) {
	debugLog("enter")
	ch := make(chan *DlmsMessage)
	msg := new(DlmsMessage)
	msg.Data = &DlmsAppLevelSendRequest{ch, invokeId, rips, pdu}
	aconn.ch <- msg
}

func (aconn *AppConn) transportReceive() {
	debugLog("enter")
	ch := make(chan *DlmsMessage)
	msg := new(DlmsMessage)
	msg.Data = &DlmsAppLevelReceiveRequest{ch}
	aconn.ch <- msg
}

func (aconn *AppConn) _transportSubmit(invokeId uint8, rips []*DlmsRequestResponse, pdu []byte) {
	aconn.transportSend(invokeId, rips, pdu)
	aconn.transportReceive()
}

func (aconn *AppConn) transportSubmit(invokeId uint8, rips []*DlmsRequestResponse, pdu []byte) {
	debugLog("enter")
	go aconn._transportSubmit(invokeId, rips, pdu)
}

func (aconn *AppConn) handleAppLevelRequests() {
	debugLog("enter")
	for msg := range aconn.ch {
		switch v := msg.Data.(type) {
		case *DlmsAppLevelSendRequest:
			debugLog("send request\n")

			aconn.rips[v.invokeId] = v.rips
			ch := make(chan *DlmsMessage)
			aconn.dconn.transportSend(ch, aconn.applicationClient, aconn.logicalDevice, v.pdu)
			_msg := <-ch
			if nil != _msg.Err {
				errorLog("closing app connection due to transport error: %v\n", _msg.Err)
				aconn.Close()
				return
			}

		case *DlmsAppLevelReceiveRequest:
			debugLog("receive request\n")

			ch := make(chan *DlmsMessage)
			aconn.dconn.transportReceive(ch, aconn.logicalDevice, aconn.applicationClient)
			_msg := <-ch

			if nil != _msg.Err {
				errorLog("closing app connection due to transport error: %v\n", _msg.Err)
				aconn.Close()
				return
			}
			m := _msg.Data.(*DlmsTransportReceiveRequestReply)
			pdu := m.pdu

			buf := bytes.NewBuffer(pdu)

			p := make([]byte, 3)
			err := binary.Read(buf, binary.BigEndian, p)
			if nil != err {
				errorLog("io.Read() failed: %v", err)
				aconn.Close()
				return
			}

			invokeId := uint8((p[2] & 0xF0) >> 4)
			debugLog("invokeId %d\n", invokeId)

			rips := aconn.rips[invokeId]
			if nil == rips {
				errorLog("no request in progresss for invokeId %d", invokeId)
				errorLog("closing app connection")
				aconn.Close()
			}
			if nil != rips[0].Dead {
				errorLog("received pdu for dead request, invokeId %d, reason for request being dead: %s\n", rips[0].invokeId, *rips[0].Dead)
				errorLog("closing app connection")
				aconn.Close()
				return
			}
			go aconn.processReply(rips, p, buf)
		default:
			panic(fmt.Sprintf("unknown request type: %T", v))
		}
	}
}

func (aconn *AppConn) killRequest(rips []*DlmsRequestResponse, err error) {
	invokeId := rips[0].invokeId
	if nil != rips[0].Dead {
		debugLog("already dead request, invokeId: %d", invokeId)
		return
	}
	for _, rip := range rips {
		rip.Dead = new(string)
		if nil != err {
			*rip.Dead = fmt.Sprintf("killed, error: %s", err.Error())
			rip.ReplyDeliveredAt = time.Now()
		} else {
			*rip.Dead = "reply delivered"
			rip.ReplyDeliveredAt = time.Now()
		}
	}
	if nil != err {
		err = fmt.Errorf("request killed, invokeId: %d, reason: %s", invokeId, *rips[0].Dead)
		errorLog("%s", err)
	}
	rips[0].Ch <- &DlmsMessage{err, DlmsResultResponse(rips)}
	close(rips[0].Ch)
	aconn.invokeIdsCh <- invokeId
}

func (aconn *AppConn) processGetResponseNormal(rips []*DlmsRequestResponse, r io.Reader, errr error) {

	err, dataAccessResult, data := decode_GetResponseNormal(r)

	rips[0].Rep = new(DlmsResponse)
	rips[0].Rep.DataAccessResult = dataAccessResult
	rips[0].Rep.Data = data

	if nil == err {
		aconn.killRequest(rips, nil)
	} else {
		if nil != errr {
			aconn.killRequest(rips, errr)
		} else {
			aconn.killRequest(rips, err)
		}
	}
}

func (aconn *AppConn) processGetResponseNormalBlock(rips []*DlmsRequestResponse, r io.Reader, errr error) {

	err, data := decode_GetResponseNormalBlock(r)

	rips[0].Rep = new(DlmsResponse)
	rips[0].Rep.DataAccessResult = dataAccessResult_success
	rips[0].Rep.Data = data

	if nil == err {
		aconn.killRequest(rips, nil)
	} else {
		if nil != errr {
			aconn.killRequest(rips, errr)
		} else {
			aconn.killRequest(rips, err)
		}
	}
}

func (aconn *AppConn) processGetResponseWithList(rips []*DlmsRequestResponse, r io.Reader, errr error) {

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
		aconn.killRequest(rips, nil)
	} else {
		if nil != errr {
			aconn.killRequest(rips, errr)
		} else {
			aconn.killRequest(rips, err)
		}
	}

}

func (aconn *AppConn) processBlockResponse(rips []*DlmsRequestResponse, r io.Reader, err error) {
	if 1 == len(rips) {
		debugLog("blocks received, processing ResponseNormal")
		aconn.processGetResponseNormalBlock(rips, r, err)
	} else {
		debugLog("blocks received, processing ResponseWithList")
		aconn.processGetResponseWithList(rips, r, err)
	}
}

func (aconn *AppConn) processSetResponseNormal(rips []*DlmsRequestResponse, r io.Reader, errr error) {

	err, dataAccessResult := decode_SetResponseNormal(r)

	rips[0].Rep = new(DlmsResponse)
	rips[0].Rep.DataAccessResult = dataAccessResult

	if nil == err {
		aconn.killRequest(rips, nil)
	} else {
		if nil != errr {
			aconn.killRequest(rips, errr)
		} else {
			aconn.killRequest(rips, err)
		}
	}
}

func (aconn *AppConn) processSetResponseWithList(rips []*DlmsRequestResponse, r io.Reader, errr error) {

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
		aconn.killRequest(rips, nil)
	} else {
		if nil != errr {
			aconn.killRequest(rips, errr)
		} else {
			aconn.killRequest(rips, err)
		}
	}
}

func (aconn *AppConn) processReply(rips []*DlmsRequestResponse, p []byte, r io.Reader) {

	invokeId := uint8((p[2] & 0xF0) >> 4)
	debugLog("invokeId %d\n", invokeId)

	if (0xC4 == p[0]) && (0x01 == p[1]) {
		debugLog("processing GetResponseNormal")

		aconn.processGetResponseNormal(rips, r, nil)

	} else if (0xC4 == p[0]) && (0x03 == p[1]) {
		debugLog("processing GetResponseWithList")

		aconn.processGetResponseWithList(rips, r, nil)

	} else if (0xC4 == p[0]) && (0x02 == p[1]) {
		// data blocks response
		debugLog("processing GetResponsewithDataBlock")

		err, lastBlock, blockNumber, dataAccessResult, rawData := decode_GetResponsewithDataBlock(r)
		if nil != err {
			aconn.killRequest(rips, err)
			return
		}
		if 0 != dataAccessResult {
			err = fmt.Errorf("error occured receiving response block, invokeId: %d, blockNumber: %d, dataAccessResult: %d", invokeId, blockNumber, dataAccessResult)
			errorLog("%s", err)
			aconn.killRequest(rips, err)
			return
		}

		if nil == rips[0].rawData {
			rips[0].rawData = rawData
		} else {
			rips[0].rawData = append(rips[0].rawData, rawData...)
		}
		_pdu := rips[0].rawData

		if lastBlock {
			aconn.processBlockResponse(rips, bytes.NewBuffer(_pdu), nil)
		} else {
			// requests next data block

			debugLog("requesting next data block after block %d", blockNumber)

			var buf bytes.Buffer
			invokeIdAndPriority := p[2]
			_, err := buf.Write([]byte{0xC0, 0x02, byte(invokeIdAndPriority)})
			if nil != err {
				aconn.killRequest(rips, err)
				return
			}

			err = encode_GetRequestForNextDataBlock(&buf, blockNumber)
			if nil != err {
				aconn.killRequest(rips, err)
				return
			}

			aconn.transportSubmit(invokeId, rips, buf.Bytes())
		}

	} else if (0xC5 == p[0]) && (0x01 == p[1]) {
		debugLog("processing SetResponseNormal")

		aconn.processSetResponseNormal(rips, r, nil)

	} else if (0xC5 == p[0]) && (0x05 == p[1]) {
		debugLog("processing SetResponseWithList")

		aconn.processSetResponseWithList(rips, r, nil)

	} else if (0xC5 == p[0]) && (0x02 == p[1]) {
		debugLog("processing SetResponseForDataBlock")

		req := rips[0].Req

		err, blockNumber := decode_SetResponseForDataBlock(r)
		if nil != err {
			aconn.killRequest(rips, err)
			return
		}
		if req.blockNumber != blockNumber {
			err = fmt.Errorf("error occured receiving response block: received unexpected blockNumber: %d, invokeId: %d ", blockNumber, invokeId)
			errorLog("%s", err)
			aconn.killRequest(rips, err)
			return
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

		var buf bytes.Buffer
		invokeIdAndPriority := p[2]
		_, err = buf.Write([]byte{0xC1, 0x03, byte(invokeIdAndPriority)})
		if nil != err {
			aconn.killRequest(rips, err)
			return
		}

		err = encode_SetRequestWithDataBlock(&buf, lastBlock, blockNumber+1, rawData)
		if nil != err {
			aconn.killRequest(rips, err)
			return
		}
		req.blockNumber += 1

		aconn.transportSubmit(invokeId, rips, buf.Bytes())

	} else if (0xC5 == p[0]) && (0x03 == p[1]) {
		debugLog("processing SetResponseForLastDataBlock")

		req := rips[0].Req

		err, dataAccessResult, blockNumber := decode_SetResponseForLastDataBlock(r)
		if nil != err {
			aconn.killRequest(rips, err)
			return
		}
		if req.blockNumber != blockNumber {
			err = fmt.Errorf("error occured receiving response block: received unexpected blockNumber: %d, invokeId: %d ", blockNumber, invokeId)
			errorLog("%s", err)
			aconn.killRequest(rips, err)
			return
		}

		rips[0].Rep = new(DlmsResponse)
		rips[0].Rep.DataAccessResult = dataAccessResult

		aconn.killRequest(rips, nil)

	} else if (0xC5 == p[0]) && (0x04 == p[1]) {
		debugLog("processing SetResponseForLastDataBlockWithList")

		req := rips[0].Req

		err, dataAccessResults, blockNumber := decode_SetResponseForLastDataBlockWithList(r)
		if nil != err {
			aconn.killRequest(rips, err)
			return
		}
		if req.blockNumber != blockNumber {
			err = fmt.Errorf("error occured receiving response block: received unexpected blockNumber: %d, invokeId: %d ", blockNumber, invokeId)
			errorLog("%s", err)
			aconn.killRequest(rips, err)
			return
		}

		if len(rips) != len(dataAccessResults) {
			err = fmt.Errorf("error occured receiving response block: received unexpected number of results: %d, expected: %d, invokeId: %d ", len(dataAccessResults), len(rips), invokeId)
			errorLog("%s", err)
			aconn.killRequest(rips, err)
			return
		}

		for i := 0; i < len(rips); i++ {
			rips[i].Rep = new(DlmsResponse)
			rips[i].Rep.DataAccessResult = dataAccessResults[i]
		}

		aconn.killRequest(rips, nil)

	} else {
		err := fmt.Errorf("received pdu discarded due to unknown tag: %02X %02X", p[0], p[1])
		errorLog("%s", err)
		return
	}
}

func (aconn *AppConn) getInvokeId() (err error, invokeId uint8) {

	debugLog("waiting for free invokeId ...")
	select {
	case _invokeId := <-aconn.invokeIdsCh:
		debugLog("invokeId: %d", _invokeId)
		return nil, _invokeId
	case <-aconn.finish:
		err = fmt.Errorf("aborted, reason: app connection closed")
		errorLog("%s", err)
		return err, 0
	}
}

func (aconn *AppConn) sendRequest(ch chan *DlmsMessage, vals []*DlmsRequest) {
	debugLog("")
	highPriority := true

	if 0 == len(vals) {
		ch <- &DlmsMessage{nil, nil}
		return
	}

	if aconn.closed {
		err := fmt.Errorf("connection closed")
		errorLog("%s", err)
		ch <- &DlmsMessage{err, nil}
		return
	}

	err, invokeId := aconn.getInvokeId()
	if nil != err {
		ch <- &DlmsMessage{err, nil}
		return
	}
	debugLog("invokeId %d\n", invokeId)

	rips := make([]*DlmsRequestResponse, len(vals))
	for i := 0; i < len(vals); i += 1 {
		rip := new(DlmsRequestResponse)
		rip.Req = vals[i]

		rip.invokeId = invokeId
		rip.RequestSubmittedAt = time.Now()
		rip.Ch = ch
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

	var buf bytes.Buffer

	if 1 == len(vals) {
		if nil == vals[0].Data {
			_, err = buf.Write([]byte{0xC0, 0x01, byte(invokeIdAndPriority)})
			if nil != err {
				errorLog("buf.Write() failed: %v\n", err)
				return
			}
			err = encode_GetRequestNormal(&buf, vals[0].ClassId, vals[0].InstanceId, vals[0].AttributeId, vals[0].AccessSelector, vals[0].AccessParameter)
			if nil != err {
				return
			}
		} else {
			if 0 == vals[0].BlockSize {
				_, err = buf.Write([]byte{0xC1, 0x01, byte(invokeIdAndPriority)})
				if nil != err {
					errorLog("buf.Write() failed: %v\n", err)
					return
				}

				err = encode_SetRequestNormal(&buf, vals[0].ClassId, vals[0].InstanceId, vals[0].AttributeId, vals[0].AccessSelector, vals[0].AccessParameter, vals[0].Data)
				if nil != err {
					return
				}
			} else {
				_, err = buf.Write([]byte{0xC1, 0x02, byte(invokeIdAndPriority)})
				if nil != err {
					errorLog("buf.Write() failed: %v\n", err)
					return
				}

				var _buf bytes.Buffer
				err = vals[0].Data.Encode(&_buf)
				if nil != err {
					return
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

				err = encode_SetRequestNormalBlock(&buf, vals[0].ClassId, vals[0].InstanceId, vals[0].AttributeId, vals[0].AccessSelector, vals[0].AccessParameter, lastBlock, vals[0].blockNumber+1, rawData)
				if nil != err {
					return
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
				return
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
			err = encode_GetRequestWithList(&buf, classIds, instanceIds, attributeIds, accessSelectors, accessParameters)
			if nil != err {
				return
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
					return
				}

				err = encode_SetRequestWithList(&buf, classIds, instanceIds, attributeIds, accessSelectors, accessParameters, datas)
				if nil != err {
					return
				}
			} else {
				_, err = buf.Write([]byte{0xC1, 0x05, byte(invokeIdAndPriority)})
				if nil != err {
					errorLog("buf.Write() failed: %v\n", err)
					return
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
						return
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

				err = encode_SetRequestWithListBlock(&buf, classIds, instanceIds, attributeIds, accessSelectors, accessParameters, lastBlock, vals[0].blockNumber+1, rawData)
				if nil != err {
					return
				} else {
					vals[0].blockNumber += 1
				}
			}
		}
	} else {
		panic("assertion failed")
	}

	aconn.transportSubmit(invokeId, rips, buf.Bytes())
}

func (aconn *AppConn) SendRequest(vals []*DlmsRequest) <-chan *DlmsMessage {

	ch := make(chan *DlmsMessage)
	go aconn.sendRequest(ch, vals)
	return ch
}
