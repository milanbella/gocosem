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

var (
	ErrDlmsTimeout      = errors.New("dlms timeout")
	ErrUnknownTransport = errors.New("unknown dlms transport")
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
	rwc                 io.ReadWriteCloser
	hdlcRwc             io.ReadWriteCloser // stream used by hdlc transport for sending and reading HDLC frames
	HdlcClient          *HdlcTransport
	transportType       int
	hdlcResponseTimeout time.Duration
	snrmTimeout         time.Duration
	discTimeout         time.Duration
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

func ipTransportSend(rwc io.ReadWriteCloser, srcWport uint16, dstWport uint16, pdu []byte) error {
	err, wpdu := makeWpdu(srcWport, dstWport, pdu)
	if nil != err {
		return err
	}
	debugLog("sending: % 02X\n", wpdu)
	_, err = rwc.Write(wpdu)
	if nil != err {
		errorLog("io.Write() failed, err: %v\n", err)
		return err
	}
	debugLog("sending: ok")
	return nil
}

func hdlcTransportSend(rwc io.ReadWriteCloser, pdu []byte) error {
	var buf bytes.Buffer
	llcHeader := []byte{0xE6, 0xE6, 0x00} // LLC sublayer header

	_, err := buf.Write(llcHeader)
	if nil != err {
		errorLog("io.Write() failed, err: %v\n", err)
		return err
	}
	_, err = buf.Write(pdu)
	if nil != err {
		errorLog("io.Write() failed, err: %v\n", err)
		return err
	}

	p := buf.Bytes()
	debugLog("sending: % 02X\n", p)
	_, err = rwc.Write(p)
	if nil != err {
		errorLog("io.Write() failed, err: %v\n", err)
		return err
	}

	debugLog("sending: ok")
	return nil
}

func (dconn *DlmsConn) transportSend(src uint16, dst uint16, pdu []byte) error {
	debugLog("trnasport type: %d, src: %d, dst: %d\n", dconn.transportType, src, dst)

	if (Transport_TCP == dconn.transportType) || (Transport_UDP == dconn.transportType) {
		return ipTransportSend(dconn.rwc, src, dst, pdu)
	} else if Transport_HDLC == dconn.transportType {
		return hdlcTransportSend(dconn.rwc, pdu)
	} else {
		panic(fmt.Sprintf("unsupported transport type: %d", dconn.transportType))
	}
}

func ipTransportReceive(rwc io.ReadWriteCloser, srcWport *uint16, dstWport *uint16) (pdu []byte, src uint16, dst uint16, err error) {
	var (
		header tWrapperHeader
	)

	debugLog("receiving pdu ...\n")
	err = binary.Read(rwc, binary.BigEndian, &header)
	if nil != err {
		errorLog("binary.Read() failed, err: %v\n", err)
		return nil, 0, 0, err
	}
	debugLog("header: ok\n")
	if (nil != srcWport) && (header.SrcWport != *srcWport) {
		err = fmt.Errorf("wrong srcWport: %d, expected: %d", header.SrcWport, srcWport)
		errorLog("%s", err)
		return nil, 0, 0, err
	}
	if (nil != dstWport) && (header.DstWport != *dstWport) {
		err = fmt.Errorf("wrong dstWport: %d, expected: %d", header.DstWport, dstWport)
		errorLog("%s", err)
		return nil, 0, 0, err
	}
	pdu = make([]byte, header.DataLength)
	err = binary.Read(rwc, binary.BigEndian, pdu)
	if nil != err {
		errorLog("binary.Read() failed, err: %v\n", err)
		return nil, 0, 0, err
	}
	debugLog("received pdu: % 02X\n", pdu)

	return pdu, header.SrcWport, header.DstWport, nil
}

func hdlcTransportReceive(rwc io.ReadWriteCloser) (pdu []byte, err error) {

	debugLog("receiving pdu ...\n")

	//TODO: Set maxSegmnetSize to AARE.user-information.server-max-receive-pdu-size.
	// AARE.user-information is of 'InitiateResponse' asn1 type and is A-XDR encoded.
	maxSegmnetSize := 3 * 1024

	p := make([]byte, maxSegmnetSize)

	// hdlc ReadWriter read returns always whole segment into 'p' or full 'p' if 'p' is not long enough to fit in all segment
	n, err := rwc.Read(p)
	if nil != err {
		errorLog("hdlc.Read() failed, err: %v\n", err)
		return nil, err
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
		return nil, err
	}
	if !bytes.Equal(llcHeader, []byte{0xE6, 0xE7, 0x00}) {
		err = fmt.Errorf("wrong LLC header")
		errorLog("%s", err)
		return nil, err
	}
	debugLog("LLC header: ok\n")

	pdu = buf.Bytes()
	debugLog("received pdu: % 02X\n", pdu)

	return pdu, nil
}

func (dconn *DlmsConn) transportReceive(src uint16, dst uint16) (pdu []byte, err error) {
	debugLog("trnascport type: %d\n", dconn.transportType)

	if (Transport_TCP == dconn.transportType) || (Transport_UDP == dconn.transportType) {
		pdu, _, _, err = ipTransportReceive(dconn.rwc, &src, &dst)
		return pdu, err
	} else if Transport_HDLC == dconn.transportType {
		return hdlcTransportReceive(dconn.rwc)
	} else {
		err := fmt.Errorf("unsupported transport type: %d", dconn.transportType)
		errorLog("%s", err)
		return nil, err
	}
}

func (dconn *DlmsConn) AppConnectWithPassword(applicationClient uint16, logicalDevice uint16, invokeId uint8, password string) (aconn *AppConn, err error) {
	var aarq = AARQ{
		appCtxt:   LogicalName_NoCiphering,
		authMech:  LowLevelSecurity,
		authValue: password,
	}
	pdu, err := aarq.encode()
	if err != nil {
		return nil, err
	}

	err = dconn.transportSend(applicationClient, logicalDevice, pdu)
	if nil != err {
		return nil, err
	}
	pdu, err = dconn.transportReceive(logicalDevice, applicationClient)
	if nil != err {
		return nil, err
	}

	var aare AARE
	err = aare.decode(pdu)
	if err != nil {
		return nil, err
	}
	if aare.result != AssociationAccepted {
		err = fmt.Errorf("app connect failed, result: %v, diagnostic: %v", aare.result, aare.diagnostic)
		errorLog("%s", err)
		return nil, err
	} else {
		aconn = NewAppConn(dconn, applicationClient, logicalDevice, invokeId)
		return aconn, nil
	}

}

func (dconn *DlmsConn) AppConnectRaw(applicationClient uint16, logicalDevice uint16, invokeId uint8, aarq []byte, aare []byte) (aconn *AppConn, err error) {
	err = dconn.transportSend(applicationClient, logicalDevice, aarq)
	if nil != err {
		return nil, err
	}
	pdu, err := dconn.transportReceive(logicalDevice, applicationClient)
	if nil != err {
		return nil, err
	}
	if !bytes.Equal(pdu, aare) {
		err = errors.New("received unexpected AARE")
		return nil, err
	} else {
		aconn = NewAppConn(dconn, applicationClient, logicalDevice, invokeId)
		return aconn, nil
	}
}

func TcpConnect(ipAddr string, port int) (dconn *DlmsConn, err error) {
	var (
		conn net.Conn
	)

	dconn = new(DlmsConn)
	dconn.transportType = Transport_TCP

	debugLog("connecting tcp transport: %s:%d\n", ipAddr, port)
	conn, err = net.Dial("tcp", fmt.Sprintf("%s:%d", ipAddr, port))
	if nil != err {
		return nil, err
	}
	dconn.rwc = conn

	debugLog("tcp transport connected: %s:%d\n", ipAddr, port)
	return dconn, nil

}

/*
'responseTimeout' should be set to network roundtrip time if hdlc is used over
    unreliable transport and it should be set to eternity hdlc is used
    over reliable tcp.
    This timeout is part of hdlc error recovery function in case of lost or delayed
    frames over unreliable transport. In case of hdlc over reliable tcp
    this 'responseTimeout' should be set to eterinty
    to avoid unnecessary sending of RR frames.

Optional 'cosemWaitTime' should be set to average time what it takes for
    cosem layer to generate request or reply. This should be used only if hdlc
    is used for cosem and it serves
    avoiding of sending unnecessary RR frames.
*/
func HdlcConnect(ipAddr string, port int, applicationClient uint16, logicalDevice uint16, physicalDevice *uint16, responseTimeout time.Duration, cosemWaitTime *time.Duration, snrmTimeout time.Duration, discTimeout time.Duration) (dconn *DlmsConn, err error) {
	var (
		conn net.Conn
	)

	dconn = new(DlmsConn)
	dconn.transportType = Transport_HDLC

	debugLog("connecting hdlc transport over tcp: %s:%d\n", ipAddr, port)
	conn, err = net.Dial("tcp", fmt.Sprintf("%s:%d", ipAddr, port))
	if nil != err {
		errorLog("net.Dial() failed: %v", err)
		return nil, err
	}
	dconn.hdlcRwc = conn

	client := NewHdlcTransport(dconn.hdlcRwc, responseTimeout, true, uint8(applicationClient), logicalDevice, physicalDevice)
	dconn.hdlcResponseTimeout = responseTimeout
	dconn.snrmTimeout = snrmTimeout
	dconn.discTimeout = discTimeout

	if nil != cosemWaitTime {
		client.SetForCosem(*cosemWaitTime)
	}

	// send SNRM
	ch := make(chan error, 1)
	go func() {
		ch <- client.SendDSNRM(nil, nil)
	}()
	select {
	case err = <-ch:
		if nil != err {
			errorLog("client.SendSNRM() failed: %v", err)
			conn.Close()
			client.Close()
			return nil, err
		}
		dconn.HdlcClient = client
		dconn.rwc = client
	case <-time.After(dconn.snrmTimeout):
		errorLog("SendSNRM(): error timeout")
		conn.Close()
		client.Close()
		return nil, ErrDlmsTimeout
	}

	return dconn, nil
}

func (dconn *DlmsConn) Close() (err error) {
	debugLog("closing transport connection")

	switch dconn.transportType {
	case Transport_TCP:
		dconn.rwc.Close()
		return nil
	case Transport_HDLC:
		// send DISC
		ch := make(chan error, 1)
		go func() {
			ch <- dconn.HdlcClient.SendDISC()
		}()
		select {
		case err = <-ch:
			if nil != err {
				errorLog("SendDISC() failed: %v", err)
			}
		case <-time.After(dconn.discTimeout):
			errorLog("SendDISC(): error timeout")
			err = ErrDlmsTimeout
		}
		dconn.hdlcRwc.Close()
		dconn.HdlcClient.Close()
		return err
	}
	return ErrUnknownTransport
}
