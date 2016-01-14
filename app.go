package gocosem

import (
	"errors"
	"fmt"
	"time"
)

type DlmsValueRequest struct {
	classId         tDlmsClassId
	instanceId      *tDlmsOid
	attributeId     tDlmsAttributeId
	accessSelector  *tDlmsAccessSelector
	accessParameter *tDlmsData
}

type DlmsValueResponse struct {
	dataAccessResult tDlmsDataAccessResult
	data             *tDlmsData
}

type DlmsValueRequestResponse struct {
	req *DlmsValueRequest
	rep *DlmsValueResponse

	invokeId           uint8
	dead               *string     // If non nil then this request is already dead from whatever reason (e.g. timeot) and MUST NOT be used anymore. String value indicates reason.
	ch                 DlmsChannel // channel to deliver reply
	requestSubmittedAt time.Time
	timeoutAt          *time.Time
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
	rips              map[uint8][]*DlmsValueRequestResponse // requests in progress
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

	return aconn
}

func (aconn *AppConn) Close() {
	if aconn.closed {
		return
	}
	aconn.closed = true
	close(aconn.finish)
	aconn.dconn.Close()
	for invokeId, _ := range aconn.rips {
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
			if nil != msg.err {
				aconn.killRequest(invokeId, msg.err)
				errorLog.Printf("%s: closing app connection due to transport error: %v\n", FNAME, msg.err)
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
	if nil == err {
		panic("assertion failed")
	}
	rips, ok := aconn.rips[invokeId]
	if !ok {
		return
	}
	if nil != rips[0].dead {
		return
	}
	for _, rip := range rips {
		rip.dead = new(string)
		*rip.dead = err.Error()
	}
	serr = fmt.Sprintf("%s: request killed, invokeId: %d, reason: %s", FNAME, invokeId, *rips[0].dead)
	errorLog.Println(serr)
	rips[0].ch <- &DlmsChannelMessage{err, nil}
	aconn.invokeIdsCh <- invokeId
}

func (aconn *AppConn) deliverReply(invokeId uint8) {
	var (
		FNAME string = "AppConn.deliverReply()"
	)
	rips, ok := aconn.rips[invokeId]
	if !ok {
		debugLog.Printf("%s: no such request, invokeId: %d", FNAME, invokeId)
		return
	}
	if nil != rips[0].dead {
		debugLog.Printf("%s: dead request, invokeId: %d", FNAME, invokeId)
		return
	}
	for _, rip := range rips {
		rip.dead = new(string)
		*rip.dead = "reply delivered"
	}
	debugLog.Printf("%s: reply delivered, invokeId: %d\n", FNAME, invokeId)
	rips[0].ch <- &DlmsChannelMessage{nil, rips}
	aconn.invokeIdsCh <- invokeId
}

func (aconn *AppConn) deliverTimeouts() {

	var deliver func()

	deliver = func() {

		select {
		case <-aconn.finish:
			return
		case <-time.After(time.Millisecond * 100):
			for invokeId, rips := range aconn.rips {

				if (nil != rips[0].timeoutAt) && rips[0].timeoutAt.After(time.Now()) {
					aconn.killRequest(invokeId, errors.New("request timeout"))
				}
			}
			go deliver()
		}

	}
	go deliver()
}

func (aconn *AppConn) processGetResponseNormal(rips []*DlmsValueRequestResponse, pdu []byte) {

	err, _, dataAccessResult, data := decode_GetResponseNormal(pdu)
	if nil != err {
		aconn.killRequest(rips[0].invokeId, err)
		return
	}

	rips[0].rep = new(DlmsValueResponse)
	rips[0].rep.dataAccessResult = dataAccessResult
	rips[0].rep.data = data

	aconn.deliverReply(rips[0].invokeId)
}

func (aconn *AppConn) processGetResponseWithList(rips []*DlmsValueRequestResponse, pdu []byte) {

	err, _, dataAccessResults, datas := decode_GetResponseWithList(pdu)
	if nil != err {
		aconn.killRequest(rips[0].invokeId, err)
		return
	}

	for i := 0; i < len(dataAccessResults); i += 1 {
		rip := rips[i]

		rip.rep = new(DlmsValueResponse)
		rip.rep.dataAccessResult = dataAccessResults[i]
		rip.rep.data = datas[i]
	}

	aconn.deliverReply(rips[0].invokeId)
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
	if nil != rips[0].dead {
		debugLog.Printf("%s: ignore pdu, request is dead, invokeId %d, reason: %s\n", FNAME, rips[0].invokeId, *rips[0].dead)
		return
	}

	if (0xC4 == pdu[0]) && (0x01 == pdu[1]) {
		errorLog.Printf("%s: processing ResponseNormal", FNAME)

		aconn.processGetResponseNormal(rips, pdu)

	} else if (0xC4 == pdu[0]) && (0x03 == pdu[1]) {
		errorLog.Printf("%s: processing ResponseWithList", FNAME)

		aconn.processGetResponseWithList(rips, pdu)

	} else if (0xC4 == pdu[0]) && (0x02 == pdu[1]) {
		// data blocks response
		errorLog.Printf("%s: processing ResponsewithDataBlock", FNAME)

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
			if (0xC4 == _pdu[0]) && (0x01 == _pdu[1]) {
				debugLog.Printf("%s: all blocks received, processing ResponseNormal", FNAME)
				aconn.processGetResponseNormal(rips, _pdu)
			} else if (0xC4 == _pdu[0]) && (0x02 == _pdu[1]) {
				debugLog.Printf("%s: all blocks received, processing ResponseWithList", FNAME)
				aconn.processGetResponseWithList(rips, _pdu)
			} else {
				panic("assertion failed")
			}
		} else {
			// requests next data block

			errorLog.Printf("%s: requesting data block: %d", FNAME, blockNumber)
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
			if nil != msg.err {
				errorLog.Printf("%s: closing app connection due to transport error: %v\n", FNAME, msg.err)
				aconn.Close()
				return
			}
			m := msg.data.(map[string]interface{})
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

func (aconn *AppConn) getRquest(ch DlmsChannel, msecTimeout int64, highPriority bool, vals []*DlmsValueRequest) {
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

		_ch := make(DlmsChannel)

		var invokeId uint8
		aconn.getInvokeId(_ch, msecTimeout)
		select {
		case msg := <-_ch:
			if nil != msg.err {
				ch <- &DlmsChannelMessage{msg.err, nil}
				return
			}
			invokeId = msg.data.(uint8)
		}
		debugLog.Printf("%s: invokeId %d\n", FNAME, invokeId)

		rips := make([]*DlmsValueRequestResponse, len(vals))
		for i := 0; i < len(vals); i += 1 {
			rip := new(DlmsValueRequestResponse)
			rip.req = vals[i]

			rip.invokeId = invokeId
			rip.requestSubmittedAt = currentTime
			rip.timeoutAt = &timeoutAt
			rip.ch = ch
			rip.highPriority = highPriority
			rips[i] = rip
		}
		aconn.rips[invokeId] = rips

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
			err, pdu = encode_GetRequestNormal(invokeIdAndPriority, vals[0].classId, vals[0].instanceId, vals[0].attributeId, vals[0].accessSelector, vals[0].accessParameter)
		} else {
			var (
				classIds         []tDlmsClassId         = make([]tDlmsClassId, len(vals))
				instanceIds      []*tDlmsOid            = make([]*tDlmsOid, len(vals))
				attributeIds     []tDlmsAttributeId     = make([]tDlmsAttributeId, len(vals))
				accessSelectors  []*tDlmsAccessSelector = make([]*tDlmsAccessSelector, len(vals))
				accessParameters []*tDlmsData           = make([]*tDlmsData, len(vals))
			)
			for i := 0; i < len(vals); i += 1 {
				classIds[i] = vals[i].classId
				instanceIds[i] = vals[i].instanceId
				attributeIds[i] = vals[i].attributeId
				accessSelectors[i] = vals[i].accessSelector
				accessParameters[i] = vals[i].accessParameter
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
