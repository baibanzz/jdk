package execute

import (
	"database/sql"
	"embed"
	"fmt"

	migrate "github.com/rubenv/sql-migrate"
	"gorm.io/gorm"
)

type TypeExecute string

const (
	TypeExecuteMysql     TypeExecute = "mysql"
	TypeExecuteSqlite    TypeExecute = "sqlite3"
	TypeExecutePostgres  TypeExecute = "postgres"
	TypeExecuteMssql     TypeExecute = "mssql"
	TypeExecuteOci8      TypeExecute = "oci8"
	TypeExecuteGodror    TypeExecute = "godror"
	TypeExecuteSnowflake TypeExecute = "snowflake"
)

// dialects sql-migrate 支持的方言集合
var dialects = map[TypeExecute]bool{
	TypeExecuteMysql: true, TypeExecuteSqlite: true, TypeExecutePostgres: true,
	TypeExecuteMssql: true, TypeExecuteOci8: true, TypeExecuteGodror: true,
	TypeExecuteSnowflake: true,
}

type Execute[T string | embed.FS] struct {
	typeSQL TypeExecute
	db      *sql.DB
	source  T
}

func NewExecute[T string | embed.FS](types TypeExecute, db *gorm.DB, source T) (*Execute[T], error) {
	if !dialects[types] {
		return nil, fmt.Errorf("unknown type: %s", types)
	}
	dbs, err := db.DB()
	if err != nil {
		return nil, err
	}
	return &Execute[T]{types, dbs, source}, nil
}

func (e *Execute[T]) ExecuteSQL() error {
	src, err := push(e.source)
	if err != nil {
		return err
	}
	_, err = migrate.Exec(e.db, string(e.typeSQL), src, migrate.Up)
	return err
}
