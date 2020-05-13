all: kpopnetd

.PHONY: kpopnetd
kpopnetd:
	go build ./cmd/kpopnetd

serve: kpopnetd testdata
	./kpopnetd

gofmt:
	go fmt ./...

testdata:
	git clone https://github.com/Kagami/go-face-testdata testdata

test: testdata
	go test -v
