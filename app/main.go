package main

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Account struct {
	ID        string
	FirstName string
	LastName  string
	Balance   int
}

func main() {
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

	// Begin a transaction on each database
	tx1, err := pools[0].Begin(context.Background())
	if err != nil {
		log.Fatalf("unable to begin transaction: %v", err)
	}
	tx2, err := pools[1].Begin(context.Background())
	if err != nil {
		log.Fatalf("unable to begin transaction: %v", err)
	}

	// Read the balance of account 0 and confirm that it has enough funds
	var fromBalance int
	err = tx1.QueryRow(context.Background(), "SELECT balance FROM account WHERE id = '0';").Scan(&fromBalance)
	if err != nil {
		log.Fatalf("QueryRow failed: %v", err)
	}

	if fromBalance < 50 {
		// Rollback the transactions and return an error
		tx1.Rollback(context.Background())
		tx2.Rollback(context.Background())
		log.Fatalf("insufficient funds in account 0")
	}

	// Subract 50 from account 0
	_, err = tx1.Exec(context.Background(), "UPDATE account SET balance = $1 WHERE id = '0';", fromBalance-50)
	if err != nil {
		tx1.Rollback(context.Background())
		tx2.Rollback(context.Background())
		log.Fatalf("Exec failed: %v", err)
	}

	// Add 50 to account 1
	var initialToBalance int
	err = tx2.QueryRow(context.Background(), "SELECT balance FROM account WHERE id = '1';").Scan(&initialToBalance)
	if err != nil {
		tx1.Rollback(context.Background())
		tx2.Rollback(context.Background())
		log.Fatalf("QueryRow failed: %v", err)
	}

	_, err = tx2.Exec(context.Background(), "UPDATE account SET balance = $1 WHERE id = '1';", initialToBalance+50)
	if err != nil {
		tx1.Rollback(context.Background())
		tx2.Rollback(context.Background())
		log.Fatalf("Exec failed: %v", err)
	}

	// Phase 1: Prepare
	_, err = tx1.Exec(context.Background(), "PREPARE TRANSACTION 'transfer_a';")
	if err != nil {
		tx1.Rollback(context.Background())
		tx2.Rollback(context.Background())
		log.Fatalf("Exec failed: %v", err)
	}
	_, err = tx2.Exec(context.Background(), "PREPARE TRANSACTION 'transfer_b';")
	if err != nil {
		tx1.Rollback(context.Background())
		tx2.Rollback(context.Background())
		log.Fatalf("Exec failed: %v", err)
	}

	// Phase 2: Commit
	_, err = tx1.Exec(context.Background(), "COMMIT PREPARED 'transfer_a';")
	if err != nil {
		log.Fatalf("Exec failed: %v", err)
	}
	_, err = tx2.Exec(context.Background(), "COMMIT PREPARED 'transfer_b';")
	if err != nil {
		log.Fatalf("Exec failed: %v", err)
	}

	// Commit the transactions
	// err = tx1.Commit(context.Background())
	// if err != nil {
	// 	log.Fatalf("unable to commit transaction: %v", err)
	// }
	// err = tx2.Commit(context.Background())
	// if err != nil {
	// 	log.Fatalf("unable to commit transaction: %v", err)
	// }

}
