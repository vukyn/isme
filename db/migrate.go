package main

import (
	"errors"
	"fmt"
	"log"
	"os"

	sqliteHistory "github.com/vukyn/isme/db/history/sqlite"
	"github.com/vukyn/isme/internal/config"

	kueryDb "github.com/vukyn/kuery/bun/db"
	"github.com/vukyn/kuery/bun/migrate"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run migrate.go <db_type> <command>")
		fmt.Println("Database types:")
		fmt.Println("  sqlite   - SQLite database (default)")
		fmt.Println("  postgres - PostgreSQL database (connection from DB_* env vars)")
		fmt.Println("Commands:")
		fmt.Println("  up       - Run all pending migrations")
		fmt.Println("  down     - Rollback last migration")
		fmt.Println("  reset    - Rollback all migrations")
		fmt.Println("  baseline - Apply the squashed dual-dialect baseline (fresh install)")
		os.Exit(1)
	}

	dbType := os.Args[1]
	command := os.Args[2]

	// The CLI db_type arg selects the driver and overrides DB_DRIVER from .env,
	// so `make migrate-up DB=sqlite` always targets SQLite regardless of the
	// active default in .env.
	switch dbType {
	case "sqlite", "postgres":
		// supported
	default:
		fmt.Printf("Unsupported database type: %s\n", dbType)
		os.Exit(1)
	}

	// Load connection settings from .env (godotenv) via the shared config loader,
	// then build the dialect-aware connection. SQLitePath honors DB_SQLITE_PATH
	// (defaults to db/app.db) so a throwaway path can be used for smoke tests.
	cfg, err := config.LoadConfig(".env")
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	db, err := kueryDb.Open(kueryDb.Config{
		Driver:      kueryDb.Driver(dbType),
		SQLitePath:  cfg.DB.SQLitePath,
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

	switch command {
	case "up":
		stats, err := migrate.Run(db, sqliteHistory.Migrations)
		if err != nil {
			log.Fatalf("Migration failed: %v", err)
		}
		printMigrateReport(stats.TotalSuccess, stats.TotalSkipped, "All migrations completed successfully")
	case "baseline":
		// Apply the squashed dual-dialect baseline as the fresh-install path:
		// full schema + seed data in one migration, then the incremental 001-029
		// set is stamped as already applied (inside the baseline) so a later `up`
		// is a clean no-op. Not part of sqliteHistory.Migrations on purpose.
		stats, err := migrate.Run(db, []migrate.Migration{sqliteHistory.BaselineMigration})
		if err != nil {
			log.Fatalf("Baseline migration failed: %v", err)
		}
		printMigrateReport(stats.TotalSuccess, stats.TotalSkipped, "Baseline applied successfully")
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
