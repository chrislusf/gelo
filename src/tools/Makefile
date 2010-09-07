
include $(GOROOT)/src/Make.inc

all:
	$(GC) gelrun.go
	$(GC) geli.go
	$(GC) turtle.go
	$(LD) -ogelo gelrun.$(O)
	$(LD) -ogeli geli.$(O)
	$(LD) -oturtle turtle.$(O)

nuke:
	rm -f *.$(O) gelo geli turtle
	rm -f *.$(O) ../examples/{gelo,geli,turtle}

install: all
	mv -t ../examples gelo geli turtle
