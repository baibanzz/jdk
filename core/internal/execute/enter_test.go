package execute

import (
	"embed"
	"testing"

	migrate "github.com/rubenv/sql-migrate"
)

//go:embed sql/*.sql
var sqlFS embed.FS

func TestExecuteSQL_DirSource(t *testing.T) {
	// 磁盘目录方式: 直接投入 sql 目录
	src := migrate.FileMigrationSource{Dir: "sql"}

	migrations, err := src.FindMigrations()
	if err != nil {
		t.Fatalf("FindMigrations 失败: %v", err)
	}

	if len(migrations) != 2 {
		t.Fatalf("期望 2 个 migration, 实际 %d", len(migrations))
	}
	if migrations[0].Id != "001_create.sql" {
		t.Fatalf("第一个 migration id 期望 001_create, 实际 %s", migrations[0].Id)
	}
	if migrations[1].Id != "002_alter.sql" {
		t.Fatalf("第二个 migration id 期望 002_alter, 实际 %s", migrations[1].Id)
	}
	if len(migrations[0].Up) == 0 || len(migrations[1].Up) == 0 {
		t.Fatal("migration 的 Up 语句不应为空")
	}
	t.Log("磁盘目录方式解析 sql 成功")
}

func TestExecuteSQL_EmbedSource(t *testing.T) {
	// embed 方式: //go:embed sql/* 后直接投入 embed.FS
	fs := &FileDirMigrationSource{sqlfss: []embed.FS{sqlFS}}

	migrations, err := fs.FindMigrations()
	if err != nil {
		t.Fatalf("FindMigrations 失败: %v", err)
	}

	if len(migrations) != 2 {
		t.Fatalf("期望 2 个 migration, 实际 %d", len(migrations))
	}
	if migrations[0].Id != "001_create.sql" {
		t.Fatalf("第一个 migration id 期望 001_create, 实际 %s", migrations[0].Id)
	}
	if migrations[1].Id != "002_alter.sql" {
		t.Fatalf("第二个 migration id 期望 002_alter, 实际 %s", migrations[1].Id)
	}
	t.Log("embed 方式解析 sql 成功")
}
