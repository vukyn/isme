package usecase

import (
	"os"
	"testing"

	"github.com/vukyn/isme/internal/testdb"
)

// TestMain starts the throwaway Postgres container the database tests share.
func TestMain(m *testing.M) {
	os.Exit(testdb.Main(m))
}
