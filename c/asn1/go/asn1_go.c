#include <stdlib.h>
#include <stdio.h>
#include <errno.h>

#include "asn1_go.h"



int consumeBytes (void *_buf, int _bufLen, void *ctx);

int consumeBytesWrap (void *_buf, int _bufLen, void *ctx) {
	return consumeBytes(_buf, _bufLen, ctx);
}

// calloc helpers

AARQ_apdu_t *hlp__calloc_AARQ_apdu_t() {
	return (AARQ_apdu_t *)calloc(1, sizeof(AARQ_apdu_t));
}

T_protocol_version_t *hlp__calloc_T_protocol_version_t() {
	return (T_protocol_version_t *)calloc(1, sizeof(T_protocol_version_t));
}

// fill in helpers

BIT_STRING_t *hlp__fill_BIT_STRING_t(BIT_STRING_t* bit_string, uint8_t *buf, int bufLen, int unusedBits) {

	if (0 == bit_string) {
		bit_string = (BIT_STRING_t *)calloc(1, sizeof(BIT_STRING_t));
	}

	if (0 != bufLen % 8) {
		fprintf(stderr, "%s:%d: bufLen must be multiple of 8", __FILE__, __LINE__);
		exit(1);
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

T_protocol_version_t *hlp__fill_T_protocol_version_t(T_protocol_version_t *T_protocol_version, uint8_t *buf, int bufLen, int unusedBits) {
	return (T_protocol_version_t *)hlp__fill_BIT_STRING_t(T_protocol_version, buf, bufLen, unusedBits);
}
