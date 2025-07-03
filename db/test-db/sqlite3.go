package main

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
)

func dumpCoinTables(db *sql.DB) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	rows, err := tx.Query(`SELECT * from coins`)
	if err != nil {
		return err
	}

	var userID string
	var coinsBefore int
	var coinsToday int

	for rows.Next() {
		err = rows.Scan(&userID, &coinsBefore, &coinsToday)

		if err != nil {
			return err
		}

		fmt.Printf("%s | %d | %d\n", userID, coinsBefore, coinsToday)
	}

	return tx.Commit()
}

func addDailyCoinForUser(db *sql.DB, userID string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`UPDATE coins SET coins_today = coins_today + 1 where user_id = ?`, userID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func rolloverCoins(db *sql.DB) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`UPDATE coins SET coins_before = coins_before + coins_today`)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`UPDATE coins SET coins_today = 0`)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func main() {
	db, err := sql.Open("sqlite3", "file:../josh.db")
	if err != nil {
		fmt.Printf("error opening database: %v\n", err)
		return
	}
	defer db.Close()

	if err := addDailyCoinForUser(db, "271750326422470656"); err != nil {
		fmt.Println(err.Error())
	}
	dumpCoinTables(db)

	fmt.Println()

	if err := rolloverCoins(db); err != nil {
		fmt.Println(err.Error())
	}
	dumpCoinTables(db)
}
