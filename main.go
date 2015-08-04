package main

import (
    "fmt"
    "net/http"
    "os"
    "log"
    "github.com/alternaDev/geomodel"
)

func handler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Hi there, I love %s! %f", r.URL.Path[1:], geomodel.Distance(0, 0, 1, 1))
}

func main() {
    port := os.Getenv("PORT")

  	if port == "" {
  		log.Fatal("$PORT must be set")
  	}

    http.HandleFunc("/", handler)
    http.ListenAndServe(":" + port, nil)

}
