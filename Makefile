
GOBIN=$(shell pwd)/bin
export GOBIN

PACK_DIR := dlserver
VER:=v1.0


.PHONY: all
all:
	go install ./src/srv/

package:
	mkdir -p $(PACK_DIR)
	cp ./bin/srv $(PACK_DIR)/dlserver
	cp -af ./images $(PACK_DIR)/ 
	cp -af ./js $(PACK_DIR)/
	cp -af app.html $(PACK_DIR)/
	tar czvf dlserver-$(VER).tar.gz $(PACK_DIR)
	rm $(PACK_DIR) -rf
