package migrations

import (
	"context"
	"database/sql"
	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(Up, Down)
}

func Up(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx,
		`CREATE TABLE songs (
    id SERIAL PRIMARY KEY ,
    group_name VARCHAR(255),
    song_name VARCHAR(255),
    song_text TEXT,
	release_date VARCHAR(10),
	link VARCHAR(255)
	);
	CREATE INDEX idx_group_name ON songs (group_name);
	CREATE INDEX idx_song_name ON songs (song_name);
	CREATE INDEX idx_release_date ON songs (release_date);
	CREATE INDEX idx_link ON songs (link);
	CREATE INDEX idx_song_text ON songs (song_text);`)
	if err != nil {
		return err
	}
	return nil
}

func Down(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, `DROP TABLE IF EXISTS songs;`)
	if err != nil {
		return err
	}
	return nil
}
