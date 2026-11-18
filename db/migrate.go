package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/vukyn/isme/db/history"
	"github.com/vukyn/isme/internal/config"

	kueryDb "github.com/vukyn/kuery/bun/db"
	"github.com/vukyn/kuery/bun/migrate"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run db/migrate.go <command>")
		fmt.Println("Connection comes from the DB_* variables (.env or environment).")
		fmt.Println("Commands:")
		fmt.Println("  up    - Run all pending migrations (a fresh database needs only this)")
		fmt.Println("  down  - Rollback last migration")
		fmt.Println("  reset - Rollback all migrations")
		os.Exit(1)
	}

	command := os.Args[1]

	cfg, err := config.LoadConfig(".env")
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	db, err := kueryDb.Open(kueryDb.Config{
		Driver:      kueryDb.DriverPostgres,
		PostgresDSN: cfg.DB.DSN,
		Host:        cfg.DB.Host,
		Port:        cfg.DB.Port,
		User:        cfg.DB.User,
		Password:    cfg.DB.Password,
		DBName:      cfg.DB.DBName,
		SSLMode:     cfg.DB.SSLMode,
	})
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	// Print where the command points, never the DSN: it carries the password.
	var database string
	if err := db.NewRaw(`SELECT current_database()`).Scan(context.Background(), &database); err != nil {
		log.Fatalf("failed to reach database: %v", err)
	}
	fmt.Printf("Target: host=%s database=%s\n", targetHost(cfg.DB.DSN, cfg.DB.Host), database)

	switch command {
	case "up":
		stats, err := migrate.Run(db, history.Migrations)
		if err != nil {
			log.Fatalf("Migration failed: %v", err)
		}
		printMigrateReport(stats.TotalSuccess, stats.TotalSkipped, "All migrations completed successfully")
	case "down":
		rolledBack, err := migrate.RollbackLast(db, history.Migrations)
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
		n, err := migrate.Reset(db, history.Migrations)
		if err != nil {
			log.Fatalf("Rollback failed: %v", err)
		}
		printMigrateReport(n, 0, "All migrations rolled back successfully")
	default:
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
	}
}

// targetHost returns the host the connection points at without any credential.
func targetHost(dsn, host string) string {
	if dsn == "" {
		return host
	}
	parsed, err := url.Parse(dsn)
	if err != nil || parsed.Hostname() == "" {
		return "(set by DB_DSN)"
	}
	return parsed.Hostname()
}

// printMigrateReport prints the migration report with statistics
func printMigrateReport(totalSuccess, totalSkipped int, msg string) {
	fmt.Printf("\n=== Migration Report ===\n")
	fmt.Printf("Total Success: %d\n", totalSuccess)
	fmt.Printf("Total Skipped: %d\n", totalSkipped)
	fmt.Println(msg)
}
