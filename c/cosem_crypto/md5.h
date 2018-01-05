#ifndef MD5_H_
#define MD5_H_


#include "tomcrypt_hash.h"


#ifdef __cplusplus
extern "C" {
#endif

int md5_init(hash_state * md);
int md5_process(hash_state * md, const unsigned char *in, unsigned long inlen);
int md5_done(hash_state * md, unsigned char *hash);
int md5_compress(hash_state *md, unsigned char *buf);


#ifdef __cplusplus
   }
#endif

#endif 
