#include <stdlib.h>
#include <stdio.h>
#include <errno.h>

#include "asn1_go.h"



int consumeBytes (void *_buf, int _bufLen, void *ctx);

int consumeBytesWrap (void *_buf, int _bufLen, void *ctx) {
	return consumeBytes(_buf, _bufLen, ctx);
}

// calloc helpers

long *hlp__calloc_long(int n) {
	return (long *)calloc((size_t)n, sizeof(long));
}

NULL_t *hlp__calloc_NULL_t() {
	return (NULL_t *)calloc(1, sizeof(NULL_t));
}

BOOLEAN_t *hlp__calloc_BOOLEAN_t() {
	return (BOOLEAN_t *)calloc(1, sizeof(BOOLEAN_t));
}

OBJECT_IDENTIFIER_t *hlp__calloc_OBJECT_IDENTIFIER_t() {
	return (OBJECT_IDENTIFIER_t *)calloc(1, sizeof(OBJECT_IDENTIFIER_t));
}

struct Authentication_value *hlp__calloc_struct_Authentication_value() {
	return (struct Authentication_value *)calloc(1, sizeof(struct Authentication_value));
}

Data_t *hlp__calloc_Data_t() {
	return (Data_t *)calloc(1, sizeof(Data_t));
}

AARQ_apdu_t *hlp__calloc_AARQ_apdu_t() {
	return (AARQ_apdu_t *)calloc(1, sizeof(AARQ_apdu_t));
}


// memory free helpers

void hlp__free_AARQ_apdu_t(AARQ_apdu_t *aarq) {
	if (0 != aarq) {
		asn_DEF_AARQ_apdu.free_struct(&asn_DEF_AARQ_apdu, aarq, 0);
	}
}

void hlp__free_AARE_apdu_t(AARE_apdu_t *aare) {
	if (0 != aare) {
		asn_DEF_AARE_apdu.free_struct(&asn_DEF_AARE_apdu, aare, 0);
	}
}


// fill in helpers

BIT_STRING_t *hlp__fill_BIT_STRING_t(BIT_STRING_t* bit_string, uint8_t *buf, int bufLen, int unusedBits) {

	if (0 == bit_string) {
		bit_string = (BIT_STRING_t *)calloc(1, sizeof(BIT_STRING_t));
	}

	bit_string->buf = buf;
	bit_string->size = bufLen;
	bit_string->bits_unused = unusedBits;

	return bit_string;
}

OBJECT_IDENTIFIER_t *hlp__fill_OBJECT_IDENTIFIER_t(OBJECT_IDENTIFIER_t *object_identifier, uint8_t *ids, int idsLen) {

	if (0 == object_identifier) {
		object_identifier =  (OBJECT_IDENTIFIER_t *)calloc(1, sizeof(OBJECT_IDENTIFIER_t));
	}

	object_identifier->buf = ids;
	object_identifier->size = idsLen;
	return object_identifier;
}

OCTET_STRING_t *hlp__fill_OCTET_STRING_t(OCTET_STRING_t *octet_string, uint8_t *buf, int bufLen) {
	if (0 == octet_string) {
		octet_string = (OCTET_STRING_t *)calloc(1, sizeof(OCTET_STRING_t));
	}
	octet_string->buf = buf;
	octet_string->size = bufLen;
	return octet_string; 
}

ANY_t *hlp__fill_ANY_t(ANY_t *any, uint8_t *buf, int bufLen) {
	if (0 == any) {
		any = (ANY_t *)calloc(1, sizeof(ANY_t));
	}
	any->buf = buf;
	any->size = bufLen;
	return any; 
}
