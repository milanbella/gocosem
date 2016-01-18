#ifndef __asn1_go_h 
#define __asn1_go_h

#include "AARQ-apdu.h"
#include "AARE-apdu.h"
#include "Data.h"

typedef struct Authentication_value_other {
	Mechanism_name_t	 other_mechanism_name;
	ANY_t	 other_mechanism_value;
	
	/* Context for parsing across buffer boundaries */
	asn_struct_ctx_t _asn_ctx;
} Authentication_value_other_t;

int consumeBytes (void *_buf, int _bufLen, void *ctx);
int consumeBytesWrap (void *_buf, int _bufLen, void *ctx);

void *hlp__calloc(size_t nmemb, size_t size);
long *hlp__calloc_long(int n);
Integer8_t *hlp__calloc_Integer8_t(int n);
Integer16_t *hlp__calloc_Integer16_t(int n);
Integer32_t *hlp__calloc_Integer32_t(int n);
Unsigned8_t *hlp__calloc_Unsigned8_t(int n);
Unsigned16_t *hlp__calloc_Unsigned16_t(int n);
OCTET_STRING_t *hlp__calloc_Float(int n);
Unsigned32_t *hlp__calloc_Unsigned32_t(int n);
OCTET_STRING_t *hlp__calloc_Float(int n);
OCTET_STRING_t *hlp__calloc_Float32(int n);
OCTET_STRING_t *hlp__calloc_Float64(int n);
OBJECT_IDENTIFIER_t *hlp__calloc_OBJECT_IDENTIFIER_t();
NULL_t *hlp__calloc_NULL_t();
BOOLEAN_t *hlp__calloc_BOOLEAN_t();
struct Authentication_value *hlp__calloc_struct_Authentication_value();
Data_t *hlp__calloc_Data_t();
AARQ_apdu_t *hlp__calloc_AARQ_apdu_t();

void hlp__free(void *ptr);
void hlp__free_AARQ_apdu_t(AARQ_apdu_t *aarq);
void hlp__free_AARE_apdu_t(AARE_apdu_t *aare);
void hlp__free_Data_t(Data_t *data);

BIT_STRING_t *hlp__fill_BIT_STRING_t(BIT_STRING_t* bit_string, uint8_t * buf, int bufLen, int unusedBits);
OBJECT_IDENTIFIER_t *hlp__fill_OBJECT_IDENTIFIER_t(OBJECT_IDENTIFIER_t *object_identifier, uint8_t *ids, int idsLen);
OCTET_STRING_t *hlp__fill_OCTET_STRING_t(OCTET_STRING_t *octet_string, uint8_t *buf, int bufLen);
ANY_t *hlp__fill_ANY_t(ANY_t *any, uint8_t *buf, int bufLen);


#endif /* __asn1_go_h */
