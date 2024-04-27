default:
	go vet
	staticcheck -checks all
	go test
