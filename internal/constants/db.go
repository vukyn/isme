package constants

// DB_FILE_PATH is the on-disk path (relative to the repo root) of the file-based
// SQLite database. It is the single source of truth shared by the DI DB opener
// and the database-backup scheduled job (which derives its backup directory from
// the same path).
const DB_FILE_PATH = "db/app.db"
