GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean

BUILD_DIR=../../build
EXT=.exe

CMDS=$(notdir $(wildcard $(CMD_DIR)/*))
EXAMPLES=$(notdir $(wildcard $(EXAMPLES_DIR)/*))
PROTOS=$(wildcard ./*/**/**/*.proto)
PROTOPB=$(PROTOS:.proto=.pb.go)

hwinfo-plugin: proto
	go build -o $(BUILD_DIR)/hwinfo-plugin.exe main.go

proto: $(PROTOPB)

$(PROTOPB): $(PROTOS)
	protoc \
 		--go_out=Mgrpc/service_config/service_config.proto=/internal/proto/grpc_service_config:. \
		--go-grpc_out=Mgrpc/service_config/service_config.proto=/internal/proto/grpc_service_config:. \
  		--go_opt=paths=source_relative \
		--go-grpc_opt=paths=source_relative \
		$(<)

clean:
	$(GOCLEAN)
	rm -f $(BUILD_DIR)/*

run:
	$(BUILD_DIR)/$(word 1, $(MAKECMDGOALS))$(EXT)

.PHONY: $(CMDS) $(EXAMPLES) proto clean run
