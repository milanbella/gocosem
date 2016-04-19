package gocosem

import (
	"bytes"
	"container/list"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

var hdlcDebug = true

const (
	HDLC_FRAME_DIRECTION_CLIENT_INBOUND  = 1
	HDLC_FRAME_DIRECTION_CLIENT_OUTBOUND = 2
	HDLC_FRAME_DIRECTION_SERVER_INBOUND  = 3
	HDLC_FRAME_DIRECTION_SERVER_OUTBOUND = 4
)

const (
	HDLC_ADDRESS_LENGTH_1 = 1
	HDLC_ADDRESS_LENGTH_2 = 2
	HDLC_ADDRESS_LENGTH_4 = 4
)

const (
	HDLC_CONTROL_I    = 1 // I frame
	HDLC_CONTROL_RR   = 2 // response ready
	HDLC_CONTROL_RNR  = 3 // response not ready
	HDLC_CONTROL_SNRM = 4 // set normal response mode
	HDLC_CONTROL_DISC = 5 // disconnect
	HDLC_CONTROL_UA   = 6 // unnumbered acknowledgement
	HDLC_CONTROL_DM   = 7 // disconnected mode
	HDLC_CONTROL_FRMR = 8 // frame reject
	HDLC_CONTROL_UI   = 9 // unnumbered information
)

type HdlcTransport struct {
	rw                         net.Conn
	responseTimeout            time.Duration
	modulus                    uint8
	maxInfoFieldLengthTransmit uint8
	maxInfoFieldLengthReceive  uint8

	windowSizeTransmit uint32
	windowSizeReceive  uint32

	serverAddrLength int // HDLC_ADDRESS_BYTE_LENGTH_1, HDLC_ADDRESS_BYTE_LENGTH_2, HDLC_ADDRESS_BYTE_LENGTH_4

	logicalDeviceId  uint16
	physicalDeviceId *uint16 // may not be present
	clientId         uint8

	writeQueue      *list.List // list of *HdlcSegment
	writeQueueMtx   *sync.Mutex
	writeAck        chan map[string]interface{}
	readQueue       *list.List // list of *HdlcSegment
	readQueueMtx    *sync.Mutex
	readAck         chan map[string]interface{}
	controlQueue    *list.List // list of *HdlcControlCommand
	controlAck      chan map[string]interface{}
	controlQueueMtx *sync.Mutex
	closed          bool
	closedMtx       *sync.Mutex
}

type HdlcClientConnection struct {
	htran *HdlcTransport
}

type HdlcServerConnection struct {
	htran *HdlcTransport
	vs    uint8 // V(S) - send sequence variable
	vr    uint8 // V(R) - receive sequence variable
}

type HdlcFrame struct {
	direction             int
	formatType            uint8
	segmentation          bool
	length                int
	logicalDeviceId       uint16
	physicalDeviceId      *uint16 // may not be present
	clientId              uint8
	poll                  bool  // poll/final bit
	nr                    uint8 // N(R) - receive sequence number
	ns                    uint8 // N(S) - send sequence number
	control               int
	fcs16                 uint16 // current fcs16 checksum
	infoField             []byte // information
	infoFieldFormat       uint8
	callingPhysicalDevice bool
	content               *bytes.Buffer
}

type HdlcSegment struct {
	p    []byte
	last bool
}

type HdlcControlCommand struct {
	control int
	snrm    *HdlcControlCommandSNRM
}

type HdlcControlCommandSNRM struct {
	maxInfoFieldLengthTransmit uint8
	maxInfoFieldLengthReceive  uint8
	windowSizeTransmit         uint32
	windowSizeReceive          uint32
}

type HdlcParameterGroup struct {
	groupId uint8
	field   []byte
}

var fcstab = []uint16{0x0000, 0x1189, 0x2312, 0x329b, 0x4624, 0x57ad, 0x6536, 0x74bf,
	0x8c48, 0x9dc1, 0xaf5a, 0xbed3, 0xca6c, 0xdbe5, 0xe97e, 0xf8f7,
	0x1081, 0x0108, 0x3393, 0x221a, 0x56a5, 0x472c, 0x75b7, 0x643e,
	0x9cc9, 0x8d40, 0xbfdb, 0xae52, 0xdaed, 0xcb64, 0xf9ff, 0xe876,
	0x2102, 0x308b, 0x0210, 0x1399, 0x6726, 0x76af, 0x4434, 0x55bd,
	0xad4a, 0xbcc3, 0x8e58, 0x9fd1, 0xeb6e, 0xfae7, 0xc87c, 0xd9f5,
	0x3183, 0x200a, 0x1291, 0x0318, 0x77a7, 0x662e, 0x54b5, 0x453c,
	0xbdcb, 0xac42, 0x9ed9, 0x8f50, 0xfbef, 0xea66, 0xd8fd, 0xc974,
	0x4204, 0x538d, 0x6116, 0x709f, 0x0420, 0x15a9, 0x2732, 0x36bb,
	0xce4c, 0xdfc5, 0xed5e, 0xfcd7, 0x8868, 0x99e1, 0xab7a, 0xbaf3,
	0x5285, 0x430c, 0x7197, 0x601e, 0x14a1, 0x0528, 0x37b3, 0x263a,
	0xdecd, 0xcf44, 0xfddf, 0xec56, 0x98e9, 0x8960, 0xbbfb, 0xaa72,
	0x6306, 0x728f, 0x4014, 0x519d, 0x2522, 0x34ab, 0x0630, 0x17b9,
	0xef4e, 0xfec7, 0xcc5c, 0xddd5, 0xa96a, 0xb8e3, 0x8a78, 0x9bf1,
	0x7387, 0x620e, 0x5095, 0x411c, 0x35a3, 0x242a, 0x16b1, 0x0738,
	0xffcf, 0xee46, 0xdcdd, 0xcd54, 0xb9eb, 0xa862, 0x9af9, 0x8b70,
	0x8408, 0x9581, 0xa71a, 0xb693, 0xc22c, 0xd3a5, 0xe13e, 0xf0b7,
	0x0840, 0x19c9, 0x2b52, 0x3adb, 0x4e64, 0x5fed, 0x6d76, 0x7cff,
	0x9489, 0x8500, 0xb79b, 0xa612, 0xd2ad, 0xc324, 0xf1bf, 0xe036,
	0x18c1, 0x0948, 0x3bd3, 0x2a5a, 0x5ee5, 0x4f6c, 0x7df7, 0x6c7e,
	0xa50a, 0xb483, 0x8618, 0x9791, 0xe32e, 0xf2a7, 0xc03c, 0xd1b5,
	0x2942, 0x38cb, 0x0a50, 0x1bd9, 0x6f66, 0x7eef, 0x4c74, 0x5dfd,
	0xb58b, 0xa402, 0x9699, 0x8710, 0xf3af, 0xe226, 0xd0bd, 0xc134,
	0x39c3, 0x284a, 0x1ad1, 0x0b58, 0x7fe7, 0x6e6e, 0x5cf5, 0x4d7c,
	0xc60c, 0xd785, 0xe51e, 0xf497, 0x8028, 0x91a1, 0xa33a, 0xb2b3,
	0x4a44, 0x5bcd, 0x6956, 0x78df, 0x0c60, 0x1de9, 0x2f72, 0x3efb,
	0xd68d, 0xc704, 0xf59f, 0xe416, 0x90a9, 0x8120, 0xb3bb, 0xa232,
	0x5ac5, 0x4b4c, 0x79d7, 0x685e, 0x1ce1, 0x0d68, 0x3ff3, 0x2e7a,
	0xe70e, 0xf687, 0xc41c, 0xd595, 0xa12a, 0xb0a3, 0x8238, 0x93b1,
	0x6b46, 0x7acf, 0x4854, 0x59dd, 0x2d62, 0x3ceb, 0x0e70, 0x1ff9,
	0xf78f, 0xe606, 0xd49d, 0xc514, 0xb1ab, 0xa022, 0x92b9, 0x8330,
	0x7bc7, 0x6a4e, 0x58d5, 0x495c, 0x3de3, 0x2c6a, 0x1ef1, 0x0f78}

const PPPINITFCS16 = uint16(0xffff) // Initial FCS value
const PPPGOODFCS16 = uint16(0xf0b8) // Good final FCS value

/*
 * Calculate a new fcs given the current fcs and the new data.
 */
func pppfcs16(fcs16 uint16, p []byte) uint16 {
	for i := 0; i < len(p); i++ {
		// fcs = (fcs >> 8) ^ fcstab[(fcs ^ *cp++) & 0xff];
		fcs16 = (fcs16 >> 8) ^ fcstab[(fcs16^uint16(p[i]))&0x00ff]
	}
	return fcs16
}

func isTimeOutErr(err error) bool {
	if nerr, ok := err.(net.Error); ok && nerr.Timeout() {
		return true
	} else {
		return false
	}
}

/*
    // How to use the fcs

   tryfcs16(cp, len)
       register unsigned char *cp;
       register int len;
   {
       u16 trialfcs;

       // add on output
       trialfcs = pppfcs16( PPPINITFCS16, cp, len );
       trialfcs ^= 0xffff;                  // complement
       cp[len] = (trialfcs & 0x00ff);       // least significant byte first
       cp[len+1] = ((trialfcs >> 8) & 0x00ff);




       // check on input
       trialfcs = pppfcs16( PPPINITFCS16, cp, len + 2 );
       if ( trialfcs == PPPGOODFCS16 )
           printf("Good FCS\n");
   }
*/

//TODO: better error reporting
var HdlcErrorMalformedSegment = errors.New("malformed segment")
var HdlcErrorInvalidValue = errors.New("invalid value")
var HdlcErrorTimeout = errors.New("time out")
var HdlcErrorConnected = errors.New("connected")
var HdlcErrorNotConnected = errors.New("not connected")
var HdlcErrorConnecting = errors.New("connecting")
var HdlcErrorNotConnecting = errors.New("not connecting")
var HdlcErrorDisconnected = errors.New("disconnected")
var HdlcErrorNotDisconnected = errors.New("not disconnected")
var HdlcErrorDisconnecting = errors.New("disconnecting")
var HdlcErrorNotDisconnecting = errors.New("not disconnecting")
var HdlcErrorNoAllowed = errors.New("not allowed")
var HdlcErrorProtocolError = errors.New("protocl error")
var HdlcErrorInfoFieldFormat = errors.New("wrong info field format")
var HdlcErrorParameterGroupId = errors.New("wrong parameter group id")
var HdlcErrorParameterValue = errors.New("wrong parameter value")
var HdlcErrorNoInfo = errors.New("frame contains no info field")
var HdlcErrorFrameRejected = errors.New("frame rejected")
var HdlcErrorNotClient = errors.New("not a client")

func NewHdlcTransport(conn net.Conn, client bool, clientId uint8, logicalDeviceId uint16, physicalDeviceId *uint16) *HdlcTransport {
	htran := new(HdlcTransport)
	htran.rw = conn
	htran.modulus = 8
	htran.maxInfoFieldLengthTransmit = 128
	htran.maxInfoFieldLengthReceive = 128
	htran.windowSizeTransmit = 7
	htran.windowSizeReceive = 7
	htran.writeQueue = list.New()
	htran.writeQueueMtx = new(sync.Mutex)
	htran.readQueue = list.New()
	htran.readQueueMtx = new(sync.Mutex)
	htran.controlQueue = list.New()
	htran.controlQueueMtx = new(sync.Mutex)
	htran.closedMtx = new(sync.Mutex)
	htran.responseTimeout = time.Duration(1) * time.Second //TODO: set it to network round trip time
	htran.serverAddrLength = HDLC_ADDRESS_LENGTH_4
	htran.clientId = clientId
	htran.logicalDeviceId = logicalDeviceId
	htran.physicalDeviceId = physicalDeviceId
	go htran.handleHdlc(client)
	return htran
}

func (htran *HdlcTransport) SendSNRM(maxInfoFieldLengthTransmit *uint8, maxInfoFieldLengthReceive *uint8, windowSizeTransmit *uint32, windowSizeReceive *uint32) (err error) {

	command := new(HdlcControlCommand)
	command.control = HDLC_CONTROL_SNRM

	command.snrm = new(HdlcControlCommandSNRM)

	if nil != maxInfoFieldLengthTransmit {
		command.snrm.maxInfoFieldLengthTransmit = *maxInfoFieldLengthTransmit
	} else {
		command.snrm.maxInfoFieldLengthTransmit = 128
	}

	if nil != maxInfoFieldLengthReceive {
		command.snrm.maxInfoFieldLengthReceive = *maxInfoFieldLengthReceive
	} else {
		command.snrm.maxInfoFieldLengthReceive = 128
	}

	if nil != windowSizeTransmit {
		command.snrm.windowSizeTransmit = *windowSizeTransmit
	} else {
		command.snrm.windowSizeTransmit = 1
	}

	if nil != windowSizeReceive {
		command.snrm.windowSizeReceive = *windowSizeReceive
	} else {
		command.snrm.windowSizeReceive = 1
	}

	htran.controlQueueMtx.Lock()
	if hdlcDebug {
		fmt.Printf("htran.SendSNRM(): sending command: %d\n", command.control)
	}
	htran.controlQueue.PushBack(command)
	htran.controlQueueMtx.Unlock()

	msg := <-htran.controlAck
	return msg["err"].(error)
}

func (htran *HdlcTransport) SendDISC() (err error) {

	command := new(HdlcControlCommand)
	command.control = HDLC_CONTROL_DISC

	htran.controlQueueMtx.Lock()
	htran.controlQueue.PushBack(command)
	htran.controlQueueMtx.Unlock()

	msg := <-htran.controlAck
	return msg["err"].(error)
}

func (htran *HdlcTransport) Write(p []byte) (n int, err error) {

	var segment *HdlcSegment
	var maxSegmentSize = htran.maxInfoFieldLengthTransmit

	n = len(p)
	for len(p) > 0 {
		segment = new(HdlcSegment)
		if len(p) >= int(maxSegmentSize) {
			segment.p = p[0:maxSegmentSize]
			p = p[len(segment.p):]
		} else {
			segment.p = p
			p = p[len(segment.p):]
			segment.last = true
		}
		htran.readQueueMtx.Lock()
		htran.readQueue.PushBack(segment)
		htran.readQueueMtx.Unlock()
	}

	msg := <-htran.readAck
	return n, msg["err"].(error)
}

func (htran *HdlcTransport) Read(p []byte) (n int, err error) {
	var segment *HdlcSegment

	msg := <-htran.writeAck
	err = msg["err"].(error)
	if nil != err {
		return 0, err
	}

	n = 0
	var l int
	for len(p) > 0 {
		htran.writeQueueMtx.Lock()
		segment = htran.writeQueue.Front().Value.(*HdlcSegment)
		htran.writeQueueMtx.Unlock()

		if len(p) >= len(segment.p) {
			l = len(segment.p) - len(p)
			copy(p, segment.p)
			n += l
			p = p[l:]
		} else {
			// partially read segment and return it shortened back to queue

			l = len(segment.p) - len(p)
			copy(p, segment.p[0:l])
			n += l
			p = p[l:]
			segment.p = segment.p[l:]

			htran.writeQueueMtx.Lock()
			htran.writeQueue.PushFront(segment)
			htran.writeQueueMtx.Unlock()
		}

		if segment.last {
			break
		}
	}
	return n, nil
}

func (htran *HdlcTransport) Close() (err error) {
	htran.controlQueueMtx.Lock()
	htran.closed = true
	htran.controlQueueMtx.Unlock()
	return nil
}

func (htran *HdlcTransport) decodeServerAddress(frame *HdlcFrame) (err error, n int) {
	var r io.Reader = htran.rw
	if hdlcDebug {
		r = frame.content
	}
	n = 0

	var b0, b1, b2, b3 byte
	p := make([]byte, 1)

	if !((HDLC_ADDRESS_LENGTH_1 == htran.serverAddrLength) || (HDLC_ADDRESS_LENGTH_2 == htran.serverAddrLength) || (HDLC_ADDRESS_LENGTH_4 == htran.serverAddrLength)) {
		panic("wrong expected server address length value")
	}

	_, err = io.ReadFull(r, p)
	if nil != err {
		if !isTimeOutErr(err) {
			errorLog("io.ReadFull() failed: %v", err)
		}
		return err, n
	}
	n += 1
	frame.fcs16 = pppfcs16(frame.fcs16, p)
	b0 = p[0]

	if b0&0x01 > 0 {
		if HDLC_ADDRESS_LENGTH_1 == htran.serverAddrLength {
			frame.logicalDeviceId = (uint16(b0) & 0x00FE) >> 1
			frame.physicalDeviceId = nil
		} else {
			errorLog("short server address")
			return HdlcErrorMalformedSegment, n
		}
	} else {
		_, err = io.ReadFull(r, p)
		if nil != err {
			if !isTimeOutErr(err) {
				errorLog("io.ReadFull() failed: %v", err)
			}
			return err, n
		}
		n += 1
		frame.fcs16 = pppfcs16(frame.fcs16, p)
		b1 = p[0]

		if b1&0x01 > 0 {
			upperMAC := (uint16(b0) & 0x00FE) >> 1
			lowerMAC := (uint16(b1) & 0x00FE) >> 1
			if HDLC_ADDRESS_LENGTH_2 == htran.serverAddrLength {
				frame.logicalDeviceId = upperMAC
				frame.physicalDeviceId = new(uint16)
				*frame.physicalDeviceId = lowerMAC
			} else if HDLC_ADDRESS_LENGTH_1 == htran.serverAddrLength {
				if 0x007F == lowerMAC {
					// all station broadcast
					frame.logicalDeviceId = lowerMAC
					frame.physicalDeviceId = nil
				} else {
					errorLog("long server address")
					return HdlcErrorMalformedSegment, n
				}
			} else if HDLC_ADDRESS_LENGTH_4 == htran.serverAddrLength {
				frame.logicalDeviceId = upperMAC
				frame.physicalDeviceId = new(uint16)
				*frame.physicalDeviceId = lowerMAC
			} else {
				panic("assertion failed")
			}
		} else {
			_, err = io.ReadFull(r, p)
			if nil != err {
				_, err = io.ReadFull(r, p)
				return err, n
			}
			n += 1
			frame.fcs16 = pppfcs16(frame.fcs16, p)
			b2 = p[0]

			if b2&0x01 > 0 {
				errorLog("short server address")
				return HdlcErrorMalformedSegment, n
			}

			_, err = io.ReadFull(r, p)
			if nil != err {
				if !isTimeOutErr(err) {
					errorLog("io.ReadFull() failed: %v", err)
				}
				return err, n
			}
			n += 1
			frame.fcs16 = pppfcs16(frame.fcs16, p)
			b3 = p[0]

			if b3&0x01 > 0 {
				upperMAC := ((uint16(b0)&0x00FE)>>1)<<7 + ((uint16(b1) & 0x00FE) >> 1)
				lowerMAC := ((uint16(b2)&0x00FE)>>1)<<7 + ((uint16(b3) & 0x00FE) >> 1)

				if HDLC_ADDRESS_LENGTH_4 == htran.serverAddrLength {

					frame.logicalDeviceId = upperMAC
					frame.physicalDeviceId = new(uint16)
					*frame.physicalDeviceId = lowerMAC

				} else if HDLC_ADDRESS_LENGTH_1 == htran.serverAddrLength {
					if (0x3FFF == upperMAC) && (0x3FFF == lowerMAC) {
						// all station broadcast 0x3FFF
						frame.logicalDeviceId = 0x3FFF
						frame.physicalDeviceId = new(uint16)
						*frame.physicalDeviceId = 0x3FFF
					} else {
						errorLog("long server address")
						return HdlcErrorMalformedSegment, n
					}
				} else if HDLC_ADDRESS_LENGTH_2 == htran.serverAddrLength {
					if (0x3FFF == upperMAC) && (0x3FFF == lowerMAC) {
						// all station broadcast 0x3FFF
						frame.logicalDeviceId = 0x3FFF
						frame.physicalDeviceId = new(uint16)
						*frame.physicalDeviceId = 0x3FFF
					} else if (upperMAC == 0x0001) && (0x3FFE == lowerMAC) && frame.callingPhysicalDevice {
						// event reporting
						frame.logicalDeviceId = upperMAC
						frame.physicalDeviceId = new(uint16)
						*frame.physicalDeviceId = lowerMAC
					} else {
						errorLog("long server address")
						return HdlcErrorMalformedSegment, n
					}
				} else {
					panic("assertion failed")
				}
			} else {
				errorLog("long server address")
				return HdlcErrorMalformedSegment, n
			}
		}
	}
	return nil, n
}

func (htran *HdlcTransport) encodeServerAddress(frame *HdlcFrame) (err error) {
	var w io.Writer = htran.rw

	var v16 uint16
	p := make([]byte, 1)

	if !((HDLC_ADDRESS_LENGTH_1 == htran.serverAddrLength) || (HDLC_ADDRESS_LENGTH_2 == htran.serverAddrLength) || (HDLC_ADDRESS_LENGTH_4 == htran.serverAddrLength)) {
		panic("wrong expected server address length value")
	}

	if HDLC_ADDRESS_LENGTH_1 == htran.serverAddrLength {
		p := make([]byte, 1)

		// logicalDeviceId

		logicalDeviceId := htran.logicalDeviceId
		if logicalDeviceId&0xFF80 > 0 {
			errorLog("logicalDeviceId exceeds limit")
			return HdlcErrorInvalidValue
		}

		v16 = (logicalDeviceId << 1) | 0x0001

		p[0] = byte(v16 & 0x00FF)
		_, err = w.Write(p)
		if nil != err {
			errorLog("binary.Write() failed: %v", err)
			return err
		}
		frame.fcs16 = pppfcs16(frame.fcs16, p)

		// physicalDeviceId

		if nil != htran.physicalDeviceId {
			errorLog("physicalDeviceId specified (expected to be nil)")
			return HdlcErrorInvalidValue
		}

	} else if HDLC_ADDRESS_LENGTH_2 == htran.serverAddrLength {

		// logicalDeviceId

		logicalDeviceId := htran.logicalDeviceId
		if logicalDeviceId&0xFF80 > 0 {
			errorLog("logicalDeviceId exceeds limit")
			return HdlcErrorInvalidValue
		}

		v16 = logicalDeviceId

		p[0] = byte(v16 & 0x00FF)
		_, err = w.Write(p)
		if nil != err {
			errorLog("binary.Write() failed: %v", err)
			return err
		}
		frame.fcs16 = pppfcs16(frame.fcs16, p)

		// physicalDeviceId

		if nil == frame.physicalDeviceId {
			errorLog("physicalDeviceId not specified")
			return HdlcErrorInvalidValue
		}

		physicalDeviceId := *htran.physicalDeviceId
		if physicalDeviceId&0xFF80 > 0 {
			errorLog("physicalDeviceId exceeds limit")
			return HdlcErrorInvalidValue
		}

		v16 = (physicalDeviceId << 1) | 0x0001

		p[0] = byte(v16 & 0x00FF)
		_, err = w.Write(p)
		if nil != err {
			errorLog("binary.Write() failed: %v", err)
			return err
		}
		frame.fcs16 = pppfcs16(frame.fcs16, p)

	} else if HDLC_ADDRESS_LENGTH_4 == htran.serverAddrLength {

		// logicalDeviceId

		logicalDeviceId := htran.logicalDeviceId
		if logicalDeviceId&0x1000 > 0 {
			errorLog("logicalDeviceId exceeds limit")
			return HdlcErrorInvalidValue
		}

		v16 = logicalDeviceId

		p[0] = byte((v16 & 0xFF00) >> 8)
		_, err = w.Write(p)
		if nil != err {
			errorLog("binary.Write() failed: %v", err)
			return err
		}
		frame.fcs16 = pppfcs16(frame.fcs16, p)

		p[0] = byte(v16 & 0x00FF)
		_, err = w.Write(p)
		if nil != err {
			errorLog("binary.Write() failed: %v", err)
			return err
		}
		frame.fcs16 = pppfcs16(frame.fcs16, p)

		// physicalDeviceId

		if nil == htran.physicalDeviceId {
			errorLog("physicalDeviceId not specified")
			return HdlcErrorInvalidValue
		}

		physicalDeviceId := *htran.physicalDeviceId
		if physicalDeviceId&0x1000 > 0 {
			errorLog("physicalDeviceId exceeds limit")
			return HdlcErrorInvalidValue
		}

		v16 = (physicalDeviceId << 1) | 0x0001

		p[0] = byte((v16 & 0xFF00) >> 8)
		_, err = w.Write(p)
		if nil != err {
			errorLog("binary.Write() failed: %v", err)
			return err
		}
		frame.fcs16 = pppfcs16(frame.fcs16, p)

		p[0] = byte(v16 & 0x00FF)
		_, err = w.Write(p)
		if nil != err {
			errorLog("binary.Write() failed: %v", err)
			return err
		}
		frame.fcs16 = pppfcs16(frame.fcs16, p)
	} else {
		panic("wrong expected server address length value")
	}

	return nil
}

func (htran *HdlcTransport) lengthServerAddress(frame *HdlcFrame) (n int) {

	n = 0

	if !((HDLC_ADDRESS_LENGTH_1 == htran.serverAddrLength) || (HDLC_ADDRESS_LENGTH_2 == htran.serverAddrLength) || (HDLC_ADDRESS_LENGTH_4 == htran.serverAddrLength)) {
		panic("wrong expected server address length value")
	}

	if HDLC_ADDRESS_LENGTH_1 == htran.serverAddrLength {
		n = 1
	} else if HDLC_ADDRESS_LENGTH_2 == htran.serverAddrLength {
		n += 2
	} else if HDLC_ADDRESS_LENGTH_4 == htran.serverAddrLength {
		n += 4
	} else {
		panic("wrong expected server address length value")
	}

	return n
}

func (htran *HdlcTransport) decodeClientAddress(frame *HdlcFrame) (err error, n int) {
	var r io.Reader = htran.rw
	if hdlcDebug {
		r = frame.content
	}
	n = 0
	var b0 byte
	p := make([]byte, 1)

	_, err = io.ReadFull(r, p)
	if nil != err {
		if !isTimeOutErr(err) {
			errorLog("io.ReadFull() failed: %v", err)
		}
		return err, n
	}
	n += 1
	frame.fcs16 = pppfcs16(frame.fcs16, p)
	b0 = p[0]

	if b0&0x01 > 0 {
		frame.clientId = (uint8(b0) & 0xFE) >> 1
	} else {
		errorLog("long client address")
		return HdlcErrorMalformedSegment, n
	}

	return nil, n
}

func (htran *HdlcTransport) encodeClientAddress(frame *HdlcFrame) (err error) {
	var w io.Writer = htran.rw
	var b0 byte
	p := make([]byte, 1)

	clientId := htran.clientId
	if clientId&0x80 > 0 {
		errorLog("clientId exceeds limit")
		return HdlcErrorInvalidValue
	}

	b0 = (clientId << 1) | 0x01

	p[0] = b0
	_, err = w.Write(p)
	if nil != err {
		errorLog("r.Write() failed: %v", err)
		return err
	}
	frame.fcs16 = pppfcs16(frame.fcs16, p)

	return nil
}

func (htran *HdlcTransport) lengthClientAddress(frame *HdlcFrame) int {
	return 1
}

func (htran *HdlcTransport) decodeFrameInfo(frame *HdlcFrame, l int) (err error, n int) {
	var r io.Reader = htran.rw
	if hdlcDebug {
		r = frame.content
	}
	p := make([]byte, 1)

	// HCS - header control sum

	_, err = io.ReadFull(r, p)
	if nil != err {
		if !isTimeOutErr(err) {
			errorLog("io.ReadFull() failed: %v", err)
		}
		return err, n
	}
	n += 1
	l += 1
	frame.fcs16 = pppfcs16(frame.fcs16, p)
	_, err = io.ReadFull(r, p)
	if nil != err {
		if !isTimeOutErr(err) {
			errorLog("io.ReadFull() failed: %v", err)
		}
		return err, n
	}
	n += 1
	l += 1
	frame.fcs16 = pppfcs16(frame.fcs16, p)

	if PPPGOODFCS16 != frame.fcs16 {
		errorLog("wrong HCS")
		return HdlcErrorMalformedSegment, n
	}

	// read information field

	infoFieldLength := frame.length - l - 2 // minus 2 bytes for FCS

	if infoFieldLength > 0 {

		if infoFieldLength > int(htran.maxInfoFieldLengthReceive) {
			errorLog("long info field")
			return HdlcErrorMalformedSegment, n
		}

		p = make([]byte, infoFieldLength)
		_, err = io.ReadFull(r, p)
		if nil != err {
			if !isTimeOutErr(err) {
				errorLog("io.ReadFull() failed: %v", err)
			}
			return err, n
		}
		n += len(p)
		frame.fcs16 = pppfcs16(frame.fcs16, p)

		frame.infoField = p

		fmt.Printf("@@@@@@@@@@@@@@@@@@@@@@@@@@@@ decode infoField: %02X\n", frame.infoField)

	} else {
		frame.infoField = nil
	}

	return nil, n
}

func (htran *HdlcTransport) encodeFrameInfo(frame *HdlcFrame) (err error) {
	var w io.Writer = htran.rw
	p := make([]byte, 1)

	// HCS - header control sum

	fcs16 := frame.fcs16
	p[0] = byte(^fcs16 & 0x00FF)
	_, err = w.Write(p)
	if nil != err {
		errorLog("w.Write() failed: %v", err)
		return err
	}
	frame.fcs16 = pppfcs16(frame.fcs16, p)
	p[0] = byte((^fcs16 & 0xFF00) >> 8)
	_, err = w.Write(p)
	if nil != err {
		errorLog("w.Write() failed: %v", err)
		return err
	}
	frame.fcs16 = pppfcs16(frame.fcs16, p)

	// write information field

	if (nil != frame.infoField) && len(frame.infoField) > 0 {
		infoFieldLength := len(frame.infoField)

		if (HDLC_FRAME_DIRECTION_CLIENT_INBOUND == frame.direction) || (HDLC_FRAME_DIRECTION_SERVER_INBOUND == frame.direction) {
			if infoFieldLength > int(htran.maxInfoFieldLengthReceive) {
				errorLog("long info field")
				return HdlcErrorMalformedSegment
			}
		} else if (HDLC_FRAME_DIRECTION_CLIENT_OUTBOUND == frame.direction) || (HDLC_FRAME_DIRECTION_SERVER_OUTBOUND == frame.direction) {
			if infoFieldLength > int(htran.maxInfoFieldLengthTransmit) {
				errorLog("long info field")
				return HdlcErrorMalformedSegment
			}
		} else {
			panic("assertion failed")
		}

		fmt.Printf("@@@@@@@@@@@@@@@@@@@@@@@@@@@@ encode infoField: %02X\n", frame.infoField)

		p = frame.infoField
		err = binary.Write(w, binary.BigEndian, p)
		if nil != err {
			errorLog("binary.Write() failed: %v", err)
			return err
		}
		frame.fcs16 = pppfcs16(frame.fcs16, p)

		// FCS - frame control sum

		fcs16 := frame.fcs16
		p[0] = byte(^fcs16 & 0x00FF)
		_, err = w.Write(p)
		if nil != err {
			errorLog("w.Write() failed: %v", err)
			return err
		}
		frame.fcs16 = pppfcs16(frame.fcs16, p)
		p[0] = byte((^fcs16 & 0xFF00) >> 8)
		_, err = w.Write(p)
		if nil != err {
			errorLog("w.Write() failed: %v", err)
			return err
		}
		frame.fcs16 = pppfcs16(frame.fcs16, p)
	}

	return nil
}

func (htran *HdlcTransport) decodeLinkParameters(frame *HdlcFrame) (err error, maxInfoFieldLengthTransmit *uint8, maxInfoFieldLengthReceive *uint8, windowSizeTransmit *uint32, windowSizeReceive *uint32) {
	r := bytes.NewBuffer(frame.infoField)

	p := make([]byte, 1)

	// format (always present)

	_, err = io.ReadFull(r, p)
	if nil != err {
		if !isTimeOutErr(err) {
			errorLog("io.ReadFull() failed: %v", err)
		}
		return err, nil, nil, nil, nil
	}
	infoFieldFormat := uint8(p[0])
	if 0x81 != infoFieldFormat {
		errorLog("wrong info field format")
		return HdlcErrorInfoFieldFormat, nil, nil, nil, nil
	}

	// group id

	_, err = io.ReadFull(r, p)
	if nil != err {
		if !isTimeOutErr(err) {
			errorLog("io.ReadFull() failed: %v", err)
		}
		return err, nil, nil, nil, nil
	}
	groupId := uint8(p[0])
	if 0x80 != groupId {
		errorLog("wrong parameter group id")
		return HdlcErrorParameterGroupId, nil, nil, nil, nil
	}

	// group length

	_, err = io.ReadFull(r, p)
	if nil != err {
		if io.EOF == err {
			return nil, nil, nil, nil, nil
		} else {
			if !isTimeOutErr(err) {
				errorLog("io.ReadFull() failed: %v", err)
			}
			return err, nil, nil, nil, nil
		}
	}
	length := uint8(p[0])
	if 0 == length {
		return nil, nil, nil, nil, nil
	}

	pp := make([]byte, length)
	_, err = io.ReadFull(r, pp)
	if nil != err {
		if !isTimeOutErr(err) {
			errorLog("io.ReadFull() failed: %v", err)
		}
		return err, nil, nil, nil, nil
	}
	rr := bytes.NewBuffer(pp)

	// parameters

	var buf *bytes.Buffer
	for {
		_, err := io.ReadFull(rr, p)
		if nil != err {
			if io.EOF == err {
				break
			}
			if !isTimeOutErr(err) {
				errorLog("io.ReadFull() failed: %v", err)
			}
			return err, nil, nil, nil, nil
		}
		parameterId := uint8(p[0])

		_, err = io.ReadFull(rr, p)
		if nil != err {
			if !isTimeOutErr(err) {
				errorLog("io.ReadFull() failed: %v", err)
			}
			return err, nil, nil, nil, nil
		}
		length = uint8(p[0])

		pp = make([]byte, length)
		_, err = io.ReadFull(rr, pp)
		if nil != err {
			if !isTimeOutErr(err) {
				errorLog("io.ReadFull() failed: %v", err)
			}
			return err, nil, nil, nil, nil
		}
		parameterValue := pp

		if 0x05 == parameterId {
			if 1 != length {
				errorLog("wrong parameter value length")
				return HdlcErrorParameterValue, nil, nil, nil, nil
			}
			maxInfoFieldLengthTransmit = new(uint8)
			*maxInfoFieldLengthTransmit = uint8(parameterValue[0])
		} else if 0x06 == parameterId {
			if 1 != length {
				errorLog("wrong parameter value length")
				return HdlcErrorParameterValue, nil, nil, nil, nil
			}
			maxInfoFieldLengthReceive = new(uint8)
			*maxInfoFieldLengthReceive = uint8(parameterValue[0])
		} else if 0x07 == parameterId {
			if 4 != length {
				errorLog("wrong parameter value length")
				return HdlcErrorParameterValue, nil, nil, nil, nil
			}
			windowSizeTransmit = new(uint32)
			buf = new(bytes.Buffer)
			err = binary.Read(buf, binary.BigEndian, windowSizeTransmit)
			if nil != err {
				if !isTimeOutErr(err) {
					errorLog("binary.Read() failed: %v", err)
				}
				return err, nil, nil, nil, nil
			}
		} else if 0x08 == parameterId {
			if 4 != length {
				errorLog("wrong parameter value length")
				return HdlcErrorParameterValue, nil, nil, nil, nil
			}
			windowSizeReceive = new(uint32)
			buf = new(bytes.Buffer)
			err = binary.Read(buf, binary.BigEndian, windowSizeReceive)
			if nil != err {
				if !isTimeOutErr(err) {
					errorLog("binary.Read() failed: %v", err)
				}
				return err, nil, nil, nil, nil
			}
		} else {
			// just ignore usupported parameter
		}
	}

	return nil, maxInfoFieldLengthTransmit, maxInfoFieldLengthReceive, windowSizeTransmit, windowSizeReceive

}

func (htran *HdlcTransport) encodeLinkParameters(frame *HdlcFrame, maxInfoFieldLengthTransmit *uint8, maxInfoFieldLengthReceive *uint8, windowSizeTransmit *uint32, windowSizeReceive *uint32) (err error) {

	w := new(bytes.Buffer)

	// format

	err = binary.Write(w, binary.BigEndian, uint8(0x81))
	if nil != err {
		errorLog("binary.Write() failed: %v", err)
		return err
	}

	// group id

	err = binary.Write(w, binary.BigEndian, uint8(0x80))
	if nil != err {
		errorLog("binary.Write() failed: %v", err)
		return err
	}

	if (nil == maxInfoFieldLengthTransmit) && (nil == maxInfoFieldLengthReceive) && (nil == windowSizeTransmit) && (nil == windowSizeReceive) {
		frame.infoField = w.Bytes()
		return nil
	}

	ww := new(bytes.Buffer)

	if nil == maxInfoFieldLengthTransmit {
		maxInfoFieldLengthTransmit = new(uint8)
		*maxInfoFieldLengthTransmit = 0
	} else {
		err = binary.Write(ww, binary.BigEndian, maxInfoFieldLengthTransmit)
	}
	err = binary.Write(ww, binary.BigEndian, uint8(0x05))
	if nil != err {
		errorLog("binary.Write() failed: %v", err)
		return err
	}
	err = binary.Write(ww, binary.BigEndian, uint8(1))
	if nil != err {
		errorLog("binary.Write() failed: %v", err)
		return err
	}
	err = binary.Write(ww, binary.BigEndian, maxInfoFieldLengthTransmit)
	if nil != err {
		errorLog("binary.Write() failed: %v", err)
		return err
	}

	if nil == maxInfoFieldLengthReceive {
		maxInfoFieldLengthReceive = new(uint8)
		*maxInfoFieldLengthReceive = 0
	} else {
		err = binary.Write(ww, binary.BigEndian, maxInfoFieldLengthReceive)
	}
	err = binary.Write(ww, binary.BigEndian, uint8(0x06))
	if nil != err {
		errorLog("binary.Write() failed: %v", err)
		return err
	}
	err = binary.Write(ww, binary.BigEndian, uint8(1))
	if nil != err {
		errorLog("binary.Write() failed: %v", err)
		return err
	}
	err = binary.Write(ww, binary.BigEndian, maxInfoFieldLengthReceive)
	if nil != err {
		errorLog("binary.Write() failed: %v", err)
		return err
	}

	if nil == windowSizeTransmit {
		windowSizeTransmit = new(uint32)
		*windowSizeTransmit = 0
	} else {
		err = binary.Write(ww, binary.BigEndian, windowSizeTransmit)
	}
	err = binary.Write(ww, binary.BigEndian, uint8(0x07))
	if nil != err {
		errorLog("binary.Write() failed: %v", err)
		return err
	}
	err = binary.Write(ww, binary.BigEndian, uint8(4))
	if nil != err {
		errorLog("binary.Write() failed: %v", err)
		return err
	}
	err = binary.Write(ww, binary.BigEndian, windowSizeTransmit)
	if nil != err {
		errorLog("binary.Write() failed: %v", err)
		return err
	}

	if nil == windowSizeReceive {
		windowSizeReceive = new(uint32)
		*windowSizeReceive = 0
	} else {
		err = binary.Write(ww, binary.BigEndian, windowSizeReceive)
	}
	err = binary.Write(ww, binary.BigEndian, uint8(0x08))
	if nil != err {
		errorLog("binary.Write() failed: %v", err)
		return err
	}
	err = binary.Write(ww, binary.BigEndian, uint8(4))
	if nil != err {
		errorLog("binary.Write() failed: %v", err)
		return err
	}
	err = binary.Write(ww, binary.BigEndian, windowSizeReceive)
	if nil != err {
		errorLog("binary.Write() failed: %v", err)
		return err
	}

	groupValue := ww.Bytes()

	// group length

	err = binary.Write(w, binary.BigEndian, uint8(len(groupValue)))
	if nil != err {
		errorLog("binary.Write() failed: %v", err)
		return err
	}

	// group value

	err = binary.Write(w, binary.BigEndian, groupValue)
	if nil != err {
		errorLog("binary.Write() failed: %v", err)
		return err
	}

	frame.infoField = w.Bytes()
	return nil

}

// decode frame address, control and information field

func (htran *HdlcTransport) decodeFrameACI(frame *HdlcFrame, l int) (err error, n int) {
	var r io.Reader = htran.rw
	if hdlcDebug {
		r = frame.content
	}
	n = 0
	var b0 byte
	var nn int

	p := make([]byte, 1)

	// dst and src address

	if (HDLC_FRAME_DIRECTION_SERVER_INBOUND == frame.direction) || (HDLC_FRAME_DIRECTION_CLIENT_OUTBOUND == frame.direction) {
		err, nn = htran.decodeServerAddress(frame)
		if nil != err {
			return err, n
		}
		n += nn
		err, nn = htran.decodeClientAddress(frame)
		if nil != err {
			return err, n
		}
		n += nn
	} else if (HDLC_FRAME_DIRECTION_CLIENT_INBOUND == frame.direction) || (HDLC_FRAME_DIRECTION_SERVER_OUTBOUND == frame.direction) {
		err, nn = htran.decodeClientAddress(frame)
		if nil != err {
			return err, n
		}
		n += nn
		err, nn = htran.decodeServerAddress(frame)
		if nil != err {
			return err, n
		}
		n += nn
	} else {
		panic("invalid frame direction")
	}

	// control

	_, err = io.ReadFull(r, p)
	if nil != err {
		if !isTimeOutErr(err) {
			errorLog("io.ReadFull() failed: %v", err)
		}
		return err, n
	}
	n += 1
	frame.fcs16 = pppfcs16(frame.fcs16, p)
	b0 = p[0]

	// P/F bit
	frame.poll = b0&0x10 > 0

	if b0&0x01 == 0 {
		frame.control = HDLC_CONTROL_I

		frame.nr = b0 & 0xE0 >> 5
		frame.ns = b0 & 0x0E >> 1

		err, nn := htran.decodeFrameInfo(frame, l+n)
		if nil != err {
			return err, n
		}
		if nil == frame.infoField {
			return HdlcErrorNoInfo, n
		}
		n += nn

		// FCS - frame control sum

		_, err = io.ReadFull(r, p)
		if nil != err {
			if !isTimeOutErr(err) {
				errorLog("io.ReadFull() failed: %v", err)
			}
			return err, n
		}
		n += 1
		frame.fcs16 = pppfcs16(frame.fcs16, p)
		_, err = io.ReadFull(r, p)
		if nil != err {
			if !isTimeOutErr(err) {
				errorLog("io.ReadFull() failed: %v", err)
			}
			return err, n
		}
		n += 1
		frame.fcs16 = pppfcs16(frame.fcs16, p)

		if PPPGOODFCS16 != frame.fcs16 {
			errorLog("wrong FCS")
			return HdlcErrorMalformedSegment, n
		}

	} else if (b0&0x08 == 0) && (b0&0x04 == 0) && (b0&0x02 == 0) && (b0&0x01 > 0) {
		frame.control = HDLC_CONTROL_RR

		// FCS - frame control sum

		_, err = io.ReadFull(r, p)
		if nil != err {
			if !isTimeOutErr(err) {
				errorLog("io.ReadFull() failed: %v", err)
			}
			return err, n
		}
		n += 1
		frame.fcs16 = pppfcs16(frame.fcs16, p)
		_, err = io.ReadFull(r, p)
		if nil != err {
			if !isTimeOutErr(err) {
				errorLog("io.ReadFull() failed: %v", err)
			}
			return err, n
		}
		n += 1
		frame.fcs16 = pppfcs16(frame.fcs16, p)

		if PPPGOODFCS16 != frame.fcs16 {
			errorLog("wrong FCS")
			return HdlcErrorMalformedSegment, n
		}
	} else if (b0&0x08 == 0) && (b0&0x04 > 0) && (b0&0x02 == 0) && (b0&0x01 > 0) {
		frame.control = HDLC_CONTROL_RNR

		// FCS - frame control sum

		_, err = io.ReadFull(r, p)
		if nil != err {
			if !isTimeOutErr(err) {
				errorLog("io.ReadFull() failed: %v", err)
			}
			return err, n
		}
		n += 1
		frame.fcs16 = pppfcs16(frame.fcs16, p)
		_, err = io.ReadFull(r, p)
		if nil != err {
			if !isTimeOutErr(err) {
				errorLog("io.ReadFull() failed: %v", err)
			}
			return err, n
		}
		n += 1
		frame.fcs16 = pppfcs16(frame.fcs16, p)

		if PPPGOODFCS16 != frame.fcs16 {
			errorLog("wrong FCS")
			return HdlcErrorMalformedSegment, n
		}
	} else if (b0&0x80 > 0) && (b0&0x40 == 0) && (b0&0x20 == 0) && (b0&0x08 == 0) && (b0&0x04 == 0) && (b0&0x02 > 0) && (b0&0x01 > 0) {

		frame.control = HDLC_CONTROL_SNRM

		err, nn := htran.decodeFrameInfo(frame, l+n)
		if nil != err {
			return err, n
		}
		n += nn

		if nil != frame.infoField {
			// FCS - frame control sum

			_, err = io.ReadFull(r, p)
			if nil != err {
				if !isTimeOutErr(err) {
					errorLog("io.ReadFull() failed: %v", err)
				}
				return err, n
			}
			n += 1
			frame.fcs16 = pppfcs16(frame.fcs16, p)
			_, err = io.ReadFull(r, p)
			if nil != err {
				if !isTimeOutErr(err) {
					errorLog("io.ReadFull() failed: %v", err)
				}
				return err, n
			}
			n += 1
			frame.fcs16 = pppfcs16(frame.fcs16, p)

			if PPPGOODFCS16 != frame.fcs16 {
				errorLog("wrong FCS")
				return HdlcErrorMalformedSegment, n
			}
		}

	} else if (b0&0x80 == 0) && (b0&0x40 > 0) && (b0&0x20 == 0) && (b0&0x08 == 0) && (b0&0x04 == 0) && (b0&0x02 > 0) && (b0&0x01 > 0) {
		frame.control = HDLC_CONTROL_DISC

		// FCS - frame control sum

		_, err = io.ReadFull(r, p)
		if nil != err {
			if !isTimeOutErr(err) {
				errorLog("io.ReadFull() failed: %v", err)
			}
			return err, n
		}
		n += 1
		frame.fcs16 = pppfcs16(frame.fcs16, p)
		_, err = io.ReadFull(r, p)
		if nil != err {
			if !isTimeOutErr(err) {
				errorLog("io.ReadFull() failed: %v", err)
			}
			return err, n
		}
		n += 1
		frame.fcs16 = pppfcs16(frame.fcs16, p)

		if PPPGOODFCS16 != frame.fcs16 {
			errorLog("wrong FCS")
			return HdlcErrorMalformedSegment, n
		}
	} else if (b0&0x80 == 0) && (b0&0x40 > 0) && (b0&0x20 > 0) && (b0&0x08 == 0) && (b0&0x04 == 0) && (b0&0x02 > 0) && (b0&0x01 > 0) {
		frame.control = HDLC_CONTROL_UA

		err, nn := htran.decodeFrameInfo(frame, l+n)
		if nil != err {
			return err, n
		}
		n += nn

		if nil == frame.infoField {
			// FCS - frame control sum

			_, err = io.ReadFull(r, p)
			if nil != err {
				if !isTimeOutErr(err) {
					errorLog("io.ReadFull() failed: %v", err)
				}
				return err, n
			}
			n += 1
			frame.fcs16 = pppfcs16(frame.fcs16, p)
			_, err = io.ReadFull(r, p)
			if nil != err {
				if !isTimeOutErr(err) {
					errorLog("io.ReadFull() failed: %v", err)
				}
				return err, n
			}
			n += 1
			frame.fcs16 = pppfcs16(frame.fcs16, p)

			if PPPGOODFCS16 != frame.fcs16 {
				errorLog("wrong FCS")
				return HdlcErrorMalformedSegment, n
			}
		}
	} else if (b0&0x80 == 0) && (b0&0x40 == 0) && (b0&0x20 == 0) && (b0&0x08 > 0) && (b0&0x04 > 0) && (b0&0x02 > 0) && (b0&0x01 > 0) {
		frame.control = HDLC_CONTROL_DM

		// FCS - frame control sum

		_, err = io.ReadFull(r, p)
		if nil != err {
			if !isTimeOutErr(err) {
				errorLog("io.ReadFull() failed: %v", err)
			}
			return err, n
		}
		n += 1
		frame.fcs16 = pppfcs16(frame.fcs16, p)
		_, err = io.ReadFull(r, p)
		if nil != err {
			if !isTimeOutErr(err) {
				errorLog("io.ReadFull() failed: %v", err)
			}
			return err, n
		}
		n += 1
		frame.fcs16 = pppfcs16(frame.fcs16, p)

		if PPPGOODFCS16 != frame.fcs16 {
			errorLog("wrong FCS")
			return HdlcErrorMalformedSegment, n
		}
	} else if (b0&0x80 > 0) && (b0&0x40 == 0) && (b0&0x20 == 0) && (b0&0x08 == 0) && (b0&0x04 > 0) && (b0&0x02 > 0) && (b0&0x01 > 0) {
		frame.control = HDLC_CONTROL_FRMR

		err, nn := htran.decodeFrameInfo(frame, l+n)
		if nil != err {
			return err, n
		}
		n += nn

		// FCS - frame control sum

		_, err = io.ReadFull(r, p)
		if nil != err {
			if !isTimeOutErr(err) {
				errorLog("io.ReadFull() failed: %v", err)
			}
			return err, n
		}
		n += 1
		frame.fcs16 = pppfcs16(frame.fcs16, p)
		_, err = io.ReadFull(r, p)
		if nil != err {
			if !isTimeOutErr(err) {
				errorLog("io.ReadFull() failed: %v", err)
			}
			return err, n
		}
		n += 1
		frame.fcs16 = pppfcs16(frame.fcs16, p)

		if PPPGOODFCS16 != frame.fcs16 {
			errorLog("wrong FCS")
			return HdlcErrorMalformedSegment, n
		}
	} else if (b0&0x80 == 0) && (b0&0x40 == 0) && (b0&0x20 == 0) && (b0&0x08 == 0) && (b0&0x04 == 0) && (b0&0x02 > 0) && (b0&0x01 > 0) {
		frame.control = HDLC_CONTROL_UI

		err, nn := htran.decodeFrameInfo(frame, l+n)
		if nil != err {
			return err, n
		}
		n += nn

		// FCS - frame control sum

		_, err = io.ReadFull(r, p)
		if nil != err {
			if !isTimeOutErr(err) {
				errorLog("io.ReadFull() failed: %v", err)
			}
			return err, n
		}
		n += 1
		frame.fcs16 = pppfcs16(frame.fcs16, p)
		_, err = io.ReadFull(r, p)
		if nil != err {
			if !isTimeOutErr(err) {
				errorLog("io.ReadFull() failed: %v", err)
			}
			return err, n
		}
		n += 1
		frame.fcs16 = pppfcs16(frame.fcs16, p)

		if PPPGOODFCS16 != frame.fcs16 {
			errorLog("wrong FCS")
			return HdlcErrorMalformedSegment, n
		}
	} else {
		errorLog("unknown control field value")
		return HdlcErrorMalformedSegment, n
	}

	return nil, n
}

// encode frame address, control and information field

func (htran *HdlcTransport) encodeFrameACI(frame *HdlcFrame) (err error) {
	var w io.Writer = htran.rw
	var b0 byte

	p := make([]byte, 1)

	// dst and src address

	if (HDLC_FRAME_DIRECTION_SERVER_OUTBOUND == frame.direction) || (HDLC_FRAME_DIRECTION_CLIENT_INBOUND == frame.direction) {
		err = htran.encodeClientAddress(frame)
		if nil != err {
			return err
		}
		err = htran.encodeServerAddress(frame)
		if nil != err {
			return err
		}
	} else if (HDLC_FRAME_DIRECTION_CLIENT_OUTBOUND == frame.direction) || (HDLC_FRAME_DIRECTION_SERVER_INBOUND == frame.direction) {
		err = htran.encodeServerAddress(frame)
		if nil != err {
			return err
		}
		err = htran.encodeClientAddress(frame)
		if nil != err {
			return err
		}
	} else {
		panic("wrond frame direction")
	}

	// control

	// P/F bit
	b0 = 0
	if frame.poll {
		b0 |= 0x10
	}

	if HDLC_CONTROL_I == frame.control {

		if frame.nr > 0x07 {
			panic("NR exceeds limit")
		}
		b0 |= frame.nr << 5

		if frame.ns > 0x07 {
			panic("NS exceeds limit")
		}
		b0 |= frame.ns << 1

		p[0] = b0
		_, err = w.Write(p)
		if nil != err {
			errorLog("w.Write() failed: %v", err)
			return err
		}
		frame.fcs16 = pppfcs16(frame.fcs16, p)

		err = htran.encodeFrameInfo(frame)
		if nil != err {
			return err
		}

	} else if HDLC_CONTROL_RR == frame.control {
		b0 |= 0x01

		if frame.nr > 0x07 {
			panic("NR exceeds limit")
		}
		b0 |= frame.nr << 5

		p[0] = b0
		_, err = w.Write(p)
		if nil != err {
			errorLog("w.Write() failed: %v", err)
			return err
		}
		frame.fcs16 = pppfcs16(frame.fcs16, p)

		// FCS - frame control sum

		fcs16 := frame.fcs16
		p[0] = byte(^fcs16 & 0x00FF)
		_, err = w.Write(p)
		if nil != err {
			errorLog("w.Write() failed: %v", err)
			return err
		}
		frame.fcs16 = pppfcs16(frame.fcs16, p)
		p[0] = byte((^fcs16 & 0xFF00) >> 8)
		_, err = w.Write(p)
		if nil != err {
			errorLog("w.Write() failed: %v", err)
			return err
		}
		frame.fcs16 = pppfcs16(frame.fcs16, p)

	} else if HDLC_CONTROL_RNR == frame.control {
		b0 |= 0x01
		b0 |= 0x04

		if frame.nr > 0x07 {
			panic("NR exceeds limit")
		}
		b0 |= frame.nr << 5

		p[0] = b0
		_, err = w.Write(p)
		if nil != err {
			errorLog("w.Write() failed: %v", err)
			return err
		}
		frame.fcs16 = pppfcs16(frame.fcs16, p)

		// FCS - frame control sum

		fcs16 := frame.fcs16
		p[0] = byte(^fcs16 & 0x00FF)
		_, err = w.Write(p)
		if nil != err {
			errorLog("w.Write() failed: %v", err)
			return err
		}
		frame.fcs16 = pppfcs16(frame.fcs16, p)
		p[0] = byte((^fcs16 & 0xFF00) >> 8)
		_, err = w.Write(p)
		if nil != err {
			errorLog("w.Write() failed: %v", err)
			return err
		}
		frame.fcs16 = pppfcs16(frame.fcs16, p)

	} else if HDLC_CONTROL_SNRM == frame.control {
		b0 |= 0x01
		b0 |= 0x02
		b0 |= 0x80

		p[0] = b0
		_, err = w.Write(p)
		if nil != err {
			errorLog("w.Write() failed: %v", err)
			return err
		}
		frame.fcs16 = pppfcs16(frame.fcs16, p)

		err = htran.encodeFrameInfo(frame)
		if nil != err {
			return err
		}

	} else if HDLC_CONTROL_DISC == frame.control {
		b0 |= 0x01
		b0 |= 0x02
		b0 |= 0x40

		p[0] = b0
		_, err = w.Write(p)
		if nil != err {
			errorLog("w.Write() failed: %v", err)
			return err
		}
		frame.fcs16 = pppfcs16(frame.fcs16, p)

		// FCS - frame control sum

		fcs16 := frame.fcs16
		p[0] = byte(^fcs16 & 0x00FF)
		_, err = w.Write(p)
		if nil != err {
			errorLog("w.Write() failed: %v", err)
			return err
		}
		frame.fcs16 = pppfcs16(frame.fcs16, p)
		p[0] = byte((^fcs16 & 0xFF00) >> 8)
		_, err = w.Write(p)
		if nil != err {
			errorLog("w.Write() failed: %v", err)
			return err
		}
		frame.fcs16 = pppfcs16(frame.fcs16, p)

	} else if HDLC_CONTROL_UA == frame.control {
		b0 |= 0x01
		b0 |= 0x02
		b0 |= 0x20
		b0 |= 0x40

		p[0] = b0
		_, err = w.Write(p)
		if nil != err {
			errorLog("w.Write() failed: %v", err)
			return err
		}
		frame.fcs16 = pppfcs16(frame.fcs16, p)

		err = htran.encodeFrameInfo(frame)
		if nil != err {
			return err
		}

	} else if HDLC_CONTROL_DM == frame.control {
		b0 |= 0x01
		b0 |= 0x02
		b0 |= 0x04
		b0 |= 0x08

		p[0] = b0
		_, err = w.Write(p)
		if nil != err {
			errorLog("w.Write() failed: %v", err)
			return err
		}
		frame.fcs16 = pppfcs16(frame.fcs16, p)

		// FCS - frame control sum

		fcs16 := frame.fcs16
		p[0] = byte(^fcs16 & 0x00FF)
		_, err = w.Write(p)
		if nil != err {
			errorLog("w.Write() failed: %v", err)
			return err
		}
		frame.fcs16 = pppfcs16(frame.fcs16, p)
		p[0] = byte((^fcs16 & 0xFF00) >> 8)
		_, err = w.Write(p)
		if nil != err {
			errorLog("w.Write() failed: %v", err)
			return err
		}
		frame.fcs16 = pppfcs16(frame.fcs16, p)

	} else if HDLC_CONTROL_FRMR == frame.control {
		b0 |= 0x01
		b0 |= 0x02
		b0 |= 0x04
		b0 |= 0x80

		p[0] = b0
		_, err = w.Write(p)
		if nil != err {
			errorLog("w.Write() failed: %v", err)
			return err
		}
		frame.fcs16 = pppfcs16(frame.fcs16, p)

		err = htran.encodeFrameInfo(frame)
		if nil != err {
			return err
		}

	} else if HDLC_CONTROL_UI == frame.control {
		b0 |= 0x01
		b0 |= 0x02

		p[0] = b0
		_, err = w.Write(p)
		if nil != err {
			errorLog("w.Write() failed: %v", err)
			return err
		}
		frame.fcs16 = pppfcs16(frame.fcs16, p)

		err = htran.encodeFrameInfo(frame)
		if nil != err {
			return err
		}
	} else {
		errorLog("invalid control field value")
		return HdlcErrorInvalidValue
	}

	return nil
}

func (htran *HdlcTransport) lengthOfFrame(frame *HdlcFrame) (n int) {
	n = 0

	// format type
	n += 2

	// src, dst address

	n += htran.lengthServerAddress(frame)
	n += htran.lengthClientAddress(frame)

	// control
	n += 1

	// HCS - header control sum
	n += 2

	if (nil != frame.infoField) && len(frame.infoField) > 0 {
		n += len(frame.infoField)
		// FCS - frame control sum
		n += 2
	}

	return n
}

// decode frame format, address, control and information field

func (htran *HdlcTransport) decodeFrameFACI(frame *HdlcFrame, l int) (err error, n int) {
	var r io.Reader = htran.rw
	n = 0

	p := make([]byte, 1)
	var b0, b1 byte

	// expect first byte of format field
	_, err = io.ReadFull(r, p)
	if nil != err {
		if !isTimeOutErr(err) {
			errorLog("io.ReadFull() failed: %v", err)
		}
		return err, n
	}
	n++
	frame.fcs16 = pppfcs16(frame.fcs16, p)

	// format field
	if 0xA0 == p[0]&0xF0 {
		b0 = p[0]

		// expect last second byte of format field
		_, err = io.ReadFull(r, p)
		if nil != err {
			if !isTimeOutErr(err) {
				errorLog("io.ReadFull() failed: %v", err)
			}
			return err, n
		}
		n++
		frame.fcs16 = pppfcs16(frame.fcs16, p)
		b1 = p[0]

		frame.formatType = 0xA0

		// test segmentation bit
		if b0&0x08 > 0 {
			frame.segmentation = true
		} else {
			frame.segmentation = false
		}

		frame.length = int((uint16(b0&0x07) << 8) + uint16(b1))
		if hdlcDebug {
			p := make([]byte, frame.length)
			_, err = io.ReadFull(htran.rw, p)
			if nil != err {
				if !isTimeOutErr(err) {
					errorLog("io.ReadFull() failed: %v", err)
				}
				return err, n
			}
			fmt.Printf("received frame: %0X%0X %0X\n", b0, b1, p)
			frame.content = bytes.NewBuffer(p)
		}

		err, nn := htran.decodeFrameACI(frame, l+n)
		n += nn
		return err, n
	} else {
		return HdlcErrorMalformedSegment, n
	}
}

// encode frame format, address, control and information field

func (htran *HdlcTransport) encodeFrameFACI(frame *HdlcFrame) (err error) {
	var w io.Writer = htran.rw

	p := make([]byte, 1)
	var b0, b1 byte

	// frame format
	b0 |= 0xA0

	// segmentation
	if frame.segmentation {
		b0 |= 0x08
	}

	length := uint16(htran.lengthOfFrame(frame))
	if length > 0x07FF {
		errorLog("frame length exceeds limt")
		return HdlcErrorInvalidValue
	}
	b0 |= byte((0xFF00 & length) >> 8)
	b1 = byte(0x00FF & length)

	p[0] = b0
	_, err = w.Write(p)
	if nil != err {
		errorLog("w.Write() failed: %v", err)
		return err
	}
	frame.fcs16 = pppfcs16(frame.fcs16, p)

	p[0] = b1
	_, err = w.Write(p)
	if nil != err {
		errorLog("w.Write() failed: %v", err)
		return err
	}
	frame.fcs16 = pppfcs16(frame.fcs16, p)

	return htran.encodeFrameACI(frame)

}

func (htran *HdlcTransport) readFrame(direction int) (err error, frame *HdlcFrame) {
	p := make([]byte, 1)
	for {
		// expect opening flag
		_, err = io.ReadFull(htran.rw, p)
		if nil != err {
			if !isTimeOutErr(err) {
				errorLog("io.ReadFull() failed: %v", err)
			}
			return err, nil
		}
		if 0x7E == p[0] { // flag
			frame := new(HdlcFrame)
			frame.direction = direction
			frame.fcs16 = PPPINITFCS16

			err, _ = htran.decodeFrameFACI(frame, 0)
			if nil != err {
				if HdlcErrorMalformedSegment == err {
					// ignore malformed segment and try read next segment
					continue
				} else {
					return err, nil
				}
			} else {
				if hdlcDebug {
					htran.printFrame(frame, "read")
				}
				return nil, frame
			}

		} else {
			// ignore everything until leading flag arrives
			continue
		}
	}
}

func (htran *HdlcTransport) writeFrame(frame *HdlcFrame) (err error) {
	var w io.Writer = htran.rw
	if hdlcDebug {
		htran.printFrame(frame, "write...")
	}

	if 0 == frame.direction {
		errorLog("frame direction not specified")
		return HdlcErrorInvalidValue
	}
	if 0 == frame.control {
		errorLog("frame controltype not specified")
		return HdlcErrorInvalidValue
	}

	p := make([]byte, 1)

	// opening flag
	p[0] = 0x7E
	_, err = w.Write(p)
	if nil != err {
		errorLog("w.Write() failed: %v", err)
		return err
	}

	err = htran.encodeFrameFACI(frame)
	if nil != err {
		errorLog("w.Write() failed: %v", err)
		return err
	}
	if hdlcDebug {
		htran.printFrame(frame, "write ok")
	}

	return nil

}

func (htran *HdlcTransport) printFrame(frame *HdlcFrame, heading string) {

	var direction string
	switch frame.direction {
	case HDLC_FRAME_DIRECTION_CLIENT_INBOUND:
		direction = "client_inbound"
	case HDLC_FRAME_DIRECTION_CLIENT_OUTBOUND:
		direction = "client_outbound"
	case HDLC_FRAME_DIRECTION_SERVER_INBOUND:
		direction = "server_inbound"
	case HDLC_FRAME_DIRECTION_SERVER_OUTBOUND:
		direction = "server_outboun"
	default:
		panic("unknown value")
	}

	var control string
	switch frame.control {
	case HDLC_CONTROL_I:
		control = "I"
	case HDLC_CONTROL_RR:
		control = "RR"
	case HDLC_CONTROL_RNR:
		control = "RNR"
	case HDLC_CONTROL_SNRM:
		control = "SNRM"
	case HDLC_CONTROL_DISC:
		control = "DISC"
	case HDLC_CONTROL_UA:
		control = "UA"
	case HDLC_CONTROL_DM:
		control = "DM"
	case HDLC_CONTROL_FRMR:
		control = "FRMR"
	case HDLC_CONTROL_UI:
		control = "UI"
	default:
		panic("unknown value")
	}

	if frame.poll {
		if nil != frame.infoField {
			fmt.Printf("%s: frame(%s, %s, P, info(%d))\n", heading, direction, control, len(frame.infoField))
		} else {
			fmt.Printf("%s: frame(%s, %s, P)\n", heading, direction, control)
		}
	} else {
		if nil != frame.infoField {
			fmt.Printf("%s: frame(%s, %s, info(%d))\n", heading, direction, control, len(frame.infoField))
		} else {
			fmt.Printf("%s: frame(%s, %s)\n", heading, direction, control)
		}
	}
}

func (htran *HdlcTransport) handleHdlc(client bool) {
	var frame *HdlcFrame
	var segment *HdlcSegment
	var command *HdlcControlCommand
	var sending bool
	var err error
	var vs uint8
	var vr uint8
	var transmittedFramesCnt uint32
	var receivedFramesCnt uint32
	var snrmCommand *HdlcControlCommand

	const (
		STATE_CONNECTING = iota
		STATE_CONNECTED
		STATE_DISCONNECTING
		STATE_DISCONNECTED
	)
	var state int = STATE_DISCONNECTED

	segmentsNoAck := list.New() // unacknowledged segments
	framesToSend := list.New()  // frames scheduled to send in next poll

mainLoop:
	for {
		htran.controlQueueMtx.Lock()
		if htran.closed {
			break mainLoop
		}
		htran.controlQueueMtx.Unlock()

		if sending {

			// flush any frames waiting for next poll

			for framesToSend.Len() > 0 {
				frame = framesToSend.Front().Value.(*HdlcFrame)
				err = htran.writeFrame(frame)
				if nil != err {
					break mainLoop
				}
				framesToSend.Remove(framesToSend.Front())
			}

			// check for any pending priority command

			htran.controlQueueMtx.Lock()
			if htran.controlQueue.Len() > 0 {
				command = htran.controlQueue.Front().Value.(*HdlcControlCommand)
				if hdlcDebug {
					fmt.Printf("htran.handleHdlc(): command: %d\n", command.control)
				}
				htran.controlQueue.Remove(htran.controlQueue.Front())
			} else {
				command = nil
			}
			htran.controlQueueMtx.Unlock()

			// check for any pending segment to transmit

			if nil == command {
				if 0 == segmentsNoAck.Len() {
					htran.readQueueMtx.Lock()
					if htran.readQueue.Len() > 0 {
						segment = htran.readQueue.Front().Value.(*HdlcSegment)
					} else {
						segment = nil
					}
					htran.readQueueMtx.Unlock()
				} else {
					// transmit again all unacknowledged frames
					for e := segmentsNoAck.Front(); e != nil; e = e.Next() {
						frame = e.Value.(*HdlcFrame)
						err = htran.writeFrame(frame)
						if nil != err {
							break mainLoop
						}
					}
					continue mainLoop
				}
			}

			if (nil != command) && (HDLC_CONTROL_SNRM == command.control) {
				if STATE_DISCONNECTED == state {
					if hdlcDebug {
						fmt.Printf("hdlc.handleHdlc(): connecting\n")
					}
					snrmCommand = command
					if !client {
						htran.controlAck <- map[string]interface{}{"err": HdlcErrorNotDisconnected}
						continue mainLoop
					}
					frame = new(HdlcFrame)
					frame.fcs16 = PPPINITFCS16
					frame.poll = true
					receivedFramesCnt = 0
					frame.direction = HDLC_FRAME_DIRECTION_CLIENT_OUTBOUND
					frame.control = HDLC_CONTROL_SNRM

					snrm := command.snrm
					err = htran.encodeLinkParameters(frame, &snrm.maxInfoFieldLengthTransmit, &snrm.maxInfoFieldLengthReceive, &snrm.windowSizeTransmit, &snrm.windowSizeTransmit)
					if nil != err {
						break mainLoop
					}

					err = htran.writeFrame(frame)
					if nil != err {
						break mainLoop
					}
					state = STATE_CONNECTING
					sending = false
				} else {
					htran.controlAck <- map[string]interface{}{"err": HdlcErrorNotDisconnected}
				}
			} else if (nil != command) && (HDLC_CONTROL_DISC == command.control) {
				if client { // only client may disconnect the line.
					if STATE_CONNECTED == state {
						frame = new(HdlcFrame)
						frame.fcs16 = PPPINITFCS16
						frame.poll = true
						receivedFramesCnt = 0
						frame.direction = HDLC_FRAME_DIRECTION_CLIENT_OUTBOUND
						frame.control = HDLC_CONTROL_DISC
						err = htran.writeFrame(frame)
						if nil != err {
							break mainLoop
						}
						state = STATE_DISCONNECTING
						sending = false
					} else {
						htran.controlAck <- map[string]interface{}{"err": HdlcErrorNotConnected}
					}
				} else {
					htran.controlAck <- map[string]interface{}{"err": HdlcErrorNoAllowed}
				}
			} else if nil != segment {
				frame = new(HdlcFrame)
				frame.fcs16 = PPPINITFCS16
				if client {
					frame.direction = HDLC_FRAME_DIRECTION_CLIENT_OUTBOUND
				} else {
					frame.direction = HDLC_FRAME_DIRECTION_SERVER_OUTBOUND
				}
				frame.segmentation = !segment.last
				frame.control = HDLC_CONTROL_I
				frame.ns = vs
				frame.nr = vr
				frame.infoField = segment.p

				if (vs+1 == htran.modulus-1) || (transmittedFramesCnt == htran.windowSizeTransmit) || segment.last { // in modulo 8 mode available sequence numbers are in range 0..7
					// ran out of available sequence numbers or exceeded transmit window or encountered segment boundary, therefore wait for acknowledgement of all segments we transmitted so far
					frame.poll = true
					receivedFramesCnt = 0
					err = htran.writeFrame(frame)
					if nil != err {
						break mainLoop
					}
					segmentsNoAck.PushBack(frame)
					if segment.last {
						htran.writeAck <- map[string]interface{}{"err": nil}
					}
					vs += 1
					sending = false
					transmittedFramesCnt += 1

				} else {
					// just transmit frame without acknowledgement
					err = htran.writeFrame(frame)
					if nil != err {
						break mainLoop
					}
					segmentsNoAck.PushBack(frame)
					vs += 1
					transmittedFramesCnt += 1
				}
			} else {
				// nothing to transmit now

				// must poll the peer to avoid deadlocking in case higher layer is not sending anything yet because it hasn't
				// received whole response from peer
				if STATE_CONNECTING == state {
					frame = new(HdlcFrame)
					frame.fcs16 = PPPINITFCS16
					frame.poll = true
					receivedFramesCnt = 0
					frame.direction = HDLC_FRAME_DIRECTION_CLIENT_OUTBOUND
					frame.control = HDLC_CONTROL_SNRM

					snrm := snrmCommand.snrm
					err = htran.encodeLinkParameters(frame, &snrm.maxInfoFieldLengthTransmit, &snrm.maxInfoFieldLengthReceive, &snrm.windowSizeTransmit, &snrm.windowSizeTransmit)
					if nil != err {
						break mainLoop
					}

					err = htran.writeFrame(frame)
					if nil != err {
						break mainLoop
					}
					sending = false
				} else if STATE_CONNECTED == state {
					frame = new(HdlcFrame)
					frame.fcs16 = PPPINITFCS16
					if client {
						frame.direction = HDLC_FRAME_DIRECTION_CLIENT_OUTBOUND
					} else {
						frame.direction = HDLC_FRAME_DIRECTION_CLIENT_OUTBOUND
					}
					frame.poll = true
					receivedFramesCnt = 0
					frame.control = HDLC_CONTROL_RR
					frame.ns = vs
					frame.nr = vr
					err := htran.writeFrame(frame)
					if nil != err {
						break mainLoop
					}
					sending = false
				} else if STATE_DISCONNECTING == state {
					frame = new(HdlcFrame)
					frame.fcs16 = PPPINITFCS16
					frame.poll = true
					receivedFramesCnt = 0
					frame.direction = HDLC_FRAME_DIRECTION_CLIENT_OUTBOUND
					frame.control = HDLC_CONTROL_DISC
					err = htran.writeFrame(frame)
					if nil != err {
						break mainLoop
					}
					sending = false
				} else if STATE_DISCONNECTED == state {
					// no need to transmit anything, just wait for client to connect again
					sending = false
				} else {
					panic("unknown state")
				}
			}

		} else {
			// receiving

			htran.rw.SetReadDeadline(time.Now().Add(htran.responseTimeout))
			if client {
				err, frame = htran.readFrame(HDLC_FRAME_DIRECTION_CLIENT_INBOUND)
			} else {
				err, frame = htran.readFrame(HDLC_FRAME_DIRECTION_SERVER_INBOUND)
			}
			if nil != err {
				if isTimeOutErr(err) { // timeout occured
					if client {
						// According ISO 13239 standard it is responsibility of client to do time-out no-reply recovery.
						// Per standard it is responsibility of client to do time-out no-reply recovery and
						// in case of timeout client may transmit  even if it did not receive the poll.
						sending = true
						continue mainLoop
					} else {
						// If server try to receive frame again (if frame does not arrive client is going to time out and must send us frame)
						continue mainLoop
					}
				} else {
					break mainLoop
				}
			}

			// Proccess received frame.

			// Transmit if received the poll.
			if frame.poll {
				sending = true
				transmittedFramesCnt = 0
			}

			if HDLC_CONTROL_I == frame.control {
				if STATE_CONNECTED == state {

					// check for any reason to reject received frame

					var reasonForReject []byte
					if frame.nr-1 > vs {
						// peer acknowledged sequence number which is ahead of our current sequnece number
						// reject frame
						reasonForReject = []byte("received acknowledgement for not yet tramsitted sequence number")
					}
					if frame.nr-1 < segmentsNoAck.Front().Value.(HdlcFrame).ns {
						reasonForReject = []byte("received acknowledgement for already acknowledged sequence number")
					}
					if nil != reasonForReject {
						frame = new(HdlcFrame)
						frame.fcs16 = PPPINITFCS16
						if client {
							frame.direction = HDLC_FRAME_DIRECTION_CLIENT_OUTBOUND
						} else {
							frame.direction = HDLC_FRAME_DIRECTION_CLIENT_OUTBOUND
						}
						frame.control = HDLC_CONTROL_FRMR
						frame.infoField = reasonForReject
						framesToSend.PushBack(frame)
						continue mainLoop
					}

					if frame.ns == vr {

						// Accept in sequence frame.
						receivedFramesCnt += 1
						if htran.modulus-1 == vr {
							vr = 0
						} else {
							vr += 1
						}

						segment = new(HdlcSegment)
						segment.p = frame.infoField
						segment.last = !frame.segmentation
						htran.writeQueueMtx.Lock()
						htran.writeQueue.PushBack(segment)
						htran.writeQueueMtx.Unlock()
						if segment.last {
							htran.writeAck <- map[string]interface{}{"err": nil}
						}

						// acknowledge received frames
						for e := segmentsNoAck.Front(); e != nil; e = e.Next() {
							if e.Value.(HdlcFrame).ns <= frame.nr-1 {
								segmentsNoAck.Remove(e)
							}
						}

					} else {
						// see: ISO/IEC 13239 - 6.11.4.2.3 Reception of incorrect frames
						// ignore segment but reuse the P/F, nr

						// this frame with correct sequence number frame.nr, it may just arrive faster ahead in sequnce or arrive deleayed in sequnce

						// acknowledge received frames
						for e := segmentsNoAck.Front(); e != nil; e = e.Next() {
							if e.Value.(HdlcFrame).ns <= frame.nr-1 {
								segmentsNoAck.Remove(e)
							}
						}

					}
				} else {
					// ignore frame
				}

			} else if HDLC_CONTROL_RR == frame.control {
				if STATE_CONNECTED == state {
					// check for any reason to reject received frame

					var reasonForReject []byte
					if frame.nr-1 > vs {
						// peer acknowledged sequence number which is ahead of our current sequnece number
						// reject frame
						reasonForReject = []byte("received acknowledgement for not yet tramsitted sequence number")
					}
					if frame.nr-1 < segmentsNoAck.Front().Value.(HdlcFrame).ns {
						reasonForReject = []byte("received acknowledgement for already acknowledged sequence number")
					}
					if nil != reasonForReject {
						frame = new(HdlcFrame)
						frame.fcs16 = PPPINITFCS16
						if client {
							frame.direction = HDLC_FRAME_DIRECTION_CLIENT_OUTBOUND
						} else {
							frame.direction = HDLC_FRAME_DIRECTION_CLIENT_OUTBOUND
						}
						frame.control = HDLC_CONTROL_FRMR
						frame.infoField = reasonForReject
						framesToSend.PushBack(frame)
						continue mainLoop
					}

					// acknowledge received frames
					for e := segmentsNoAck.Front(); e != nil; e = e.Next() {
						if e.Value.(HdlcFrame).ns <= frame.nr-1 {
							segmentsNoAck.Remove(e)
						}
					}
				} else {
					// ignore frame
				}
			} else if HDLC_CONTROL_RNR == frame.control {
				if STATE_CONNECTED == state {
					// check for any reason to reject received frame

					var reasonForReject []byte
					if frame.nr-1 > vs {
						// peer acknowledged sequence number which is ahead of our current sequnece number
						// reject frame
						reasonForReject = []byte("received acknowledgement for not yet tramsitted sequence number")
					}
					if frame.nr-1 < segmentsNoAck.Front().Value.(HdlcFrame).ns {
						reasonForReject = []byte("received acknowledgement for already acknowledged sequence number")
					}
					if nil != reasonForReject {
						frame = new(HdlcFrame)
						frame.fcs16 = PPPINITFCS16
						if client {
							frame.direction = HDLC_FRAME_DIRECTION_CLIENT_OUTBOUND
						} else {
							frame.direction = HDLC_FRAME_DIRECTION_CLIENT_OUTBOUND
						}
						frame.control = HDLC_CONTROL_FRMR
						frame.infoField = reasonForReject
						framesToSend.PushBack(frame)
						continue mainLoop
					}

					// acknowledge received frames
					for e := segmentsNoAck.Front(); e != nil; e = e.Next() {
						if e.Value.(HdlcFrame).ns <= frame.nr-1 {
							segmentsNoAck.Remove(e)
						}
					}
				} else {
					// ignore frame
				}
			} else if HDLC_CONTROL_SNRM == frame.control {
				if STATE_DISCONNECTED == state {

					err, maxInfoFieldLengthTransmit, maxInfoFieldLengthReceive, windowSizeTransmit, windowSizeReceive := htran.decodeLinkParameters(frame)
					if nil != err {
						break mainLoop
					}

					frame = new(HdlcFrame)
					frame.fcs16 = PPPINITFCS16
					frame.poll = true
					frame.direction = HDLC_FRAME_DIRECTION_SERVER_OUTBOUND // only client may send SNRM
					frame.control = HDLC_CONTROL_UA

					// negotiate link parameters

					if nil != maxInfoFieldLengthTransmit {
						if *maxInfoFieldLengthTransmit > htran.maxInfoFieldLengthReceive {
							*maxInfoFieldLengthTransmit = htran.maxInfoFieldLengthReceive
						} else {
							htran.maxInfoFieldLengthReceive = *maxInfoFieldLengthTransmit
						}
					} else {
						*maxInfoFieldLengthTransmit = htran.maxInfoFieldLengthReceive
					}

					if nil != maxInfoFieldLengthReceive {
						if *maxInfoFieldLengthReceive > htran.maxInfoFieldLengthTransmit {
							*maxInfoFieldLengthReceive = htran.maxInfoFieldLengthTransmit
						} else {
							htran.maxInfoFieldLengthTransmit = *maxInfoFieldLengthReceive
						}
					} else {
						*maxInfoFieldLengthReceive = htran.maxInfoFieldLengthTransmit
					}

					if nil != windowSizeTransmit {
						if *windowSizeTransmit > htran.windowSizeTransmit {
							*windowSizeTransmit = htran.windowSizeTransmit
						} else {
							htran.windowSizeTransmit = *windowSizeTransmit
						}
					} else {
						*windowSizeTransmit = htran.windowSizeTransmit
					}

					if nil != windowSizeReceive {
						if *windowSizeReceive > htran.windowSizeReceive {
							*windowSizeReceive = htran.windowSizeReceive
						} else {
							htran.windowSizeReceive = *windowSizeReceive
						}
					} else {
						*windowSizeReceive = htran.windowSizeReceive
					}

					err = htran.encodeLinkParameters(frame, maxInfoFieldLengthTransmit, maxInfoFieldLengthReceive, windowSizeTransmit, windowSizeTransmit)
					if nil != err {
						break mainLoop
					}

					vs = 0
					vr = 0
					client = false
					fmt.Printf("@@@@@@@@@@@@@@@@@@@@@@@ cp 200\n")
					framesToSend.PushBack(frame)
				}
			} else if HDLC_CONTROL_DISC == frame.control {

				if STATE_CONNECTED == state {
					frame = new(HdlcFrame)
					frame.fcs16 = PPPINITFCS16
					frame.poll = true
					frame.direction = HDLC_FRAME_DIRECTION_SERVER_OUTBOUND // only client may send DISC
					frame.control = HDLC_CONTROL_UA
					state = STATE_DISCONNECTED
					framesToSend.PushBack(frame)
				} else if STATE_DISCONNECTED == state {
					frame = new(HdlcFrame)
					frame.fcs16 = PPPINITFCS16
					frame.poll = true
					frame.direction = HDLC_FRAME_DIRECTION_SERVER_OUTBOUND // only client may send DISC
					frame.control = HDLC_CONTROL_DM
					framesToSend.PushBack(frame)
				} else {
					// ignore frame
				}
				framesToSend.PushBack(frame)
			} else if HDLC_CONTROL_DM == frame.control {
				if STATE_DISCONNECTING == state {
					state = STATE_DISCONNECTED
				} else {
					// ignore frame
				}
			} else if HDLC_CONTROL_UA == frame.control {
				if STATE_DISCONNECTING == state {
					state = STATE_DISCONNECTED
					htran.controlAck <- map[string]interface{}{"err": nil}
				} else if STATE_CONNECTING == state {

					// negotiate link parameters

					err, maxInfoFieldLengthTransmit, maxInfoFieldLengthReceive, windowSizeTransmit, windowSizeReceive := htran.decodeLinkParameters(frame)
					if nil != err {
						break mainLoop
					}
					if nil != maxInfoFieldLengthTransmit {
						htran.maxInfoFieldLengthTransmit = *maxInfoFieldLengthTransmit
					} else {
						htran.maxInfoFieldLengthTransmit = 128
					}
					if nil != maxInfoFieldLengthReceive {
						htran.maxInfoFieldLengthReceive = *maxInfoFieldLengthReceive
					} else {
						htran.maxInfoFieldLengthReceive = 128
					}
					if nil != windowSizeTransmit {
						htran.windowSizeTransmit = *windowSizeTransmit
					} else {
						htran.windowSizeTransmit = 1
					}
					if nil != windowSizeReceive {
						htran.windowSizeReceive = *windowSizeReceive
					} else {
						htran.windowSizeReceive = 1
					}

					state = STATE_CONNECTED
					vs = 0
					vr = 0
					client = true
					htran.controlAck <- map[string]interface{}{"err": nil}
				} else {
					// ignore frame
				}
			} else if HDLC_CONTROL_UI == frame.control {
				if STATE_CONNECTED == state {
					panic("handling of UI frames is not implementd")
				} else {
					// ignore frame
				}
			} else if HDLC_CONTROL_FRMR == frame.control {
				errorLog("frame rejected, reason: %s", string(frame.infoField))
				err = HdlcErrorFrameRejected
				break mainLoop
			} else {
				// ignore frame
			}
		}
	}
	if nil != err {
		go func() {
			htran.writeAck <- map[string]interface{}{"err": err}
			close(htran.writeAck)
		}()
		go func() {
			htran.readAck <- map[string]interface{}{"err": err}
			close(htran.readAck)
		}()
		go func() {
			htran.controlAck <- map[string]interface{}{"err": err}
			close(htran.controlAck)
		}()
	}
}
