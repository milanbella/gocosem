package gocosem

const (
	HDLC_FRAME_DIRECTION_CLIENT_INBOUND = iota
	HDLC_FRAME_DIRECTION_CLIENT_OUTBOUND
	HDLC_FRAME_DIRECTION_SERVER_INBOUND
	HDLC_FRAME_DIRECTION_SERVER_OUTBOUND
)

type HdlcTransport struct {
	rwc            io.ReadWriteCloser
	ch             chan *DlmsMessage
	linkConnected  bool
	readFrameState int
}

type HdlcFrame struct {
	direction        int
	formatType       uint8
	segmentation     bool
	length           uint16
	logicalDeviceId  uint16
	physicalDeviceId *uint16 // may not be present
	clientId         uint8
}

var ErrorMalformedSegment = errors.New("ErrorMalformedSegment")

func NewHdlcTransport(rwc io.ReadWriteCloser) *HdlcTransport {
	htran = new(HdlcTrandport)
	htran.linkConnected = false
	htran.readFrameState = HDLC_READ_FRAME_STATE_FIRST_SEGMENT
	return htran
}

func (htran *HDLCTransport) parseServerAddress(frame *HdlcFrame, r io.Reader) (err error) {

	var b0, b1, b2, b3 byte

	//TODO: implemet cosem: 8.4.2.5 Handling inopportune address lengths in the server

	err = binary.Read(r, binary.BigEndian, &b0)
	if nil != err {
		errorLog("binary.Read() failed, err: %v", err)
		return err
	}
	if b0&0x01 > 0 {
		frame.dst.logicalDeviceId = (uint16(b0) & 0x00FE) >> 1
		frame.dst.physicalDeviceId = nil
	} else {
		err = binary.Read(r, binary.BigEndian, &b1)
		if nil != err {
			errorLog("binary.Read() failed, err: %v", err)
			return err
		}
		if b1&0x01 > 0 {
			frame.dst.logicalDeviceId = (uint16(b0) & 0x00FE) >> 1
			physicalDeviceId := new(uint16)
			physicalDeviceId = (uint16(b1) & 0x00FE) >> 1
			frame.dst.physicalDeviceId = physicalDeviceId
		} else {
			err = binary.Read(r, binary.BigEndian, &b2)
			if nil != err {
				errorLog("binary.Read() failed, err: %v", err)
				return err
			}
			if b2&0x01 > 0 {
				errorLog("server address is shorter then 4 bytes")
				return ErrorMalformedSegment
			}
			err = binary.Read(r, binary.BigEndian, &b3)
			if nil != err {
				errorLog("binary.Read() failed, err: %v", err)
				return err
			}
			if b3&0x01 > 0 {
				frame.dst.logicalDeviceId = ((uint16(b0)&0x00FE)>>1)<<7 + ((uint16(b1) & 0x00FE) >> 1)
				physicalDeviceId := new(uint16)
				physicalDeviceId = ((uint16(b2)&0x00FE)>>1)<<7 + ((uint16(b3) & 0x00FE) >> 1)
				frame.dst.physicalDeviceId = physicalDeviceId
			} else {
				errorLog("server address is longer then 4 bytes")
				return ErrorMalformedSegment
			}
		}
	}
}

func (htran *HDLCTransport) parseClientAddress(frame *HdlcFrame, r io.Reader) (err error) {
	var b0 byte

	err = binary.Read(r, binary.BigEndian, &b0)
	if nil != err {
		errorLog("binary.Read() failed, err: %v", err)
		return err
	}
	if b0&0x01 > 0 {
		frame.clientId = (uint8(b0) & 0xFE) >> 1
	} else {
		errorLog("client address is longer then 1 byte")
		return ErrorMalformedSegment
	}
}

func (htran *HDLCTransport) parseFrameBody(frame *HdlcFrame, p []byte) (err error) {
	r := bytes.NewBuffer(p)

	if HDLC_FRAME_DIRECTION_SERVER_INBOUND == frame.direction {
		err = htran.parseServerAddress(r)
		if nil != err {
			return err
		}
		err = htran.parseClientAddress(r)
		if nil != err {
			return err
		}
	} else if HDLC_FRAME_DIRECTION_CLIENT_INBOUND == frame.direction {
		err = htran.parseClientAddress(r)
		if nil != err {
			return err
		}
		err = htran.parseServerAddress(r)
		if nil != err {
			return err
		}
	} else {
		panic("frame direction is not inbound")
	}
}

func (htran *HDLCTransport) readFrameBody(direction int) (err error, frame *HdlcFrame) {
	var n int

	p := make([]byte, 1)
	var b0, b1 byte

	// expect first byte of format field
	n, err = rwc.Read(p)
	if nil != err {
		errorLog("rwc.Read() failed: %v", err)
		return err, nil
	}
	if 0xA0 == p[0]&0xF0 {
		b0 = p[0]

		// expect last second byte of format field
		n, err = rwc.Read(p)
		if nil != err {
			errorLog("rwc.Read() failed: %v", err)
			return err, nil
		}
		b1 = p[0]

		frame := new(HdlcFrame)
		frame.direction = direction

		// read remainder of frame body
		frame.formatType = 0xA0
		frame.length = (uint16(b0&0x70) << 8) + uint16(b1)
		if b0&0x08 > 0 {
			frame.segmentation = true
		} else {
			frame.segmentation = false
		}
		buf := new(bytes.Buffer)
		p = make([]byte, frame.length-2)
		err = binary.Read(htran.rwc, binary.BigEndian, p)
		if nil != err {
			errorLog(" binary.Read() failed, err: %v\n", err)
			return err, nil
		}
		err = htrans.parseFrameBody(frame, p)
		return err, frame
	} else {
		return ErrorMalformedSegment, nil
	}
}

func (htran *HDLCTransport) readFrame(direction int) (err error, frame *HdlcFrame) {
	var n int
	p := make([]byte, 1)
	for {
		// expect flag
		n, err = rwc.Read(p)
		if nil != err {
			errorLog("rwc.Read() failed: %v", err)
			return err, nil
		}
		if 0x7E == p[0] {
			err, frame = readFrameBody(direction)
			if nil != err {
				if ErrorMalformedSegment == err {
					// ignore malformed segment and try read next segment
					continue
				} else {
					return nil, frame
				}
			}
		} else {
			// ignore everything until leading flag arrives
			continue
		}
	}
}
