build:
	docker buildx build --platform linux/amd64 -t beebeewijaya/agar-data-collection .

build-local:
	docker buildx build -t beebeewijaya/agar-data-collection .

push: build
	docker push beebeewijaya/agar-data-collection

serial-log:
	caffeinate -i python3 tools/serial_log.py $(NAME)

PHONY: build build-local push serial-log
