PROJECT_DIR = $(CURDIR)
PROJECT_BIN = ${PROJECT_DIR}/bin

.PHONY: bin.build
bin.build:
	mkdir -p ${PROJECT_DIR}/build
	rm -f ${PROJECT_DIR}/build/blk
	go build -ldflags="-s -w" -o ${PROJECT_DIR}/build/blk ${PROJECT_DIR}/cmd/app/main.go

.PHONY: up
up: 
	sudo docker compose up --build -d

.PHONY: run.local
run.local: bin.build
	${PROJECT_DIR}/build/blk

.PHONY: start.d
start.d:
	sudo systemctl start docker