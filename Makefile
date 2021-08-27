.PHONY: startNetwork
startNetwork: stopNetwork
	/bin/bash ./scripts/startNetwork.sh

.PHONY: stopNetwork
stopNetwork:
	/bin/bash ./scripts/stopNetwork.sh

.PHONY: deployCC
deployCC: startNetwork
	/bin/bash ./scripts/deployCC.sh

.PHONY: startServer
startServer: 
	go run cmd/server/main.go