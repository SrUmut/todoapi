package main

import (
	"database/sql"

	_ "github.com/lib/pq"
)

type Storage interface {
	CreateAccount(*Account) error
	GetAccount() ([]*Account, error)
	GetTodoByID(int64) ([]*Todo, error)
	AddTodoWithID(int64, []TodoAddReq) error
	StatusChangeTodoWithID(int64, []TodoStatusChangeReq) error
	DeleteTodoWithID(int64, []int) error
	DeleteAccountByID(int64) error
	GetAccountByID(int64) (*string, error)
}

type PostgresStore struct {
	db *sql.DB
}

var lastID int64

func NewPostgresStore() (*PostgresStore, error) {
	connStr := "user=postgres dbname=postgres password=mysecretpassword sslmode=disable" // example configuration from go pq doc
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	return &PostgresStore{db}, nil
}

func (s *PostgresStore) Init() error {
	if err := s.createAccountTable(); err != nil {
		return err
	}
	if err := s.createTODOTable(); err != nil {
		return err
	}
	if err := s.getLastIDFromDB(&lastID); err != nil {
		return err
	}
	return nil
}

func (s *PostgresStore) GetAccount() ([]*Account, error) {
	query := `SELECT * FROM account`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	accounts := make([]*Account, 0)
	for rows.Next() {
		account := new(Account)
		err := rows.Scan(
			&account.ID,
			&account.FirstName,
			&account.LastName,
			&account.EncryptedPassword,
		)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, account)
	}

	return accounts, err
}

func (s *PostgresStore) CreateAccount(account *Account) error {
	query := `
	INSERT INTO account
	VALUES ($1, $2, $3, $4)
	`
	_, err := s.db.Exec(
		query,
		account.ID,
		account.FirstName,
		account.LastName,
		account.EncryptedPassword,
	)
	return err
}

func (s *PostgresStore) GetTodoByID(id int64) ([]*Todo, error) {
	query := `
		SELECT number, done, context FROM todo
		WHERE account_id=$1
	`
	rows, err := s.db.Query(query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	todos := make([]*Todo, 0)
	for rows.Next() {
		todo := new(Todo)
		err := rows.Scan(
			&todo.Number,
			&todo.Done,
			&todo.Context,
		)
		if err != nil {
			return nil, err
		}
		todos = append(todos, todo)
	}
	return todos, nil
}

func (s *PostgresStore) AddTodoWithID(id int64, reqs []TodoAddReq) error {
	query := `
	INSERT INTO todo (account_id, context, done)
	VALUES ($1, $2, $3)`
	for _, req := range reqs {
		if _, err := s.db.Exec(query, id, req.Context, req.Done); err != nil {
			return err
		}
	}

	return nil
}

func (s *PostgresStore) StatusChangeTodoWithID(id int64, reqs []TodoStatusChangeReq) error {
	query := `
	UPDATE todo
	SET done = $1
	WHERE account_id = $2 and number = $3`
	for _, req := range reqs {
		if _, err := s.db.Exec(query, req.Done, id, req.Number); err != nil {
			return err
		}
	}

	return nil
}

func (s *PostgresStore) DeleteTodoWithID(id int64, numbers []int) error {
	query := `DELETE FROM todo WHERE account_id = $1 and number = $2`
	for _, number := range numbers {
		if _, err := s.db.Exec(query, id, number); err != nil {
			return err
		}
	}

	return nil
}

func (s *PostgresStore) DeleteAccountByID(id int64) error {
	query := `DELETE FROM account WHERE id = $1;`
	if _, err := s.db.Exec(query, id); err != nil {
		return err
	}

	query = `DELETE FROM todo WHERE account_id = $1;`
	if _, err := s.db.Exec(query, id); err != nil {
		return err
	}

	return nil
}

func (s *PostgresStore) GetAccountByID(id int64) (*string, error) {
	query := `SELECT encrypted_password FROM account WHERE id = $1`
	row, err := s.db.Query(query, id)
	if err != nil {
		return nil, err
	}

	row.Next()
	enc_pass := new(string)
	if err := row.Scan(enc_pass); err != nil {
		return nil, err
	}

	return enc_pass, nil
}

func (s *PostgresStore) createAccountTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS account (
			id SERIAL PRIMARY KEY,
			first_name VARCHAR(50),
			last_name VARCHAR(50),
			encrypted_password VARCHAR(100)
		)`
	_, err := s.db.Exec(query)
	return err
}

func (s *PostgresStore) createTODOTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS todo (
			account_id BIGINT,
			number SERIAL,
			context VARCHAR(100),
			done BOOLEAN,
			PRIMARY KEY (account_id, number)
		)`
	_, err := s.db.Exec(query)
	return err
}

func (s *PostgresStore) getLastIDFromDB(id *int64) error {
	query := `SELECT MAX(id) FROM account`
	row := s.db.QueryRow(query)
	var lastIDNull sql.NullInt64
	if err := row.Scan(&lastIDNull); err != nil {
		return err
	}

	if !lastIDNull.Valid {
		*id = 0
	} else {
		*id = lastIDNull.Int64
	}
	return nil
}

func GetAndIncrementLastID() int64 {
	defer func() {
		lastID++
	}()
	return lastID
}
