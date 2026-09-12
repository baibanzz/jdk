package core

import (
	"embed"
	"errors"

	"github.com/baibanzz/jdk/core/internal/execute"
	"github.com/baibanzz/jdk/core/internal/mysql"
	"github.com/baibanzz/jdk/core/internal/postgres"
	redispkg "github.com/baibanzz/jdk/core/internal/redis"
	"github.com/baibanzz/jdk/core/internal/sqlite"
	"github.com/baibanzz/jdk/model"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewDB[T model.Mysql | model.Sqlite3 | model.PostgreSql](t T, loggers logger.Interface) (*gorm.DB, error) {
	switch v := any(t).(type) {
	case model.Mysql:
		return mysql.NewMysql(v, loggers)
	case model.Sqlite3:
		return sqlite.NewSqlite3(v, loggers)
	case model.PostgreSql:
		return postgres.NewPostgreSql(v, loggers)
	default:
		return nil, errors.New("类型错误")
	}
}

func NewRedis(r model.Redis) (*redis.Client, error) {
	return redispkg.NewRedis(r)
}

type TypeExecute = execute.TypeExecute

const (
	TypeExecuteMysql     = execute.TypeExecuteMysql
	TypeExecuteSqlite    = execute.TypeExecuteSqlite
	TypeExecutePostgres  = execute.TypeExecutePostgres
	TypeExecuteMssql     = execute.TypeExecuteMssql
	TypeExecuteOci8      = execute.TypeExecuteOci8
	TypeExecuteGodror    = execute.TypeExecuteGodror
	TypeExecuteSnowflake = execute.TypeExecuteSnowflake
)

type Execute[T string | embed.FS] = execute.Execute[T]

func NewExecute[T string | embed.FS](types TypeExecute, db *gorm.DB, source T) (*Execute[T], error) {
	return execute.NewExecute(types, db, source)
}
