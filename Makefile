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

test: testdata
	go test -v ./facerec

deploy: kpopnetd testdata
	docker build -t kpopnet .
	docker save kpopnet | pv | ssh -C ${KPOPNET_DEPLOY_HOST} 'docker load'
	ssh ${KPOPNET_DEPLOY_HOST} 'systemctl restart docker-compose@kpopnet.service'
