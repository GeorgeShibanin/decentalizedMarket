package main

import (
	"decentralizedProject/handlers"
	"decentralizedProject/storage/mongostorage"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
	"time"
)

func NewServer() *http.Server {
	r := mux.NewRouter()
	handler := &handlers.HTTPHandler{}
	mongoUrl := os.Getenv("MONGO_URL")
	//mongoUrl := "mongodb://localhost:27017"
	mongoStorage := mongostorage.NewStorage(mongoUrl)
	handler = &handlers.HTTPHandler{
		Storage: mongoStorage,
	}

	r.HandleFunc("/", handlers.HandleRoot).Methods("GET", "POST")
	r.HandleFunc("/deliveries", handler.HandleGetOrder).Methods("GET", "POST")
	r.HandleFunc("/confirm", handler.HandleAcceptDelivery).Methods(http.MethodPost)

	port := os.Getenv("SERVER_PORT")
	return &http.Server{
		Handler:      r,
		Addr:         ":" + port,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
}

func main() {
	srv := NewServer()
	log.Printf("Start serving on %s", srv.Addr)
	log.Fatal(srv.ListenAndServe())
}
