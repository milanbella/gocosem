// +build ignore

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

func uintEquals(t *testing.T, a []uint, b []uint, report bool) bool {
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
	t.Logf("TestAsn1_encode_AARQapdu_for_tcp_meter()")
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

	err, b := encode_AARQapdu(&aarq)
	if nil != err {
		t.Fatalf("encode_AARQapdu() failed")
	}
	printBuffer(t, b)

	expectB := []byte{0x60, 0x36, 0xA1, 0x09, 0x06, 0x07, 0x60, 0x85, 0x74, 0x05, 0x08, 0x01, 0x01, 0x8A, 0x02, 0x07, 0x80, 0x8B, 0x07, 0x60, 0x85, 0x74, 0x05, 0x08, 0x02, 0x01, 0xAC, 0x0A, 0x80, 0x08, 0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37, 0x38, 0xBE, 0x10, 0x04, 0x0E, 0x01, 0x00, 0x00, 0x00, 0x06, 0x5F, 0x1F, 0x04, 0x00, 0x00, 0x7E, 0x1F, 0x04, 0xB0}

	if !byteEquals(t, b, expectB, true) {
		t.Fatalf("bytes don't match")
	}
}

func TestAsn1_decode_AAREapdu_for_tcp_meter(t *testing.T) {
	t.Logf("TestAsn1_decode_AAREapdu_for_tcp_meter()")
	/*

		XDLMS-APDU CHOICE
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
		        0x0800065f1f040000181f08000007
		Successfully decoded 43 bytes.
		rec1value XDLMS-APDU ::= aare :
		  {
		    application-context-name { 2 16 756 5 8 1 1 },
		    result accepted,
		    result-source-diagnostic acse-service-user : null,
		    user-information '0800065F1F040000181F08000007'H
		  }

	*/

	b := []byte{0x61, 0x29, 0xA1, 0x09, 0x06, 0x07, 0x60, 0x85, 0x74, 0x05, 0x08, 0x01, 0x01, 0xA2, 0x03, 0x02, 0x01, 0x00, 0xA3, 0x05, 0xA1, 0x03, 0x02, 0x01, 0x00, 0xBE, 0x10, 0x04, 0x0E, 0x08, 0x00, 0x06, 0x5F, 0x1F, 0x04, 0x00, 0x00, 0x18, 0x1F, 0x08, 0x00, 0x00, 0x07}
	err, aare := decode_AAREapdu(b)
	if nil != err {
		t.Fatalf("decode_AAREapdu() failed")
	}

	t.Logf("%v", aare.applicationContextName)
	t.Logf("%v", aare.result)
	t.Logf("%v", aare.resultSourceDiagnostic.getTag())
	t.Logf("%v", aare.resultSourceDiagnostic.getVal())
	printBuffer(t, *aare.userInformation)

	if !uintEquals(t, aare.applicationContextName, []uint{2, 16, 756, 5, 8, 1, 1}, true) {
		t.Fatalf("aare.applicationContextName don't match")
	}
	if aare.result != 0 {
		t.Fatalf("aare.result don't match")
	}
	if 1 != aare.resultSourceDiagnostic.getTag() {
		t.Fatalf("aare.resultSourceDiagnostic.tag don't match")
	}
	if 0 != aare.resultSourceDiagnostic.getVal().(int) {
		t.Fatalf("aare.resultSourceDiagnostic.val don't match")
	}
	if !byteEquals(t, *aare.userInformation, []byte{0x08, 0x00, 0x06, 0x5F, 0x1F, 0x04, 0x00, 0x00, 0x18, 0x1F, 0x08, 0x00, 0x00, 0x07}, true) {
		t.Fatalf(" are.userInformation don't match")
	}
}

func TestAsn1_encode_AARQapdu_for_hdlc_meter(t *testing.T) {
	t.Logf("TestAsn1_encode_AARQapdu_for_hdlc_meter()")
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
	userInformation := tAsn1OctetString([]byte{0x01, 0x00, 0x00, 0x00, 0x06, 0x5F, 0x1F, 0x04, 0x00, 0xFF, 0xFF, 0xFF, 0x02, 0x00})
	aarq.userInformation = &userInformation

	err, b := encode_AARQapdu(&aarq)
	if nil != err {
		t.Fatalf("encode_AARQapdu() failed")
	}
	printBuffer(t, b)

	expectB := []byte{0x60, 0x36, 0xA1, 0x09, 0x06, 0x07, 0x60, 0x85, 0x74, 0x05, 0x08, 0x01, 0x01, 0x8A, 0x02, 0x07, 0x80, 0x8B, 0x07, 0x60, 0x85, 0x74, 0x05, 0x08, 0x02, 0x01, 0xAC, 0x0A, 0x80, 0x08, 0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37, 0x38, 0xBE, 0x10, 0x04, 0x0E, 0x01, 0x00, 0x00, 0x00, 0x06, 0x5F, 0x1F, 0x04, 0x00, 0xFF, 0xFF, 0xFF, 0x02, 0x00}

	if !byteEquals(t, b, expectB, true) {
		t.Fatalf("bytes don't match")
	}
}

func TestAsn1_decode_AAREapdu_for_hdlc_meter(t *testing.T) {
	t.Logf("TestAsn1_decode_AAREapdu_for_tcp_meter()")
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

	b := []byte{0x61, 0x29, 0xA1, 0x09, 0x06, 0x07, 0x60, 0x85, 0x74, 0x05, 0x08, 0x01, 0x01, 0xA2, 0x03, 0x02, 0x01, 0x00, 0xA3, 0x05, 0xA1, 0x03, 0x02, 0x01, 0x00, 0xBE, 0x10, 0x04, 0x0E, 0x08, 0x00, 0x06, 0x5F, 0x1F, 0x04, 0x00, 0x00, 0xFE, 0x1D, 0x00, 0xEF, 0x00, 0x07}
	err, aare := decode_AAREapdu(b)
	if nil != err {
		t.Fatalf("decode_AAREapdu() failed")
	}

	t.Logf("%v", aare.applicationContextName)
	t.Logf("%v", aare.result)
	t.Logf("%v", aare.resultSourceDiagnostic.getTag())
	t.Logf("%v", aare.resultSourceDiagnostic.getVal())
	printBuffer(t, *aare.userInformation)

	if !uintEquals(t, aare.applicationContextName, []uint{2, 16, 756, 5, 8, 1, 1}, true) {
		t.Fatalf("aare.applicationContextName don't match")
	}
	if aare.result != 0 {
		t.Fatalf("aare.result don't match")
	}
	if 1 != aare.resultSourceDiagnostic.getTag() {
		t.Fatalf("aare.resultSourceDiagnostic.tag don't match")
	}
	if 0 != aare.resultSourceDiagnostic.getVal().(int) {
		t.Fatalf("aare.resultSourceDiagnostic.val don't match")
	}
	if !byteEquals(t, *aare.userInformation, []byte{0x08, 0x00, 0x06, 0x5F, 0x1F, 0x04, 0x00, 0x00, 0xFE, 0x1D, 0x00, 0xEF, 0x00, 0x07}, true) {
		t.Fatalf(" are.userInformation don't match")
	}
}
