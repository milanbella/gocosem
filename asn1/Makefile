objects = $(filter-out converter-sample.o, $(patsubst %.c,%.o,$(wildcard *.c)))
libCosemPdu.a : $(objects)
	ar rcs $@ $<
%.o : %.c
	$(CC) -I. -c $(CFLAGS) $(CPPFLAGS) $< -o $@
clean:
	rm *.o
