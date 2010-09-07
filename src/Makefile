
include $(GOROOT)/src/Make.inc

all: pkgs tools

pkgs:
	cd gelo; make install
	cd gelo/extensions; make install
	cd gelo/commands; make install

tools: pkgs
	cd tools; make install

nuke:
	rm -f *.$(O) run
	cd tools; make nuke
	cd gelo; make nuke
	cd gelo/extensions; make nuke
	cd gelo/commands; make nuke
