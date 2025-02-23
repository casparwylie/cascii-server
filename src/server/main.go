package main

import (
	"log"
	"net/http"
)

func main() {

	dbFactory := DbFactory{maxConns: 5, maxIdleConns: 5}
	dbClient := dbFactory.Get()
	defer dbClient.Close()

	http.Handle("/", Router(&Servicers{db: dbClient}))

	// TODO: Look into timeout configs
	log.Fatal(http.ListenAndServe(":8000", nil))
}
