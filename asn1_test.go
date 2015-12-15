package gocosem

import (
	"bytes"
	"fmt"
	"io"
	"testing"
)

func printBuffer(buf bytes.Buffer) {
	for {
		c, err := buf.ReadByte()
		if nil == err {
			fmt.Printf("%02X ", c)
		} else if io.EOF == err {
			break
		} else {
			panic(fmt.Sprintf("buf.RaedByte() failed, err: %v", err))
		}
	}
	fmt.Println()
}

func bytesEquals(t *testing.T, a []byte, b []byte, report bool) bool {
	if len(a) != len(b) {
		if report {
			t.Logf("length differs")
		}
		return false
	}
	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			if report {
				t.Logf("bytes differ, index: %v", i)
			}
			return false
		}
	}
	return true
}

func TestX_encode_AARQapdu(t *testing.T) {
	fmt.Println("TestEncodeAARQapdu()")
	/*

		60 36 A1 09 06 07 60 85 74 05 08 01 01 8A 02 07 80 8B 07 60 85 74 05 08 02 01 AC 0A 80 08 31 32 33 34 35 36 37 38 BE 10 04 0E 01 00 00 00 06 5F 1F 04 00 00 7E 1F 04 B0

						DLMS-APDU CHOICE
						  aarq AARQ-apdu SEQUENCE: tag = [APPLICATION 0] constructed; length = 54
						    application-context-name : tag = [1] constructed; length = 9
						      Application-context-name OBJECT IDENTIFIER: tag = [UNIVERSAL 6] primitive; length = 7
						        { 2 16 756 5 8 1 1 }
						    sender-acse-requirements ACSE-requirements BIT STRING: tag = [10] primitive; length = 2
						      0x0780
						    mechanism-name Mechanism-name OBJECT IDENTIFIER: tag = [11] primitive; length = 7
						      { 2 16 756 5 8 2 1 }
						    calling-authentication-value : tag = [12] constructed; length = 10
						      Authentication-value CHOICE
						        charstring GraphicString: tag = [0] primitive; length = 8
						          "12345678"
						    user-information : tag = [30] constructed; length = 16
						      Association-information OCTET STRING: tag = [UNIVERSAL 4] primitive; length = 14
						        0x01000000065f1f0400007e1f04b0
						Successfully decoded 56 bytes.
						rec1value XDLMS-APDU ::= aarq :
						  {
						    application-context-name { 2 16 756 5 8 1 1 },
						    sender-acse-requirements { authentication },
						    mechanism-name { 2 16 756 5 8 2 1 },
						    calling-authentication-value charstring : "12345678",
						    user-information '01000000065F1F0400007E1F04B0'H
						  }
	*/
	var buf bytes.Buffer
	var aarq AARQapdu

	aarq.applicationContextName = tAsn1ObjectIdentifier([]uint{2, 16, 756, 5, 8, 1, 1})
	aarq.senderAcseRequirements = &tAsn1BitString{
		buf:        []byte{0x80},
		bitsUnused: 7,
	}
	mechanismName := (tAsn1ObjectIdentifier)([]uint{2, 16, 756, 5, 8, 2, 1})
	aarq.mechanismName = &mechanismName
	aarq.callingAuthenticationValue = new(tAsn1Choice)
	password := tAsn1GraphicString([]byte("12345678"))
	aarq.callingAuthenticationValue.setVal(int(C_Authentication_value_PR_charstring), &password)
	userInformation := tAsn1OctetString([]byte{0x01, 0x00, 0x00, 0x00, 0x06, 0x5F, 0x1F, 0x04, 0x00, 0x00, 0x7E, 0x1F, 0x04, 0xB0})
	aarq.userInformation = &userInformation

	buf = encode_AARQapdu(&aarq)
	printBuffer(buf)

	expectBuf := []byte{0x60, 0x36, 0xA1, 0x09, 0x06, 0x07, 0x60, 0x85, 0x74, 0x05, 0x08, 0x01, 0x01, 0x8A, 0x02, 0x07, 0x80, 0x8B, 0x07, 0x60, 0x85, 0x74, 0x05, 0x08, 0x02, 0x01, 0xAC, 0x0A, 0x80, 0x08, 0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37, 0x38, 0xBE, 0x10, 0x04, 0x0E, 0x01, 0x00, 0x00, 0x00, 0x06, 0x5F, 0x1F, 0x04, 0x00, 0x00, 0x7E, 0x1F, 0x04, 0xB0}

	if !bytesEquals(t, buf.Bytes(), expectBuf, true) {
		t.Errorf("bytes don't match")
	}
}
