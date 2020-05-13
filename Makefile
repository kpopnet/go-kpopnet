all: kpopnetd

db/bin_data.go: $(wildcard db/sql/*.sql)
	go generate ./db

.PHONY: kpopnetd
kpopnetd: db/bin_data.go
	go build ./cmd/kpopnetd

serve: kpopnetd testdata
	./kpopnetd

gofmt:
	go fmt ./...

testdata:
	git clone https://github.com/Kagami/go-face-testdata testdata

test: kpopnetd testdata
	go test -v ./facerec
