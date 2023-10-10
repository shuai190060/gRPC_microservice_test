package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"

	"github.com/jackc/pgx/v4"
	"github.com/prometheus/client_golang/prometheus"
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

	// // fmt.Printf("%+v\n", *store) // test the db connection
	// run Rest Api service
	go startRESTApiService(store)

	//----------------------------------------------------------------------------------
	// gRPC
	//----------------------------------------------------------------------------------

	// server service
	startGRPCServerService()

	//----------------------------------------------------------------------------------
	// metrics
	//----------------------------------------------------------------------------------

}

func startRESTApiService(store *PostgresStore) {
	reg := prometheus.NewRegistry()
	m := NewMetrics(reg)
	server := NewApiServer(":3000", store, m)
	server.Run()
}

func startGRPCServerService() {

	// Fetch environment variables
	user := os.Getenv("DB_USER")
	dbname := os.Getenv("DB_NAME")
	host := os.Getenv("DB_HOST")
	sslmode := os.Getenv("DB_SSLMODE")
	pass_db := os.Getenv("POSTGRES_PASSWORD")

	database_url := fmt.Sprintf("postgresql://%s:%s@%s:5432/%s?sslmode=%s", user, pass_db, host, dbname, sslmode)

	conn, err := pgx.Connect(context.Background(), database_url)
	if err != nil {
		log.Fatalf("unable to establish connection:%v", err)

	}
	defer conn.Close(context.Background())

	var account_server *AccountServer = NewAccountServer()
	account_server.conn = conn
	if err := account_server.Run(); err != nil {
		log.Fatalf("failed to server:%v", err)
	}
}

// func startGRPCClientService() {
// 	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
// 	if err != nil {
// 		log.Fatalf("did not connect: %v", err)
// 	}
// 	defer conn.Close()

// 	// create new client with this connection
// 	c := pb.NewAccountManagementClient(conn)
// 	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
// 	defer cancel()

// 	first_name := "bob"
// 	last_name := "jack"
// 	r, err := c.CreateAccount(ctx, &pb.NewAccount{
// 		FirstName: first_name,
// 		LastName:  last_name,
// 	})
// 	if err != nil {
// 		log.Fatalf("could not create new account:%v", err)
// 	}
// 	log.Printf(`Account details:
// 	First_name: %s
// 	Last_name: %s
// 	Number: %d
// 	`, r.GetFirstName(), r.GetLastName(), r.GetNumber())

// 	params := &pb.GetAccountParams{}
// 	res_acc_list, err := c.GetAccount(ctx, params)
// 	if err != nil {
// 		log.Fatalf("could not retrieve accounts: %v", err)
// 	}
// 	log.Print("\nuser list is:\n")
// 	fmt.Printf("r.GetAccount():%v", res_acc_list)

// }
