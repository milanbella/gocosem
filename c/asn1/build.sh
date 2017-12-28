#see asn1c compiler: https://github.com/vlm/asn1c
set -xe
asn1c -no-gen-example -pdu=XDLMS-APDU cosem.asn1
#cp go/asn_internal.h .
make
