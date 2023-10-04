package main

import (
	"log"
	"math/rand"
)

func init() {
	rand.New(rand.NewSource(1000000))
}

func main() {

	store, err := NewPostgreStore()
	if err != nil {
		log.Fatal(err)
	}
	if err := store.Init(); err != nil {
		log.Fatal(err)
	}

	// fmt.Printf("%+v\n", *store) // test the db connection
	server := NewApiServer(":3000", store)
	server.Run()

}
