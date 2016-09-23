build:
	go build -v .
	# needed for ICMP ping on macOS and Linux
	sudo chown root graping
	sudo chmod u+s graping
