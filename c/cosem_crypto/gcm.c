#include "tomcrypt.h"

#ifdef LTC_GCM_MODE
struct ltc_cipher_descriptor cipher_descriptor=
{
    "aes",
    6,
    16, 32, 16, 10,
    SETUP, ECB_ENC, ECB_DEC, NULL, ECB_DONE, ECB_KS,
    NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL
};

/**
  Initialize a GCM state
  @param gcm     The GCM state to initialize
  @param cipher  The index of the cipher to use
  @param key     The secret key
  @param keylen  The length of the secret key
  @return CRYPT_OK on success
 */
int gcm_init(gcm_state *gcm,
			 int cipher,
             const unsigned char *key,
			 int keylen)
{
   int           err;
   unsigned char B[16];
#ifdef LTC_GCM_TABLES
   int           x, y, z, t;
#endif



#ifdef LTC_FAST
   if (16 % sizeof(LTC_FAST_TYPE)) {
      return CRYPT_INVALID_ARG;
   }
#endif

   /* is cipher valid? */
   if (cipher_descriptor.name == NULL)
   {
     return CRYPT_INVALID_CIPHER;
   }

   if (cipher_descriptor.block_length != 16)
   {
      return CRYPT_INVALID_CIPHER;
   }

   /* schedule key */
   if ((err = cipher_descriptor.setup(key, keylen, 0, &gcm->K)) != CRYPT_OK)
   {
      return err;
   }

   /* H = E(0) */
   zeromem(B, 16);
   if ((err = cipher_descriptor.ecb_encrypt(B, gcm->H, &gcm->K)) != CRYPT_OK)
   {
      return err;
   }

   /* setup state */
   zeromem(gcm->buf, sizeof(gcm->buf));
   zeromem(gcm->X,   sizeof(gcm->X));
   gcm->cipher   = cipher;
   gcm->mode     = LTC_GCM_MODE_IV;
   gcm->ivmode   = 0;
   gcm->buflen   = 0;
   gcm->totlen   = 0;
   gcm->pttotlen = 0;

#ifdef LTC_GCM_TABLES
   /* setup tables */

   /* generate the first table as it has no shifting (from which we make the other tables) */
   zeromem(B, 16);
   for (y = 0; y < 256; y++)
   {
        B[0] = y;
        gcm_gf_mult(gcm->H, B, &gcm->PC[0][y][0]);
   }

   /* now generate the rest of the tables based the previous table */
   for (x = 1; x < 16; x++)
   {
      for (y = 0; y < 256; y++)
	  {
         /* now shift it right by 8 bits */
         t = gcm->PC[x-1][y][15];
         for (z = 15; z > 0; z--)
		 {
             gcm->PC[x][y][z] = gcm->PC[x-1][y][z-1];
         }
         gcm->PC[x][y][0] = gcm_shift_table[t<<1];
         gcm->PC[x][y][1] ^= gcm_shift_table[(t<<1)+1];
     }
  }

#endif

   return CRYPT_OK;
}

/**
  Add IV data to the GCM state
  @param gcm    The GCM state
  @param IV     The initial value data to add
  @param IVlen  The length of the IV
  @return CRYPT_OK on success
 */
int gcm_add_iv(gcm_state *gcm,
               const unsigned char *IV,
			   unsigned long IVlen)
{
   unsigned long x, y;
   if (IVlen > 0)
   {
   }
   /* must be in IV mode */
   if (gcm->mode != LTC_GCM_MODE_IV)
   {
      return CRYPT_INVALID_ARG;
   }

   if (gcm->buflen >= 16 || gcm->buflen < 0)
   {
      return CRYPT_INVALID_ARG;
   }

   if (cipher_descriptor.name == NULL)
   {
     return CRYPT_INVALID_CIPHER;
   }
   /* trip the ivmode flag */
   if (IVlen + gcm->buflen > 12)
   {
      gcm->ivmode |= 1;
   }

   x = 0;
#ifdef LTC_FAST
   if (gcm->buflen == 0) {
      for (x = 0; x < (IVlen & ~15); x += 16) {
          for (y = 0; y < 16; y += sizeof(LTC_FAST_TYPE)) {
              *((LTC_FAST_TYPE*)(&gcm->X[y])) ^= *((LTC_FAST_TYPE*)(&IV[x + y]));
          }
          gcm_mult_h(gcm, gcm->X);
          gcm->totlen += 128;
      }
      IV += x;
   }
#endif

   /* start adding IV data to the state */
   for (; x < IVlen; x++)
   {
       gcm->buf[gcm->buflen++] = *IV++;

       if (gcm->buflen == 16)
	   {
         /* GF mult it */
         for (y = 0; y < 16; y++)
		 {
             gcm->X[y] ^= gcm->buf[y];
         }
         gcm_mult_h(gcm, gcm->X);
         gcm->buflen = 0;
         gcm->totlen += 128;
      }
   }

   return CRYPT_OK;
}

int gcm_add_aad(gcm_state *gcm,
               const unsigned char *adata,
			   unsigned long adatalen)
{
   unsigned long x;
#ifdef LTC_FAST
   unsigned long y;
#endif

   if (adatalen > 0)
   {

   }

   if (gcm->buflen > 16 || gcm->buflen < 0)
   {
      return CRYPT_INVALID_ARG;
   }

   if (cipher_descriptor.name == NULL)
   {
      return CRYPT_INVALID_CIPHER;
   }

   /* in IV mode? */
   if (gcm->mode == LTC_GCM_MODE_IV)
   {
      /* let's process the IV */
      if (gcm->ivmode || gcm->buflen != 12)
	  {
         for (x = 0; x < (unsigned long)gcm->buflen; x++)
		 {
             gcm->X[x] ^= gcm->buf[x];
         }
         if (gcm->buflen)
		 {
            gcm->totlen += gcm->buflen * CONST64(8);
            gcm_mult_h(gcm, gcm->X);
         }

         /* mix in the length */
         zeromem(gcm->buf, 8);
         STORE64H(gcm->totlen, gcm->buf+8);
         for (x = 0; x < 16; x++)
		 {
             gcm->X[x] ^= gcm->buf[x];
         }
         gcm_mult_h(gcm, gcm->X);

         /* copy counter out */
         XMEMCPY(gcm->Y, gcm->X, 16);
         zeromem(gcm->X, 16);
      }
	  else
	  {
         XMEMCPY(gcm->Y, gcm->buf, 12);
         gcm->Y[12] = 0;
         gcm->Y[13] = 0;
         gcm->Y[14] = 0;
         gcm->Y[15] = 1;
      }
      XMEMCPY(gcm->Y_0, gcm->Y, 16);
      zeromem(gcm->buf, 16);
      gcm->buflen = 0;
      gcm->totlen = 0;
      gcm->mode   = LTC_GCM_MODE_AAD;
   }

   if (gcm->mode != LTC_GCM_MODE_AAD || gcm->buflen >= 16)
   {
      return CRYPT_INVALID_ARG;
   }

   x = 0;
   /* start adding AAD data to the state */
   for (; x < adatalen; x++)
   {
       gcm->X[gcm->buflen++] ^= *adata++;
       if (gcm->buflen == 16)
	   {
         /* GF mult it */
         gcm_mult_h(gcm, gcm->X);
         gcm->buflen = 0;
         gcm->totlen += 128;
      }
   }

   return CRYPT_OK;
}

/**
  Process plaintext/ciphertext through GCM
  @param gcm       The GCM state
  @param pt        The plaintext
  @param ptlen     The plaintext length (ciphertext length is the same)
  @param ct        The ciphertext
  @param direction Encrypt or Decrypt mode (GCM_ENCRYPT or GCM_DECRYPT)
  @return CRYPT_OK on success
 */
int gcm_process(gcm_state *gcm,
                     unsigned char *pt,
					 unsigned long ptlen,
                     unsigned char *ct,
                     int direction)
{
   unsigned long x;
   int           y, err;
   unsigned char b;
   if (ptlen > 0)
   {

   }

   if (gcm->buflen > 16 || gcm->buflen < 0)
   {
      return CRYPT_INVALID_ARG;
   }

   if (cipher_descriptor.name == NULL)
   {
     return CRYPT_INVALID_CIPHER;
   }

   /* in AAD mode? */
   if (gcm->mode == LTC_GCM_MODE_AAD)
   {
      /* let's process the AAD */
      if (gcm->buflen)
	  {
         gcm->totlen += gcm->buflen * CONST64(8);
         gcm_mult_h(gcm, gcm->X);
      }

      /* increment counter */
      for (y = 15; y >= 12; y--)
	  {
          if (++gcm->Y[y] & 255) { break; }
      }
      /* encrypt the counter */
      if ((err = cipher_descriptor.ecb_encrypt(gcm->Y, gcm->buf, &gcm->K)) != CRYPT_OK)
	  {
         return err;
      }

      gcm->buflen = 0;
      gcm->mode   = LTC_GCM_MODE_TEXT;
   }

   if (gcm->mode != LTC_GCM_MODE_TEXT)
   {
      return CRYPT_INVALID_ARG;
   }

   x = 0;
   /* process text */
   for (; x < ptlen; x++)
   {
       if (gcm->buflen == 16)
	   {
          gcm->pttotlen += 128;
          gcm_mult_h(gcm, gcm->X);

          /* increment counter */
          for (y = 15; y >= 12; y--)
		  {
              if (++gcm->Y[y] & 255) { break; }
          }
          if ((err = cipher_descriptor.ecb_encrypt(gcm->Y, gcm->buf, &gcm->K)) != CRYPT_OK)
		  {
             return err;
          }
          gcm->buflen = 0;
       }

       if (direction == GCM_ENCRYPT)
	   {
          b = ct[x] = pt[x] ^ gcm->buf[gcm->buflen];
       }
	   else
	   {
          b = ct[x];
          pt[x] = ct[x] ^ gcm->buf[gcm->buflen];
       }
       gcm->X[gcm->buflen++] ^= b;
   }

   return CRYPT_OK;
}

/**
  Terminate a GCM stream
  @param gcm     The GCM state
  @param tag     [out] The destination for the MAC tag
  @param taglen  [in/out]  The length of the MAC tag
  @return CRYPT_OK on success
 */
int gcm_done(gcm_state *gcm,
             unsigned char *tag,
			 unsigned long *taglen)
{
   unsigned long x;
   int err;
   if (gcm->buflen > 16 || gcm->buflen < 0)
   {
      return CRYPT_INVALID_ARG;
   }

   if (cipher_descriptor.name == NULL)
   {
     return CRYPT_INVALID_CIPHER;
   }
   if (gcm->mode != LTC_GCM_MODE_TEXT)
   {
      return CRYPT_INVALID_ARG;
   }

   /* handle remaining ciphertext */
   if (gcm->buflen)
   {
      gcm->pttotlen += gcm->buflen * CONST64(8);
      gcm_mult_h(gcm, gcm->X);
   }

   /* length */
   STORE64H(gcm->totlen, gcm->buf);
   STORE64H(gcm->pttotlen, gcm->buf+8);
   for (x = 0; x < 16; x++)
   {
       gcm->X[x] ^= gcm->buf[x];
   }
   gcm_mult_h(gcm, gcm->X);
   /* encrypt original counter */
   if ((err = cipher_descriptor.ecb_encrypt(gcm->Y_0, gcm->buf, &gcm->K)) != CRYPT_OK)
   {
      return err;
   }
   for (x = 0; x < 16 && x < *taglen; x++)
   {
       tag[x] = gcm->buf[x] ^ gcm->X[x];
   }
   *taglen = x;
   cipher_descriptor.done(&gcm->K);
   return CRYPT_OK;
}
/**
  Reset a GCM state to as if you just called gcm_init().  This saves the initialization time.
  @param gcm   The GCM state to reset
  @return CRYPT_OK on success
*/
int gcm_reset(gcm_state *gcm)
{
   zeromem(gcm->buf, sizeof(gcm->buf));
   zeromem(gcm->X,   sizeof(gcm->X));
   gcm->mode     = LTC_GCM_MODE_IV;
   gcm->ivmode   = 0;
   gcm->buflen   = 0;
   gcm->totlen   = 0;
   gcm->pttotlen = 0;
   return CRYPT_OK;
}

/**
  Process an entire GCM packet in one call.
  @param cipher            Index of cipher to use
  @param key               The secret key
  @param keylen            The length of the secret key
  @param IV                The initial vector
  @param IVlen             The length of the initial vector
  @param adata             The additional authentication data (header)
  @param adatalen          The length of the adata
  @param pt                The plaintext
  @param ptlen             The length of the plaintext (ciphertext length is the same)
  @param ct                The ciphertext
  @param tag               [out] The MAC tag
  @param taglen            [in/out] The MAC tag length
  @param direction         Encrypt or Decrypt mode (GCM_ENCRYPT or GCM_DECRYPT)
  @return CRYPT_OK on success
 */
int gcm_memory(int           cipher,
               const unsigned char *key,
			   unsigned long keylen,
               const unsigned char *IV,
			   unsigned long IVlen,
               const unsigned char *adata,
			   unsigned long adatalen,
               unsigned char *pt,
			   unsigned long ptlen,
               unsigned char *ct,
               unsigned char *tag,
			   unsigned long *taglen,
               int direction)
{
    void      *orig;
    gcm_state *gcm;
    int        err;

    if (cipher_descriptor.name == NULL)
	{
      return CRYPT_INVALID_CIPHER;
    }

    if (cipher_descriptor.accel_gcm_memory != NULL)
	{
       return
         cipher_descriptor.accel_gcm_memory
                                          (key,   keylen,
                                           IV,    IVlen,
                                           adata, adatalen,
                                           pt,    ptlen,
                                           ct,
                                           tag,   taglen,
                                           direction);
    }



#ifndef LTC_GCM_TABLES_SSE2
    orig = gcm = XMALLOC(sizeof(*gcm));
#else
    orig = gcm = XMALLOC(sizeof(*gcm) + 16);
#endif
    if (gcm == NULL)
	{
        return CRYPT_MEM;
    }

   /* Force GCM to be on a multiple of 16 so we can use 128-bit aligned operations
    * note that we only modify gcm and keep orig intact.  This code is not portable
    * but again it's only for SSE2 anyways, so who cares?
    */
#ifdef LTC_GCM_TABLES_SSE2
   if ((unsigned long)gcm & 15) {
      gcm = (gcm_state *)((unsigned long)gcm + (16 - ((unsigned long)gcm & 15)));
   }
#endif

    if ((err = gcm_init(gcm, cipher, key, keylen)) != CRYPT_OK)
	{
       goto LTC_ERR;
    }
    if ((err = gcm_add_iv(gcm, IV, IVlen)) != CRYPT_OK)
	{
       goto LTC_ERR;
    }
    if ((err = gcm_add_aad(gcm, adata, adatalen)) != CRYPT_OK)
	{
       goto LTC_ERR;
    }
    if ((err = gcm_process(gcm, pt, ptlen, ct, direction)) != CRYPT_OK)
	{
       goto LTC_ERR;
    }
    err = gcm_done(gcm, tag, taglen);
LTC_ERR:
    XFREE(orig);
    return err;
}
#endif
