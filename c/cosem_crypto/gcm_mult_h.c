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

/**
   @file gcm_mult_h.c
   GCM implementation, do the GF mult, by Tom St Denis
*/
#include "tomcrypt.h"

#if defined(LTC_GCM_MODE)
/**
  GCM multiply by H
  @param gcm   The GCM state which holds the H value
  @param I     The value to multiply H by
 */
void gcm_mult_h(gcm_state *gcm, unsigned char *I)
{
   unsigned char T[16];
#ifdef LTC_GCM_TABLES
   int x, y;
   XMEMCPY(T, &gcm->PC[0][I[0]][0], 16);
   for (x = 1; x < 16; x++) 
   {
       for (y = 0; y < 16; y++) 
	   {
           T[y] ^= gcm->PC[x][I[x]][y];
       }
   }
#else     
   gcm_gf_mult(gcm->H, I, T); 
#endif
   XMEMCPY(I, T, 16);
}
#endif

/* $Source: /cvs/libtom/libtomcrypt/src/encauth/gcm/gcm_mult_h.c,v $ */
/* $Revision: 1.6 $ */
/* $Date: 2007/05/12 14:32:35 $ */
