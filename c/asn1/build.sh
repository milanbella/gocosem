#see asn1c compiler: https://github.com/vlm/asn1c
set -x
asn1c cosem.asn1
cp go/asn_internal.h .
make
