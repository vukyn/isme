package history

import (
	"context"
	"database/sql"
	"testing"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/sqliteshim"
)

// newMigratedTestDB opens an in-memory SQLite database and applies every
// migration that comes before the named one.
func newMigratedTestDB(t *testing.T, before string) *bun.DB {
	t.Helper()

	sqldb, err := sql.Open(sqliteshim.ShimName, ":memory:")
	if err != nil {
		t.Fatalf("open in-memory sqlite: %v", err)
	}
	// a single connection keeps every query on the same in-memory database
	sqldb.SetMaxOpenConns(1)
	db := bun.NewDB(sqldb, sqlitedialect.New())
	t.Cleanup(func() { db.Close() })

	for _, migration := range Migrations {
		if migration.Name == before {
			return db
		}
		if err := migration.Up(db); err != nil {
			t.Fatalf("migration %s failed: %v", migration.Name, err)
		}
	}
	t.Fatalf("migration %s is not registered", before)
	return nil
}

type userRow struct {
	ID         string
	Name       string
	Email      string
	Password   string
	Status     int
	IsVerified int
	AvatarURL  sql.NullString
	DeletedAt  sql.NullString
	DeletedBy  sql.NullString
}

func readUsers(t *testing.T, db *bun.DB) map[string]userRow {
	t.Helper()
	rows, err := db.QueryContext(context.Background(), `
		SELECT id, name, email, password, status, is_verified, avatar_url, deleted_at, deleted_by
		FROM users`)
	if err != nil {
		t.Fatalf("read users: %v", err)
	}
	defer rows.Close()

	users := map[string]userRow{}
	for rows.Next() {
		var row userRow
		if err := rows.Scan(&row.ID, &row.Name, &row.Email, &row.Password, &row.Status, &row.IsVerified, &row.AvatarURL, &row.DeletedAt, &row.DeletedBy); err != nil {
			t.Fatalf("scan user: %v", err)
		}
		users[row.ID] = row
	}
	return users
}

// emailUniqueIndexes lists the unique indexes on users that cover email, with
// whether each is partial.
func emailUniqueIndexes(t *testing.T, db *bun.DB) map[string]bool {
	t.Helper()
	rows, err := db.QueryContext(context.Background(), `
		SELECT il.name, il.partial
		FROM pragma_index_list('users') AS il
		JOIN pragma_index_info(il.name) AS ii
		WHERE il."unique" = 1 AND ii.name = 'email'`)
	if err != nil {
		t.Fatalf("list indexes: %v", err)
	}
	defer rows.Close()

	indexes := map[string]bool{}
	for rows.Next() {
		var name string
		var partial int
		if err := rows.Scan(&name, &partial); err != nil {
			t.Fatalf("scan index: %v", err)
		}
		indexes[name] = partial == 1
	}
	return indexes
}

func insertUser(db *bun.DB, id, email string, deleted bool) error {
	deletedAt := any(nil)
	if deleted {
		deletedAt = "2026-01-01 00:00:00"
	}
	_, err := db.ExecContext(context.Background(), `
		INSERT INTO users (id, name, email, password, status, deleted_at)
		VALUES (?, ?, ?, 'hash', 1, ?)`, id, "name-"+id, email, deletedAt)
	return err
}

func TestM032UsersEmailUniqueLiveOnly(t *testing.T) {
	db := newMigratedTestDB(t, m032UsersEmailUniqueLiveOnly.Name)
	ctx := context.Background()

	seed := []string{
		`INSERT INTO users (id, name, email, password, status, is_verified, avatar_url)
			VALUES ('usr_live', 'Live', 'live@hasaki.vn', 'hash-live', 1, 1, 'https://cdn/a.png')`,
		`INSERT INTO users (id, name, email, password, status, is_verified, deleted_at, deleted_by)
			VALUES ('usr_gone', 'Gone', 'gone@hasaki.vn', 'hash-gone', 2, 1, '2026-01-01 00:00:00', 'usr_admin')`,
		`INSERT INTO users (id, name, email, password, status, is_verified)
			VALUES ('usr_unverified', 'Unverified', 'unverified@hasaki.vn', 'hash-unverified', 1, 0)`,
	}
	for _, statement := range seed {
		if _, err := db.ExecContext(ctx, statement); err != nil {
			t.Fatalf("seed: %v", err)
		}
	}
	before := readUsers(t, db)

	if err := m032UsersEmailUniqueLiveOnly.Up(db); err != nil {
		t.Fatalf("up: %v", err)
	}

	after := readUsers(t, db)
	if len(after) != len(before) {
		t.Fatalf("expected %d users after rebuild, got %d", len(before), len(after))
	}
	for id, row := range before {
		if after[id] != row {
			t.Errorf("user %s changed across the rebuild: before %+v after %+v", id, row, after[id])
		}
	}

	indexes := emailUniqueIndexes(t, db)
	if len(indexes) != 1 || !indexes["users_email_live_uidx"] {
		t.Fatalf("expected only the partial users_email_live_uidx on email, got %v", indexes)
	}

	// a deleted user's email is free again, a second live holder is not
	if err := insertUser(db, "usr_reinvited", "gone@hasaki.vn", false); err != nil {
		t.Errorf("expected a live user to take a soft-deleted user's email, got %v", err)
	}
	if err := insertUser(db, "usr_duplicate", "live@hasaki.vn", false); err == nil {
		t.Error("expected a second live user with the same email to be rejected")
	}
	if err := insertUser(db, "usr_gone_again", "live@hasaki.vn", true); err != nil {
		t.Errorf("expected a soft-deleted row to share a live email, got %v", err)
	}
}

func TestM032DownRestoresTableWideUnique(t *testing.T) {
	db := newMigratedTestDB(t, m032UsersEmailUniqueLiveOnly.Name)

	if err := insertUser(db, "usr_one", "one@hasaki.vn", false); err != nil {
		t.Fatalf("seed: %v", err)
	}
	if err := m032UsersEmailUniqueLiveOnly.Up(db); err != nil {
		t.Fatalf("up: %v", err)
	}
	if err := m032UsersEmailUniqueLiveOnly.Down(db); err != nil {
		t.Fatalf("down: %v", err)
	}

	indexes := emailUniqueIndexes(t, db)
	if len(indexes) != 1 {
		t.Fatalf("expected one table-wide unique index on email after down, got %v", indexes)
	}
	for name, partial := range indexes {
		if partial {
			t.Errorf("expected index %s to be table-wide after down", name)
		}
	}
	if len(readUsers(t, db)) != 1 {
		t.Error("expected the user to survive down")
	}
	if err := insertUser(db, "usr_two", "one@hasaki.vn", true); err == nil {
		t.Error("expected the table-wide constraint to reject a soft-deleted duplicate after down")
	}
}
