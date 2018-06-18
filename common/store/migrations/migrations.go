package migrations

import (
	"github.com/mattes/migrate"
	_ "github.com/mattes/migrate/database/postgres"
	_ "github.com/mattes/migrate/source/file"
)

func Up(options ApplyOptions) (res ApplyResult) {
	var m *migrate.Migrate
	m, res.Err = migrate.New(options.SourceURL, options.DatabaseURL)
	if res.Err != nil {
		return
	}

	if err := m.Up(); err != nil {
		if err == migrate.ErrNoChange {
			return
		}
		res.Err = err
		return
	}

	res.Changes = true
	return
}

type ApplyOptions struct {
	SourceURL   string
	DatabaseURL string
}

type ApplyResult struct {
	Err     error
	Changes bool
}
