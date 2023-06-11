INSTALL_FILES = \
	/etc/systemd/system/sockactex.socket \
	/etc/systemd/system/sockactex.service \
	/var/lib/sockactex/bin/sockactex

install: $(INSTALL_FILES)
	sudo systemctl restart sockactex

/etc/systemd/system/sockactex.socket: sockactex.socket
	sudo install sockactex.socket /etc/systemd/system/sockactex.socket
	sudo systemctl daemon-reload

/etc/systemd/system/sockactex.service: sockactex.service
	sudo install sockactex.service /etc/systemd/system/sockactex.service
	sudo systemctl daemon-reload

/var/lib/sockactex/bin/sockactex: main.go
	sudo mkdir -p /var/lib/sockactex/bin
	go build -trimpath -tags netgo -o /tmp/sockactex
	sudo install /tmp/sockactex /var/lib/sockactex/bin/sockactex

clean:
	sudo rm -f $(INSTALL_FILES)
	sudo systemctl daemon-reload

.PHONY: install clean
