

all: pkgs tools

pkgs:
	cd gelo; make install
	cd gelo/commands; make install

tools: pkgs
	cd tools; make install

nuke:
	rm -f *.8 run
	cd tools; make nuke
	cd gelo; make nuke
	cd gelo/commands; make nuke
