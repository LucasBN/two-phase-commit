package main

import (
	"context"
	"log"
	"lucasbn/two-phase-commit/app/participant"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Account struct {
	ID        string
	FirstName string
	LastName  string
	Balance   int
}

func main() {

	participant.Run("postgres://lucas@localhost:5432/accounts_a")

	return

	databases := []string{
		"postgres://lucas@localhost:5432/accounts_a",
		"postgres://lucas@localhost:5432/accounts_b",
	}

	pools := make([]*pgxpool.Pool, 0, len(databases))

	for _, db := range databases {
		pool, err := pgxpool.New(context.Background(), db)
		if err != nil {
			log.Fatalf("unable to connect to database: %v", err)
		}
		defer pool.Close()
		pools = append(pools, pool)
	}


	

	// Example: transfer 50 from account 0 (which lives in the first database)
	// to account 1 (which lives in the second database)
	tx1, err := pools[0].Begin(context.Background())
	if err != nil {
		log.Fatalf("unable to begin transaction: %v", err)
	}

	tx2, err := pools[1].Begin(context.Background())
	if err != nil {
		log.Fatalf("unable to begin transaction: %v", err)
	}

	var fromBalance int
	err = tx1.QueryRow(context.Background(), "SELECT balance FROM account WHERE id = '0';").Scan(&fromBalance)
	if err != nil {
		log.Fatalf("QueryRow failed: %v", err)
	}

	var toBalance int
	err = tx2.QueryRow(context.Background(), "SELECT balance FROM account WHERE id = '1';").Scan(&toBalance)
	if err != nil {
		log.Fatalf("QueryRow failed: %v", err)
	}

	if fromBalance < 50 {
		log.Fatalf("insufficient funds in account 0")
	}

	_, err = tx1.Exec(context.Background(), "UPDATE account SET balance = $1 WHERE id = '0';", fromBalance-50)
	if err != nil {
		log.Fatalf("Exec failed: %v", err)
	}

	_, err = tx2.Exec(context.Background(), "UPDATE account SET balance = $1 WHERE id = '1';", toBalance+50)
	if err != nil {
		log.Fatalf("Exec failed: %v", err)
	}

	_, err = tx1.Exec(context.Background(), "PREPARE TRANSACTION 'a_test';")
	if err != nil {
		log.Fatalf("Exec failed: %v", err)
	}

	_, err = tx2.Exec(context.Background(), "PREPARE TRANSACTION 'b_test';")
	if err != nil {
		log.Fatalf("Exec failed: %v", err)
	}

	err = tx1.Commit(context.Background())
	if err != nil {
		log.Fatalf("unable to commit transaction: %v", err)
	}

	err = tx2.Commit(context.Background())
	if err != nil {
		log.Fatalf("unable to commit transaction: %v", err)
	}
}
