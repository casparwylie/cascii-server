package main

import (
    "fmt"
    "log"
    "net/http"
)

func main() {


    http.HandleFunc("/hi", func(w http.ResponseWriter, r *http.Request){
        fmt.Fprintf(w, "Hi")
    })

    http.Handle("/", http.FileServer(http.Dir("./frontend")))
    log.Fatal(http.ListenAndServe(":8000", nil))

}
