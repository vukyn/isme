package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"

	sqliteHistory "github.com/vukyn/isme/db/history/sqlite"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/sqliteshim"
	"github.com/vukyn/kuery/bun/migrate"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run migrate.go <db_type> <command>")
		fmt.Println("Database types:")
		fmt.Println("  sqlite - SQLite database")
		fmt.Println("Commands:")
		fmt.Println("  up     - Run all pending migrations")
		fmt.Println("  down   - Rollback last migration")
		fmt.Println("  reset  - Rollback all migrations")
		os.Exit(1)
	}

	dbType := os.Args[1]
	command := os.Args[2]

	// Set database path based on type
	var dbPath string
	switch dbType {
	case "sqlite":
		dbPath = "db/app.db" // change to db/sqlite/app.db
	default:
		fmt.Printf("Unsupported database type: %s\n", dbType)
		os.Exit(1)
	}

	// Open database connection (creates file if not exists)
	sqldb, err := sql.Open(sqliteshim.ShimName, dbPath)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer sqldb.Close()

	db := bun.NewDB(sqldb, sqlitedialect.New())

	switch command {
	case "up":
		stats, err := migrate.Run(db, sqliteHistory.Migrations)
		if err != nil {
			log.Fatalf("Migration failed: %v", err)
		}
		printMigrateReport(stats.TotalSuccess, stats.TotalSkipped, "All migrations completed successfully")
	case "down":
		rolledBack, err := migrate.RollbackLast(db, sqliteHistory.Migrations)
		if err != nil {
			if errors.Is(err, migrate.ErrNoMigrationsToRollback) {
				printMigrateReport(0, 0, "No migrations to rollback")
			} else {
				log.Fatalf("Rollback failed: %v", err)
			}
		} else {
			totalSuccess := 0
			if rolledBack {
				totalSuccess = 1
			}
			printMigrateReport(totalSuccess, 0, "Last migration rolled back successfully")
		}
	case "reset":
		n, err := migrate.Reset(db, sqliteHistory.Migrations)
		if err != nil {
			log.Fatalf("Rollback failed: %v", err)
		}
		printMigrateReport(n, 0, "All migrations rolled back successfully")
	default:
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
	}
}

// printMigrateReport prints the migration report with statistics
func printMigrateReport(totalSuccess, totalSkipped int, msg string) {
	fmt.Printf("\n=== Migration Report ===\n")
	fmt.Printf("Total Success: %d\n", totalSuccess)
	fmt.Printf("Total Skipped: %d\n", totalSkipped)
	fmt.Println(msg)
}
