
all:
	8g gelrun.go
	8g geli.go
	8g turtle.go
	8l -ogelo gelrun.8
	8l -ogeli geli.8
	8l -oturtle turtle.8

nuke:
	rm -f *.8 gelo geli turtle
	rm -f *.8 ../examples/{gelo,geli,turtle}

install: all
	mv -t ../examples gelo geli turtle
