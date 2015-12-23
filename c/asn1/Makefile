objects = $(filter-out converter-sample.o, $(patsubst %.c,%.o,$(wildcard *.c)))

CPPFLAGS = -Wall -I. -I/usr/local/include/gc

libCosemPdu.a : $(objects) go/asn1_go.o
	echo $(objects)
	ar rcs $@ $^
%.o : %.c
	$(CC) -I. -c $(CFLAGS) $(CPPFLAGS) $< -o $@
clean:
	rm *.o
	rm go/*.o
