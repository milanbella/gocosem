enum color {ORANGE, PUCE, TURQUOISE};

struct list {
	string data<>;
	int key;
	color col;
	list *next;
};

/*
InitiateRequest ::= SEQUENCE
{
	-- shall not be encoded in DLMS without ciphering
	dedicated-key OCTET STRING OPTIONAL,
	response-allowed BOOLEAN DEFAULT TRUE,
	proposed-quality-of-service IMPLICIT Integer8 OPTIONAL,
	proposed-dlms-version-number Unsigned8,
	proposed-conformance Conformance,
	client-max-receive-pdu-size Unsigned16
}
*/

struct InitiateRequest {
	octet    	dedicatedKey<>;
	bool    	responseAllowed;
	char    	proposedQualityOfService<1>;
	octet    	proposedDlmsVersionNumber;
	octet    	proposedConformance[4];
	uint16_t   	clientMaxReceivePduSize;
};

/*

InitiateResponse ::= SEQUENCE
{
	negotiated-quality-of-service IMPLICIT Integer8 OPTIONAL,
	negotiated-dlms-version-number Unsigned8,
	negotiated-conformance Conformance,
	server-max-receive-pdu-size Unsigned16,
	vaa-name ObjectName
}

*/

struct InitiateResponse {
	octet 			negotiatedQualityOfService<1>;
	unsigned char 	negotiatedDlmsVersionNumber;
	octet 			negotiatedConformance[4];
	uint16_t 		serverMaxReceivePduSize;
	uint16_t 		vaaName;
};

program PRINTER {
	version PRINTER_V1 {
		int PRINT_LIST(list) = 1;
		int SUM_LIST(list) = 2;
	} = 1;
} = 0x2fffffff;
