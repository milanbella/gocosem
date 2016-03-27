package gocosem

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type appContext uint8

const (
	LogicalName_NoCiphering   appContext = 1
	ShortName_NoCiphering                = 2
	LogicalName_WithCiphering            = 3
	ShortName_WithCiphering              = 4
)

type authMechanism uint8

const (
	NoSecurity        authMechanism = 0
	LowLevelSecurity                = 1
	HighLevelSecurity               = 2
)

type AARQ struct {
	appCtxt   appContext
	authMech  authMechanism
	authValue string
}

func (aarq *AARQ) encode() ([]byte, error) {
	appCtxtName := []byte{0xA1, 0x09, 0x06, 0x07, 0x60, 0x85, 0x74, 0x05, 0x08, 0x01,
		byte(aarq.appCtxt), // application context
	}
	senderAcse := []byte{0x8A, 0x02, 0x07,
		0x80, // requirement of authentication
	}
	authMechName := []byte{0x8B, 0x07, 0x60, 0x85, 0x74, 0x05, 0x08, 0x02,
		byte(aarq.authMech), // authentication mechanism
	}
	vlen := byte(len(aarq.authValue))
	callingAuth := append([]byte{0xAC, vlen + 2, 0x80, vlen},
		[]byte(aarq.authValue)..., // calling auth value
	)
	userInfo := []byte{0xBE, 0x10, 0x04, 0x0E,
		0x01, 0x00, 0x00, 0x00, 0x06, 0x5F, 0x1F, 0x04, 0x00, // initiate request
		0x00, 0x7E, 0x1F, // conformance block
		0x04, 0xB0, // max receive pdu length
	}

	var buf bytes.Buffer
	buf.Write([]byte{0x60, 46 + vlen})
	buf.Write(appCtxtName)
	buf.Write(senderAcse)
	buf.Write(authMechName)
	buf.Write(callingAuth)
	buf.Write(userInfo)

	return buf.Bytes(), nil
}

type assocResult uint8

const (
	AssociationAccepted          assocResult = 0
	AssociationRejectedPermanent             = 1
	AssociationRejectedTransient             = 2
)

type assocDiagnostic uint8

const (
	DiagNull                       assocDiagnostic = 0
	DiagNoReason                                   = 1
	DiagAppContextNotSupported                     = 2
	DiagAuthMechanismNotRecognized                 = 11
	DiagAuthMechanismRequired                      = 12
	DiagAuthenticationFailure                      = 13
	DiagAuthenticationRequired                     = 13
)

type AARE struct {
	appCtxt    appContext
	result     assocResult
	diagnostic assocDiagnostic
}

func (aare *AARE) decode(b []byte) (err error) {
	var (
		tag         byte
		size        uint8
		blockHeader = make([]byte, 3)
		initResp    = make([]byte, 7)
		confBlock   = make([]byte, 3)
		maxPduSize  uint16
		vaaName     uint16
	)

	var buf = bytes.NewReader(b)
	read := func(d interface{}) {
		if err == nil {
			err = binary.Read(buf, binary.BigEndian, d)
		}
	}
	skip := func(n int64) {
		if err == nil {
			_, err = buf.Seek(n, 1)
		}
	}
	read(&tag)
	read(&size)
	skip(10)
	read(&aare.appCtxt)
	skip(4)
	read(&aare.result)
	skip(6)
	read(&aare.diagnostic)
	skip(1)
	read(&blockHeader)
	read(&initResp)
	read(&confBlock)
	read(&maxPduSize)
	read(&vaaName)

	if tag != 0x61 {
		return fmt.Errorf("invalid AARE")
	}

	return err
}
