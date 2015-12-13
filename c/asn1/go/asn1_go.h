#ifndef __asn1_go_h 
#define __asn1_go_h

#include "AARQ-apdu.h"

int consumeBytes (void *_buf, int _bufLen, void *ctx);
int consumeBytesWrap (void *_buf, int _bufLen, void *ctx);

BIT_STRING_t *hlp__calloc_BIT_STRING_t();
AARQ_apdu_t *hlp__calloc_AARQ_apdu_t();

BIT_STRING_t *hlp__fill_BIT_STRING_t(BIT_STRING_t* bit_string, uint8_t * buf, int bufLen, int unusedBits);
OBJECT_IDENTIFIER_t *hlp__fill_OBJECT_IDENTIFIER_t(OBJECT_IDENTIFIER_t *object_identifier, uint8_t *ids, int idsLen);

T_protocol_version_t *hlp__fill_T_protocol_version_t(T_protocol_version_t *T_protocol_version, uint8_t *buf, int bufLen, int unusedBits);

#endif /* __asn1_go_h */
