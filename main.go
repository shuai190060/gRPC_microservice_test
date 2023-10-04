package main

import (
	"log"
)

func main() {

	store, err := NewPostgreStore()
	if err != nil {
		log.Fatal(err)
	}

	// fmt.Printf("%+v\n", *store) // test the db connection
	server := NewApiServer(":3000", store)
	server.Run()

}
