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
	TypeExecuteMysql TypeExecute = "mysql"
)

type Execute[T string | embed.FS] struct {
	typeSQL TypeExecute
	db      *sql.DB
	source  T
}

func NewExecute[T string | embed.FS](types TypeExecute, db *gorm.DB, source T) (*Execute[T], error) {
	switch types {
	case TypeExecuteMysql:
		dbs, err := db.DB()
		if err != nil {
			return nil, err
		}
		return &Execute[T]{types, dbs, source}, nil
	default:
		return nil, fmt.Errorf("unknown type: %s", types)
	}
}

func (e *Execute[T]) ExecuteSQL() error {
	var src, err = push(e.source)
	_, err = migrate.Exec(e.db, "mysql", src, migrate.Up)
	return err
}
