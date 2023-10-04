package main

import (
	"math/rand"
	"time"
)

type Account struct {
	ID        int       `json:"id"`
	FirstName string    `json:"firstName"`
	LastName  string    `json:"lastName"`
	Number    int64     `json:"number"`
	Balance   int64     `json:"balance"`
	CreatedAT time.Time `json:"createdAt"`
}

func NewAccount(firstName, lastName string) *Account {
	s := rand.NewSource(time.Now().UnixNano())
	r := rand.New(s)
	return &Account{
		// ID:        rand.Intn(1000000),
		FirstName: firstName,
		LastName:  lastName,
		Number:    int64(r.Intn(1000000)),
		Balance:   0,
		CreatedAT: time.Now().UTC(),
	}
}

type CreateAccountRequest struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

type TransferRequest struct {
	ToAccount int `json:"toAccount"`
	Amount    int `json:"amount"`
}
