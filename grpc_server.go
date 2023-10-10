package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"time"

	"github.com/jackc/pgx/v4"
	pb "github.com/shuai1900/gRPC_microservice/account_proto"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	port = ":50051"
)

type AccountServer struct {
	pb.UnimplementedAccountManagementServer
	conn    *pgx.Conn
	metrics *grpcMetrics
	// account_list *pb.AccountList
}

func (s *AccountServer) CreateAccount(ctx context.Context, in *pb.NewAccount) (*pb.Account, error) {

	log.Printf("received:%v", in.GetFirstName())

	createSQL := `
		create table if not exists account (
			id serial primary key,
			first_name varchar(50),
			last_name varchar(50),
			number serial,
			balance serial,
			created_at timestamp
		)
	`

	_, err := s.conn.Exec(context.Background(), createSQL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "table creation failed: %v\n", err)
		os.Exit(1)
	}

	var account_id int32 = int32(rand.Intn(10000))
	created_account := &pb.Account{
		Id:        account_id,
		FirstName: in.GetFirstName(),
		LastName:  in.GetLastName(),
		Number:    int64(rand.Intn(1000000)),
		Balance:   0,
		CreatedAt: timestamppb.Now(),
	}

	//convert the timestamppb.Now() to time.Time

	// start a transaction
	tx, err := s.conn.Begin(context.Background())
	if err != nil {
		log.Fatalf("conn.Begin failed:%v", err)
	}
	insert := `
		insert into account (
			first_name, last_name, number, balance, created_at
		) values (
			$1,$2,$3,$4,$5
		)
	`
	_, err = tx.Exec(context.Background(), insert, created_account.FirstName, created_account.LastName, created_account.Number, created_account.Balance, TimestampProtoToTime(created_account.CreatedAt))
	if err != nil {
		log.Fatalf("tx.exec failed: %v", err)
	}
	tx.Commit(context.Background())

	return created_account, nil

}

// convert timestamppb to time
func TimestampProtoToTime(ts *timestamppb.Timestamp) time.Time {
	return ts.AsTime()
}

// func TimeToTimestampProto(t time.Time) *timestamppb.Timestamp {
// 	return timestamppb.New(t)
// }

func NewAccountServer(metrics *grpcMetrics) *AccountServer {
	return &AccountServer{
		conn:    nil,
		metrics: metrics,
	}
}

func (s *AccountServer) Run() error {

	lis, err := net.Listen("tcp", port)
	if err != nil {
		return err
	}
	//init new grpc server
	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(s.metrics.UnaryInterceptor))
	pb.RegisterAccountManagementServer(grpcServer, s)
	log.Printf("server listening at %v", lis.Addr())
	return grpcServer.Serve(lis)

}

func (s *AccountServer) GetAccount(ctx context.Context, in *pb.GetAccountParams) (*pb.AccountList, error) {
	query := `
		select * from account
	`
	var account_list *pb.AccountList = &pb.AccountList{}
	rows, err := s.conn.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		account := pb.Account{}
		var createdAt time.Time // to convert timestamppb.now into time.time
		err = rows.Scan(&account.Id, &account.FirstName, &account.LastName, &account.Number, &account.Balance, &createdAt)
		if err != nil {
			return nil, err
		}
		account.CreatedAt = timestamppb.New(createdAt) // rewrite the account.CreatedAt with the converted time
		account_list.Accounts = append(account_list.Accounts, &account)
	}

	return account_list, nil

}
