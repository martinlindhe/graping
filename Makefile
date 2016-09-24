GOPATH = $(shell echo $$GOPATH)

all: build

build:
	go build -v .
	# needed for ICMP ping on macOS and Linux
	sudo chown root ./graping
	sudo chmod u+s ./graping

install: build
	go install .
	sudo chown root $(GOPATH)/bin/graping
	sudo chmod u+s $(GOPATH)/bin/graping
