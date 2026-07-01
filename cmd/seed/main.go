// Command seed bootstraps a fresh local isme database with admin users and the
// downstream app_services (medioa, rainy) plus their permission catalogs, so a
// just-migrated dev DB is usable without manual clicking.
//
// Idempotent: re-running resets the admin passwords and tops up any missing
// perms/roles. An app_service is created only once (its plaintext secret is
// shown ONCE on creation — rotate via the UI if you lose it).
//
// Environment-aware (driven by APP_ENV / cfg.App.Env):
//   - non-prod: admin emails use the `.local` suffix and share the fixed dev
//     password "123456789"; secrets are printed to stdout.
//   - production: admin emails use the `.prod` suffix, each admin gets a strong
//     random password, and all sensitive values (admin passwords + minted app
//     secrets) are written to ../SEED_CREDENTIALS_PROD.md (platform root, 0600,
//     gitignored) INSTEAD of stdout.
//
// Usage: go run cmd/seed/main.go
package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/vukyn/isme/internal/config"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect"
	kueryDb "github.com/vukyn/kuery/bun/db"
	"github.com/vukyn/kuery/cryp"
	"github.com/vukyn/kuery/cryp/aes"
	"github.com/vukyn/kuery/cryp/rand"
)

const (
	// created_by/updated_by are left empty to match the migration-seeded system
	// rows (app_isme, system roles). A non-empty marker like "seed" would point
	// app_service/role list endpoints at a non-existent user → GetByID 500.
	seedActor   = ""
	adminPass   = "123456789"
	appCtxInfo  = "authen" // mirrors the isme self-app; used as AES additional-data for the secret
	statusActive = 1
)

// permission is one (resource, action) pair scoped to an app.
type permission struct {
	resource string
	action   string
}

// appSeed describes a downstream app_service + its admin role + perm catalog.
type appSeed struct {
	id          string // app_service id (referenced as app_id by perms/roles)
	code        string // app_code — must match the consumer's AUTH_APP_CODE
	name        string
	redirectURL string
	icon        string
	color       string
	adminRoleID string
	adminEmail  string // the admin user granted full perms on this app
	perms       []permission
	icons       map[string]string // resource -> icon key (per-resource icon, migration 019)
}

// isPostgres reports whether the bun DB targets Postgres (vs SQLite).
func isPostgres(db *bun.DB) bool {
	return db.Dialect().Name() == dialect.PG
}

// insertIgnore renders a dialect-aware idempotent insert. On SQLite it uses
// `INSERT OR IGNORE INTO`; on Postgres it uses `INSERT INTO ... ON CONFLICT
// (<conflictCols>) DO NOTHING`. body is the `<table> (cols) VALUES (...)` part
// (without the leading INSERT keyword). conflictCols is the conflict target for
// Postgres (e.g. "code, action" or "id"); ignored on SQLite.
func insertIgnore(db *bun.DB, body, conflictCols string) string {
	if isPostgres(db) {
		return "INSERT INTO " + body + " ON CONFLICT (" + conflictCols + ") DO NOTHING"
	}
	return "INSERT OR IGNORE INTO " + body
}

func crud(resources ...string) []permission {
	out := []permission{}
	for _, r := range resources {
		for _, a := range []string{"read", "create", "update", "delete"} {
			out = append(out, permission{r, a})
		}
	}
	return out
}

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	if cfg.AES.Secret == "" {
		log.Fatal("AES_SECRET is empty — cannot encrypt app secrets")
	}

	// Dialect-aware open driven by the same config the app uses: DB_DRIVER=postgres
	// targets Postgres, otherwise SQLite (default). Mirrors internal/di/di_db.go.
	db, err := kueryDb.Open(kueryDb.Config{
		Driver:      kueryDb.Driver(cfg.DB.Driver),
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
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()
	ctx := context.Background()

	// Environment-aware behavior: prod uses `.prod` admin emails + per-account
	// random passwords + a credentials file; everything else keeps the dev
	// defaults (`.local` emails, shared password, stdout secrets).
	isProd := cfg.App.Env == "production"
	suffix := "local"
	if isProd {
		suffix = "prod"
	}
	ismeEmail := "admin@isme." + suffix
	medioaEmail := "admin@medioa." + suffix
	rainyEmail := "admin@rainy." + suffix

	// redirect_url is left blank in prod (set per-environment later via the admin
	// UI / consumer config); dev keeps the local callback URLs.
	medioaRedirect := "http://app.medioa.local:8082/auth/callback"
	rainyRedirect := "http://app.rainy.local:8083/auth/callback"
	if isProd {
		medioaRedirect = ""
		rainyRedirect = ""
	}

	// adminPasswords collects each admin's plaintext password so prod runs can
	// write them to the credentials file. In non-prod every entry is adminPass.
	adminPasswords := map[string]string{}
	// passwordFor returns the password to seed for an admin email: a strong
	// per-account random string in prod, the shared dev constant otherwise.
	passwordFor := func(email string) string {
		if isProd {
			return rand.RandMixedString(20, true, true)
		}
		return adminPass
	}

	// --- admin users (per-app: each admin owns its namesake app) ---
	// Fixed ULIDs so a wipe-and-reseed reproduces the SAME ids (other systems /
	// downstream data reference these). On an existing DB the id is keyed by
	// email and never changes. (A `.prod` email is simply a new row that still
	// gets the fixed id — dev and prod are separate DBs.)
	users := []struct{ id, name, email string }{
		{"01KTKDKNXTZDSGH5YKG151J877", "ISME Admin", ismeEmail},
		{"01KBYG3MYVVSYEKTRDJ4VT3DK6", "Medioa Admin", medioaEmail},
		{"01KTR9CB27MT4PJQ3TZ4P6SCCX", "Rainy Admin", rainyEmail},
	}
	userIDs := map[string]string{}
	for _, u := range users {
		password := passwordFor(u.email)
		id, err := upsertUser(ctx, db, u.id, u.name, u.email, password)
		if err != nil {
			log.Fatalf("upsert user %s: %v", u.email, err)
		}
		userIDs[u.email] = id
		adminPasswords[u.email] = password
		fmt.Printf("user  ok  %-20s %s\n", u.email, id)
	}

	// --- isme admin -> existing isme admin role (seeded by migrations) ---
	if err := assignRole(ctx, db, userIDs[ismeEmail], "rol_admin", "app_isme"); err != nil {
		log.Fatalf("assign isme admin: %v", err)
	}
	fmt.Printf("role  ok  %s -> rol_admin @ app_isme\n", ismeEmail)

	// --- downstream apps: medioa + rainy ---
	apps := []appSeed{
		{
			id: "app_medioa", code: "medioa", name: "Medioa",
			redirectURL: medioaRedirect, icon: "layers", color: "sky",
			adminRoleID: "rol_admin_medioa", adminEmail: medioaEmail,
			perms: medioaPerms(),
			icons: map[string]string{
				"object":   "file",
				"bucket":   "folder",
				"storage":  "database",
				"api_key":  "key",
				"settings": "settings",
			},
		},
		{
			id: "app_rainy", code: "rainy", name: "Rainy",
			redirectURL: rainyRedirect, icon: "cloud-rain", color: "magenta",
			adminRoleID: "rol_admin_rainy", adminEmail: rainyEmail,
			perms: append(crud("playlist", "station", "track", "album", "artist"),
				permission{"settings", "read"},
				permission{"settings", "update"},
			),
			icons: map[string]string{
				"album":    "album",
				"artist":   "user",
				"playlist": "playlist",
				"station":  "station",
				"track":    "music",
				"settings": "settings",
			},
		},
	}

	secrets := map[string]string{}
	for _, a := range apps {
		secret, created, err := ensureAppService(ctx, db, cfg.AES.Secret, a)
		if err != nil {
			log.Fatalf("app_service %s: %v", a.code, err)
		}
		if created {
			secrets[a.code] = secret
			fmt.Printf("app   new %-8s %s (secret printed below)\n", a.code, a.id)
		} else {
			fmt.Printf("app   ok  %-8s %s (exists — rotate via UI to get a secret)\n", a.code, a.id)
		}

		if err := ensureRole(ctx, db, a.adminRoleID, a.id, "admin", "Admin", "Full access to every resource", true); err != nil {
			log.Fatalf("role %s: %v", a.adminRoleID, err)
		}
		granted, err := seedPermsAndGrant(ctx, db, a.id, a.adminRoleID, a.perms)
		if err != nil {
			log.Fatalf("perms %s: %v", a.code, err)
		}
		if err := assignRole(ctx, db, userIDs[a.adminEmail], a.adminRoleID, a.id); err != nil {
			log.Fatalf("assign %s admin: %v", a.code, err)
		}
		// per-resource icons (migration 019: one icon per (app_id, resource))
		for resource, icon := range a.icons {
			if err := setResourceIcon(ctx, db, a.id, resource, icon); err != nil {
				log.Fatalf("icon %s/%s: %v", a.code, resource, err)
			}
		}
		fmt.Printf("role  ok  %s -> %s @ %s (%d perms, %d icons)\n", a.adminEmail, a.adminRoleID, a.id, granted, len(a.icons))
	}

	// --- medioa member role: full storage surface, granted to the isme + rainy admins ---
	if err := ensureRole(ctx, db, "rol_member_medioa", "app_medioa", "member", "Member", "Read/write access to the storage surface", false); err != nil {
		log.Fatalf("medioa member role: %v", err)
	}
	// appearance: user icon + blue; set via UPDATE so re-runs apply it even
	// though ensureRole is INSERT OR IGNORE.
	if _, err := db.ExecContext(ctx,
		`UPDATE roles SET icon = 'user', color = 'sky', updated_at = current_timestamp WHERE id = 'rol_member_medioa'`); err != nil {
		log.Fatalf("medioa member role appearance: %v", err)
	}
	// full storage surface + full api_key (medioa api_key = create/read/delete, no update)
	memberPermSet := append(crud("storage"),
		permission{"api_key", "create"},
		permission{"api_key", "read"},
		permission{"api_key", "delete"},
	)
	memberPerms, err := seedPermsAndGrant(ctx, db, "app_medioa", "rol_member_medioa", memberPermSet)
	if err != nil {
		log.Fatalf("medioa member perms: %v", err)
	}
	for _, email := range []string{ismeEmail, rainyEmail} {
		if err := assignRole(ctx, db, userIDs[email], "rol_member_medioa", "app_medioa"); err != nil {
			log.Fatalf("assign medioa member %s: %v", email, err)
		}
		fmt.Printf("role  ok  %s -> rol_member_medioa @ app_medioa (%d perms)\n", email, memberPerms)
	}

	// --- summary ---
	fmt.Printf("\n=== Seed complete ===\n")
	if isProd {
		// Prod: never print passwords (or, for safety, secrets) to stdout —
		// funnel every sensitive value into the credentials file at the
		// platform root.
		path, err := writeProdCredentials(cfg.App.Env, adminPasswords, secrets)
		if err != nil {
			log.Fatalf("write credentials file: %v", err)
		}
		fmt.Printf("Production seed: admin passwords + app secrets WRITTEN to %s (0600, gitignored).\n", path)
		fmt.Printf("That file holds the sensitive values — store it securely, then delete it.\n")
		fmt.Printf("Admin passwords are RANDOM per account; admins must change them after first login.\n")
		return
	}

	fmt.Printf("All 3 admins password: %s\n", adminPass)
	if len(secrets) > 0 {
		fmt.Printf("\nApp secrets (shown ONCE — copy into the consumer's .env, or rotate later):\n")
		for code, s := range secrets {
			fmt.Printf("  %-8s app_secret = %s\n", code, s)
		}
		fmt.Printf("\nConsumer .env: set AUTH_APP_CODE + the app_secret above (medioa->medioa2/.env, rainy->rainy/.env).\n")
		fmt.Printf("Verify each app_service redirect_url in the isme admin UI matches the consumer's callback.\n")
	}
}

// writeProdCredentials writes a 0600 Markdown file at the platform root
// (../SEED_CREDENTIALS_PROD.md — the seeder runs from the isme/ dir) holding
// every sensitive value minted this run: per-admin plaintext passwords and the
// freshly-created app secrets. It returns the path written. The file is the
// SOLE record of these secrets in prod (they are not printed to stdout), so it
// must be stored securely and then deleted.
func writeProdCredentials(appEnv string, adminPasswords, secrets map[string]string) (string, error) {
	const path = "../SEED_CREDENTIALS_PROD.md"

	var b strings.Builder
	b.WriteString("# isme Seed Credentials (PRODUCTION)\n\n")
	b.WriteString("> WARNING: This file contains PLAINTEXT secrets (admin passwords + app secrets).\n")
	b.WriteString("> It is gitignored and must NOT be committed. Store these values in your secret\n")
	b.WriteString("> manager, then DELETE this file.\n\n")
	b.WriteString("- Generated: " + time.Now().UTC().Format(time.RFC3339) + "\n")
	b.WriteString("- APP_ENV: " + appEnv + "\n\n")

	b.WriteString("## Admin accounts\n\n")
	b.WriteString("| Email | Password |\n")
	b.WriteString("| --- | --- |\n")
	for _, email := range []string{"admin@isme.prod", "admin@medioa.prod", "admin@rainy.prod"} {
		if pw, ok := adminPasswords[email]; ok {
			b.WriteString("| " + email + " | `" + pw + "` |\n")
		}
	}
	b.WriteString("\n")

	b.WriteString("## App secrets (minted this run)\n\n")
	if len(secrets) == 0 {
		b.WriteString("_None minted this run (app_services already existed — rotate via the isme admin UI to get a new secret)._\n\n")
	} else {
		b.WriteString("| App code | app_secret |\n")
		b.WriteString("| --- | --- |\n")
		for code, s := range secrets {
			b.WriteString("| " + code + " | `" + s + "` |\n")
		}
		b.WriteString("\n")
	}

	b.WriteString("## Next steps\n\n")
	b.WriteString("- Set the `medioa` app_secret as `AUTH_APP_SECRET` in medioa2's environment.\n")
	b.WriteString("- Set the `rainy` app_secret as `AUTH_APP_SECRET` in rainy's environment.\n")
	b.WriteString("- Admins must change their password after first login.\n")

	if err := os.WriteFile(path, []byte(b.String()), 0600); err != nil {
		return "", err
	}
	return path, nil
}

// medioaPerms is the exact catalog medioa2 enforces (appCode=medioa). NOTE:
// api_key has no `update`, and bucket carries invite/member_read/quota beyond CRUD.
func medioaPerms() []permission {
	perms := crud("object", "bucket", "storage")
	perms = append(perms,
		permission{"object", "share"},
		permission{"bucket", "invite"},
		permission{"bucket", "member_read"},
		permission{"bucket", "quota"},
		permission{"api_key", "create"},
		permission{"api_key", "read"},
		permission{"api_key", "delete"},
		permission{"settings", "read"},
		permission{"settings", "update"},
		permission{"analytics", "read"},
	)
	return perms
}

// upsertUser creates the user (verified + active) or, if the email exists,
// resets its password/status so re-running the seed restores a known login.
func upsertUser(ctx context.Context, db *bun.DB, id, name, email, password string) (string, error) {
	if id == "" {
		id = cryp.ULID()
	}
	_, err := db.ExecContext(ctx, `
		INSERT INTO users (id, name, email, password, status, is_verified, created_by, updated_by)
		VALUES (?, ?, ?, ?, ?, 1, ?, ?)
		ON CONFLICT(email) DO UPDATE SET
			password = excluded.password,
			status = excluded.status,
			is_verified = 1,
			updated_at = current_timestamp,
			updated_by = excluded.updated_by
	`, id, name, email, cryp.HashArgon2id(password), statusActive, seedActor, seedActor)
	if err != nil {
		return "", err
	}
	var storedID string
	if err := db.QueryRowContext(ctx, `SELECT id FROM users WHERE email = ?`, email).Scan(&storedID); err != nil {
		return "", err
	}
	return storedID, nil
}

// ensureAppService creates the app_service once with a freshly generated,
// AES-encrypted secret. Returns the plaintext + created=true only on first
// creation (the plaintext is unrecoverable afterwards — rotate via the UI).
func ensureAppService(ctx context.Context, db *bun.DB, aesSecret string, a appSeed) (string, bool, error) {
	var existing string
	err := db.QueryRowContext(ctx, `SELECT id FROM app_services WHERE app_code = ?`, a.code).Scan(&existing)
	if err == nil {
		return "", false, nil // already exists, leave it (secret stays as-is)
	}
	if err != sql.ErrNoRows {
		return "", false, err
	}

	plain := rand.RandMixedString(8, true, true)
	if plain == "" {
		return "", false, fmt.Errorf("failed to generate app_secret")
	}
	encrypted, err := aes.Encrypt(plain, aesSecret, appCtxInfo)
	if err != nil {
		return "", false, fmt.Errorf("encrypt app_secret: %w", err)
	}
	// Guard the insert with ON CONFLICT DO NOTHING (app_code is UNIQUE) so a
	// re-run can't violate the constraint even though the SELECT above already
	// returned ErrNoRows for this code.
	_, err = db.ExecContext(ctx, insertIgnore(db,
		`app_services
			(id, app_code, app_name, app_secret, redirect_url, ctx_info, status, icon, color, created_by, updated_by)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, "app_code"),
		a.id, a.code, a.name, encrypted, a.redirectURL, appCtxInfo, statusActive, a.icon, a.color, seedActor, seedActor)
	if err != nil {
		return "", false, err
	}
	return plain, true, nil
}

// setResourceIcon sets the per-resource icon on all of a resource's permission
// rows (migration 019: icon is shared across the same (app_id, resource)).
func setResourceIcon(ctx context.Context, db *bun.DB, appID, resource, icon string) error {
	_, err := db.ExecContext(ctx,
		`UPDATE permissions SET icon = ? WHERE app_id = ? AND resource = ?`, icon, appID, resource)
	return err
}

func ensureRole(ctx context.Context, db *bun.DB, id, appID, code, name, description string, isSystem bool) error {
	systemFlag := 0
	if isSystem {
		systemFlag = 1
	}
	_, err := db.ExecContext(ctx, insertIgnore(db,
		`roles (id, app_id, code, name, description, is_system, created_by, updated_by)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`, "id"),
		id, appID, code, name, description, systemFlag, seedActor, seedActor)
	return err
}

// seedPermsAndGrant inserts each app-scoped permission and grants it to the role.
func seedPermsAndGrant(ctx context.Context, db *bun.DB, appID, roleID string, perms []permission) (int, error) {
	granted := 0
	for _, p := range perms {
		if _, err := db.ExecContext(ctx,
			insertIgnore(db, `permissions (app_id, resource, action) VALUES (?, ?, ?)`, "app_id, resource, action"),
			appID, p.resource, p.action); err != nil {
			return granted, err
		}
		var permID int64
		if err := db.QueryRowContext(ctx,
			`SELECT id FROM permissions WHERE app_id = ? AND resource = ? AND action = ?`,
			appID, p.resource, p.action).Scan(&permID); err != nil {
			return granted, err
		}
		if _, err := db.ExecContext(ctx,
			insertIgnore(db, `role_permissions (role_id, permission_id) VALUES (?, ?)`, "role_id, permission_id"),
			roleID, permID); err != nil {
			return granted, err
		}
		granted++
	}
	return granted, nil
}

// assignRole grants a user a role scoped to an app (idempotent via the
// (user_id, role_id, app_service_id) unique index from migration 009).
func assignRole(ctx context.Context, db *bun.DB, userID, roleID, appServiceID string) error {
	if userID == "" {
		return fmt.Errorf("empty user id")
	}
	// The idempotency key is the unique index (user_id, role_id,
	// COALESCE(app_service_id, '')) from migration 009; the seeder always passes a
	// concrete app_service_id, so the conflict target matches that expression index.
	_, err := db.ExecContext(ctx, insertIgnore(db,
		`user_roles (id, user_id, role_id, app_service_id, created_by)
		VALUES (?, ?, ?, ?, ?)`, "user_id, role_id, COALESCE(app_service_id, '')"),
		cryp.ULID(), userID, roleID, appServiceID, seedActor)
	return err
}
