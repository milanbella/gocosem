package gocosem

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"time"
	"unsafe"
)

type tDlmsInvokeIdAndPriority uint8
type tDlmsClassId uint16
type tDlmsOid [6]uint8
type tDlmsAttributeId uint8
type tDlmsAccessSelector uint8
type tDlmsData tAsn1Choice

type tDlmsDataAccessResult uint8

const (
	dataAccessResult_success                 = 0
	dataAccessResult_hardwareFault           = 1
	dataAccessResult_temporaryFailure        = 2
	dataAccessResult_readWriteDenied         = 3
	dataAccessResult_objectUndefined         = 4
	dataAccessResult_objectClassInconsistent = 9
	dataAccessResult_objectUnavailable       = 11
	dataAccessResult_typeUnmatched           = 12
	dataAccessResult_scopeOfAccessViolated   = 13
	dataAccessResult_dataBlockUnavailable    = 14
	dataAccessResult_longGetAborted          = 15
	dataAccessResult_noLongGetInProgress     = 16
	dataAccessResult_longSetAborted          = 17
	dataAccessResult_noLongSetInProgress     = 18
	dataAccessResult_dataBlockNumberInvalid  = 19
	dataAccessResult_otherReason             = 250
)

var errorLog *log.Logger = getErrorLogger()
var debugLog *log.Logger = getDebugLogger()

func getRequest(classId tDlmsClassId, instanceId *tDlmsOid, attributeId tDlmsAttributeId, accessSelector *tDlmsAccessSelector, accessParameters *tDlmsData) (err error, pdu []byte) {
	var FNAME = "getRequest()"

	var w bytes.Buffer

	var buf bytes.Buffer
	err = binary.Write(&buf, binary.BigEndian, classId)
	if nil != err {
		errorLog.Printf(fmt.Sprintf("%s: binary.Write() failed, err: %s\n", err))
		return err, nil
	}
	b := buf.Bytes()
	_, err = w.Write(b)
	if nil != err {
		errorLog.Printf("%s: w.Wite() failed, err: %v\n", FNAME, err)
		return err, nil
	}

	_, err = w.Write((*instanceId)[0:6])
	if nil != err {
		errorLog.Printf("%s: w.Wite() failed, err: %v\n", FNAME, err)
		return err, nil
	}

	_, err = w.Write([]byte{byte(attributeId)})
	if nil != err {
		errorLog.Printf("%s: w.Wite() failed, err: %v\n", FNAME, err)
		return err, nil
	}

	if 0 != attributeId {
		var as []byte
		var ap []byte
		if nil == accessSelector {
			as = []byte{0}
		} else {
			as = []byte{byte(*accessSelector)}
		}
		if nil != accessParameters {
			err, ap = encode_Data((*tAsn1Choice)(accessParameters))
			if nil != err {
				return err, nil
			}
		} else {
			ap = make([]byte, 0)
		}

		_, err = w.Write(as)
		if nil != err {
			errorLog.Printf("%s: w.Wite() failed, err: %v\n", FNAME, err)
			return err, nil
		}

		_, err = w.Write(ap)
		if nil != err {
			errorLog.Printf("%s: w.Wite() failed, err: %v\n", FNAME, err)
			return err, nil
		}
	}

	return nil, w.Bytes()
}

func getResponse(pdu []byte) (err error, n int, dataAccessResult tDlmsDataAccessResult, data *tDlmsData) {
	var FNAME = "getResponse()"
	var serr string
	var nn = 0

	b := pdu[0:]
	n = 0

	if len(b) < 1 {
		serr = fmt.Sprintf("%s: short pdu", FNAME)
		errorLog.Println(serr)
		return errors.New(serr), 0, 0, nil
	}
	dataAccessResult = tDlmsDataAccessResult(b[0])
	b = b[1:]
	n += 1

	var cdata *tAsn1Choice
	if dataAccessResult_success == dataAccessResult {
		err, cdata, nn = decode_Data(b)
		if nil != err {
			return err, n + nn, 0, nil
		}
		n += nn
	}

	return nil, n, dataAccessResult, (*tDlmsData)(cdata)
}

func encode_GetRequestNormal(invokeIdAndPriority tDlmsInvokeIdAndPriority, classId tDlmsClassId, instanceId *tDlmsOid, attributeId tDlmsAttributeId, accessSelector *tDlmsAccessSelector, accessParameters *tDlmsData) (err error, pdu []byte) {
	var FNAME = "encode_GetRequestNormal()"

	var w bytes.Buffer

	_, err = w.Write([]byte{0xc0, 0x01})
	if nil != err {
		errorLog.Printf("%s: w.Wite() failed, err: %v\n", FNAME, err)
		return err, nil
	}

	_, err = w.Write([]byte{byte(invokeIdAndPriority)})
	if nil != err {
		errorLog.Printf("%s: w.Wite() failed, err: %v\n", FNAME, err)
		return err, nil
	}

	err, pdu = getRequest(classId, instanceId, attributeId, accessSelector, accessParameters)
	if nil != err {
		errorLog.Printf("%s: getRequest() failed, err: %v\n", FNAME, err)
		return err, nil
	}

	_, err = w.Write(pdu)
	if nil != err {
		errorLog.Printf("%s: w.Wite() failed, err: %v\n", FNAME, err)
		return err, nil
	}

	return nil, w.Bytes()
}

func decode_GetResponseNormal(pdu []byte) (err error, invokeIdAndPriority tDlmsInvokeIdAndPriority, dataAccessResult tDlmsDataAccessResult, data *tDlmsData) {
	var FNAME = "decode_GetResponsenormal()"
	var serr string
	b := pdu[0:]

	if len(b) < 2 {
		return errors.New("short pdu"), 0, 0, nil
	}
	if !bytes.Equal(b[0:2], []byte{0xC4, 0x01}) {
		errorLog.Printf("%s: pdu is not GetResponsenormal: 0x%02X 0x%02X\n", FNAME, b[0], b[1])
		return errors.New("pdu is not GetResponsenormal"), 0, 0, nil
	}
	b = b[2:]

	if len(b) < 1 {
		serr = fmt.Sprintf("%s: short pdu", FNAME)
		errorLog.Println(serr)
		return errors.New(serr), 0, 0, nil
	}
	invokeIdAndPriority = tDlmsInvokeIdAndPriority(b[0])
	b = b[1:]

	err, _, dataAccessResult, data = getResponse(b)
	if nil != err {
		return err, 0, 0, nil
	}

	return nil, invokeIdAndPriority, dataAccessResult, data
}

func encode_GetRequestWithList(invokeIdAndPriority tDlmsInvokeIdAndPriority, classIds []tDlmsClassId, instanceIds []*tDlmsOid, attributeIds []tDlmsAttributeId, accessSelectors []*tDlmsAccessSelector, accessParameters []*tDlmsData) (err error, pdu []byte) {
	var FNAME = "encode_GetRequestWithList()"

	var w bytes.Buffer

	_, err = w.Write([]byte{0xc0, 0x03})
	if nil != err {
		errorLog.Printf("%s: w.Wite() failed, err: %v\n", FNAME, err)
		return err, nil
	}

	_, err = w.Write([]byte{byte(invokeIdAndPriority)})
	if nil != err {
		errorLog.Printf("%s: w.Wite() failed, err: %v\n", FNAME, err)
		return err, nil
	}

	count := len(classIds) // count of get requests

	_, err = w.Write([]byte{byte(count)})
	if nil != err {
		errorLog.Printf("%s: w.Wite() failed, err: %v\n", FNAME, err)
		return err, nil
	}

	for i := 0; i < count; i += 1 {

		err, pdu = getRequest(classIds[i], instanceIds[i], attributeIds[i], accessSelectors[i], accessParameters[i])
		if nil != err {
			errorLog.Printf("%s: getRequest() failed, err: %v\n", FNAME, err)
			return err, nil
		}

		_, err = w.Write(pdu)
		if nil != err {
			errorLog.Printf("%s: w.Wite() failed, err: %v\n", FNAME, err)
			return err, nil
		}
	}

	return nil, w.Bytes()
}

func decode_GetResponseWithList(pdu []byte) (err error, invokeIdAndPriority tDlmsInvokeIdAndPriority, dataAccessResults []tDlmsDataAccessResult, datas []*tDlmsData) {
	var FNAME = "decode_GetResponseWithList()"
	var serr string
	b := pdu[0:]

	if len(b) < 2 {
		serr = fmt.Sprintf("%s: short pdu", FNAME)
		errorLog.Println(serr)
		return errors.New(serr), 0, nil, nil
	}
	if !bytes.Equal(b[0:2], []byte{0xC4, 0x03}) {
		errorLog.Printf("%s: pdu is not GetResponseWithList: 0x%02X 0x%02X\n", FNAME, b[0], b[1])
		return errors.New("pdu is not GetResponseWithList"), 0, nil, nil
	}
	b = b[2:]

	if len(b) < 1 {
		serr = fmt.Sprintf("%s: short pdu", FNAME)
		errorLog.Println(serr)
		return errors.New(serr), 0, nil, nil
	}
	invokeIdAndPriority = tDlmsInvokeIdAndPriority(b[0])
	b = b[1:]

	if len(b) < 1 {
		serr = fmt.Sprintf("%s: short pdu", FNAME)
		errorLog.Println(serr)
		return errors.New(serr), 0, nil, nil
	}
	count := int(b[0])
	b = b[1:]

	dataAccessResults = make([]tDlmsDataAccessResult, count)
	datas = make([]*tDlmsData, count)

	var dataAccessResult tDlmsDataAccessResult
	var data *tDlmsData
	var n int
	for i := 0; i < count; i += 1 {
		err, n, dataAccessResult, data = getResponse(b)
		if nil != err {
			return err, 0, nil, nil
		}
		b = b[n:]
		dataAccessResults[i] = dataAccessResult
		datas[i] = data
	}

	return nil, invokeIdAndPriority, dataAccessResults, datas
}

func decode_GetResponsewithDataBlock(pdu []byte) (err error, invokeIdAndPriority tDlmsInvokeIdAndPriority, lastBlock bool, blockNumber uint32, dataAccessResult tDlmsDataAccessResult, rawData []byte) {
	var FNAME = "decode_GetResponsewithDataBlock()"
	var serr string
	b := pdu[0:]

	if len(b) < 2 {
		serr = fmt.Sprintf("%s: short pdu", FNAME)
		errorLog.Println(serr)
		return errors.New(serr), 0, false, 0, 0, nil
	}
	if !bytes.Equal(b[0:2], []byte{0xC4, 0x02}) {
		serr = fmt.Sprintf("%s: pdu is not GetResponsewithDataBlock: 0x%02X 0x%02X ", FNAME, b[0], b[1])
		errorLog.Println(serr)
		return errors.New(serr), 0, false, 0, 0, nil
	}
	b = b[2:]

	if len(b) < 1 {
		serr = fmt.Sprintf("%s: short pdu", FNAME)
		errorLog.Println(serr)
		return errors.New(serr), 0, false, 0, 0, nil
	}
	invokeIdAndPriority = tDlmsInvokeIdAndPriority(b[0])
	b = b[1:]

	if len(b) < 1 {
		serr = fmt.Sprintf("%s: short pdu", FNAME)
		errorLog.Println(serr)
		return errors.New(serr), 0, false, 0, 0, nil
	}
	if 0 == b[0] {
		lastBlock = false
	} else {
		lastBlock = true
	}
	b = b[1:]

	if len(b) < 4 {
		serr = fmt.Sprintf("%s: short pdu", FNAME)
		errorLog.Println(serr)
		return errors.New(serr), 0, false, 0, 0, nil
	}
	err = binary.Read(bytes.NewBuffer(b[0:4]), binary.BigEndian, &blockNumber)
	if nil != err {
		errorLog.Println("%s: binary.Read() failed, err: %v", err)
		return err, 0, false, 0, 0, nil
	}
	b = b[4:]

	if len(b) < 1 {
		serr = fmt.Sprintf("%s: short pdu", FNAME)
		errorLog.Println(serr)
		return errors.New(serr), 0, false, 0, 0, nil
	}
	dataAccessResult = tDlmsDataAccessResult(b[0])
	b = b[1:]

	if len(b) < 1 {
		serr = fmt.Sprintf("%s: short pdu", FNAME)
		errorLog.Println(serr)
		return errors.New(serr), 0, false, 0, 0, nil
	}
	tag := b[0]
	b = b[1:]

	if 0x1E != tag {
		serr = fmt.Sprintf("%s: wrong raw data tag: 0X%02X", FNAME)
		errorLog.Println(serr)
		return errors.New(serr), 0, false, 0, 0, nil
	}

	rawData = b

	return nil, invokeIdAndPriority, lastBlock, blockNumber, dataAccessResult, rawData
}

func encode_GetRequestForNextDataBlock(invokeIdAndPriority tDlmsInvokeIdAndPriority, blockNumber uint32) (err error, pdu []byte) {
	var FNAME = "encode_GetRequestForNextDataBlock()"

	var w bytes.Buffer

	_, err = w.Write([]byte{0xc0, 0x02})
	if nil != err {
		errorLog.Printf("%s: w.Wite() failed, err: %v\n", FNAME, err)
		return err, nil
	}

	_, err = w.Write([]byte{byte(invokeIdAndPriority)})
	if nil != err {
		errorLog.Printf("%s: w.Wite() failed, err: %v\n", FNAME, err)
		return err, nil
	}

	var buf bytes.Buffer
	err = binary.Write(&buf, binary.BigEndian, blockNumber)
	if nil != err {
		errorLog.Printf(fmt.Sprintf("%s: binary.Write() failed, err: %s\n", err))
		return err, nil
	}
	b := buf.Bytes()
	_, err = w.Write(b)
	if nil != err {
		errorLog.Printf("%s: w.Wite() failed, err: %v\n", FNAME, err)
		return err, nil
	}

	return nil, w.Bytes()
}

const (
	COSEM_lowest_level_security_mechanism_name           = uint(0)
	COSEM_low_level_security_mechanism_name              = uint(1)
	COSEM_high_level_security_mechanism_name             = uint(2)
	COSEM_high_level_security_mechanism_name_using_MD5   = uint(3)
	COSEM_high_level_security_mechanism_name_using_SHA_1 = uint(4)
	COSEM_High_Level_Security_Mechanism_Name_Using_GMAC  = uint(5)
)

const (
	Logical_Name_Referencing_No_Ciphering   = uint(1)
	Short_Name_Referencing_No_Ciphering     = uint(2)
	Logical_Name_Referencing_With_Ciphering = uint(3)
	Short_Name_Referencing_With_Ciphering   = uint(4)
)

const (
	ACSE_Requirements_authentication = byte(0x80) // bit 0
)

const (
	Transport_HLDC = int(1)
	Transport_UDP  = int(2)
	Transport_TCP  = int(3)
)

type DlmsChannelMessage struct {
	err  error
	data interface{}
}

type DlmsChannel chan *DlmsChannelMessage

type tWrapperHeader struct {
	protocolVersion uint16
	srcWport        uint16
	dstWport        uint16
	dataLength      uint16
}

func (header *tWrapperHeader) String() string {
	return fmt.Sprintf("tWrapperHeader {protocolVersion: %d, srcWport: %s, dstWport: %d, dataLength: %d}")
}

type tWrapperPdu struct {
	tWrapperHeader
	pdu []byte
}

type tDlmsValueRequest struct {
	classId         tDlmsClassId
	instanceId      *tDlmsOid
	attributeId     tDlmsAttributeId
	accessSelector  *tDlmsAccessSelector
	accessParameter *tDlmsData
}

//func encode_GetRequestNormal(invokeIdAndPriority tDlmsInvokeIdAndPriority, classId tDlmsClassId, instanceId *tDlmsOid, attributeId tDlmsAttributeId, accessSelector *tDlmsAccessSelector, accessParameters *tDlmsData) (err error, pdu []byte) {
type tDlmsRequest struct {
	invokeIdAndPriority tDlmsInvokeIdAndPriority
	valueRequests       []*tDlmsValueRequest
}

type DlmsConn struct {
	rwc           io.ReadWriteCloser
	transportType int
	ch            DlmsChannel // channel used to serialize inbound requests
}

type DlmsTransportSendRequest struct {
	ch                DlmsChannel // reply channel
	applicationClient uint16
	logicalDevice     uint16
	pdu               []byte
}

type DlmsTransportReceiveRequest struct {
	ch DlmsChannel // reply channel
}

var ErrorDlmsTimeout = errors.New("ErrorDlmsTimeout")

// Never call this method directly or else you risk race condtitions on io.Writer() in case of paralell call.
// Use instead proxy variant 'transportSend()' which queues this method call on sync channel.

func (dconn *DlmsConn) doTransportSend(ch DlmsChannel, applicationClient uint16, logicalDevice uint16, pdu []byte) {
	var (
		FNAME string = "transportSend()"
	)

	if (Transport_TCP == dconn.transportType) || (Transport_UDP == dconn.transportType) {
		err, wpdu := dconn.makeWpdu(applicationClient, logicalDevice, pdu)
		if nil != err {
			ch <- &DlmsChannelMessage{err, nil}
			return
		}
		_, err = dconn.rwc.Write(wpdu)
		if nil != err {
			errorLog.Printf("%s: io.Write() failed, err: %v\n", FNAME, err)
			ch <- &DlmsChannelMessage{err, nil}
			return
		}
		ch <- &DlmsChannelMessage{nil, nil}
	} else {
		panic(fmt.Sprintf("%s: unsupported transport type: %d", FNAME, dconn.transportType))
	}
	panic("assertion failed")
}

func (dconn *DlmsConn) transportSend(ch DlmsChannel, applicationClient uint16, logicalDevice uint16, pdu []byte) {
	msg := new(DlmsChannelMessage)

	data := new(DlmsTransportSendRequest)
	data.ch = ch
	data.applicationClient = applicationClient
	data.logicalDevice = logicalDevice
	data.pdu = pdu

	msg.data = data

	dconn.ch <- msg
}

func readLength(r io.Reader, length int) (err error, data []byte) {
	var (
		buf bytes.Buffer
		n   int
	)

	p := make([]byte, length)
	for {
		n, err = r.Read(p[0 : length-buf.Len()])
		if n > 0 {
			buf.Write(p[0:n])
			if length == buf.Len() {
				return nil, data
			} else if length < buf.Len() {
				panic("assertion failed")
			} else {
			}
		} else if 0 == n {
			if nil != err {
				errorLog.Printf("%s: io.Read() failed, err: %v", err)
				return err, data
			} else {
				panic("assertion failed")
			}
		} else {
			panic("assertion failed")
		}
	}
	panic("assertion failed")
}

// Never call this method directly or else you risk race condtitions on io.Writer() in case of paralell call.
// Use instead proxy variant 'transportReceive()' which queues this method call on sync channel.

func (dconn *DlmsConn) doTransportReceive(ch DlmsChannel) {
	var (
		FNAME     string = "transportRecive()"
		serr      string
		err       error
		headerPdu []byte
		header    tWrapperHeader
	)

	if (Transport_TCP == dconn.transportType) || (Transport_UDP == dconn.transportType) {
		err, headerPdu = readLength(dconn.rwc, int(unsafe.Sizeof(header)))
		if nil != err {
			ch <- &DlmsChannelMessage{err, nil}
			return
		}
		err = binary.Read(bytes.NewBuffer(headerPdu), binary.BigEndian, header)
		if nil != err {
			errorLog.Printf("%s: binary.Read() failed, err: %v", err)
			ch <- &DlmsChannelMessage{err, nil}
			return
		}
		if header.dataLength <= 0 {
			serr = fmt.Sprintf("%s: wrong pdu length: %d", FNAME, header.dataLength)
			errorLog.Println(serr)
			ch <- &DlmsChannelMessage{errors.New(serr), nil}
			return
		}
		debugLog.Printf("%s: %s", FNAME, header)
		err, pdu := readLength(dconn.rwc, int(header.dataLength))
		if nil != err {
			ch <- &DlmsChannelMessage{err, nil}
			return
		}
		debugLog.Printf("%s: pdu: %X\n", FNAME, pdu)
		ch <- &DlmsChannelMessage{nil, pdu}
		return

	} else {
		serr = fmt.Sprintf("%s: unsupported transport type: %d", FNAME, dconn.transportType)
		errorLog.Println(serr)
		ch <- &DlmsChannelMessage{errors.New(serr), nil}
		return
	}
}

func (dconn *DlmsConn) transportReceive(ch DlmsChannel) {
	msg := new(DlmsChannelMessage)
	data := new(DlmsTransportReceiveRequest)
	data.ch = ch
	msg.data = data
	dconn.ch <- msg
}

func (dconn *DlmsConn) handleTransportRequests() {

	for {
		msg := <-dconn.ch
		switch v := msg.data.(type) {
		case *DlmsTransportSendRequest:
			dconn.transportSend(v.ch, v.applicationClient, v.logicalDevice, v.pdu)
		case *DlmsTransportReceiveRequest:
			dconn.transportReceive(v.ch)
		default:
			panic(fmt.Sprintf("unknown request type: %T", v))
		}
	}
}

func (dconn *DlmsConn) AppConnectWithPassword(ch DlmsChannel, msecTimeout int64, applicationClient uint16, logicalDevice uint16, password string) {
	var (
		serr string
		err  error
		aarq AARQapdu
		pdu  []byte
	)

	_ch := make(DlmsChannel)
	go func() {

		aarq.applicationContextName = tAsn1ObjectIdentifier([]uint{2, 16, 756, 5, 8, 1, Logical_Name_Referencing_No_Ciphering})
		aarq.senderAcseRequirements = &tAsn1BitString{
			buf:        []byte{ACSE_Requirements_authentication},
			bitsUnused: 7,
		}
		mechanismName := (tAsn1ObjectIdentifier)([]uint{2, 16, 756, 5, 8, 2, COSEM_low_level_security_mechanism_name})
		aarq.mechanismName = &mechanismName
		aarq.callingAuthenticationValue = new(tAsn1Choice)
		_password := tAsn1GraphicString([]byte(password))
		aarq.callingAuthenticationValue.setVal(int(C_Authentication_value_PR_charstring), &_password)

		//TODO A-XDR encoding of userInformation
		userInformation := tAsn1OctetString([]byte{0x01, 0x00, 0x00, 0x00, 0x06, 0x5F, 0x1F, 0x04, 0x00, 0x00, 0x7E, 0x1F, 0x04, 0xB0})

		aarq.userInformation = &userInformation

		err, pdu = encode_AARQapdu(&aarq)
		if nil != err {
			_ch <- &DlmsChannelMessage{err, nil}
			return
		}

		dconn.transportSend(_ch, applicationClient, logicalDevice, pdu)
		msg := <-_ch
		if nil != msg.err {
			_ch <- &DlmsChannelMessage{msg.err, nil}
			return
		}
		dconn.transportReceive(_ch)
		msg = <-ch
		if nil != msg.err {
			_ch <- &DlmsChannelMessage{msg.err, nil}
			return
		}
		err, aare := decode_AAREapdu((msg.data).([]byte))
		if nil != err {
			_ch <- &DlmsChannelMessage{msg.err, nil}
			return
		}
		if C_Association_result_accepted != int(aare.result) {
			serr = fmt.Sprintf("%s: app connect failed, aare.result %d, aare.resultSourceDiagnostic: %d", aare.result, aare.resultSourceDiagnostic)
			errorLog.Println(serr)
			_ch <- &DlmsChannelMessage{errors.New(serr), nil}
			return
		} else {
			_ch <- &DlmsChannelMessage{nil, nil}
		}

	}()

	select {
	case msg := <-_ch:
		if nil == msg.err {
			ch <- &DlmsChannelMessage{msg.err, &AppConn{dconn, applicationClient, logicalDevice}}
		} else {
			ch <- &DlmsChannelMessage{msg.err, nil}
		}
	case <-time.After(time.Millisecond * time.Duration(msecTimeout)):
		ch <- &DlmsChannelMessage{ErrorDlmsTimeout, nil}
	}

}

func (dconn *DlmsConn) makeWpdu(applicationClient uint16, logicalDevice uint16, pdu []byte) (err error, wpdu []byte) {
	var (
		FNAME   string = "makeWpdu()"
		buf     bytes.Buffer
		wrapper tWrapperPdu
	)

	wrapper.protocolVersion = 0x00001
	wrapper.srcWport = applicationClient
	wrapper.dstWport = logicalDevice
	wrapper.dataLength = uint16(len(pdu))
	wrapper.pdu = pdu

	err = binary.Write(&buf, binary.BigEndian, &wpdu)
	if nil != err {
		errorLog.Printf("%s:  binary.Write() failed, err: %v\n", FNAME, err)
		return err, nil
	}
	return nil, buf.Bytes()

}

func TcpConnect(ch DlmsChannel, msecTimeout int64, ipAddr string, port int) {
	var (
		FNAME string = "connectTCP()"
		conn  net.Conn
		err   error
	)

	dconn := new(DlmsConn)
	_ch := make(DlmsChannel)

	go func() {

		debugLog.Printf("%s: connecting tcp transport: %s:%d\n", ipAddr, port)
		conn, err = net.Dial("tcp", fmt.Sprintf("%s:%d", ipAddr, port))
		if nil != err {
			errorLog.Printf("%s: net.Dial() failed, err: %v", FNAME, err)
			_ch <- &DlmsChannelMessage{err, nil}
			return
		}
		//./dlms.go:731: undefined: dconn in dconn.rwc
		dconn.rwc = conn
	}()

	select {
	case msg := <-_ch:
		if nil == msg.err {
			debugLog.Printf("%s: tcp transport connected: %s:%d\n", ipAddr, port)
			dconn.ch = make(DlmsChannel)
			go dconn.handleTransportRequests()
		} else {
			debugLog.Printf("%s: tcp transport connection failed: %s:%d, err: %v\n", ipAddr, port, msg.err)
			ch <- &DlmsChannelMessage{msg.err, msg.data}
		}
	case <-time.After(time.Millisecond * time.Duration(msecTimeout)):
		errorLog.Printf("%s: tcp transport connection time out: %s:%d\n", ipAddr, port)
		ch <- &DlmsChannelMessage{ErrorDlmsTimeout, nil}
	}

}
