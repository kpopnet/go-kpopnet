all: kpopnetd

.PHONY: kpopnetd
kpopnetd:
	go build -o . ./...

serve: kpopnetd
	./kpopnetd serve

gofmt:
	go fmt ./...

testdata:
	git clone https://github.com/Kagami/go-face-testdata testdata

test: testdata
	go test -v
