package main

import (
	"context"
	"log"
	"math/rand"

	pb "github.com/shuai1900/gRPC_microservice/account_proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	port = ":50051"
)

type AccountServer struct {
	pb.UnimplementedAccountManagementServer
}

func (s *AccountServer) CreateAccount(ctx context.Context, in *pb.NewAccount) (*pb.Account, error) {
	log.Printf("received:%v", in.GetFirstName())

	var account_id int32 = int32(rand.Intn(1000000))
	return &pb.Account{
		Id:        account_id,
		FirstName: in.GetFirstName(),
		LastName:  in.GetLastName(),
		Number:    int64(rand.Intn(1000000)),
		Balance:   0,
		CreatedAt: timestamppb.Now(),
	}, nil

}
