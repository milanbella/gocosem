package gocosem

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
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

const (
	lowest_level_security_mechanism           = int(0)
	low_level_security_mechanism              = int(1)
	high_level_security_mechanism             = int(2)
	high_level_security_mechanism_using_MD5   = int(3)
	high_level_security_mechanism_using_SHA_1 = int(4)
	high_level_security_mechanism_using_GMAC  = int(5)
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
	systemTitle         []byte
	securityMechanismId int
	AK                  []byte // authentication key
	EK                  []byte // encryption key
	frameCounter        uint32
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
	debugLog("sending: %02X\n", p)
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

func cosemTagToGsmTag(tag1 byte) (err error, tag byte) {
	if tag1 == 1 {
		return nil, 33
	}
	if tag1 == 5 {
		return nil, 37
	}
	if tag1 == 6 {
		return nil, 38
	}
	if tag1 == 8 {
		return nil, 40
	}
	if tag1 == 12 {
		return nil, 44
	}
	if tag1 == 13 {
		return nil, 45
	}
	if tag1 == 14 {
		return nil, 46
	}
	if tag1 == 22 {
		return nil, 54
	}
	if tag1 == 24 {
		return nil, 56
	}
	if tag1 == 192 {
		return nil, 200
	}
	if tag1 == 193 {
		return nil, 201
	}
	if tag1 == 194 {
		return nil, 202
	}
	if tag1 == 199 {
		return nil, 203
	}
	if tag1 == 196 {
		return nil, 204
	}
	if tag1 == 197 {
		return nil, 205
	}
	if tag1 == 199 {
		return nil, 207
	}
	err = fmt.Errorf("unknown tag")
	errorLog("%s", err)
	return err, 0
}

func gsmTagToCosemTag(tag1 byte) (err error, tag byte) {
	if tag1 == 33 {
		return nil, 1
	}
	if tag1 == 37 {
		return nil, 5
	}
	if tag1 == 38 {
		return nil, 6
	}
	if tag1 == 40 {
		return nil, 8
	}
	if tag1 == 44 {
		return nil, 12
	}
	if tag1 == 45 {
		return nil, 13
	}
	if tag1 == 46 {
		return nil, 14
	}
	if tag1 == 54 {
		return nil, 22
	}
	if tag1 == 56 {
		return nil, 24
	}
	if tag1 == 200 {
		return nil, 192
	}
	if tag1 == 201 {
		return nil, 193
	}
	if tag1 == 202 {
		return nil, 194
	}
	if tag1 == 203 {
		return nil, 199
	}
	if tag1 == 204 {
		return nil, 196
	}
	if tag1 == 205 {
		return nil, 197
	}
	if tag1 == 207 {
		return nil, 199
	}
	err = fmt.Errorf("unknown tag")
	errorLog("%s", err)
	return err, 0
}
func (dconn *DlmsConn) encryptPduGSM(pdu []byte) (err error, epdu []byte) {

	// enforce 128 bit keys
	if len(dconn.AK) != 16 {
		err = fmt.Errorf("authentication key length is not 16")
		errorLog("%s", err)
		return err, nil
	}
	if len(dconn.EK) != 16 {
		err = fmt.Errorf("encryption key length is not 16")
		errorLog("%s", err)
		return err, nil
	}

	// tag
	err, tag := cosemTagToGsmTag(pdu[0])
	if nil != err {
		return err, nil
	}

	// security control
	SC := byte(0x30) // security control

	// frame counter
	dconn.frameCounter += 1
	FC := make([]byte, 4)
	FC[0] = byte(dconn.frameCounter >> 24 & 0xFF)
	FC[1] = byte(dconn.frameCounter >> 16 & 0xFF)
	FC[2] = byte(dconn.frameCounter >> 8 & 0xFF)
	FC[3] = byte(dconn.frameCounter & 0xFF)

	// initialization vector
	IV := make([]byte, 12) // initialization vector
	if len(dconn.systemTitle) != 8 {
		err = fmt.Errorf("system title length is not 8")
		errorLog("%s", err)
		return err, nil
	}
	copy(IV, dconn.systemTitle)

	// additional authenticated data
	AAD := make([]byte, 1+len(dconn.AK))
	AAD[0] = SC
	copy(AAD[1:], dconn.AK)

	block, err := aes.NewCipher(dconn.EK)
	if err != nil {
		return err, nil
	}

	aesgcm, err := cipher.NewGCMWithNonceSize(block, len(IV))
	if err != nil {
		return err, nil
	}
	ciphertext := aesgcm.Seal(nil, IV, pdu, AAD)

	length := 1 + len(FC) + len(ciphertext)

	epdu = make([]byte, 1+1+1+len(FC)+len(ciphertext))

	epdu[0] = tag
	if length > 0xFF {
		warnLog("length exceeds 255")
	}
	epdu[1] = byte(length)
	epdu[2] = SC
	copy(epdu[3:], FC)
	copy(epdu[3+len(FC):], ciphertext)

	return nil, epdu
}

func (dconn *DlmsConn) encryptPdu(authenticationMechanismId int, pdu []byte) (err error, epdu []byte) {
	if authenticationMechanismId == high_level_security_mechanism_using_GMAC {
		return dconn.encryptPduGSM(pdu)
	} else {
		err = fmt.Errorf("authentication mechanism %v not supported", authenticationMechanismId)
		errorLog("%s", err)
		return err, nil
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

func (dconn *DlmsConn) AppConnectWithSecurity5(applicationClient uint16, logicalDevice uint16, invokeId uint8, applicationContextName []uint32, callingAPtitle []byte, clientToServerChallenge string, userInformation []byte) (aconn *AppConn, err error) {

	var aarq AARQapdu

	aarq.applicationContextName = tAsn1ObjectIdentifier(applicationContextName)
	_callingAPtitle := tAsn1OctetString(callingAPtitle)
	aarq.callingAPtitle = &_callingAPtitle
	aarq.senderAcseRequirements = &tAsn1BitString{
		buf:        []byte{0x80},
		bitsUnused: 7,
	}
	mechanismName := (tAsn1ObjectIdentifier)([]uint32{2, 16, 756, 5, 8, 2, 5})
	aarq.mechanismName = &mechanismName
	aarq.callingAuthenticationValue = new(tAsn1Choice)
	aarq.callingAuthenticationValue.setVal(0, tAsn1GraphicString([]byte(clientToServerChallenge)))
	_userInformation := tAsn1OctetString(userInformation)
	aarq.userInformation = &_userInformation

	var buf *bytes.Buffer
	buf = new(bytes.Buffer)
	err = encode_AARQapdu(buf, &aarq)
	if nil != err {
		return nil, err
	}
	pdu := buf.Bytes()

	err = dconn.transportSend(applicationClient, logicalDevice, pdu)
	if nil != err {
		return nil, err
	}
	pdu, err = dconn.transportReceive(logicalDevice, applicationClient)
	if nil != err {
		return nil, err
	}

	buf = bytes.NewBuffer(pdu)
	err, aare := decode_AAREapdu(buf)
	if nil != err {
		return nil, err
	}

	// verify AARE

	if aare.result != 0 {
		err = fmt.Errorf("app connect failed: verify AARE: result %v", aare.result)
		errorLog("%s", err)
		return nil, err
	}
	if !(aare.resultSourceDiagnostic.tag == 1 && aare.resultSourceDiagnostic.val.(tAsn1Integer) == tAsn1Integer(14)) { // 14 - authentication-required
		err = fmt.Errorf("app connect failed: verify AARE: meter did not require authentication")
		errorLog("%s", err)
		return nil, err
	}
	if aare.mechanismName == nil {
		err = fmt.Errorf("app connect failed: verify AARE: meter did not require proper authentication mechanism id: mechanism_id(5)")
		errorLog("%s", err)
		return nil, err
	}
	oi := ([]uint32)(*aare.mechanismName)
	if !(oi[0] == 2 && oi[1] == 16 && oi[2] == 756 && oi[3] == 5 && oi[4] == 8 && oi[5] == 2 && oi[6] == 5) {
		err = fmt.Errorf("app connect failed: verify AARE: meter did not require proper authentication mechanism id: mechanism_id(5)")
		errorLog("%s", err)
		return nil, err
	}

	dconn.systemTitle = callingAPtitle
	dconn.securityMechanismId = high_level_security_mechanism_using_GMAC

	aconn = NewAppConn(dconn, applicationClient, logicalDevice, invokeId)
	return aconn, nil

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
