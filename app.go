package gocosem

import (
	"errors"
	"fmt"
	"time"
)

var ErrorRequestTimeout = errors.New("request timeout")
var ErrorBlockTimeout = errors.New("block receive timeout")

type DlmsValueRequest struct {
	ClassId         DlmsClassId
	InstanceId      *DlmsOid
	AttributeId     DlmsAttributeId
	AccessSelector  *DlmsAccessSelector
	AccessParameter *DlmsData
}

type DlmsValueResponse struct {
	DataAccessResult DlmsDataAccessResult
	Data             *DlmsData
}

type DlmsValueRequestResponse struct {
	Req *DlmsValueRequest
	Rep *DlmsValueResponse

	invokeId           uint8
	Dead               *string     // If non nil then this request is already dead from whatever reason (e.g. timeot) and MUST NOT be used anymore. String value indicates reason.
	Ch                 DlmsChannel // channel to deliver reply
	RequestSubmittedAt time.Time
	ReplyDeliveredAt   time.Time
	TimeoutAt          *time.Time
	msecBlockTimeout   int64
	blockTimeoutAt     *time.Time
	highPriority       bool
	rawData            []byte
}

type AppConn struct {
	closed            bool
	dconn             *DlmsConn
	applicationClient uint16
	logicalDevice     uint16
	invokeIdsCh       chan uint8
	finish            chan string
	rips              map[uint8][]*DlmsValueRequestResponse // Requests in progress. Map key is invokeId. In case of GetRequestNormal value array will comtain just one item. In case of  GetRequestWithList array lengh will be equal to number of values requested.
}

type DlmsResponse []*DlmsValueRequestResponse

func (rep DlmsResponse) RequestAt(i int) (req *DlmsValueRequest) {
	return rep[i].Req
}

func (rep DlmsResponse) DataAt(i int) *DlmsData {
	return rep[i].Rep.Data
}

func (rep DlmsResponse) DataAccessResultAt(i int) DlmsDataAccessResult {
	return rep[i].Rep.DataAccessResult
}

func (rep DlmsResponse) DeliveredIn() time.Duration {
	return rep[0].ReplyDeliveredAt.Sub(rep[0].RequestSubmittedAt)
}

func NewAppConn(dconn *DlmsConn, applicationClient uint16, logicalDevice uint16) (aconn *AppConn) {
	aconn = new(AppConn)
	aconn.closed = false
	aconn.dconn = dconn
	aconn.applicationClient = applicationClient
	aconn.logicalDevice = logicalDevice

	aconn.invokeIdsCh = make(chan uint8, 0x0F+1)
	for i := 0; i <= 0x0F; i += 1 {
		aconn.invokeIdsCh <- uint8(i)
	}

	aconn.rips = make(map[uint8][]*DlmsValueRequestResponse)

	aconn.finish = make(chan string)

	aconn.receiveReplies()
	aconn.deliverTimeouts()

	return aconn
}

func (aconn *AppConn) Close() {
	if aconn.closed {
		return
	}
	aconn.closed = true
	close(aconn.finish)
	aconn.dconn.Close()
	for invokeId, rips := range aconn.rips {
		if nil != rips[0].Dead {
			continue
		}
		aconn.killRequest(invokeId, errors.New("app connection closed"))
	}
}

func (aconn *AppConn) transportSend(invokeId uint8, pdu []byte) {
	go func() {
		var (
			FNAME string = "AppConn.transportSend()"
		)
		ch := make(DlmsChannel)
		aconn.dconn.transportSend(ch, aconn.applicationClient, aconn.logicalDevice, pdu)
		select {
		case msg := <-ch:
			if nil != msg.Err {
				aconn.killRequest(invokeId, msg.Err)
				errorLog.Printf("%s: closing app connection due to transport error: %v\n", FNAME, msg.Err)
				aconn.Close()
				return
			}
		}
	}()
}

func (aconn *AppConn) killRequest(invokeId uint8, err error) {
	var (
		FNAME string = "AppConn.killRequest()"
		serr  string
	)
	rips, ok := aconn.rips[invokeId]
	if !ok {
		debugLog.Printf("%s: no such request, invokeId: %d", FNAME, invokeId)
		return
	}
	if nil != rips[0].Dead {
		debugLog.Printf("%s: already dead request, invokeId: %d", FNAME, invokeId)
		return
	}
	for _, rip := range rips {
		rip.Dead = new(string)
		if nil != err {
			*rip.Dead = fmt.Sprintf("killed, error: %s", err.Error())
		} else {
			*rip.Dead = "reply delivered"
			rip.ReplyDeliveredAt = time.Now()
		}
	}
	serr = fmt.Sprintf("%s: request killed, invokeId: %d, reason: %s", FNAME, invokeId, *rips[0].Dead)
	errorLog.Println(serr)
	rips[0].Ch <- &DlmsChannelMessage{err, DlmsResponse(rips)}
	aconn.invokeIdsCh <- invokeId
}

func (aconn *AppConn) deliverTimeouts() {
	var FNAME string = "AppConn.deliverTimeouts()"

	var deliver func()

	deliver = func() {

		select {
		case <-aconn.finish:
			return
		case <-time.After(time.Millisecond * 100):
			currentTime := time.Now()
			for invokeId, rips := range aconn.rips {
				if nil != rips[0].Dead {
					continue
				}

				if (nil != rips[0].TimeoutAt) && (currentTime.After(*rips[0].TimeoutAt)) {
					errorLog.Printf("%s request invokeId %d timed out, killed after %v", FNAME, invokeId, currentTime.Sub(rips[0].RequestSubmittedAt))
					if nil != rips[0].rawData {
						// If in the middle of receiving block response try to parse received data so far.
						aconn.processBlockResponse(rips, rips[0].rawData, ErrorRequestTimeout)
					} else {
						aconn.killRequest(invokeId, ErrorRequestTimeout)
					}
				}
				if (nil != rips[0].blockTimeoutAt) && (currentTime.After(*rips[0].blockTimeoutAt)) {
					errorLog.Printf("%s request invokeId %d block timed out, killed after %v", FNAME, invokeId, currentTime.Sub(rips[0].RequestSubmittedAt))
					if nil != rips[0].rawData {
						// If in the middle of receiving block response try to parse received data so far.
						aconn.processBlockResponse(rips, rips[0].rawData, ErrorBlockTimeout)
					} else {
						aconn.killRequest(invokeId, ErrorBlockTimeout)
					}
				}
			}
			go deliver()
		}

	}
	go deliver()
}

func (aconn *AppConn) processGetResponseNormal(rips []*DlmsValueRequestResponse, pdu []byte, errr error) {

	err, _, dataAccessResult, data := decode_GetResponseNormal(pdu)

	rips[0].Rep = new(DlmsValueResponse)
	rips[0].Rep.DataAccessResult = dataAccessResult
	rips[0].Rep.Data = data

	if nil == err {
		aconn.killRequest(rips[0].invokeId, nil)
	} else {
		if nil != errr {
			aconn.killRequest(rips[0].invokeId, errr)
		} else {
			aconn.killRequest(rips[0].invokeId, err)
		}
	}
}

func (aconn *AppConn) processGetResponseWithList(rips []*DlmsValueRequestResponse, pdu []byte, errr error) {

	err, _, dataAccessResults, datas := decode_GetResponseWithList(pdu)

	for i := 0; i < len(dataAccessResults); i += 1 {
		rip := rips[i]
		rip.Rep = new(DlmsValueResponse)
		rip.Rep.DataAccessResult = dataAccessResults[i]
		rip.Rep.Data = datas[i]
	}

	if nil == err {
		aconn.killRequest(rips[0].invokeId, nil)
	} else {
		if nil != errr {
			aconn.killRequest(rips[0].invokeId, errr)
		} else {
			aconn.killRequest(rips[0].invokeId, err)
		}
	}

}

func (aconn *AppConn) processBlockResponse(rips []*DlmsValueRequestResponse, pdu []byte, err error) {
	var (
		FNAME string = "AppConn.processBlockResponse()"
		serr  string
	)

	if (0xC4 == pdu[0]) && (0x01 == pdu[1]) {
		debugLog.Printf("%s: blocks received, processing ResponseNormal", FNAME)
		aconn.processGetResponseNormal(rips, pdu, err)
	} else if (0xC4 == pdu[0]) && (0x03 == pdu[1]) {
		debugLog.Printf("%s: blocks received, processing ResponseWithList", FNAME)
		aconn.processGetResponseWithList(rips, pdu, err)
	} else {
		serr = fmt.Sprintf("%s: assembled pdu discarded due to unknown tag: %02X %02X", pdu[0], pdu[1])
		errorLog.Println(serr)
		aconn.killRequest(rips[0].invokeId, errors.New(serr))
		return
	}
}

func (aconn *AppConn) processReply(pdu []byte) {
	var (
		FNAME string = "processReply()"
		serr  string
	)

	invokeId := uint8((pdu[2] & 0xF0) >> 4)
	debugLog.Printf("%s: invokeId %d\n", FNAME, invokeId)

	rips := aconn.rips[invokeId]
	if nil == rips {
		errorLog.Printf("%s: no request in progresss for invokeId %d, pdu is discarded\n", FNAME, invokeId)
		return
	}
	if nil != rips[0].Dead {
		debugLog.Printf("%s: ignore pdu, request is dead, invokeId %d, reason: %s\n", FNAME, rips[0].invokeId, *rips[0].Dead)
		return
	}

	if (0xC4 == pdu[0]) && (0x01 == pdu[1]) {
		debugLog.Printf("%s: processing ResponseNormal", FNAME)

		aconn.processGetResponseNormal(rips, pdu, nil)

	} else if (0xC4 == pdu[0]) && (0x03 == pdu[1]) {
		debugLog.Printf("%s: processing ResponseWithList", FNAME)

		aconn.processGetResponseWithList(rips, pdu, nil)

	} else if (0xC4 == pdu[0]) && (0x02 == pdu[1]) {
		// data blocks response
		debugLog.Printf("%s: processing ResponsewithDataBlock", FNAME)

		err, invokeIdAndPriority, lastBlock, blockNumber, dataAccessResult, rawData := decode_GetResponsewithDataBlock(pdu)
		if nil != err {
			aconn.killRequest(rips[0].invokeId, err)
			return
		}
		if 0 != dataAccessResult {
			serr = fmt.Sprintf("%s: error occured receiving response block, invokeId: %d, blockNumber: %d, dataAccessResult: %d", FNAME, invokeId, blockNumber, dataAccessResult)
			errorLog.Println(serr)
			aconn.killRequest(rips[0].invokeId, errors.New(serr))
			return
		}

		if nil == rips[0].rawData {
			rips[0].rawData = rawData
		} else {
			rips[0].rawData = append(rips[0].rawData, rawData...)
		}
		_pdu := rips[0].rawData

		if lastBlock {
			aconn.processBlockResponse(rips, _pdu, nil)
		} else {
			// requests next data block

			blockTimeoutAt := time.Now().Add(time.Millisecond * time.Duration(rips[0].msecBlockTimeout))

			if 0 != rips[0].msecBlockTimeout {
				rips[0].blockTimeoutAt = &blockTimeoutAt
			}

			debugLog.Printf("%s: requesting data block: %d", FNAME, blockNumber)
			err, _pdu := encode_GetRequestForNextDataBlock(invokeIdAndPriority, blockNumber)
			if nil != err {
				aconn.killRequest(rips[0].invokeId, err)
				return
			}
			aconn.transportSend(rips[0].invokeId, _pdu)
		}

	} else {
		serr = fmt.Sprintf("%s: received pdu discarded due to unknown tag: %02X %02X", pdu[0], pdu[1])
		errorLog.Println(serr)
		return
	}
}

func (aconn *AppConn) receiveReplies() {
	go func() {
		var (
			FNAME string = "AppConn.receiveReplies()"
			serr  string
		)

		for {
			if aconn.closed {
				break
			}
			ch := make(DlmsChannel)
			aconn.dconn.transportReceive(ch, aconn.logicalDevice, aconn.applicationClient)
			msg := <-ch
			if nil != msg.Err {
				errorLog.Printf("%s: closing app connection due to transport error: %v\n", FNAME, msg.Err)
				aconn.Close()
				return
			}
			m := msg.Data.(map[string]interface{})
			if m["srcWport"] != aconn.logicalDevice {
				serr = fmt.Sprintf("%s: incorret srcWport in received pdu: ", FNAME, m["srcWport"])
				errorLog.Println(serr)
				aconn.Close()
				return
			}
			if m["dstWport"] != aconn.applicationClient {
				serr = fmt.Sprintf("%s: incorret dstWport in received pdu: ", FNAME, m["dstWport"])
				errorLog.Println(serr)
				aconn.Close()
				return
			}
			pdu := m["pdu"].([]byte)
			go aconn.processReply(pdu)
		}
	}()
}

func (aconn *AppConn) getInvokeId(ch DlmsChannel, msecTimeout int64) {
	go func() {
		var (
			FNAME string = "AppConn.getInvokeId()"
			serr  string
		)

		select {
		case invokeId := <-aconn.invokeIdsCh:
			ch <- &DlmsChannelMessage{nil, invokeId}
		case <-aconn.finish:
			serr = fmt.Sprintf("%s: aborted, reason: app connection closed", FNAME)
			errorLog.Println(serr)
			ch <- &DlmsChannelMessage{errors.New(serr), nil}
			return
		case <-time.After(time.Millisecond * time.Duration(msecTimeout)):
			serr = fmt.Sprintf("%s: aborted, reason timeout", FNAME)
			errorLog.Println(serr)
			ch <- &DlmsChannelMessage{errors.New(serr), nil}
			return
		}
	}()
}

func (aconn *AppConn) getRquest(ch DlmsChannel, msecTimeout int64, msecBlockTimeout int64, highPriority bool, vals []*DlmsValueRequest) {
	go func() {
		var (
			FNAME string = "AppConn.getRquest()"
		)

		if 0 == len(vals) {
			ch <- &DlmsChannelMessage{nil, nil}
			return
		}

		if aconn.closed {
			serr := fmt.Sprintf("%s: connection closed", FNAME)
			errorLog.Println(serr)
			ch <- &DlmsChannelMessage{errors.New(serr), nil}
			return
		}

		currentTime := time.Now()
		timeoutAt := currentTime.Add(time.Millisecond * time.Duration(msecTimeout))
		if 0 != msecTimeout {
			timeoutAt = currentTime.Add(time.Millisecond * time.Duration(msecTimeout))
		}

		_ch := make(DlmsChannel)

		var invokeId uint8
		aconn.getInvokeId(_ch, msecTimeout)
		select {
		case msg := <-_ch:
			if nil != msg.Err {
				ch <- &DlmsChannelMessage{msg.Err, nil}
				return
			}
			invokeId = msg.Data.(uint8)
		}
		debugLog.Printf("%s: invokeId %d\n", FNAME, invokeId)

		rips := make([]*DlmsValueRequestResponse, len(vals))
		for i := 0; i < len(vals); i += 1 {
			rip := new(DlmsValueRequestResponse)
			rip.Req = vals[i]

			rip.invokeId = invokeId
			rip.RequestSubmittedAt = currentTime
			if 0 != msecTimeout {
				rip.TimeoutAt = &timeoutAt
			}
			rip.Ch = ch
			rip.highPriority = highPriority
			rips[i] = rip
		}
		aconn.rips[invokeId] = rips

		rips[0].msecBlockTimeout = msecBlockTimeout

		// build and forward pdu to transport

		var invokeIdAndPriority tDlmsInvokeIdAndPriority
		if highPriority {
			invokeIdAndPriority = tDlmsInvokeIdAndPriority((invokeId << 4) | 0x01)
		} else {
			invokeIdAndPriority = tDlmsInvokeIdAndPriority(invokeId << 4)
		}

		var (
			err error
			pdu []byte
		)

		if 1 == len(vals) {
			err, pdu = encode_GetRequestNormal(invokeIdAndPriority, vals[0].ClassId, vals[0].InstanceId, vals[0].AttributeId, vals[0].AccessSelector, vals[0].AccessParameter)
		} else {
			var (
				classIds         []DlmsClassId         = make([]DlmsClassId, len(vals))
				instanceIds      []*DlmsOid            = make([]*DlmsOid, len(vals))
				attributeIds     []DlmsAttributeId     = make([]DlmsAttributeId, len(vals))
				accessSelectors  []*DlmsAccessSelector = make([]*DlmsAccessSelector, len(vals))
				accessParameters []*DlmsData           = make([]*DlmsData, len(vals))
			)
			for i := 0; i < len(vals); i += 1 {
				classIds[i] = vals[i].ClassId
				instanceIds[i] = vals[i].InstanceId
				attributeIds[i] = vals[i].AttributeId
				accessSelectors[i] = vals[i].AccessSelector
				accessParameters[i] = vals[i].AccessParameter
			}
			err, pdu = encode_GetRequestWithList(invokeIdAndPriority, classIds, instanceIds, attributeIds, accessSelectors, accessParameters)
		}

		if nil != err {
			aconn.killRequest(invokeId, err)
			return
		}
		aconn.transportSend(invokeId, pdu)
	}()

}
