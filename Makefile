
.PHONY: 
all: help

## license: Adds license header to missing files.
license:
	@echo "  >  \033[32mAdding license headers...\033[0m "
	go get -u github.com/google/addlicense
	addlicense -v -c "Sygma" -f ./scripts/header.txt -y 2023 -ignore ".idea/**"  .

## license-check: Checks for missing license headers
license-check:
	@echo "  >  \033[Checking for license headers...\033[0m "
	GO111MODULE=off go get -u github.com/google/addlicense
	addlicense -check -c "Sygma" -f ./scripts/header.txt -y 2021 -ignore ".idea/**" .

coverage:
	go tool cover -func cover.out | grep total | awk '{print $3}'

test:
	./scripts/tests.sh

e2e-test:
	./scripts/e2e-test.sh

genmocks:
	mockgen -source=./chains/evm/listener/handlers/stateRoot.go -destination=./mock/stateRoot.go -package mock
	mockgen -source=./chains/evm/message/stateRoot.go -destination=./mock/stateRootMessage.go -package mock
	mockgen -destination=./mock/store.go -package mock github.com/sygmaprotocol/sygma-core/store KeyValueReaderWriter
	mockgen -source=./chains/evm/proof/receipt.go -destination=./mock/proof.go -package mock 
	mockgen -source=./chains/evm/proof/root.go -destination=./mock/root.go -package mock 



PLATFORMS := linux/amd64 darwin/amd64 darwin/arm64 linux/arm
temp = $(subst /, ,$@)
os = $(word 1, $(temp))
arch = $(word 2, $(temp))

$(PLATFORMS):
	GOOS=$(os) GOARCH=$(arch) go build -ldflags "-X google.golang.org/protobuf/reflect/protoregistry.conflictPolicy=ignore" -o 'build/${os}-${arch}/relayer'; \

build-all: $(PLATFORMS)
