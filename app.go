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

type DlmsValueReply struct {
	req                *DlmsValueRequest
	invokeId           uint8
	ch                 DlmsChannel // channel to deliver reply
	requestSubmittedAt time.Time
}

type AppConn struct {
	dconn             *DlmsConn
	applicationClient uint16
	logicalDevice     uint16
	invokeIdsCh       chan uint8
	reqCh             DlmsChannel
	repCh             DlmsChannel
	requests          map[uint8]*DlmsValueReply // requests waiting for reply to arrive
}

func NewAppConn(dconn *DlmsConn, applicationClient uint16, logicalDevice uint16) (aconn *AppConn) {
	aconn = new(AppConn)
	aconn.applicationClient = applicationClient
	aconn.logicalDevice = logicalDevice

	aconn.invokeIdsCh = make(chan uint8, 0x0F+1)
	for i := 0; i <= 0x0F; i += 1 {
		aconn.invokeIdsCh <- uint8(i)
	}
	return aconn
}

func (aconn *AppConn) getRquest(ch DlmsChannel, msecTimeout int64, highPriority bool, req *DlmsValueRequest) {
	var (
		FNAME string = "AppConn.getRquest()"
		serr  string
	)

	finish := make(chan string)
	_ch := make(DlmsChannel)

	go func() {

		rep := new(DlmsValueReply)
		rep.req = req

		// wait for free invokeId slot
		var invokeId uint8
		select {
		case invokeId = <-aconn.invokeIdsCh:
		case reason := <-finish:
			serr = fmt.Sprintf("%s: aborted, reason: %v\n", FNAME, reason)
			_ch <- &DlmsChannelMessage{errors.New(serr), nil}
			return
		}
		rep.invokeId = invokeId
		rep.requestSubmittedAt = time.Now()
		rep.ch = ch
		aconn.requests[rep.invokeId] = rep

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
