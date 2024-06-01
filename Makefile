PROJECT_DIR = $(CURDIR)
PROJECT_BIN = ${PROJECT_DIR}/bin
TOOLS_BIN = ${PROJECT_DIR}/tools

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

.PHONY: test
test:
	sudo docker compose -f docker-compose.test.yaml up --build --abort-on-container-exit
	sudo docker compose -f docker-compose.test.yaml down --volumes

.PHONY: get.tools
get.tools:
	mkdir -p ${TOOLS_BIN}
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b ${TOOLS_BIN} v1.59.0

.PHONY: lint
lint:
	${TOOLS_BIN}/golangci-lint run --config ./.golangci-lint.yaml ./...


