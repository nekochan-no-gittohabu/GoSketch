package main

import (
	"fmt"
	"log"
	"net/http"

	"personal/gosketch/service/config"
)

const (
	port = ":8080"
)

func main() {
	db, srv, err := config.Run()
	if err != nil {
		log.Fatalf(err.Error())
	}

	defer db.Close()

	err = http.ListenAndServe(port, srv)
	if err != nil {
		fmt.Println(err)
		return
	}
}
