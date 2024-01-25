EXECNAME=pushme
EXECPATH=/usr/local/bin

# TODO: Support Windows installation
install: 
	go build -o ${EXECNAME} main.go
	sudo mv ${EXECNAME} ${EXECPATH}
	@echo "installed!"

uninstall:
	sudo rm ${EXECPATH}/${EXECNAME}
	@echo "uninstalled!"

