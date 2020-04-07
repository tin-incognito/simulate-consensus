#makefile

GOCMD=go
GORUN=$(GOCMD) run
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
BINARY_NAME=sawtooth

run:
	$(GORUN) *.go
build:
	$(GOBUILD) -o $(BINARY_NAME) -v
test:
	$(GOTEST) -v ./...
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
build-run:
	$(GOBUILD) -o $(BINARY_NAME) -v ./...
	./$(BINARY_NAME)
glide:
	$(GOGET) github.com/Masterminds/glide
	glide install
sawtooth:
	# $(GORUN) sawtooth.go message.go blockchain.go message.go main.go node.go
	$(GORUN) *.go

race:
	$(GORUN) -race *.go