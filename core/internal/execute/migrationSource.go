package execute

import (
	"embed"
	"fmt"
	"io"
	"path"
	"sort"
	"strings"

	migrate "github.com/rubenv/sql-migrate"
)

func push[T string | embed.FS](source T) (migrate.MigrationSource, error) {
	switch v := any(source).(type) {
	case string:
		return migrate.FileMigrationSource{Dir: v}, nil
	case embed.FS:
		return &fileDirMigrationSource{sqlfss: []embed.FS{v}}, nil
	default:
		return nil, fmt.Errorf("unknown type %T", v)
	}
}

// fileDirMigrationSource 文件目录源
type fileDirMigrationSource struct {
	sqlfss []embed.FS
}

// FindMigrations FindMigrations
func (f fileDirMigrationSource) FindMigrations() ([]*migrate.Migration, error) {

	if len(f.sqlfss) == 0 {
		return nil, nil
	}
	migrations := make([]*migrate.Migration, 0, 100)

	for _, sqlfs := range f.sqlfss {
		err := f.findMigrations(sqlfs, &migrations)
		if err != nil {
			return nil, err
		}
	}

	// Make sure migrations are sorted
	sort.Sort(byID(migrations))

	return migrations, nil
}
func (f fileDirMigrationSource) findMigrations(fs embed.FS, migrations *[]*migrate.Migration) error {

	files, err := fs.ReadDir("sql")
	if err != nil {
		return err
	}
	for _, info := range files {

		if strings.HasSuffix(info.Name(), ".sql") {
			file, err := fs.Open(path.Join("sql", info.Name()))
			if err != nil {
				return fmt.Errorf("error while opening %s: %s", info.Name(), err)
			}

			migration, err := migrate.ParseMigration(info.Name(), file.(io.ReadSeeker))
			if err != nil {
				return fmt.Errorf("error while parsing %s: %s", info.Name(), err)
			}
			*migrations = append(*migrations, migration)

		}
	}

	return nil
}

type byID []*migrate.Migration

func (b byID) Len() int           { return len(b) }
func (b byID) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }
func (b byID) Less(i, j int) bool { return b[i].Less(b[j]) }
