// -build ignore

package gocosem

import (
	"bytes"
	"fmt"
	"io"
	"testing"
)

func printBuffer(t *testing.T, inb []byte) {
	buf := bytes.NewBuffer(inb)
	str := ""
	for {
		c, err := buf.ReadByte()
		if nil == err {
			str += fmt.Sprintf("% 02X ", c)
		} else if io.EOF == err {
			break
		} else {
			panic(fmt.Sprintf("buf.RaedByte() failed, err: %v", err))
		}
	}
	t.Logf("%s", str)
}

func byteEquals(t *testing.T, a []byte, b []byte, report bool) bool {
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

func uint32Equals(t *testing.T, a []uint32, b []uint32, report bool) bool {
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

func TestAsn1_encode_AARQapdu_for_tcp_meter(t *testing.T) {
	t.Logf("TestAsn1_encode_AARQapdu_for_tcp_meter")
	/*

			60 36 A1 09 06 07 60 85 74 05 08 01 01 8A 02 07 80 8B 07 60 85 74 05 08 02 01 AC 0A 80 08 31 32 33 34 35 36 37 38 BE 10 04 0E 01 00 00 00 06 5F 1F 04 00 00 7E 1F 04 B0
			60    36    A1    09    06    07    60    85    74    05    08    01    01    8A    02    07    80    8B    07    60    85    74    05    08    02    01    AC    0A    80    08    31    32    33    34    35    36    37    38    BE    10    04    0E    01    00    00    00    06    5F    1F    04    00    00    7E    1F    04    B0
		   0x60, 0x36, 0xA1, 0x09, 0x06, 0x07, 0x60, 0x85, 0x74, 0x05, 0x08, 0x01, 0x01, 0x8A, 0x02, 0x07, 0x80, 0x8B, 0x07, 0x60, 0x85, 0x74, 0x05, 0x08, 0x02, 0x01, 0xAC, 0x0A, 0x80, 0x08, 0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37, 0x38, 0xBE, 0x10, 0x04, 0x0E, 0x01, 0x00, 0x00, 0x00, 0x06, 0x5F, 0x1F, 0x04, 0x00, 0x00, 0x7E, 0x1F, 0x04, 0xB0}

							XDLMS-APDU CHOICE
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
	var aarq AARQapdu

	aarq.applicationContextName = tAsn1ObjectIdentifier([]uint32{2, 16, 756, 5, 8, 1, 1})
	aarq.senderAcseRequirements = &tAsn1BitString{
		buf:        []byte{0x80},
		bitsUnused: 7,
	}
	mechanismName := (tAsn1ObjectIdentifier)([]uint32{2, 16, 756, 5, 8, 2, 1})
	aarq.mechanismName = &mechanismName
	password := tAsn1GraphicString([]byte("12345678"))
	aarq.callingAuthenticationValue = new(tAsn1Choice)
	aarq.callingAuthenticationValue.setVal(0, password)
	userInformation := tAsn1OctetString([]byte{0x01, 0x00, 0x00, 0x00, 0x06, 0x5F, 0x1F, 0x04, 0x00, 0x00, 0x7E, 0x1F, 0x04, 0xB0})
	aarq.userInformation = &userInformation

	var buf bytes.Buffer
	err := encode_AARQapdu(&buf, &aarq)
	if nil != err {
		t.Fatalf("encode_AARQapdu() failed")
	}
	b := buf.Bytes()
	printBuffer(t, b)

	expectB := []byte{0x60, 0x36, 0xA1, 0x09, 0x06, 0x07, 0x60, 0x85, 0x74, 0x05, 0x08, 0x01, 0x01, 0x8A, 0x02, 0x07, 0x80, 0x8B, 0x07, 0x60, 0x85, 0x74, 0x05, 0x08, 0x02, 0x01, 0xAC, 0x0A, 0x80, 0x08, 0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37, 0x38, 0xBE, 0x10, 0x04, 0x0E, 0x01, 0x00, 0x00, 0x00, 0x06, 0x5F, 0x1F, 0x04, 0x00, 0x00, 0x7E, 0x1F, 0x04, 0xB0}

	if !byteEquals(t, b, expectB, true) {
		t.Fatalf("bytes don't match")
	}
}

func TestAsn1_decode_AARQapdu_for_tcp_meter(t *testing.T) {
	t.Logf("TestAsn1_decode_AARQapdu_for_tcp_meter")
	/*

		60 36 A1 09 06 07 60 85 74 05 08 01 01 8A 02 07 80 8B 07 60 85 74 05 08 02 01 AC 0A 80 08 31 32 33 34 35 36 37 38 BE 10 04 0E 01 00 00 00 06 5F 1F 04 00 00 7E 1F 04 B0

						XDLMS-APDU CHOICE
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

	b := []byte{0x60, 0x36, 0xA1, 0x09, 0x06, 0x07, 0x60, 0x85, 0x74, 0x05, 0x08, 0x01, 0x01, 0x8A, 0x02, 0x07, 0x80, 0x8B, 0x07, 0x60, 0x85, 0x74, 0x05, 0x08, 0x02, 0x01, 0xAC, 0x0A, 0x80, 0x08, 0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37, 0x38, 0xBE, 0x10, 0x04, 0x0E, 0x01, 0x00, 0x00, 0x00, 0x06, 0x5F, 0x1F, 0x04, 0x00, 0x00, 0x7E, 0x1F, 0x04, 0xB0}
	buf := bytes.NewReader(b)
	err, aarq := decode_AARQapdu(buf)
	if nil != err {
		t.Fatalf("decode_AARQapdu() failed")
	}

	if !uint32Equals(t, aarq.applicationContextName, []uint32{2, 16, 756, 5, 8, 1, 1}, true) {
		t.Fatalf("aarq.applicationContextName don't match")
	}
	if !byteEquals(t, ([]byte)(aarq.senderAcseRequirements.buf), []byte{0x80}, true) {
		t.Fatalf("aarq.senderAcseRequirements don't match")
	}
	if aarq.senderAcseRequirements.bitsUnused != 7 {
		t.Fatalf("aarq.senderAcseRequirements don't match")
	}
	if !uint32Equals(t, *(*[]uint32)(aarq.mechanismName), []uint32{2, 16, 756, 5, 8, 2, 1}, true) {
		t.Fatalf("aarq.mechanismName don't match")
	}
	if aarq.callingAuthenticationValue.tag != 0 {
		t.Fatalf("aarq.callingAuthenticationValue don't match")
	}
	if !byteEquals(t, ([]byte)(aarq.callingAuthenticationValue.val.(tAsn1GraphicString)), []byte("12345678"), true) {
		t.Fatalf("aarq.callingAuthenticationValue don't match")
	}
	if !byteEquals(t, *(*[]byte)(aarq.userInformation), []byte{0x01, 0x00, 0x00, 0x00, 0x06, 0x5F, 0x1F, 0x04, 0x00, 0x00, 0x7E, 0x1F, 0x04, 0xB0}, true) {
		t.Fatalf("aarq.userInformation don't match")
	}

}

func TestAsn1_encode_AARQapdu_for_hdlc_meter(t *testing.T) {
	t.Logf("TestAsn1_encode_AARQapdu_for_hdlc_meter")
	/*

		60 36 A1 09 06 07 60 85 74 05 08 01 01 8A 02 07 80 8B 07 60 85 74 05 08 02 01 AC 0A 80 08 31 32 33 34 35 36 37 38 BE 10 04 0E 01 00 00 00 06 5F 1F 04 00 FF FF FF 02 00

		XDLMS-APDU CHOICE
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
		        0x01000000065f1f0400ffffff0200
		Successfully decoded 56 bytes.
		rec1value XDLMS-APDU ::= aarq :
		  {
		    application-context-name { 2 16 756 5 8 1 1 },
		    sender-acse-requirements { authentication },
		    mechanism-name { 2 16 756 5 8 2 1 },
		    calling-authentication-value charstring : "12345678",
		    user-information '01000000065F1F0400FFFFFF0200'H
		  }

	*/
	var aarq AARQapdu

	aarq.applicationContextName = tAsn1ObjectIdentifier([]uint32{2, 16, 756, 5, 8, 1, 1})
	aarq.senderAcseRequirements = &tAsn1BitString{
		buf:        []byte{0x80},
		bitsUnused: 7,
	}
	mechanismName := (tAsn1ObjectIdentifier)([]uint32{2, 16, 756, 5, 8, 2, 1})
	aarq.mechanismName = &mechanismName
	password := tAsn1GraphicString([]byte("12345678"))
	aarq.callingAuthenticationValue = new(tAsn1Choice)
	aarq.callingAuthenticationValue.setVal(0, password)
	userInformation := tAsn1OctetString([]byte{0x01, 0x00, 0x00, 0x00, 0x06, 0x5F, 0x1F, 0x04, 0x00, 0xFF, 0xFF, 0xFF, 0x02, 0x00})
	aarq.userInformation = &userInformation

	var buf bytes.Buffer
	err := encode_AARQapdu(&buf, &aarq)
	if nil != err {
		t.Fatalf("encode_AARQapdu() failed")
	}
	b := buf.Bytes()
	printBuffer(t, b)

	expectB := []byte{0x60, 0x36, 0xA1, 0x09, 0x06, 0x07, 0x60, 0x85, 0x74, 0x05, 0x08, 0x01, 0x01, 0x8A, 0x02, 0x07, 0x80, 0x8B, 0x07, 0x60, 0x85, 0x74, 0x05, 0x08, 0x02, 0x01, 0xAC, 0x0A, 0x80, 0x08, 0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37, 0x38, 0xBE, 0x10, 0x04, 0x0E, 0x01, 0x00, 0x00, 0x00, 0x06, 0x5F, 0x1F, 0x04, 0x00, 0xFF, 0xFF, 0xFF, 0x02, 0x00}

	if !byteEquals(t, b, expectB, true) {
		t.Fatalf("bytes don't match")
	}
}

func TestAsn1_decode_AARQapdu_for_hdlc_meter(t *testing.T) {
	t.Logf("TestAsn1_decode_AARQapdu_for_hdlc_meter")
	/*

		60 36 A1 09 06 07 60 85 74 05 08 01 01 8A 02 07 80 8B 07 60 85 74 05 08 02 01 AC 0A 80 08 31 32 33 34 35 36 37 38 BE 10 04 0E 01 00 00 00 06 5F 1F 04 00 FF FF FF 02 00

		XDLMS-APDU CHOICE
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
		        0x01000000065f1f0400ffffff0200
		Successfully decoded 56 bytes.
		rec1value XDLMS-APDU ::= aarq :
		  {
		    application-context-name { 2 16 756 5 8 1 1 },
		    sender-acse-requirements { authentication },
		    mechanism-name { 2 16 756 5 8 2 1 },
		    calling-authentication-value charstring : "12345678",
		    user-information '01000000065F1F0400FFFFFF0200'H
		  }

	*/
	b := []byte{0x60, 0x36, 0xA1, 0x09, 0x06, 0x07, 0x60, 0x85, 0x74, 0x05, 0x08, 0x01, 0x01, 0x8A, 0x02, 0x07, 0x80, 0x8B, 0x07, 0x60, 0x85, 0x74, 0x05, 0x08, 0x02, 0x01, 0xAC, 0x0A, 0x80, 0x08, 0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37, 0x38, 0xBE, 0x10, 0x04, 0x0E, 0x01, 0x00, 0x00, 0x00, 0x06, 0x5F, 0x1F, 0x04, 0x00, 0xFF, 0xFF, 0xFF, 0x02, 0x00}
	buf := bytes.NewReader(b)
	err, aarq := decode_AARQapdu(buf)
	if nil != err {
		t.Fatalf("decode_AARQapdu() failed")
	}

	if !uint32Equals(t, aarq.applicationContextName, []uint32{2, 16, 756, 5, 8, 1, 1}, true) {
		t.Fatalf("aarq.applicationContextName don't match")
	}
	if !byteEquals(t, ([]byte)(aarq.senderAcseRequirements.buf), []byte{0x80}, true) {
		t.Fatalf("aarq.senderAcseRequirements don't match")
	}
	if aarq.senderAcseRequirements.bitsUnused != 7 {
		t.Fatalf("aarq.senderAcseRequirements don't match")
	}
	if !uint32Equals(t, *(*[]uint32)(aarq.mechanismName), []uint32{2, 16, 756, 5, 8, 2, 1}, true) {
		t.Fatalf("aarq.mechanismName don't match")
	}
	if aarq.callingAuthenticationValue.tag != 0 {
		t.Fatalf("aarq.callingAuthenticationValue don't match")
	}
	if !byteEquals(t, ([]byte)(aarq.callingAuthenticationValue.val.(tAsn1GraphicString)), []byte("12345678"), true) {
		t.Fatalf("aarq.callingAuthenticationValue don't match")
	}
	if !byteEquals(t, *(*[]byte)(aarq.userInformation), []byte{0x01, 0x00, 0x00, 0x00, 0x06, 0x5F, 0x1F, 0x04, 0x00, 0xFF, 0xFF, 0xFF, 0x02, 0x00}, true) {
		t.Fatalf("aarq.userInformation don't match")
	}
}

func TestAsn1_encode_AAREapdu_for_hdlc_meter(t *testing.T) {
	t.Logf("TestAsn1_encode_AAREapdu_for_hdlc_meter()")
	/*


			0x61, 0x29, 0xA1, 0x09, 0x06, 0x07, 0x60, 0x85, 0x74, 0x05, 0x08, 0x01, 0x01, 0xA2, 0x03, 0x02, 0x01, 0x00, 0xA3, 0x05, 0xA1, 0x03, 0x02, 0x01, 0x00, 0xBE, 0x10, 0x04, 0x0E, 0x08, 0x00, 0x06, 0x5F, 0x1F, 0x04, 0x00, 0x00, 0xFE, 0x1D, 0x00, 0xEF, 0x00, 0x07
			61 29 A1 09 06 07 60 85 74 05 08 01 01 A2 03 02 01 00 A3 05 A1 03 02 01 00 BE 10 04 0E 08 00 06 5F 1F 04 00 00 FE 1D 00 EF 00 07


		-APDU CHOICE
		  aare AARE-apdu SEQUENCE: tag = [APPLICATION 1] constructed; length = 41
			application-context-name : tag = [1] constructed; length = 9
			  Application-context-name OBJECT IDENTIFIER: tag = [UNIVERSAL 6] primitive; length = 7
				{ 2 16 756 5 8 1 1 }
			result : tag = [2] constructed; length = 3
			  Association-result INTEGER: tag = [UNIVERSAL 2] primitive; length = 1
				0
			result-source-diagnostic : tag = [3] constructed; length = 5
			  Associate-source-diagnostic CHOICE
				acse-service-user : tag = [1] constructed; length = 3
				  INTEGER: tag = [UNIVERSAL 2] primitive; length = 1
					0
			user-information : tag = [30] constructed; length = 16
			  Association-information OCTET STRING: tag = [UNIVERSAL 4] primitive; length = 14
				0x0800065f1f040000fe1d00ef0007
		Successfully decoded 43 bytes.
		rec1value XDLMS-APDU ::= aare :
		  {
			application-context-name { 2 16 756 5 8 1 1 },
			result accepted,
			result-source-diagnostic acse-service-user : null,
			user-information '0800065F1F040000FE1D00EF0007'H
		  }



	*/
	var aare AAREapdu
	aareBytes := []byte{0x61, 0x29, 0xA1, 0x09, 0x06, 0x07, 0x60, 0x85, 0x74, 0x05, 0x08, 0x01, 0x01, 0xA2, 0x03, 0x02, 0x01, 0x00, 0xA3, 0x05, 0xA1, 0x03, 0x02, 0x01, 0x00, 0xBE, 0x10, 0x04, 0x0E, 0x08, 0x00, 0x06, 0x5F, 0x1F, 0x04, 0x00, 0x00, 0xFE, 0x1D, 0x00, 0xEF, 0x00, 0x07}

	aare.applicationContextName = tAsn1ObjectIdentifier([]uint32{2, 16, 756, 5, 8, 1, 1})
	aare.result = tAsn1Integer(0)
	aare.resultSourceDiagnostic.tag = 1
	aare.resultSourceDiagnostic.val = tAsn1Integer(0)
	userInformation := tAsn1OctetString([]byte{0x08, 0x00, 0x06, 0x5F, 0x1F, 0x04, 0x00, 0x00, 0xFE, 0x1D, 0x00, 0xEF, 0x00, 0x07})
	aare.userInformation = &userInformation

	var buf bytes.Buffer
	err := encode_AAREapdu(&buf, &aare)
	if nil != err {
		t.Fatalf("encode_AAREapdu_1() failed")
	}

	b := buf.Bytes()
	if !byteEquals(t, b, aareBytes, true) {
		t.Logf("%X\n", b)
		t.Fatalf("encode_AAREapdu_1() failed")
	}

}

func TestAsn1_decode_AAREapdu_for_hdlc_meter(t *testing.T) {
	t.Logf("TestAsn1_decode_AAREapdu_for_hdlc_meter()")
	/*


			0x61, 0x29, 0xA1, 0x09, 0x06, 0x07, 0x60, 0x85, 0x74, 0x05, 0x08, 0x01, 0x01, 0xA2, 0x03, 0x02, 0x01, 0x00, 0xA3, 0x05, 0xA1, 0x03, 0x02, 0x01, 0x00, 0xBE, 0x10, 0x04, 0x0E, 0x08, 0x00, 0x06, 0x5F, 0x1F, 0x04, 0x00, 0x00, 0xFE, 0x1D, 0x00, 0xEF, 0x00, 0x07
			61 29 A1 09 06 07 60 85 74 05 08 01 01 A2 03 02 01 00 A3 05 A1 03 02 01 00 BE 10 04 0E 08 00 06 5F 1F 04 00 00 FE 1D 00 EF 00 07


		-APDU CHOICE
		  aare AARE-apdu SEQUENCE: tag = [APPLICATION 1] constructed; length = 41
			application-context-name : tag = [1] constructed; length = 9
			  Application-context-name OBJECT IDENTIFIER: tag = [UNIVERSAL 6] primitive; length = 7
				{ 2 16 756 5 8 1 1 }
			result : tag = [2] constructed; length = 3
			  Association-result INTEGER: tag = [UNIVERSAL 2] primitive; length = 1
				0
			result-source-diagnostic : tag = [3] constructed; length = 5
			  Associate-source-diagnostic CHOICE
				acse-service-user : tag = [1] constructed; length = 3
				  INTEGER: tag = [UNIVERSAL 2] primitive; length = 1
					0
			user-information : tag = [30] constructed; length = 16
			  Association-information OCTET STRING: tag = [UNIVERSAL 4] primitive; length = 14
				0x0800065f1f040000fe1d00ef0007
		Successfully decoded 43 bytes.
		rec1value XDLMS-APDU ::= aare :
		  {
			application-context-name { 2 16 756 5 8 1 1 },
			result accepted,
			result-source-diagnostic acse-service-user : null,
			user-information '0800065F1F040000FE1D00EF0007'H
		  }



	*/
	aareBytes := []byte{0x61, 0x29, 0xA1, 0x09, 0x06, 0x07, 0x60, 0x85, 0x74, 0x05, 0x08, 0x01, 0x01, 0xA2, 0x03, 0x02, 0x01, 0x00, 0xA3, 0x05, 0xA1, 0x03, 0x02, 0x01, 0x00, 0xBE, 0x10, 0x04, 0x0E, 0x08, 0x00, 0x06, 0x5F, 0x1F, 0x04, 0x00, 0x00, 0xFE, 0x1D, 0x00, 0xEF, 0x00, 0x07}

	buf := bytes.NewReader(aareBytes)
	err, aare := decode_AAREapdu(buf)
	if nil != err {
		t.Fatalf("decode_AAREapdu_1() failed")
	}

	if !uint32Equals(t, aare.applicationContextName, []uint32{2, 16, 756, 5, 8, 1, 1}, true) {
		t.Fatalf("aare.applicationContextName don't match")
	}
	if aare.result != 0 {
		t.Fatalf("aare.result don't match")
	}
	if 1 != aare.resultSourceDiagnostic.tag {
		t.Fatalf("aare.resultSourceDiagnostic.tag don't match")
	}
	if 0 != aare.resultSourceDiagnostic.val.(tAsn1Integer) {
		t.Fatalf("aare.resultSourceDiagnostic.val don't match")
	}
	if !byteEquals(t, *aare.userInformation, []byte{0x08, 0x00, 0x06, 0x5F, 0x1F, 0x04, 0x00, 0x00, 0xFE, 0x1D, 0x00, 0xEF, 0x00, 0x07}, true) {
		t.Fatalf(" are.userInformation don't match")
	}

}

func TestAsn1_encode_AARQapdu_for_hdlc_meter_with_sec(t *testing.T) {
	t.Logf("TestAsn1_encode_AARQapdu_for_hdlc_meter_with_sec")
	/*
		60 1D A1 09 06 07 60 85 74 05 08 01 01 BE 10 04 0E 01 00 00 00 06 5F 1F 04 00 00 FC 1F FF FF

		DLMS-APDU CHOICE
		  aarq AARQ-apdu SEQUENCE: tag = [APPLICATION 0] constructed; length = 29
		    application-context-name : tag = [1] constructed; length = 9
		      Application-context-name OBJECT IDENTIFIER: tag = [UNIVERSAL 6] primitive; length = 7
		        { 2 16 756 5 8 1 1 }
		    user-information : tag = [30] constructed; length = 16
		      Association-information OCTET STRING: tag = [UNIVERSAL 4] primitive; length = 14
		        0x01000000065f1f040000fc1fffff
		Successfully decoded 31 bytes.
		rec1value XDLMS-APDU ::= aarq :
		  {
		    application-context-name { 2 16 756 5 8 1 1 },
		    user-information '01000000065F1F040000FC1FFFFF'H
		  }

	*/
	var aarq AARQapdu

	aarq.applicationContextName = tAsn1ObjectIdentifier([]uint32{2, 16, 756, 5, 8, 1, 1})
	userInformation := tAsn1OctetString([]byte{0x01, 0x00, 0x00, 0x00, 0x06, 0x5F, 0x1F, 0x04, 0x00, 0x00, 0xFC, 0x1F, 0xFF, 0xFF})
	aarq.userInformation = &userInformation

	var buf bytes.Buffer
	err := encode_AARQapdu(&buf, &aarq)
	if nil != err {
		t.Fatalf("encode_AARQapdu() failed")
	}
	b := buf.Bytes()
	printBuffer(t, b)

	expectB := []byte{0x60, 0x1D, 0xA1, 0x09, 0x06, 0x07, 0x60, 0x85, 0x74, 0x05, 0x08, 0x01, 0x01, 0xBE, 0x10, 0x04, 0x0E, 0x01, 0x00, 0x00, 0x00, 0x06, 0x5F, 0x1F, 0x04, 0x00, 0x00, 0xFC, 0x1F, 0xFF, 0xFF}

	if !byteEquals(t, b, expectB, true) {
		t.Fatalf("bytes don't match")
	}
}

func TestAsn1_decode_AARQapdu_for_hdlc_meter_with_sec(t *testing.T) {
	t.Logf("TestAsn1_decode_AARQapdu_for_hdlc_meter_with_sec")
	/*

		60 1D A1 09 06 07 60 85 74 05 08 01 01 BE 10 04 0E 01 00 00 00 06 5F 1F 04 00 00 FC 1F FF FF

		DLMS-APDU CHOICE
		  aarq AARQ-apdu SEQUENCE: tag = [APPLICATION 0] constructed; length = 29
		    application-context-name : tag = [1] constructed; length = 9
		      Application-context-name OBJECT IDENTIFIER: tag = [UNIVERSAL 6] primitive; length = 7
		        { 2 16 756 5 8 1 1 }
		    user-information : tag = [30] constructed; length = 16
		      Association-information OCTET STRING: tag = [UNIVERSAL 4] primitive; length = 14
		        0x01000000065f1f040000fc1fffff
		Successfully decoded 31 bytes.
		rec1value XDLMS-APDU ::= aarq :
		  {
		    application-context-name { 2 16 756 5 8 1 1 },
		    user-information '01000000065F1F040000FC1FFFFF'H
		  }

	*/

	b := []byte{0x60, 0x1D, 0xA1, 0x09, 0x06, 0x07, 0x60, 0x85, 0x74, 0x05, 0x08, 0x01, 0x01, 0xBE, 0x10, 0x04, 0x0E, 0x01, 0x00, 0x00, 0x00, 0x06, 0x5F, 0x1F, 0x04, 0x00, 0x00, 0xFC, 0x1F, 0xFF, 0xFF}
	buf := bytes.NewReader(b)
	err, aarq := decode_AARQapdu(buf)
	if nil != err {
		t.Fatalf("decode_AARQapdu() failed")
	}

	if !uint32Equals(t, aarq.applicationContextName, []uint32{2, 16, 756, 5, 8, 1, 1}, true) {
		t.Fatalf("aarq.applicationContextName don't match")
	}
	if !byteEquals(t, *(*[]byte)(aarq.userInformation), []byte{0x01, 0x00, 0x00, 0x00, 0x06, 0x5F, 0x1F, 0x04, 0x00, 0x00, 0xFC, 0x1F, 0xFF, 0xFF}, true) {
		t.Fatalf("aarq.userInformation don't match")
	}

}

func TestAsn1_encode_AARQapdu_for_hdlc_meter_with_sec_5(t *testing.T) {
	t.Logf("TestAsn1_encode_AARQapdu_for_hdlc_meter_with_sec_5")
	/*
		60 55 A1 09 06 07 60 85 74 05 08 01 03 A6 0A 04 08 4D 45 4C 00 00 00 00 01 8A 02 07 80 8B 07 60 85 74 05 08 02 05 AC 0A 80 08 29 48 42 2B 30 46 30 34 BE 23 04 21 21 1F 30 24 50 7E 1E C4 C0 DB B9 52 C7 0E 7B 3F F0 A2 96 2B B8 86 5A B9 E5 67 A0 C3 81 D6 EB F5 C3

		XDLMS-APDU CHOICE
		  aarq AARQ-apdu SEQUENCE: tag = [APPLICATION 0] constructed; length = 85
			application-context-name : tag = [1] constructed; length = 9
			  Application-context-name OBJECT IDENTIFIER: tag = [UNIVERSAL 6] primitive; length = 7
				{ 2 16 756 5 8 1 3 }
			calling-AP-title : tag = [6] constructed; length = 10
			  AP-title OCTET STRING: tag = [UNIVERSAL 4] primitive; length = 8
				0x4d454c0000000001
			sender-acse-requirements ACSE-requirements BIT STRING: tag = [10] primitive; length = 2
			  0x0780
			mechanism-name Mechanism-name OBJECT IDENTIFIER: tag = [11] primitive; length = 7
			  { 2 16 756 5 8 2 5 }
			calling-authentication-value : tag = [12] constructed; length = 10
			  Authentication-value CHOICE
				charstring GraphicString: tag = [0] primitive; length = 8
				  ")HB+0F04"
			user-information : tag = [30] constructed; length = 35
			  Association-information OCTET STRING: tag = [UNIVERSAL 4] primitive; length = 33
				0x211f3024507e1ec4c0dbb952c70e7b3ff0 ...
		Successfully decoded 87 bytes.
		rec1value XDLMS-APDU ::= aarq :
		  {
			application-context-name { 2 16 756 5 8 1 3 },
			calling-AP-title '4D454C0000000001'H,
			sender-acse-requirements { authentication },
			mechanism-name { 2 16 756 5 8 2 5 },
			calling-authentication-value charstring : ")HB+0F04",
			user-information '211F3024507E1EC4C0DBB952C70E7B3FF0 ...'H
		  }
	*/
	var aarq AARQapdu

	aarq.applicationContextName = tAsn1ObjectIdentifier([]uint32{2, 16, 756, 5, 8, 1, 3})
	callingAPtitle := tAsn1OctetString([]byte{0x4D, 0x45, 0x4C, 0x00, 0x00, 0x00, 0x00, 0x01})
	aarq.callingAPtitle = &callingAPtitle
	aarq.senderAcseRequirements = &tAsn1BitString{
		buf:        []byte{0x80},
		bitsUnused: 7,
	}
	mechanismName := (tAsn1ObjectIdentifier)([]uint32{2, 16, 756, 5, 8, 2, 5})
	aarq.mechanismName = &mechanismName
	password := tAsn1GraphicString([]byte(")HB+0F04"))
	aarq.callingAuthenticationValue = new(tAsn1Choice)
	aarq.callingAuthenticationValue.setVal(0, password)
	userInformation := tAsn1OctetString([]byte{0x21, 0x1F, 0x30, 0x24, 0x50, 0x7E, 0x1E, 0xC4, 0xC0, 0xDB, 0xB9, 0x52, 0xC7, 0x0E, 0x7B, 0x3F, 0xF0, 0xA2, 0x96, 0x2B, 0xB8, 0x86, 0x5A, 0xB9, 0xE5, 0x67, 0xA0, 0xC3, 0x81, 0xD6, 0xEB, 0xF5, 0xC3})
	aarq.userInformation = &userInformation

	var buf bytes.Buffer
	err := encode_AARQapdu(&buf, &aarq)
	if nil != err {
		t.Fatalf("encode_AARQapdu() failed")
	}
	b := buf.Bytes()
	printBuffer(t, b)

	expectB := []byte{0x60, 0x55, 0xA1, 0x09, 0x06, 0x07, 0x60, 0x85, 0x74, 0x05, 0x08, 0x01, 0x03, 0xA6, 0x0A, 0x04, 0x08, 0x4D, 0x45, 0x4C, 0x00, 0x00, 0x00, 0x00, 0x01, 0x8A, 0x02, 0x07, 0x80, 0x8B, 0x07, 0x60, 0x85, 0x74, 0x05, 0x08, 0x02, 0x05, 0xAC, 0x0A, 0x80, 0x08, 0x29, 0x48, 0x42, 0x2B, 0x30, 0x46, 0x30, 0x34, 0xBE, 0x23, 0x04, 0x21, 0x21, 0x1F, 0x30, 0x24, 0x50, 0x7E, 0x1E, 0xC4, 0xC0, 0xDB, 0xB9, 0x52, 0xC7, 0x0E, 0x7B, 0x3F, 0xF0, 0xA2, 0x96, 0x2B, 0xB8, 0x86, 0x5A, 0xB9, 0xE5, 0x67, 0xA0, 0xC3, 0x81, 0xD6, 0xEB, 0xF5, 0xC3}

	if !byteEquals(t, b, expectB, true) {
		t.Fatalf("bytes don't match")
	}
}

func TestAsn1_decode_AARQapdu_for_hdlc_meter_with_sec_5(t *testing.T) {
	t.Logf("TestAsn1_decode_AARQapdu_for_hdlc_meter_with_sec_5")
	/*
		60 55 A1 09 06 07 60 85 74 05 08 01 03 A6 0A 04 08 4D 45 4C 00 00 00 00 01 8A 02 07 80 8B 07 60 85 74 05 08 02 05 AC 0A 80 08 29 48 42 2B 30 46 30 34 BE 23 04 21 21 1F 30 24 50 7E 1E C4 C0 DB B9 52 C7 0E 7B 3F F0 A2 96 2B B8 86 5A B9 E5 67 A0 C3 81 D6 EB F5 C3

		XDLMS-APDU CHOICE
		  aarq AARQ-apdu SEQUENCE: tag = [APPLICATION 0] constructed; length = 85
			application-context-name : tag = [1] constructed; length = 9
			  Application-context-name OBJECT IDENTIFIER: tag = [UNIVERSAL 6] primitive; length = 7
				{ 2 16 756 5 8 1 3 }
			calling-AP-title : tag = [6] constructed; length = 10
			  AP-title OCTET STRING: tag = [UNIVERSAL 4] primitive; length = 8
				0x4d454c0000000001
			sender-acse-requirements ACSE-requirements BIT STRING: tag = [10] primitive; length = 2
			  0x0780
			mechanism-name Mechanism-name OBJECT IDENTIFIER: tag = [11] primitive; length = 7
			  { 2 16 756 5 8 2 5 }
			calling-authentication-value : tag = [12] constructed; length = 10
			  Authentication-value CHOICE
				charstring GraphicString: tag = [0] primitive; length = 8
				  ")HB+0F04"
			user-information : tag = [30] constructed; length = 35
			  Association-information OCTET STRING: tag = [UNIVERSAL 4] primitive; length = 33
				0x211f3024507e1ec4c0dbb952c70e7b3ff0 ...
		Successfully decoded 87 bytes.
		rec1value XDLMS-APDU ::= aarq :
		  {
			application-context-name { 2 16 756 5 8 1 3 },
			calling-AP-title '4D454C0000000001'H,
			sender-acse-requirements { authentication },
			mechanism-name { 2 16 756 5 8 2 5 },
			calling-authentication-value charstring : ")HB+0F04",
			user-information '211F3024507E1EC4C0DBB952C70E7B3FF0 ...'H
		  }
	*/

	b := []byte{0x60, 0x55, 0xA1, 0x09, 0x06, 0x07, 0x60, 0x85, 0x74, 0x05, 0x08, 0x01, 0x03, 0xA6, 0x0A, 0x04, 0x08, 0x4D, 0x45, 0x4C, 0x00, 0x00, 0x00, 0x00, 0x01, 0x8A, 0x02, 0x07, 0x80, 0x8B, 0x07, 0x60, 0x85, 0x74, 0x05, 0x08, 0x02, 0x05, 0xAC, 0x0A, 0x80, 0x08, 0x29, 0x48, 0x42, 0x2B, 0x30, 0x46, 0x30, 0x34, 0xBE, 0x23, 0x04, 0x21, 0x21, 0x1F, 0x30, 0x24, 0x50, 0x7E, 0x1E, 0xC4, 0xC0, 0xDB, 0xB9, 0x52, 0xC7, 0x0E, 0x7B, 0x3F, 0xF0, 0xA2, 0x96, 0x2B, 0xB8, 0x86, 0x5A, 0xB9, 0xE5, 0x67, 0xA0, 0xC3, 0x81, 0xD6, 0xEB, 0xF5, 0xC3}
	buf := bytes.NewReader(b)
	err, aarq := decode_AARQapdu(buf)
	if nil != err {
		t.Fatalf("decode_AARQapdu() failed")
	}

	if !uint32Equals(t, aarq.applicationContextName, []uint32{2, 16, 756, 5, 8, 1, 3}, true) {
		t.Fatalf("aarq.applicationContextName don't match")
	}
	if !byteEquals(t, *(*[]byte)(aarq.callingAPtitle), []byte{0x4D, 0x45, 0x4C, 0x00, 0x00, 0x00, 0x00, 0x01}, true) {
		t.Fatalf("aarq.callingAPtitle don't match")
	}
	if !byteEquals(t, ([]byte)(aarq.senderAcseRequirements.buf), []byte{0x80}, true) {
		t.Fatalf("aarq.senderAcseRequirements don't match")
	}
	if !uint32Equals(t, *(*[]uint32)(aarq.mechanismName), []uint32{2, 16, 756, 5, 8, 2, 5}, true) {
		t.Fatalf("aarq.mechanismName don't match")
	}
	if aarq.callingAuthenticationValue.tag != 0 {
		t.Fatalf("aarq.callingAuthenticationValue don't match")
	}
	if !byteEquals(t, ([]byte)(aarq.callingAuthenticationValue.val.(tAsn1GraphicString)), []byte(")HB+0F04"), true) {
		t.Fatalf("aarq.callingAuthenticationValue don't match")
	}
	if !byteEquals(t, *(*[]byte)(aarq.userInformation), []byte{0x21, 0x1F, 0x30, 0x24, 0x50, 0x7E, 0x1E, 0xC4, 0xC0, 0xDB, 0xB9, 0x52, 0xC7, 0x0E, 0x7B, 0x3F, 0xF0, 0xA2, 0x96, 0x2B, 0xB8, 0x86, 0x5A, 0xB9, 0xE5, 0x67, 0xA0, 0xC3, 0x81, 0xD6, 0xEB, 0xF5, 0xC3}, true) {
		t.Fatalf("aarq.userInformation don't match")
	}

}

func TestAsn1_encode_AAREapdu_for_hdlc_meter_with_sec_5(t *testing.T) {
	t.Logf("TestAsn1_encode_AAREapdu_for_hdlc_meter_with_sec_5")
	/*

		61 61 A1 09 06 07 60 85 74 05 08 01 03 A2 03 02 01 00 A3 05 A1 03 02 01 0E A4 0A 04 08 4D 45 4C 65 70 A0 37 B2 88 02 07 80 89 07 60 85 74 05 08 02 05 AA 0A 80 08 28 47 33 63 6E 50 7E 73 BE 23 04 21 28 1F 30 00 00 00 2F F9 F1 4F 54 98 BD 2A 0B B0 00 7F DB 93 18 B7 79 77 48 5F 54 C4 EE 12 10 1B B1

		XDLMS-APDU CHOICE
		  aare AARE-apdu SEQUENCE: tag = [APPLICATION 1] constructed; length = 97
		    application-context-name : tag = [1] constructed; length = 9
		      Application-context-name OBJECT IDENTIFIER: tag = [UNIVERSAL 6] primitive; length = 7
		        { 2 16 756 5 8 1 3 }
		    result : tag = [2] constructed; length = 3
		      Association-result INTEGER: tag = [UNIVERSAL 2] primitive; length = 1
		        0
		    result-source-diagnostic : tag = [3] constructed; length = 5
		      Associate-source-diagnostic CHOICE
		        acse-service-user : tag = [1] constructed; length = 3
		          INTEGER: tag = [UNIVERSAL 2] primitive; length = 1
		            14
		    responding-AP-title : tag = [4] constructed; length = 10
		      AP-title OCTET STRING: tag = [UNIVERSAL 4] primitive; length = 8
		        0x4d454c6570a037b2
		    responder-acse-requirements ACSE-requirements BIT STRING: tag = [8] primitive; length = 2
		      0x0780
		    mechanism-name Mechanism-name OBJECT IDENTIFIER: tag = [9] primitive; length = 7
		      { 2 16 756 5 8 2 5 }
		    responding-authentication-value : tag = [10] constructed; length = 10
		      Authentication-value CHOICE
		        charstring GraphicString: tag = [0] primitive; length = 8
		          "(G3cnP~s"
		    user-information : tag = [30] constructed; length = 35
		      Association-information OCTET STRING: tag = [UNIVERSAL 4] primitive; length = 33
		        0x281f300000002ff9f14f5498bd2a0bb000 ...
		Successfully decoded 99 bytes.
		rec1value XDLMS-APDU ::= aare :
		  {
		    application-context-name { 2 16 756 5 8 1 3 },
		    result accepted,
		    result-source-diagnostic acse-service-user : authentication-required,
		    responding-AP-title '4D454C6570A037B2'H,
		    responder-acse-requirements { authentication },
		    mechanism-name { 2 16 756 5 8 2 5 },
		    responding-authentication-value charstring : "(G3cnP~s",
		    user-information '281F300000002FF9F14F5498BD2A0BB000 ...'H
		  }


	*/
	var aare AAREapdu
	aareBytes := []byte{0x61, 0x61, 0xA1, 0x09, 0x06, 0x07, 0x60, 0x85, 0x74, 0x05, 0x08, 0x01, 0x03, 0xA2, 0x03, 0x02, 0x01, 0x00, 0xA3, 0x05, 0xA1, 0x03, 0x02, 0x01, 0x0E, 0xA4, 0x0A, 0x04, 0x08, 0x4D, 0x45, 0x4C, 0x65, 0x70, 0xA0, 0x37, 0xB2, 0x88, 0x02, 0x07, 0x80, 0x89, 0x07, 0x60, 0x85, 0x74, 0x05, 0x08, 0x02, 0x05, 0xAA, 0x0A, 0x80, 0x08, 0x28, 0x47, 0x33, 0x63, 0x6E, 0x50, 0x7E, 0x73, 0xBE, 0x23, 0x04, 0x21, 0x28, 0x1F, 0x30, 0x00, 0x00, 0x00, 0x2F, 0xF9, 0xF1, 0x4F, 0x54, 0x98, 0xBD, 0x2A, 0x0B, 0xB0, 0x00, 0x7F, 0xDB, 0x93, 0x18, 0xB7, 0x79, 0x77, 0x48, 0x5F, 0x54, 0xC4, 0xEE, 0x12, 0x10, 0x1B, 0xB1}

	aare.applicationContextName = tAsn1ObjectIdentifier([]uint32{2, 16, 756, 5, 8, 1, 3})
	aare.result = tAsn1Integer(0)
	aare.resultSourceDiagnostic.tag = 1
	aare.resultSourceDiagnostic.val = tAsn1Integer(14)
	respondingAPtitle := tAsn1OctetString([]byte{0x4D, 0x45, 0x4C, 0x65, 0x70, 0xA0, 0x37, 0xB2})
	aare.respondingAPtitle = &respondingAPtitle
	aare.responderAcseRequirements = &tAsn1BitString{
		buf:        []byte{0x80},
		bitsUnused: 7,
	}
	mechanismName := (tAsn1ObjectIdentifier)([]uint32{2, 16, 756, 5, 8, 2, 5})
	aare.mechanismName = &mechanismName
	password := tAsn1GraphicString([]byte("(G3cnP~s"))
	aare.respondingAuthenticationValue = new(tAsn1Choice)
	aare.respondingAuthenticationValue.setVal(0, password)
	userInformation := tAsn1OctetString([]byte{0x28, 0x1F, 0x30, 0x00, 0x00, 0x00, 0x2F, 0xF9, 0xF1, 0x4F, 0x54, 0x98, 0xBD, 0x2A, 0x0B, 0xB0, 0x00, 0x7F, 0xDB, 0x93, 0x18, 0xB7, 0x79, 0x77, 0x48, 0x5F, 0x54, 0xC4, 0xEE, 0x12, 0x10, 0x1B, 0xB1})
	aare.userInformation = &userInformation

	var buf bytes.Buffer
	err := encode_AAREapdu(&buf, &aare)
	if nil != err {
		t.Fatalf("encode_AAREapdu_1() failed")
	}

	b := buf.Bytes()
	if !byteEquals(t, b, aareBytes, true) {
		t.Logf("%X\n", b)
		t.Fatalf("encode_AAREapdu_1() failed")
	}

}

func TestAsn1_decode_AAREapdu_for_hdlc_meter_with_sec_5(t *testing.T) {
	t.Logf("TestAsn1_decode_AAREapdu_for_hdlc_meter_with_sec_5")
	/*

		61 61 A1 09 06 07 60 85 74 05 08 01 03 A2 03 02 01 00 A3 05 A1 03 02 01 0E A4 0A 04 08 4D 45 4C 65 70 A0 37 B2 88 02 07 80 89 07 60 85 74 05 08 02 05 AA 0A 80 08 28 47 33 63 6E 50 7E 73 BE 23 04 21 28 1F 30 00 00 00 2F F9 F1 4F 54 98 BD 2A 0B B0 00 7F DB 93 18 B7 79 77 48 5F 54 C4 EE 12 10 1B B1

	*/
	aareBytes := []byte{0x61, 0x61, 0xA1, 0x09, 0x06, 0x07, 0x60, 0x85, 0x74, 0x05, 0x08, 0x01, 0x03, 0xA2, 0x03, 0x02, 0x01, 0x00, 0xA3, 0x05, 0xA1, 0x03, 0x02, 0x01, 0x0E, 0xA4, 0x0A, 0x04, 0x08, 0x4D, 0x45, 0x4C, 0x65, 0x70, 0xA0, 0x37, 0xB2, 0x88, 0x02, 0x07, 0x80, 0x89, 0x07, 0x60, 0x85, 0x74, 0x05, 0x08, 0x02, 0x05, 0xAA, 0x0A, 0x80, 0x08, 0x28, 0x47, 0x33, 0x63, 0x6E, 0x50, 0x7E, 0x73, 0xBE, 0x23, 0x04, 0x21, 0x28, 0x1F, 0x30, 0x00, 0x00, 0x00, 0x2F, 0xF9, 0xF1, 0x4F, 0x54, 0x98, 0xBD, 0x2A, 0x0B, 0xB0, 0x00, 0x7F, 0xDB, 0x93, 0x18, 0xB7, 0x79, 0x77, 0x48, 0x5F, 0x54, 0xC4, 0xEE, 0x12, 0x10, 0x1B, 0xB1}

	buf := bytes.NewReader(aareBytes)
	err, aare := decode_AAREapdu(buf)
	if nil != err {
		t.Fatalf("decode_AAREapdu_1() failed")
	}

	if !uint32Equals(t, aare.applicationContextName, []uint32{2, 16, 756, 5, 8, 1, 3}, true) {
		t.Fatalf("aare.applicationContextName don't match")
	}
	if aare.result != 0 {
		t.Fatalf("aare.result don't match")
	}
	if 1 != aare.resultSourceDiagnostic.tag {
		t.Fatalf("aare.resultSourceDiagnostic.tag don't match")
	}
	if 14 != aare.resultSourceDiagnostic.val.(tAsn1Integer) {
		t.Fatalf("aare.resultSourceDiagnostic.val don't match")
	}
	if !byteEquals(t, *(*[]byte)(aare.respondingAPtitle), []byte{0x4D, 0x45, 0x4C, 0x65, 0x70, 0xA0, 0x37, 0xB2}, true) {
		t.Fatalf("aarq.respondingAPtitle don't match")
	}
	if !byteEquals(t, ([]byte)(aare.responderAcseRequirements.buf), []byte{0x80}, true) {
		t.Fatalf("aarq.responderAcseRequirements don't match")
	}
	if !uint32Equals(t, *(*[]uint32)(aare.mechanismName), []uint32{2, 16, 756, 5, 8, 2, 5}, true) {
		t.Fatalf("aarq.mechanismName don't match")
	}
	if aare.respondingAuthenticationValue.tag != 0 {
		t.Fatalf("aarq.respondingAuthenticationValue don't match")
	}
	if !byteEquals(t, ([]byte)(aare.respondingAuthenticationValue.val.(tAsn1GraphicString)), []byte("(G3cnP~s"), true) {
		t.Fatalf("aarq.respondingAuthenticationValue don't match")
	}
	if !byteEquals(t, *aare.userInformation, []byte{0x28, 0x1F, 0x30, 0x00, 0x00, 0x00, 0x2F, 0xF9, 0xF1, 0x4F, 0x54, 0x98, 0xBD, 0x2A, 0x0B, 0xB0, 0x00, 0x7F, 0xDB, 0x93, 0x18, 0xB7, 0x79, 0x77, 0x48, 0x5F, 0x54, 0xC4, 0xEE, 0x12, 0x10, 0x1B, 0xB1}, true) {
		t.Fatalf(" are.userInformation don't match")
	}

}
