package main

import (
    "fmt"
    "net/http"
    "os"
    "log"
    "github.com/alternaDev/geomodel"
    "github.com/gorilla/mux"
)

func handler(w http.ResponseWriter, r *http.Request) {
  vars := mux.Vars(r)

  fmt.Fprintf(w, "Hi there, I love %s! %f", vars["name"], geomodel.Distance(0, 0, 1, 1))
}

func setupRouter() *mux.Router {
  r := mux.NewRouter()

  r.HandleFunc("/{name}", handler)

  return r
}

func listen(port string) {
  r := setupRouter()

  http.Handle("/", r)
  http.ListenAndServe(":" + port, nil)
}

func main() {
    port := os.Getenv("PORT")

    if port == "" {
      log.Fatal("$PORT must be set")
    }

    listen(port)
}
