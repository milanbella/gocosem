
#if defined(LRW_MODE) || defined(LTC_GCM_MODE)
void gcm_gf_mult(const unsigned char *a, const unsigned char *b, unsigned char *c);
#endif


/* table shared between GCM and LRW */
#if defined(LTC_GCM_TABLES) || defined(LRW_TABLES) || ((defined(LTC_GCM_MODE) || defined(LTC_GCM_MODE)) && defined(LTC_FAST))
extern const unsigned char gcm_shift_table[];
#endif

#ifdef LTC_GCM_MODE

#define GCM_ENCRYPT 0
#define GCM_DECRYPT 1

#define LTC_GCM_MODE_IV    0
#define LTC_GCM_MODE_AAD   1
#define LTC_GCM_MODE_TEXT  2

typedef struct { 
   symmetric_key       K;
   unsigned char       H[16],        /* multiplier */
                       X[16],        /* accumulator */
                       Y[16],        /* counter */
                       Y_0[16],      /* initial counter */
                       buf[16];      /* buffer for stuff */

   int                 cipher,       /* which cipher */
                       ivmode,       /* Which mode is the IV in? */
                       mode,         /* mode the GCM code is in */
                       buflen;       /* length of data in buf */

   ulong64             totlen,       /* 64-bit counter used for IV and AAD */
                       pttotlen;     /* 64-bit counter for the PT */

#ifdef LTC_GCM_TABLES
   unsigned char       PC[16][256][16]  /* 16 tables of 8x128 */
#ifdef LTC_GCM_TABLES_SSE2
__attribute__ ((aligned (16)))
#endif
;
#endif  
} gcm_state;

void gcm_mult_h(gcm_state *gcm, unsigned char *I);

int gcm_init(gcm_state *gcm, int cipher,
             const unsigned char *key, int keylen);

int gcm_reset(gcm_state *gcm);

int gcm_add_iv(gcm_state *gcm, 
               const unsigned char *IV,     unsigned long IVlen);

int gcm_add_aad(gcm_state *gcm,
               const unsigned char *adata,  unsigned long adatalen);

int gcm_process(gcm_state *gcm,
                     unsigned char *pt,     unsigned long ptlen,
                     unsigned char *ct,
                     int direction);

int gcm_done(gcm_state *gcm, 
                     unsigned char *tag,    unsigned long *taglen);

int gcm_memory(      int           cipher,
               const unsigned char *key,    unsigned long keylen,
               const unsigned char *IV,     unsigned long IVlen,
               const unsigned char *adata,  unsigned long adatalen,
                     unsigned char *pt,     unsigned long ptlen,
                     unsigned char *ct, 
                     unsigned char *tag,    unsigned long *taglen,
                               int direction);
int gcm_test(void);

#endif /* LTC_GCM_MODE */

/* $Source: /cvs/libtom/libtomcrypt/src/headers/tomcrypt_mac.h,v $ */
/* $Revision: 1.23 $ */
/* $Date: 2007/05/12 14:37:41 $ */
