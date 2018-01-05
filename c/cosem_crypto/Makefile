objects = $(patsubst %.c,%.o,$(wildcard *.c))

CPPFLAGS = -Wall -I.

libcosemcrypto.a : $(objects)
	echo $(objects)
	ar rcs $@ $^
%.o : %.c
	$(CC) -I. -c $(CFLAGS) $(CPPFLAGS) $< -o $@
clean:
	rm *.o
