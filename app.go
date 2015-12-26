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
	finishChannels    []chan string
	rips              map[uint8][]*DlmsValueRequestResponse // requests in progress
}

func NewAppConn(dconn *DlmsConn, applicationClient uint16, logicalDevice uint16) (aconn *AppConn) {
	aconn = new(AppConn)
	aconn.closed = false
	aconn.applicationClient = applicationClient
	aconn.logicalDevice = logicalDevice

	aconn.invokeIdsCh = make(chan uint8, 0x0F+1)
	for i := 0; i <= 0x0F; i += 1 {
		aconn.invokeIdsCh <- uint8(i)
	}

	aconn.finishChannels = make([]chan string, 0, 10)
	return aconn
}

func (aconn *AppConn) Close() {
	for i := 0; i < len(aconn.finishChannels); i++ {
		aconn.finishChannels[i] <- "close"
	}
	aconn.closed = true
	aconn.dconn.Close()
}

func (aconn *AppConn) deliverTimeouts() {

	var (
		serr string
	)

	finish := make(chan string)
	aconn.finishChannels = append(aconn.finishChannels, finish)

	var deliver func()

	deliver = func() {

		select {
		case <-finish:
			return
		case <-time.After(time.Millisecond * 100):
			for invokeId, rips := range aconn.rips {

				rip := rips[0]
				ch := rips[0].ch

				if (nil != rip.timeoutAt) && rip.timeoutAt.After(time.Now()) {
					serr = fmt.Sprintf("%s: request timeout, invokeId: %d", invokeId)
					errorLog.Println(serr)

					delete(aconn.rips, invokeId)
					aconn.invokeIdsCh <- invokeId

					ch <- &DlmsChannelMessage{errors.New(serr), nil}
				}
			}
			deliver()
		}

	}
	deliver()
}

func (aconn *AppConn) processGetResponseNormal(rips []*DlmsValueRequestResponse, pdu []byte) {

	rip := rips[0]
	ch := rip.ch

	err, _, dataAccessResult, data := decode_GetResponseNormal(pdu)
	if nil != err {
		delete(aconn.rips, rip.invokeId)
		aconn.invokeIdsCh <- rip.invokeId
		ch <- &DlmsChannelMessage{err, nil}
		return
	}

	rip.rep.dataAccessResult = dataAccessResult
	rip.rep.data = data
	delete(aconn.rips, rip.invokeId)
	aconn.invokeIdsCh <- rip.invokeId
	ch <- &DlmsChannelMessage{nil, rips}

	return
}

func (aconn *AppConn) processGetResponseWithList(rips []*DlmsValueRequestResponse, pdu []byte) {

	err, _, dataAccessResults, datas := decode_GetResponseWithList(pdu)
	if nil != err {
		delete(aconn.rips, rips[0].invokeId)
		aconn.invokeIdsCh <- rips[0].invokeId
		rips[0].ch <- &DlmsChannelMessage{err, nil}
		return
	}

	for i := 0; i < len(dataAccessResults); i += 1 {
		rip := rips[i]

		rip.rep = new(DlmsValueResponse)
		rip.rep.dataAccessResult = dataAccessResults[i]
		rip.rep.data = datas[i]
	}

	delete(aconn.rips, rips[0].invokeId)
	aconn.invokeIdsCh <- rips[0].invokeId
	rips[0].ch <- &DlmsChannelMessage{nil, rips}
}

func (aconn *AppConn) processReply(pdu []byte) {
	var (
		serr string
	)

	invokeId := uint8((pdu[2] & 0xF0) >> 4)

	rips := aconn.rips[invokeId]
	if nil == rips {
		errorLog.Printf("%s: no request in progresss for invokeId %d, pdu is discarded", invokeId)
		return
	}

	if (0xC4 == pdu[0]) && (0x01 == pdu[1]) {

		aconn.processGetResponseNormal(rips, pdu)

	} else if (0xC4 == pdu[0]) && (0x03 == pdu[1]) {

		aconn.processGetResponseWithList(rips, pdu)

	} else if (0xC4 == pdu[0]) && (0x02 == pdu[1]) {
		// data blocks response

		err, invokeIdAndPriority, lastBlock, blockNumber, dataAccessResult, rawData := decode_GetResponsewithDataBlock(pdu)
		if nil != err {
			delete(aconn.rips, rips[0].invokeId)
			aconn.invokeIdsCh <- invokeId
			rips[0].ch <- &DlmsChannelMessage{err, nil}
			return
		}
		if 0 != dataAccessResult {
			serr = fmt.Sprintf("%s: error occured receiving response block, invokeId: %d, blockNumber: %d, dataAccessResult: %d", invokeId, blockNumber, dataAccessResult)
			errorLog.Println(serr)
			delete(aconn.rips, rips[0].invokeId)
			aconn.invokeIdsCh <- invokeId
			rips[0].ch <- &DlmsChannelMessage{errors.New(serr), nil}
			return
		}

		if nil != rips[0].rawData {
			rips[0].rawData = rawData
		} else {
			rips[0].rawData = append(rips[0].rawData, rawData...)
		}

		if lastBlock {
			if 1 == len(rips) {
				// normal get
				h := []byte{0xC4, 0x01, byte(invokeIdAndPriority), 0x00}
				_pdu := make([]byte, 0, len(h)+len(rips[0].rawData))
				_pdu = append(_pdu, h...)
				_pdu = append(_pdu, rips[0].rawData...)
				aconn.processGetResponseNormal(rips, _pdu)
			} else {
				// get with list
				h := []byte{0xC4, 0x02, byte(invokeIdAndPriority), 0x00}
				_pdu := make([]byte, 0, len(h)+len(rips[0].rawData))
				_pdu = append(_pdu, h...)
				_pdu = append(_pdu, rips[0].rawData...)
				aconn.processGetResponseWithList(rips, _pdu)
			}
		} else {
			// requests next data block
			//func encode_GetRequestForNextDataBlock(invokeIdAndPriority tDlmsInvokeIdAndPriority, blockNumber uint32) (err error, pdu []byte) {
			err, _pdu := encode_GetRequestForNextDataBlock(invokeIdAndPriority, blockNumber)
			if nil != err {
				delete(aconn.rips, rips[0].invokeId)
				aconn.invokeIdsCh <- invokeId
				rips[0].ch <- &DlmsChannelMessage{err, nil}
				return
			}

			_ch := make(DlmsChannel)
			aconn.dconn.transportSend(_ch, aconn.applicationClient, aconn.logicalDevice, _pdu)
			msg := <-_ch
			if nil != msg.err {
				delete(aconn.rips, rips[0].invokeId)
				aconn.invokeIdsCh <- rips[0].invokeId
				rips[0].ch <- &DlmsChannelMessage{err, nil}
				return
			}
		}

	} else {
		serr = fmt.Sprintf("%s: received pdu discarded due to unknown tag: %02X %02X", pdu[0], pdu[1])
		errorLog.Println(serr)
		return
	}
}

func (aconn *AppConn) receiveReplies() {
	ch := make(DlmsChannel)
	for {
		if aconn.closed {
			return
		}
		aconn.dconn.transportReceive(ch)
		msg := <-ch
		if nil != msg.err {
			aconn.Close()
			return
		}
		pdu := msg.data.([]byte)
		go aconn.processReply(pdu)
	}
}

func (aconn *AppConn) getRquest(ch DlmsChannel, msecTimeout int64, highPriority bool, vals []*DlmsValueRequest) {
	var (
		FNAME string = "AppConn.getRquest()"
		serr  string
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

	//map[uint8][]*DlmsValueRequestResponse
	rips := make([]*DlmsValueRequestResponse, len(vals))
	for i := 0; i < len(vals); i += 1 {
		rip := new(DlmsValueRequestResponse)
		rip.req = vals[i]

		rip.requestSubmittedAt = currentTime
		rip.timeoutAt = &timeoutAt
		rip.ch = ch
		rip.highPriority = highPriority
		rips[i] = rip
	}

	finish := make(chan string)
	_ch := make(DlmsChannel)

	go func() {

		// wait for free invokeId slot
		var invokeId uint8
		select {
		case invokeId = <-aconn.invokeIdsCh:
		case reason := <-finish:
			serr = fmt.Sprintf("%s: aborted, reason: %v\n", FNAME, reason)
			errorLog.Println(serr)
			_ch <- &DlmsChannelMessage{errors.New(serr), nil}
			return
		}
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
			delete(aconn.rips, rips[0].invokeId)
			aconn.invokeIdsCh <- rips[0].invokeId
			_ch <- &DlmsChannelMessage{err, nil}
			return
		}

		aconn.dconn.transportSend(_ch, aconn.applicationClient, aconn.logicalDevice, pdu)
		select {
		case msg := <-_ch:
			if nil != msg.err {
				delete(aconn.rips, rips[0].invokeId)
				aconn.invokeIdsCh <- rips[0].invokeId
				_ch <- &DlmsChannelMessage{err, nil}
				return
			}
			_ch <- &DlmsChannelMessage{nil, nil}
		case reason := <-finish:
			serr = fmt.Sprintf("%s: aborted, reason: %v\n", FNAME, reason)
			errorLog.Println(serr)
			delete(aconn.rips, rips[0].invokeId)
			aconn.invokeIdsCh <- rips[0].invokeId
			_ch <- &DlmsChannelMessage{errors.New(serr), nil}
			return
		}

	}()

	select {
	case msg := <-_ch:
		if nil == msg.err {
			// just return, reply will be forarded to channel 'ch'
			return
		} else {
			ch <- &DlmsChannelMessage{msg.err, msg.data}
		}
	case <-time.After(time.Millisecond * time.Duration(msecTimeout)):
		finish <- "timeout"
	}
}
