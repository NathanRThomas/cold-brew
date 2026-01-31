
GOCMD=go
GOBUILD=$(GOCMD) build -buildvcs=false
GOTEST=$(GOCMD) test -failfast -run
GOPATH=/usr/local/bin
DIR=$(shell pwd)

build: clear
build: build-api
build: build-qb

clear:
	@clear

build-api:
	@echo "building ColdBrew API..."
	@$(GOBUILD) -o $(GOPATH)/coldBrew ./cmd/api

build-qb:
	@echo "building ColdBrew QB..."
	@$(GOBUILD) -o $(GOPATH)/coldBrew-qb ./cmd/qb
	
update:
	clear
	@echo "updating dependencies..."
	@go get -u -t ./...
	@go mod tidy 

test:
	@clear 
	@echo "testing QA..."
	@$(GOTEST) QA ./...

test-net:
	@clear 
	@echo "testing NET..."
	@$(GOTEST) Net ./...

run: export CONFIG=$(DIR)/testing/config.json
run: export TEMPLATES=$(DIR)/templates/
run: build-api
run:
	@coldBrew -p 8080

run-qb: export CONFIG=$(DIR)/testing/config.json
run-qb: export TEMPLATES=$(DIR)/templates/
run-qb: build-qb
run-qb:
	@coldBrew-qb -p 8081

# for deploying on the vms
deploy: build
deploy:
	systemctl restart coldbrew
	systemctl restart coldbrewqb
	
