package main

import "golang.org/x/crypto/bcrypt"

type Account struct {
	ID                int64  `json:"id"`
	FirstName         string `json:"first_name"`
	LastName          string `json:"last_name"`
	EncryptedPassword string `json:"-"`
}

type CreateAccountReq struct {
	FirtName string `json:"first_name"`
	LastName string `json:"last_name"`
	Password string `json:"password"`
}

type Todo struct {
	Number  int    `json:"number"`
	Done    bool   `json:"done"`
	Context string `json:"context"`
}

type TodoAddReq struct {
	Done    bool   `json:"done"`
	Context string `json:"context"`
}

type TodoStatusChangeReq struct {
	Number int  `json:"number"`
	Done   bool `json:"done"`
}

type TodoModifyReq struct {
	Add          []TodoAddReq          `json:"add"`
	StatusChange []TodoStatusChangeReq `json:"status_change"`
	Delete       []int                 `json:"delete"`
}

type DeleteAccountReq struct {
	ID int64 `json:"id"`
}

type PasswordReq struct {
	Password string `json:"password"`
}

func NewAccount(fname, lname string, pass string) (*Account, error) {
	enc_pwd, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	return &Account{
		ID:                GetAndIncrementLastID(),
		FirstName:         fname,
		LastName:          lname,
		EncryptedPassword: string(enc_pwd),
	}, nil
}
