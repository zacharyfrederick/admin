.PHONY: startNetwork
startNetwork: stopNetwork
	/bin/bash ./scripts/startNetwork.sh

.PHONY: stopNetwork
stopNetwork:
	/bin/bash ./scripts/stopNetwork.sh

.PHONY: deployCC
deployCC: cleanWallet startNetwork
	/bin/bash ./scripts/deployCC.sh

.PHONY: startServer
startServer: 
	go run cmd/server/main.go

.PHONY: cleanWallet
cleanWallet:
	rm -rf ${GOPATH}/src/github.com/zacharyfrederick/admin/cmd/server/wallet/*.id