package main

import (
	"context"
	"log"
	"math/rand"
	"net"
	"os"

	pb "github.com/shuai1900/gRPC_microservice/account_proto"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	port = ":50051"
)

type AccountServer struct {
	pb.UnimplementedAccountManagementServer
	// account_list *pb.AccountList
}

func (s *AccountServer) CreateAccount(ctx context.Context, in *pb.NewAccount) (*pb.Account, error) {
	log.Printf("received:%v", in.GetFirstName())

	readBytes, err := os.ReadFile("accounts.json")

	var account_list *pb.AccountList = &pb.AccountList{}

	var account_id int32 = int32(rand.Intn(10000))
	created_account := &pb.Account{
		Id:        account_id,
		FirstName: in.GetFirstName(),
		LastName:  in.GetLastName(),
		Number:    int64(rand.Intn(1000000)),
		Balance:   0,
		CreatedAt: timestamppb.Now(),
	}
	if err != nil {
		if os.IsNotExist(err) {
			log.Println("file not found, creating a new json file")
			account_list.Accounts = append(account_list.Accounts, created_account)
			jsonBytes, err := protojson.Marshal(account_list)
			if err != nil {
				log.Fatalf("json marshaling failed: %v", err)
			}
			if err := os.WriteFile("accounts.json", jsonBytes, 0644); err != nil {
				log.Fatalf("failed write to file: %v", err)
			}
			return created_account, nil
		} else {
			log.Fatalf("error reading file:%v", err)
		}
	}
	if err := protojson.Unmarshal(readBytes, account_list); err != nil {
		log.Fatalf("failed to parse user list: %v", err)
	}
	account_list.Accounts = append(account_list.Accounts, created_account)
	jsonBytes, err := protojson.Marshal(account_list)
	if err != nil {
		log.Fatalf("json marshaling failed: %v", err)
	}
	if err := os.WriteFile("accounts.json", jsonBytes, 0644); err != nil {
		log.Fatalf("failed write to file: %v", err)
	}
	return created_account, nil

}

func NewAccountServer() *AccountServer {
	return &AccountServer{}
}

func (server *AccountServer) Run() error {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		return err
	}
	//init new grpc server
	s := grpc.NewServer()
	pb.RegisterAccountManagementServer(s, server)
	log.Printf("server listening at %v", lis.Addr())
	return s.Serve(lis)

}

func (s *AccountServer) GetAccount(ctx context.Context, in *pb.GetAccountParams) (*pb.AccountList, error) {
	jsonBytes, err := os.ReadFile("accounts.json")
	if err != nil {
		log.Fatalf("failed read from file: %v", err)
	}
	var account_list *pb.AccountList = &pb.AccountList{}

	if err := protojson.Unmarshal(jsonBytes, account_list); err != nil {
		log.Fatalf("failed to ummarshal: %v", err)

	}
	return account_list, nil

}
