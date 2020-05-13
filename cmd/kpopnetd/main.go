package main

import (
	"fmt"
	"log"

	"github.com/kpopnet/go-kpopnet/db"
	"github.com/kpopnet/go-kpopnet/facerec"
	"github.com/kpopnet/go-kpopnet/server"

	"github.com/docopt/docopt-go"
)

// VERSION is current version.
const VERSION = "0.0.0"

// USAGE is usage help in docopt DSL.
const USAGE = `
K-pop face recognition backend.

Usage:
  kpopnetd [options]
  kpopnetd [-h | --help]
  kpopnetd [-V | --version]

Options:
  -h --help     Show this screen.
  -V --version  Show version.
  -H <host>     Host to listen on [default: 127.0.0.1].
  -p <port>     Port to listen on [default: 8002].
  -c <conn>     PostgreSQL connection string
                [default: user=meguca password=meguca dbname=meguca sslmode=disable].
  -d <datadir>  Data directory location [default: ./testdata].
`

type config struct {
	Host    string `docopt:"-H"`
	Port    int    `docopt:"-p"`
	Conn    string `docopt:"-c"`
	DataDir string `docopt:"-d"`
}

func serve(conf config) {
	if err := db.StartDB(nil, conf.Conn); err != nil {
		log.Fatal(err)
	}
	if err := facerec.StartFaceRec(conf.DataDir); err != nil {
		log.Fatal(err)
	}
	opts := server.ServerOptions{
		Address: fmt.Sprintf("%v:%v", conf.Host, conf.Port),
	}
	log.Printf("Listening on %v", opts.Address)
	log.Fatal(server.StartServer(opts))
}

func main() {
	opts, err := docopt.ParseArgs(USAGE, nil, VERSION)
	if err != nil {
		log.Fatal(err)
	}
	var conf config
	if err := opts.Bind(&conf); err != nil {
		log.Fatal(err)
	}
	serve(conf)
}
