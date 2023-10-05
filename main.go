package main

import (
	"context"
	"log"
	"math/rand"
	"net"
	"time"

	pb "github.com/shuai1900/gRPC_microservice/account_proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func init() {
	rand.New(rand.NewSource(1000000))
}

const (
	address = "localhost:50051"
)

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
	go startGRPCServerService()

	// client service to write to postgresql
	time.Sleep(1 * time.Second)
	startGRPCClientService()
	select {}

}

func startRESTApiService(store *PostgresStore) {
	server := NewApiServer(":3000", store)
	server.Run()
}

func startGRPCServerService() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	//init new grpc server
	s := grpc.NewServer()
	pb.RegisterAccountManagementServer(s, &AccountServer{})
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to server:%v", err)
	}
}

func startGRPCClientService() {
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	// create new client with this connection
	c := pb.NewAccountManagementClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	first_name := "bob"
	last_name := "jack"
	r, err := c.CreateAccount(ctx, &pb.NewAccount{
		FirstName: first_name,
		LastName:  last_name,
	})
	if err != nil {
		log.Fatalf("could not create new account:%v", err)
	}
	log.Printf(`Account details:
	First_name: %s
	Last_name: %s
	Number: %d
	`, r.GetFirstName(), r.GetLastName(), r.GetNumber())

}
