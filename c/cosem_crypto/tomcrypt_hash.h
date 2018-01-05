#ifndef TOMCRYPT_HASH_H_
#define TOMCRYPT_HASH_H_

#ifdef __cplusplus
extern "C" {
#endif

#define BLOCK_SIZE 64
#include "tomcrypt_macros.h"



 /* ---- HASH FUNCTIONS ---- */
struct sha1_state {
    ulong64 length;
    unsigned long state[5];
	unsigned long curlen;
    unsigned char buf[64];
};

struct md5_state {
    ulong64 length;
    unsigned long state[4];
	unsigned long curlen;
    unsigned char buf[64];
};

 /* ---- HASH FUNCTIONS ---- */
typedef union Hash_state {
    char dummy[1];
    struct sha1_state   sha1;
	struct md5_state    md5;
    void *data;
} hash_state;


/** hash descriptor */
extern  struct ltc_hash_descriptor {
    /** name of hash */
    char *name;
    /** internal ID */
    unsigned char ID;
    /** Size of digest in octets */
    unsigned long hashsize;
    /** Input block size in octets */
    unsigned long blocksize;

} hash_descriptor[];

/* $Source: /cvs/libtom/libtomcrypt/src/headers/tomcrypt_hash.h,v $ */
/* $Revision: 1.22 $ */
/* $Date: 2007/05/12 14:32:35 $ */
#ifdef __cplusplus
   }
#endif

#endif
