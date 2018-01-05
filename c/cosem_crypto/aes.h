#ifndef AES_H_
#define AES_H_

#ifdef __cplusplus
extern "C" {
#endif

#define SETUP    rijndael_setup
#define ECB_ENC  rijndael_ecb_encrypt
#define ECB_DEC  rijndael_ecb_decrypt
#define ECB_DONE rijndael_done
#define ECB_KS   rijndael_keysize
 

#ifdef __cplusplus
   }
#endif

#endif
