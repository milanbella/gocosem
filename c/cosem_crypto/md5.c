/* LibTomCrypt, modular cryptographic library -- Tom St Denis
 *
 * LibTomCrypt is a library that provides various cryptographic
 * algorithms in a highly modular and flexible manner.
 *
 * The library is free for all purposes without any express
 * guarantee it works.
 *
 * Tom St Denis, tomstdenis@gmail.com, http://libtom.org
 */

#include "md5.h"
#include <stdio.h>
#include <string.h>
#include <stdlib.h>
/**
  @file md5.c
  LTC_MD5 hash function by Tom St Denis 
*/



const struct ltc_hash_descriptor md5_desc =
{
	"md5",
    3,
    16,
    64
};

#define F(x,y,z)  (z ^ (x & (y ^ z)))
#define G(x,y,z)  (y ^ (z & (y ^ x)))
#define H(x,y,z)  (x^y^z)
#define I(x,y,z)  (y^(x|(~z)))


#define FF(a,b,c,d,M,s,t) \
    a = (a + F(b,c,d) + M + t); a = ROLc(a, s) + b;

#define GG(a,b,c,d,M,s,t) \
    a = (a + G(b,c,d) + M + t); a = ROLc(a, s) + b;

#define HH(a,b,c,d,M,s,t) \
    a = (a + H(b,c,d) + M + t); a = ROLc(a, s) + b;

#define II(a,b,c,d,M,s,t) \
    a = (a + I(b,c,d) + M + t); a = ROLc(a, s) + b;


/* a simple macro for making hash "process" functions */
/*#define HASH_PROCESS(md5_process, md5_compress, md5, BLOCK_SIZE) */                      


int  md5_compress(hash_state *md, unsigned char *buf)

{
    unsigned long i, W[16], a, b, c, d;
    /* copy the state into 512-bits into W[0..15] */
    for (i = 0; i < 16; i++) 
	{
        LOAD32L(W[i], buf + (4*i));
    }
 
    /* copy state */
    a = md->md5.state[0];
    b = md->md5.state[1];
    c = md->md5.state[2];
    d = md->md5.state[3];


    FF(a,b,c,d,W[0],7,0xd76aa478UL)
    FF(d,a,b,c,W[1],12,0xe8c7b756UL)
    FF(c,d,a,b,W[2],17,0x242070dbUL)
    FF(b,c,d,a,W[3],22,0xc1bdceeeUL)
    FF(a,b,c,d,W[4],7,0xf57c0fafUL)
    FF(d,a,b,c,W[5],12,0x4787c62aUL)
    FF(c,d,a,b,W[6],17,0xa8304613UL)
    FF(b,c,d,a,W[7],22,0xfd469501UL)
    FF(a,b,c,d,W[8],7,0x698098d8UL)
    FF(d,a,b,c,W[9],12,0x8b44f7afUL)
    FF(c,d,a,b,W[10],17,0xffff5bb1UL)
    FF(b,c,d,a,W[11],22,0x895cd7beUL)
    FF(a,b,c,d,W[12],7,0x6b901122UL)
    FF(d,a,b,c,W[13],12,0xfd987193UL)
    FF(c,d,a,b,W[14],17,0xa679438eUL)
    FF(b,c,d,a,W[15],22,0x49b40821UL)
    GG(a,b,c,d,W[1],5,0xf61e2562UL)
    GG(d,a,b,c,W[6],9,0xc040b340UL)
    GG(c,d,a,b,W[11],14,0x265e5a51UL)
    GG(b,c,d,a,W[0],20,0xe9b6c7aaUL)
    GG(a,b,c,d,W[5],5,0xd62f105dUL)
    GG(d,a,b,c,W[10],9,0x02441453UL)
    GG(c,d,a,b,W[15],14,0xd8a1e681UL)
    GG(b,c,d,a,W[4],20,0xe7d3fbc8UL)
    GG(a,b,c,d,W[9],5,0x21e1cde6UL)
    GG(d,a,b,c,W[14],9,0xc33707d6UL)
    GG(c,d,a,b,W[3],14,0xf4d50d87UL)
    GG(b,c,d,a,W[8],20,0x455a14edUL)
    GG(a,b,c,d,W[13],5,0xa9e3e905UL)
    GG(d,a,b,c,W[2],9,0xfcefa3f8UL)
    GG(c,d,a,b,W[7],14,0x676f02d9UL)
    GG(b,c,d,a,W[12],20,0x8d2a4c8aUL)
    HH(a,b,c,d,W[5],4,0xfffa3942UL)
    HH(d,a,b,c,W[8],11,0x8771f681UL)
    HH(c,d,a,b,W[11],16,0x6d9d6122UL)
    HH(b,c,d,a,W[14],23,0xfde5380cUL)
    HH(a,b,c,d,W[1],4,0xa4beea44UL)
    HH(d,a,b,c,W[4],11,0x4bdecfa9UL)
    HH(c,d,a,b,W[7],16,0xf6bb4b60UL)
    HH(b,c,d,a,W[10],23,0xbebfbc70UL)
    HH(a,b,c,d,W[13],4,0x289b7ec6UL)
    HH(d,a,b,c,W[0],11,0xeaa127faUL)
    HH(c,d,a,b,W[3],16,0xd4ef3085UL)
    HH(b,c,d,a,W[6],23,0x04881d05UL)
    HH(a,b,c,d,W[9],4,0xd9d4d039UL)
    HH(d,a,b,c,W[12],11,0xe6db99e5UL)
    HH(c,d,a,b,W[15],16,0x1fa27cf8UL)
    HH(b,c,d,a,W[2],23,0xc4ac5665UL)
    II(a,b,c,d,W[0],6,0xf4292244UL)
    II(d,a,b,c,W[7],10,0x432aff97UL)
    II(c,d,a,b,W[14],15,0xab9423a7UL)
    II(b,c,d,a,W[5],21,0xfc93a039UL)
    II(a,b,c,d,W[12],6,0x655b59c3UL)
    II(d,a,b,c,W[3],10,0x8f0ccc92UL)
    II(c,d,a,b,W[10],15,0xffeff47dUL)
    II(b,c,d,a,W[1],21,0x85845dd1UL)
    II(a,b,c,d,W[8],6,0x6fa87e4fUL)
    II(d,a,b,c,W[15],10,0xfe2ce6e0UL)
    II(c,d,a,b,W[6],15,0xa3014314UL)
    II(b,c,d,a,W[13],21,0x4e0811a1UL)
    II(a,b,c,d,W[4],6,0xf7537e82UL)
    II(d,a,b,c,W[11],10,0xbd3af235UL)
    II(c,d,a,b,W[2],15,0x2ad7d2bbUL)
    II(b,c,d,a,W[9],21,0xeb86d391UL)


    md->md5.state[0] = md->md5.state[0] + a;
    md->md5.state[1] = md->md5.state[1] + b;
    md->md5.state[2] = md->md5.state[2] + c;
    md->md5.state[3] = md->md5.state[3] + d;

    return CRYPT_OK;
}

int md5_process (hash_state * md, const unsigned char *in, unsigned long inlen)               
{                                                                                           
    unsigned long n;                                                                        
    int           err;                                                                      
    if (md-> md5 .curlen > sizeof(md-> md5 .buf)) 
	{                             
       return CRYPT_INVALID_ARG;                                                            
    }                                                                                       
    while (inlen > 0) 
	{                                                                     
        if (md-> md5 .curlen == 0 && inlen >= BLOCK_SIZE) 
		{                           
           if ((err = md5_compress (md, (unsigned char *)in)) != CRYPT_OK) 
		   {               
              return err;                                                                   
           }                                                                                
           md-> md5 .length += BLOCK_SIZE * 8;                                        
           in             += BLOCK_SIZE;                                                    
           inlen          -= BLOCK_SIZE;                                                    
        } 
		else 
		{                                                                            
           n = MIN(inlen, (BLOCK_SIZE - md-> md5 .curlen));                           
           memcpy(md-> md5 .buf + md-> md5.curlen, in, (size_t)n);              
           md-> md5 .curlen += n;                                                     
           in             += n;                                                             
           inlen          -= n;                                                             
           if (md-> md5 .curlen == BLOCK_SIZE) 
		   {                                      
              if ((err = md5_compress (md, md-> md5 .buf)) != CRYPT_OK) 
			  {            
                 return err;                                                                
              }                                                                             
              md-> md5 .length += 8*BLOCK_SIZE;                                       
              md-> md5 .curlen = 0;                                                   
           }                                                                                
       }                                                                                    
    }                                                                                       
    return CRYPT_OK;                                                                        
}
/**
   Initialize the hash state
   @param md   The hash state you wish to initialize
   @return CRYPT_OK if successful
*/
int md5_init(hash_state * md)
{
   md->md5.state[0] = 0x67452301UL;
   md->md5.state[1] = 0xefcdab89UL;
   md->md5.state[2] = 0x98badcfeUL;
   md->md5.state[3] = 0x10325476UL;
   md->md5.curlen = 0;
   md->md5.length = 0;
   return CRYPT_OK;
}

/**
   Terminate the hash to get the digest
   @param md  The hash state
   @param out [out] The destination of the hash (16 bytes)
   @return CRYPT_OK if successful
*/
int md5_done(hash_state * md, unsigned char *out)
{
    int i;
    if (md->md5.curlen >= sizeof(md->md5.buf)) 
	{
       return CRYPT_INVALID_ARG;
    }


    /* increase the length of the message */
    md->md5.length += md->md5.curlen * 8;

    /* append the '1' bit */
    md->md5.buf[md->md5.curlen++] = (unsigned char)0x80;

    /* if the length is currently above 56 bytes we append zeros
     * then compress.  Then we can fall back to padding zeros and length
     * encoding like normal.
     */
    if (md->md5.curlen > 56) 
	{
        while (md->md5.curlen < 64) 
		{
            md->md5.buf[md->md5.curlen++] = (unsigned char)0;
        }
        md5_compress(md, md->md5.buf);
        md->md5.curlen = 0;
    }

    /* pad upto 56 bytes of zeroes */
    while (md->md5.curlen < 56) 
	{
        md->md5.buf[md->md5.curlen++] = (unsigned char)0;
    }

    /* store length */
    STORE64L(md->md5.length, md->md5.buf+56);
    md5_compress(md, md->md5.buf);

    /* copy output */
    for (i = 0; i < 4; i++)
	{
        STORE32L(md->md5.state[i], out+(4*i));
    }

    return CRYPT_OK;
}

/* $Source: /cvs/libtom/libtomcrypt/src/hashes/md5.c,v $ */
/* $Revision: 1.10 $ */
/* $Date: 2007/05/12 14:25:28 $ */

