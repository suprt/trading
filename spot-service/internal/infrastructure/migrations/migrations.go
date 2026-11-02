package migrations

import (
	"context"
	"database/sql"
	"embed"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

//go:embed *.sql
var FS embed.FS

func Up(ctx context.Context, dsn string) (err error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return err
	}
	defer func() {
		if errDb := db.Close(); errDb != nil {
			if err == nil {
				err = errDb
			}
		}

	}()

	goose.SetBaseFS(FS)

	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	if err := goose.UpContext(ctx, db, "."); err != nil {
		return err
	}
	return nil
}
