package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func AddMainRoutes(router *mux.Router) {
	router.PathPrefix("/static/").Handler(
		http.StripPrefix("/static/", http.FileServer(http.Dir("./frontend"))),
	)
	router.HandleFunc("/{any:.*}",
		func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, "./frontend/cascii-core/cascii.html")
		},
	)
}

func main() {

	dbFactory := DbFactory{maxConns: 5, maxIdleConns: 5}
	dbClient := dbFactory.Get()
	defer dbClient.Close()

	router := mux.NewRouter()

	AddApiRoutes(router, &Servicers{db: dbClient})
	AddMainRoutes(router)

	http.Handle("/", router)

	// TODO: Look into timeout configs
	log.Fatal(http.ListenAndServe(":8000", nil))
}
