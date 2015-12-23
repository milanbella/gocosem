#include <stdlib.h>
#include <stdio.h>
#include <errno.h>

#include "asn1_go.h"



int consumeBytes (void *_buf, int _bufLen, void *ctx);

int consumeBytesWrap (void *_buf, int _bufLen, void *ctx) {
	return consumeBytes(_buf, _bufLen, ctx);
}

// calloc helpers

void *hlp__gc_calloc(size_t nmemb, size_t size) {
	return CALLOC(nmemb, size);
}

long *hlp__calloc_long(int n) {
	return (long *)CALLOC((size_t)n, sizeof(long));
}

Integer8_t *hlp__calloc_Integer8_t(int n) {
	return (Integer8_t *)CALLOC((size_t)n, sizeof(Integer8_t));
}

Integer16_t *hlp__calloc_Integer16_t(int n) {
	return (Integer16_t *)CALLOC((size_t)n, sizeof(Integer16_t));
}

Integer32_t *hlp__calloc_Integer32_t(int n) {
	return (Integer32_t *)CALLOC((size_t)n, sizeof(Integer32_t));
}

Unsigned8_t *hlp__calloc_Unsigned8_t(int n) {
	return (Unsigned8_t *)CALLOC((size_t)n, sizeof(Unsigned8_t));
}

Unsigned16_t *hlp__calloc_Unsigned16_t(int n) {
	return (Unsigned16_t *)CALLOC((size_t)n, sizeof(Unsigned16_t));
}

Unsigned32_t *hlp__calloc_Unsigned32_t(int n) {
	return (Unsigned32_t *)CALLOC((size_t)n, sizeof(Unsigned32_t));
}

OCTET_STRING_t *hlp__calloc_Float(int n) {
	OCTET_STRING_t *o =  (OCTET_STRING_t *)CALLOC(1, sizeof(OCTET_STRING_t));
	o->buf = CALLOC(4*n, sizeof(uint8_t));
	o->size = 4*n;
	return o;
}

OCTET_STRING_t *hlp__calloc_Float32(int n) {
	OCTET_STRING_t *o =  (OCTET_STRING_t *)CALLOC(1, sizeof(OCTET_STRING_t));
	o->buf = CALLOC(4*n, sizeof(uint8_t));
	o->size = 4*n;
	return o;
}

OCTET_STRING_t *hlp__calloc_Float64(int n) {
	OCTET_STRING_t *o =  (OCTET_STRING_t *)CALLOC(1, sizeof(OCTET_STRING_t));
	o->buf = CALLOC(8*n, sizeof(uint8_t));
	o->size = 8*n;
	return o;
}

NULL_t *hlp__calloc_NULL_t() {
	return (NULL_t *)CALLOC(1, sizeof(NULL_t));
}

BOOLEAN_t *hlp__calloc_BOOLEAN_t() {
	return (BOOLEAN_t *)CALLOC(1, sizeof(BOOLEAN_t));
}

OBJECT_IDENTIFIER_t *hlp__calloc_OBJECT_IDENTIFIER_t() {
	return (OBJECT_IDENTIFIER_t *)CALLOC(1, sizeof(OBJECT_IDENTIFIER_t));
}

struct Authentication_value *hlp__calloc_struct_Authentication_value() {
	return (struct Authentication_value *)CALLOC(1, sizeof(struct Authentication_value));
}

Data_t *hlp__calloc_Data_t() {
	return (Data_t *)CALLOC(1, sizeof(Data_t));
}

AARQ_apdu_t *hlp__calloc_AARQ_apdu_t() {
	return (AARQ_apdu_t *)CALLOC(1, sizeof(AARQ_apdu_t));
}


// memory free helpers

void hlp__gc_free(void *ptr) {
	FREEMEM(ptr);
}

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

void hlp__free_Data_t(Data_t *data) {
	if (0 != data) {
		asn_DEF_Data.free_struct(&asn_DEF_Data, data, 0);
	}
}


// fill in helpers

BIT_STRING_t *hlp__fill_BIT_STRING_t(BIT_STRING_t* bit_string, uint8_t *buf, int bufLen, int unusedBits) {

	if (0 == bit_string) {
		bit_string = (BIT_STRING_t *)CALLOC(1, sizeof(BIT_STRING_t));
	}

	bit_string->buf = buf;
	bit_string->size = bufLen;
	bit_string->bits_unused = unusedBits;

	return bit_string;
}

OBJECT_IDENTIFIER_t *hlp__fill_OBJECT_IDENTIFIER_t(OBJECT_IDENTIFIER_t *object_identifier, uint8_t *ids, int idsLen) {

	if (0 == object_identifier) {
		object_identifier =  (OBJECT_IDENTIFIER_t *)CALLOC(1, sizeof(OBJECT_IDENTIFIER_t));
	}

	object_identifier->buf = ids;
	object_identifier->size = idsLen;
	return object_identifier;
}

OCTET_STRING_t *hlp__fill_OCTET_STRING_t(OCTET_STRING_t *octet_string, uint8_t *buf, int bufLen) {
	if (0 == octet_string) {
		octet_string = (OCTET_STRING_t *)CALLOC(1, sizeof(OCTET_STRING_t));
	}
	octet_string->buf = buf;
	octet_string->size = bufLen;
	return octet_string; 
}

ANY_t *hlp__fill_ANY_t(ANY_t *any, uint8_t *buf, int bufLen) {
	if (0 == any) {
		any = (ANY_t *)CALLOC(1, sizeof(ANY_t));
	}
	any->buf = buf;
	any->size = bufLen;
	return any; 
}
