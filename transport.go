package gocosem

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"time"
)

const (
	Transport_TCP  = int(1)
	Transport_UDP  = int(2)
	Transport_HDLC = int(3)
)

type DlmsMessage struct {
	Err  error
	Data interface{}
}

type DlmsChannel chan *DlmsMessage
type DlmsReplyChannel <-chan *DlmsMessage

type tWrapperHeader struct {
	ProtocolVersion uint16
	SrcWport        uint16
	DstWport        uint16
	DataLength      uint16
}

/*func (header *tWrapperHeader) String() string {
	return fmt.Sprintf("tWrapperHeader %+v", header)
}*/

type DlmsConn struct {
	closed        bool
	closedAck     chan error
	rwc           io.ReadWriteCloser
	hdlcRwc       io.ReadWriteCloser // stream used by hdlc transport for sending and reading HDLC frames
	HdlcClient    *HdlcTransport
	transportType int
	ch            chan *DlmsMessage // channel to handle transport level requests/replies
}

type DlmsTransportSendRequest struct {
	ch  chan *DlmsMessage // reply channel
	src uint16            // source address
	dst uint16            // destination address
	pdu []byte
}
type DlmsTransportSendRequestReply struct {
}

type DlmsTransportReceiveRequest struct {
	ch  chan *DlmsMessage // reply channel
	src uint16            // source address
	dst uint16            // destination address
}
type DlmsTransportReceiveRequestReply struct {
	src uint16 // source address
	dst uint16 // destination address
	pdu []byte
}

var ErrorDlmsTimeout = errors.New("ErrorDlmsTimeout")

func makeWpdu(srcWport uint16, dstWport uint16, pdu []byte) (err error, wpdu []byte) {
	var (
		buf    bytes.Buffer
		header tWrapperHeader
	)

	header.ProtocolVersion = 0x00001
	header.SrcWport = srcWport
	header.DstWport = dstWport
	header.DataLength = uint16(len(pdu))

	err = binary.Write(&buf, binary.BigEndian, &header)
	if nil != err {
		errorLog(" binary.Write() failed, err: %v\n", err)
		return err, nil
	}
	_, err = buf.Write(pdu)
	if nil != err {
		errorLog(" binary.Write() failed, err: %v\n", err)
		return err, nil
	}
	return nil, buf.Bytes()

}

func _ipTransportSend(ch chan *DlmsMessage, rwc io.ReadWriteCloser, srcWport uint16, dstWport uint16, pdu []byte) {
	err, wpdu := makeWpdu(srcWport, dstWport, pdu)
	if nil != err {
		ch <- &DlmsMessage{err, nil}
		return
	}
	debugLog("sending: % 02X\n", wpdu)
	_, err = rwc.Write(wpdu)
	if nil != err {
		errorLog("io.Write() failed, err: %v\n", err)
		ch <- &DlmsMessage{err, nil}
		return
	}
	debugLog("sending: ok")
	ch <- &DlmsMessage{nil, &DlmsTransportSendRequestReply{}}
}

func ipTransportSend(ch chan *DlmsMessage, rwc io.ReadWriteCloser, srcWport uint16, dstWport uint16, pdu []byte) {
	go _ipTransportSend(ch, rwc, srcWport, dstWport, pdu)
}

func _hdlcTransportSend(ch chan *DlmsMessage, rwc io.ReadWriteCloser, pdu []byte) {
	var buf bytes.Buffer
	llcHeader := []byte{0xE6, 0xE6, 0x00} // LLC sublayer header

	_, err := buf.Write(llcHeader)
	if nil != err {
		errorLog("io.Write() failed, err: %v\n", err)
		ch <- &DlmsMessage{err, nil}
		return
	}
	_, err = buf.Write(pdu)
	if nil != err {
		errorLog("io.Write() failed, err: %v\n", err)
		ch <- &DlmsMessage{err, nil}
		return
	}

	p := buf.Bytes()
	debugLog("sending: %02X\n", p)
	_, err = rwc.Write(p)
	if nil != err {
		errorLog("io.Write() failed, err: %v\n", err)
		ch <- &DlmsMessage{err, nil}
		return
	}

	debugLog("sending: ok")
	ch <- &DlmsMessage{nil, &DlmsTransportSendRequestReply{}}
}

func hdlcTransportSend(ch chan *DlmsMessage, rwc io.ReadWriteCloser, pdu []byte) {
	go _hdlcTransportSend(ch, rwc, pdu)
}

// Never call this method directly or else you risk race condtitions on io.Writer() in case of paralell call.
// Use instead proxy variant 'transportSend()' which queues this method call on sync channel.

func (dconn *DlmsConn) doTransportSend(ch chan *DlmsMessage, src uint16, dst uint16, pdu []byte) {
	go dconn._doTransportSend(ch, src, dst, pdu)
}

func (dconn *DlmsConn) _doTransportSend(ch chan *DlmsMessage, src uint16, dst uint16, pdu []byte) {
	debugLog("trnasport type: %d, src: %d, dst: %d\n", dconn.transportType, src, dst)

	if (Transport_TCP == dconn.transportType) || (Transport_UDP == dconn.transportType) {
		ipTransportSend(ch, dconn.rwc, src, dst, pdu)
	} else if Transport_HDLC == dconn.transportType {
		hdlcTransportSend(ch, dconn.rwc, pdu)
	} else {
		panic(fmt.Sprintf("unsupported transport type: %d", dconn.transportType))
	}
}

func (dconn *DlmsConn) transportSend(ch chan *DlmsMessage, src uint16, dst uint16, pdu []byte) {
	// enqueue send request
	go func() {
		msg := new(DlmsMessage)

		data := new(DlmsTransportSendRequest)
		data.ch = ch
		data.src = src
		data.dst = dst
		data.pdu = pdu

		msg.Data = data

		dconn.ch <- msg
	}()
}

func ipTransportReceiveForApp(ch chan *DlmsMessage, rwc io.ReadWriteCloser, srcWport uint16, dstWport uint16) {
	ipTransportReceive(ch, rwc, &srcWport, &dstWport)
}

func ipTransportReceiveForAny(ch chan *DlmsMessage, rwc io.ReadWriteCloser) {
	ipTransportReceive(ch, rwc, nil, nil)
}

func _ipTransportReceive(ch chan *DlmsMessage, rwc io.ReadWriteCloser, srcWport *uint16, dstWport *uint16) {
	var (
		err    error
		header tWrapperHeader
	)

	debugLog("receiving pdu ...\n")
	err = binary.Read(rwc, binary.BigEndian, &header)
	if nil != err {
		errorLog("binary.Read() failed, err: %v\n", err)
		ch <- &DlmsMessage{err, nil}
		return
	}
	debugLog("header: ok\n")
	if (nil != srcWport) && (header.SrcWport != *srcWport) {
		err = fmt.Errorf("wrong srcWport: %d, expected: %d", header.SrcWport, *srcWport)
		errorLog("%s", err)
		ch <- &DlmsMessage{err, nil}
		return
	}
	if (nil != dstWport) && (header.DstWport != *dstWport) {
		err = fmt.Errorf("wrong dstWport: %d, expected: %d", header.DstWport, *dstWport)
		errorLog("%s", err)
		ch <- &DlmsMessage{err, nil}
		return
	}
	pdu := make([]byte, header.DataLength)
	err = binary.Read(rwc, binary.BigEndian, pdu)
	if nil != err {
		errorLog("binary.Read() failed, err: %v\n", err)
		ch <- &DlmsMessage{err, nil}
		return
	}
	debugLog("received pdu: % 02X\n", pdu)

	// send reply
	ch <- &DlmsMessage{nil, &DlmsTransportReceiveRequestReply{header.SrcWport, header.DstWport, pdu}}

	return
}

func ipTransportReceive(ch chan *DlmsMessage, rwc io.ReadWriteCloser, srcWport *uint16, dstWport *uint16) {
	go _ipTransportReceive(ch, rwc, srcWport, dstWport)
}

func _hdlcTransportReceive(ch chan *DlmsMessage, rwc io.ReadWriteCloser) {
	var (
		err error
	)

	debugLog("receiving pdu ...\n")

	//TODO: Set maxSegmnetSize to AARE.user-information.server-max-receive-pdu-size.
	// AARE.user-information is of 'InitiateResponse' asn1 type and is A-XDR encoded.
	maxSegmnetSize := 3 * 1024

	p := make([]byte, maxSegmnetSize)

	// hdlc ReadWriter read returns always whole segment into 'p' or full 'p' if 'p' is not long enough to fit in all segment
	n, err := rwc.Read(p)
	if nil != err {
		errorLog("hdlc.Read() failed, err: %v\n", err)
		ch <- &DlmsMessage{err, nil}
		return
	}
	// Guard against read buffer being shorter then maximum possible segment size.
	if len(p) == n {
		panic("short read suspected, increase buffer size!")
	}

	buf := bytes.NewBuffer(p[0:n])

	llcHeader := make([]byte, 3) // LLC sublayer header
	err = binary.Read(buf, binary.BigEndian, llcHeader)
	if nil != err {
		errorLog("binary.Read() failed, err: %v\n", err)
		ch <- &DlmsMessage{err, nil}
		return
	}
	if !bytes.Equal(llcHeader, []byte{0xE6, 0xE7, 0x00}) {
		err = fmt.Errorf("wrong LLC header")
		errorLog("%s", err)
		ch <- &DlmsMessage{err, nil}
		return
	}
	debugLog("LLC header: ok\n")

	pdu := buf.Bytes()
	debugLog("received pdu: % 02X\n", pdu)

	// send reply
	ch <- &DlmsMessage{nil, &DlmsTransportReceiveRequestReply{0, 0, pdu}}

	return
}

func hdlcTransportReceive(ch chan *DlmsMessage, rwc io.ReadWriteCloser) {
	go _hdlcTransportReceive(ch, rwc)
}

// Never call this method directly or else you risk race condtitions on io.Reader() in case of paralell call.
// Use instead proxy variant 'transportReceive()' which queues this method call on sync channel.

func (dconn *DlmsConn) doTransportReceive(ch chan *DlmsMessage, src uint16, dst uint16) {
	debugLog("trnascport type: %d\n", dconn.transportType)

	if (Transport_TCP == dconn.transportType) || (Transport_UDP == dconn.transportType) {
		ipTransportReceiveForApp(ch, dconn.rwc, src, dst)
	} else if Transport_HDLC == dconn.transportType {
		hdlcTransportReceive(ch, dconn.rwc)
	} else {
		err := fmt.Errorf("unsupported transport type: %d", dconn.transportType)
		errorLog("%s", err)
		ch <- &DlmsMessage{err, nil}
		return
	}
}

func (dconn *DlmsConn) transportReceive(ch chan *DlmsMessage, src uint16, dst uint16) {
	// enqueue receive request
	go func() {
		data := new(DlmsTransportReceiveRequest)
		data.ch = ch
		data.src = src
		data.dst = dst
		msg := new(DlmsMessage)
		msg.Data = data
		dconn.ch <- msg
	}()
}

func (dconn *DlmsConn) handleTransportRequests() {
	var err error
	debugLog("start")
	for msg := range dconn.ch {
		switch v := msg.Data.(type) {
		case *DlmsTransportSendRequest:
			debugLog("send request\n")
			if dconn.closed {
				err = fmt.Errorf("tansport send request ignored, transport connection closed")
				errorLog("%s", err)
				v.ch <- &DlmsMessage{err, nil}
			}
			dconn.doTransportSend(v.ch, v.src, v.dst, v.pdu)
		case *DlmsTransportReceiveRequest:
			debugLog("receive request\n")
			if dconn.closed {
				err = fmt.Errorf("transport receive request ignored, transport connection closed")
				errorLog("%s", err)
				v.ch <- &DlmsMessage{err, nil}
			}
			dconn.doTransportReceive(v.ch, v.src, v.dst)
		default:
			panic(fmt.Sprintf("unknown request type: %T", v))
		}
	}
	debugLog("finish")

	// cleanup

	if (Transport_TCP == dconn.transportType) || (Transport_UDP == dconn.transportType) {
		err = dconn.rwc.Close()
	} else if Transport_HDLC == dconn.transportType {
		err = dconn.rwc.Close()
	} else {
		panic(fmt.Sprintf("unsupported transport type: %d", dconn.transportType))
	}
	dconn.closedAck <- err
}

func (dconn *DlmsConn) AppConnectWithPassword(applicationClient uint16, logicalDevice uint16, password string) <-chan *DlmsMessage {
	ch := make(chan *DlmsMessage)
	go func() {
		defer close(ch)
		var aarq = AARQ{
			appCtxt:   LogicalName_NoCiphering,
			authMech:  LowLevelSecurity,
			authValue: password,
		}
		pdu, err := aarq.encode()
		if err != nil {
			ch <- &DlmsMessage{err, nil}
			return
		}

		_ch := make(chan *DlmsMessage)
		dconn.transportSend(_ch, applicationClient, logicalDevice, pdu)
		msg := <-_ch
		if nil != msg.Err {
			ch <- &DlmsMessage{msg.Err, nil}
			return
		}
		dconn.transportReceive(_ch, logicalDevice, applicationClient)
		msg = <-_ch
		if nil != msg.Err {
			_ch <- &DlmsMessage{msg.Err, nil}
			return
		}
		m := msg.Data.(*DlmsTransportReceiveRequestReply)
		pdu = m.pdu

		var aare AARE
		err = aare.decode(pdu)
		if err != nil {
			ch <- &DlmsMessage{err, nil}
			return
		}
		if aare.result != AssociationAccepted {
			err = fmt.Errorf("app connect failed, result: %v, diagnostic: %v", aare.result, aare.diagnostic)
			errorLog("%s", err)
			ch <- &DlmsMessage{err, nil}
			return
		} else {
			aconn := NewAppConn(dconn, applicationClient, logicalDevice)
			ch <- &DlmsMessage{nil, aconn}
		}

	}()
	return ch
}

func (dconn *DlmsConn) AppConnectRaw(applicationClient uint16, logicalDevice uint16, aarq []byte) <-chan *DlmsMessage {
	ch := make(chan *DlmsMessage)
	go func() {
		defer close(ch)

		_ch := make(chan *DlmsMessage)
		dconn.transportSend(_ch, applicationClient, logicalDevice, aarq)
		msg := <-_ch
		if nil != msg.Err {
			ch <- &DlmsMessage{msg.Err, nil}
			return
		}
		dconn.transportReceive(_ch, logicalDevice, applicationClient)
		msg = <-_ch
		if nil != msg.Err {
			ch <- &DlmsMessage{msg.Err, nil}
			return
		}
		m := msg.Data.(*DlmsTransportReceiveRequestReply)
		ch <- &DlmsMessage{nil, m.pdu}
	}()
	return ch
}

func _TcpConnect(ch chan *DlmsMessage, ipAddr string, port int) {
	var (
		conn net.Conn
		err  error
	)

	defer close(ch)

	dconn := new(DlmsConn)
	dconn.closed = false
	dconn.closedAck = make(chan error)
	dconn.ch = make(chan *DlmsMessage)
	dconn.transportType = Transport_TCP

	debugLog("connecting tcp transport: %s:%d\n", ipAddr, port)
	conn, err = net.Dial("tcp", fmt.Sprintf("%s:%d", ipAddr, port))
	if nil != err {
		errorLog("net.Dial(%s:%d) failed, err: %v", ipAddr, port, err)
		ch <- &DlmsMessage{err, nil}
		return
	}
	dconn.rwc = conn

	debugLog("tcp transport connected: %s:%d\n", ipAddr, port)
	go dconn.handleTransportRequests()
	ch <- &DlmsMessage{nil, dconn}

}

func TcpConnect(ipAddr string, port int) <-chan *DlmsMessage {

	ch := make(chan *DlmsMessage)
	go _TcpConnect(ch, ipAddr, port)
	return ch
}

func _HdlcConnect(ch chan *DlmsMessage, ipAddr string, port int, applicationClient uint16, logicalDevice uint16, responseTimeout time.Duration) {
	var (
		conn net.Conn
		err  error
	)

	defer close(ch)

	dconn := new(DlmsConn)
	dconn.closed = false
	dconn.closedAck = make(chan error)
	dconn.ch = make(chan *DlmsMessage)
	dconn.transportType = Transport_HDLC

	debugLog("connecting hdlc transport over tcp: %s:%d\n", ipAddr, port)
	conn, err = net.Dial("tcp", fmt.Sprintf("%s:%d", ipAddr, port))
	if nil != err {
		errorLog("net.Dial(%s:%d) failed, err: %v", ipAddr, port, err)
		ch <- &DlmsMessage{err, nil}
		return
	}
	dconn.hdlcRwc = conn

	client := NewHdlcTransport(dconn.hdlcRwc, responseTimeout, true, uint8(applicationClient), logicalDevice, nil)
	err = client.SendSNRM(nil, nil)
	if nil != err {
		conn.Close()
		ch <- &DlmsMessage{err, nil}
		return
	}
	dconn.HdlcClient = client
	dconn.rwc = client

	debugLog("hdlc transport connected over tcp: %s:%d\n", ipAddr, port)

	go dconn.handleTransportRequests()
	ch <- &DlmsMessage{nil, dconn}

}

func HdlcConnect(ipAddr string, port int, applicationClient uint16, logicalDevice uint16, networkRoundtripTime time.Duration) <-chan *DlmsMessage {

	ch := make(chan *DlmsMessage)
	go _HdlcConnect(ch, ipAddr, port, applicationClient, logicalDevice, networkRoundtripTime)
	return ch
}

func (dconn *DlmsConn) Close() {
	if dconn.closed {
		return
	}
	debugLog("closing transport connection")
	close(dconn.ch)
	<-dconn.closedAck
	dconn.closed = true
}
