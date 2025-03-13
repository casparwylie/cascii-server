package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

func IsProd() bool {
	return os.Getenv("CASCII_ENV") == "prod"
}

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

	log.Printf("Starting server - prod: %t", IsProd())
	log.Fatal(http.ListenAndServe(":8000", nil))
}
