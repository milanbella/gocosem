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

	ch                 DlmsChannel // channel to deliver reply
	requestSubmittedAt time.Time
	timeoutAt          *time.Time
}

type DlmsValueReply struct {
	req              *DlmsValueRequest
	invokeId         uint8
	dataAccessResult tDlmsDataAccessResult
	data             *tDlmsData
}

type AppConn struct {
	closed            bool
	dconn             *DlmsConn
	applicationClient uint16
	logicalDevice     uint16
	invokeIdsCh       chan uint8
	requests          map[uint8][]DlmsValueReply // requests waiting for reply to arrive
	finishChannels    []chan string
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
			for invokeId, replies := range aconn.requests {

				// all replies point to same request so we just can pick request from first reply
				//./app.go:74: invalid operation: replies[0] (type *[]DlmsValueReply does not support indexing)
				req := replies[0].req

				if nil == req {
					panic("assertion failed")
				}
				ch := req.ch
				if nil == ch {
					panic("assertion failed")
				}

				if (nil != req.timeoutAt) && req.timeoutAt.After(time.Now()) {
					serr = fmt.Sprintf("%s: request timeout, invokeId: %d", invokeId)
					errorLog.Println(serr)

					delete(aconn.requests, invokeId)
					aconn.invokeIdsCh <- invokeId

					ch <- &DlmsChannelMessage{errors.New(serr), nil}
				}
			}
			deliver()
		}

	}
	deliver()
}

func (aconn *AppConn) processReply(pdu []byte) {
	var (
		serr string
	)

	invokeId := uint8((pdu[2] & 0xF0) >> 4)

	replies := aconn.requests[invokeId]

	if nil == replies {
		errorLog.Printf("%s: no request for invokeId %d, pdu is discarded", invokeId)
		return
	}
	if len(replies) < 1 {
		panic("assertion failed")
	}
	rep := replies[0]
	if nil == replies[0].req {
		panic("assertion failed")
	}
	req := replies[0].req
	if nil == req.ch {
		panic("assertion failed")
	}
	ch := req.ch

	if (0xC4 == pdu[0]) && (0x01 == pdu[1]) {
		err, invokeIdAndPriority, dataAccessResult, data := decode_GetResponseNormal(pdu)
		if invokeId != uint8((invokeIdAndPriority&0xF0)>>4) {
			panic("ivoke id differs")
		}
		if nil != err {
			ch <- &DlmsChannelMessage{errors.New(serr), nil}
		}
		rep.dataAccessResult = dataAccessResult
		rep.data = data
	} else {
		serr = fmt.Sprintf("%s: unknown tag: %02X %02X", pdu[0], pdu[1])
		errorLog.Println(serr)
		ch <- &DlmsChannelMessage{errors.New(serr), nil}
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
			ch <- &DlmsChannelMessage{msg.err, nil}
			aconn.Close()
			return
		}
		pdu := msg.data.([]byte)
		go aconn.processReply(pdu)
	}
}

func (aconn *AppConn) getRquest(ch DlmsChannel, msecTimeout int64, highPriority bool, classId tDlmsClassId, instanceId *tDlmsOid, attributeId tDlmsAttributeId, accessSelector *tDlmsAccessSelector, accessParameter *tDlmsData) {
	var (
		FNAME string = "AppConn.getRquest()"
		serr  string
	)

	if aconn.closed {
		serr := fmt.Sprintf("%s: connection closed", FNAME)
		errorLog.Println(serr)
		ch <- &DlmsChannelMessage{errors.New(serr), nil}
		return
	}

	req := new(DlmsValueRequest)

	req.classId = classId
	req.instanceId = instanceId
	req.attributeId = attributeId
	req.accessSelector = accessSelector
	req.accessParameter = accessParameter

	req.requestSubmittedAt = time.Now()
	timeoutAt := req.requestSubmittedAt.Add(time.Millisecond * time.Duration(msecTimeout))
	req.timeoutAt = &timeoutAt
	req.ch = ch

	finish := make(chan string)
	_ch := make(DlmsChannel)

	go func() {

		// wait for free invokeId slot
		var invokeId uint8
		select {
		case invokeId = <-aconn.invokeIdsCh:
		case reason := <-finish:
			serr = fmt.Sprintf("%s: aborted, reason: %v\n", FNAME, reason)
			_ch <- &DlmsChannelMessage{errors.New(serr), nil}
			return
		}
		aconn.requests[invokeId] = make([]DlmsValueReply, 1)
		rep := &aconn.requests[invokeId][0]

		rep.invokeId = invokeId

		// build and forward pdu to transport

		var invokeIdAndPriority tDlmsInvokeIdAndPriority
		if highPriority {
			invokeIdAndPriority = tDlmsInvokeIdAndPriority((invokeId << 4) | 0x01)
		} else {
			invokeIdAndPriority = tDlmsInvokeIdAndPriority(invokeId << 4)
		}
		err, pdu := encode_GetRequestNormal(invokeIdAndPriority, req.classId, req.instanceId, req.attributeId, req.accessSelector, req.accessParameter)
		if nil != err {
			errorLog.Printf("%s: encode_GetRequestNormal() failed, err: %v\n", FNAME, err)
			delete(aconn.requests, rep.invokeId)
			aconn.invokeIdsCh <- invokeId
			_ch <- &DlmsChannelMessage{err, nil}
			return
		}
		aconn.dconn.transportSend(_ch, aconn.applicationClient, aconn.logicalDevice, pdu)
		select {
		case msg := <-_ch:
			if nil != msg.err {
				errorLog.Printf("%s: encode_GetRequestNormal() failed, err: %v\n", FNAME, err)
				delete(aconn.requests, rep.invokeId)
				aconn.invokeIdsCh <- invokeId
				_ch <- &DlmsChannelMessage{err, nil}
				return
			}
			_ch <- &DlmsChannelMessage{nil, nil}
		case reason := <-finish:
			serr = fmt.Sprintf("%s: aborted, reason: %v\n", FNAME, reason)
			delete(aconn.requests, rep.invokeId)
			aconn.invokeIdsCh <- invokeId
			_ch <- &DlmsChannelMessage{errors.New(serr), nil}
			return
		}

		// wait for reply to arrive

	}()

	select {
	case msg := <-_ch:
		if nil == msg.err {
			// just return, reply shall be forarded to channel 'ch' later upon arrival of reply pdu
			return
		} else {
			ch <- &DlmsChannelMessage{msg.err, msg.data}
		}
	case <-time.After(time.Millisecond * time.Duration(msecTimeout)):
		finish <- "timeout"
	}
}
