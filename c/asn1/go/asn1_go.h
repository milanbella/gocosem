#ifndef __asn1_go_h 
#define __asn1_go_h

#include "AARQ-apdu.h"

typedef struct Authentication_value_other {
	Mechanism_name_t	 other_mechanism_name;
	ANY_t	 other_mechanism_value;
	
	/* Context for parsing across buffer boundaries */
	asn_struct_ctx_t _asn_ctx;
} Authentication_value_other_t;

int consumeBytes (void *_buf, int _bufLen, void *ctx);
int consumeBytesWrap (void *_buf, int _bufLen, void *ctx);

OBJECT_IDENTIFIER_t *hlp__calloc_OBJECT_IDENTIFIER_t();
struct Authentication_value *hlp__calloc_struct_Authentication_value();
AARQ_apdu_t *hlp__calloc_AARQ_apdu_t();

AARQ_apdu_t *hlp__free_AARQ_apdu_t(AARQ_apdu_t *aarq);

BIT_STRING_t *hlp__fill_BIT_STRING_t(BIT_STRING_t* bit_string, uint8_t * buf, int bufLen, int unusedBits);
OBJECT_IDENTIFIER_t *hlp__fill_OBJECT_IDENTIFIER_t(OBJECT_IDENTIFIER_t *object_identifier, uint8_t *ids, int idsLen);
OCTET_STRING_t *hlp__fill_OCTET_STRING_t(OCTET_STRING_t *octet_string, uint8_t *buf, int bufLen);
ANY_t *hlp__fill_ANY_t(ANY_t *any, uint8_t *buf, int bufLen);


#endif /* __asn1_go_h */
