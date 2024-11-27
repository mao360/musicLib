package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mao360/musicLib/models"
	"github.com/sirupsen/logrus"
	"strings"
)

var (
	songAlreadyInDB = errors.New("song already in db")
	songNotFound    = errors.New("song not found")
	changedColumns  = errors.New("unexpected number of changed columns")
)

type DBInterface interface {
	GetSongsByFilterDB(filters map[string]string, limit, offset int) ([]models.Song, error)
	SearchSongByIDDB(id int) (models.Song, error)
	DeleteSongByIDDB(id int) error
	ChangeSongByIDDB(id int, song models.Song) error
	AddSongDB(song models.Song) error
}

type DB struct {
	conn   *pgxpool.Pool
	logger *logrus.Logger
}

func NewDB(conn *pgxpool.Pool, logger *logrus.Logger) *DB {
	return &DB{conn, logger}
}

func (db *DB) GetSongsByFilterDB(filters map[string]string, limit, offset int) ([]models.Song, error) {
	db.logger.Debugf("len(filters)=%d limit=%d offset=%d", len(filters), limit, offset)
	query := "SELECT group_name, song_name, song_text, release_date, link FROM songs"
	values := make([]interface{}, 0)
	placeholderNum := 1
	if len(filters) != 0 {
		query += ` WHERE `
		for k, v := range filters {
			if k == "release_date" {
				query += fmt.Sprintf("%s LIKE '%%%s' AND ", k, v)
				continue
			} else {
				query += fmt.Sprintf("%s=$%d AND ", k, placeholderNum)
			}
			placeholderNum++
			values = append(values, v)
		}
		query = strings.TrimSuffix(query, " AND ")
	}

	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d;", placeholderNum, placeholderNum+1)
	values = append(values, limit, offset)

	db.logger.Debugf("SQL query: %s", query)
	rows, err := db.conn.Query(context.Background(), query, values...)

	defer rows.Close()
	if err != nil {
		db.logger.WithFields(logrus.Fields{
			"layer":       "db",
			"function":    "GetSongsByFilterDB",
			"subFunction": "Query()",
		}).Errorf("error: %s", err.Error())
		return nil, err
	}
	songs := make([]models.Song, 0)
	for rows.Next() {
		song := models.Song{}
		err = rows.Scan(&song.Group, &song.Song, &song.Text, &song.ReleaseDate, &song.Link)
		if err != nil {
			db.logger.WithFields(logrus.Fields{
				"layer":       "db",
				"function":    "GetSongsByFilterDB",
				"subFunction": "Scan()",
			}).Errorf("error: %s", err.Error())
			return nil, err
		}
		songs = append(songs, song)
	}
	db.logger.Debugf("len of songs list=%d", len(songs))
	return songs, nil
}

func (db *DB) SearchSongByIDDB(id int) (models.Song, error) {
	db.logger.Debugf("search id=%d", id)
	rows, err := db.conn.Query(context.Background(),
		`SELECT group_name, song_name, song_text, release_date, link
	FROM songs
	WHERE id=$1;`, id)
	defer rows.Close()
	if err != nil {
		db.logger.WithFields(logrus.Fields{
			"layer":       "db",
			"function":    "SearchSongByIDDB",
			"subFunction": "Query()",
		}).Errorf("error: %s", err.Error())
		return models.Song{}, err
	}
	if !rows.Next() {
		db.logger.WithFields(logrus.Fields{
			"layer":    "db",
			"function": "SearchSongByIDDB",
		}).Errorf("error: %s", songNotFound)
		return models.Song{}, songNotFound
	}
	song := models.Song{}
	err = rows.Scan(&song.Group, &song.Song, &song.Text, &song.ReleaseDate, &song.Link)
	if err != nil {
		db.logger.WithFields(logrus.Fields{
			"layer":       "db",
			"function":    "SearchSongByIDDB",
			"subFunction": "Scan()",
		}).Errorf("error: %s", err.Error())
		return models.Song{}, err
	}
	db.logger.Debugf("searched song: %s %s %s", song.Group, song.Song, song.Text)
	return song, nil
}

func (db *DB) DeleteSongByIDDB(id int) error {
	db.logger.Debugf("searched id=%d", id)
	rows, err := db.conn.Query(context.Background(),
		`SELECT id FROM songs WHERE id=$1;`, id)
	if err != nil {
		db.logger.WithFields(logrus.Fields{
			"layer":       "db",
			"function":    "DeleteSongByIDDB",
			"subFunction": "Query()",
		}).Errorf("error: %s", err.Error())
		return err
	}
	defer rows.Close()
	if !rows.Next() {
		db.logger.WithFields(logrus.Fields{
			"layer":    "db",
			"function": "DeleteSongByIDDB",
		}).Errorf("error: %s", songNotFound.Error())
		return songNotFound
	}
	commandTag, err := db.conn.Exec(context.Background(),
		`DELETE FROM songs
	WHERE id=$1;`, id)
	if err != nil {
		db.logger.WithFields(logrus.Fields{
			"layer":       "db",
			"function":    "DeleteSongByIDDB",
			"subFunction": "Exec()",
		}).Errorf("error: %s", err.Error())
		return err
	}
	db.logger.Debugf("rows affected=%d", commandTag.RowsAffected())
	if commandTag.RowsAffected() != 1 {
		db.logger.WithFields(logrus.Fields{
			"layer":    "db",
			"function": "DeleteSongByIDDB",
		}).Errorf("error: %s", changedColumns.Error())
		return changedColumns
	}
	return nil
}

func (db *DB) ChangeSongByIDDB(id int, song models.Song) error {
	rows, err := db.conn.Query(context.Background(),
		`SELECT id FROM songs WHERE id=$1;`, id)
	if err != nil {
		db.logger.WithFields(logrus.Fields{
			"layer":       "db",
			"function":    "ChangeSongByIDDB",
			"subFunction": "Query()",
		}).Errorf("error: %s", err.Error())
		return err
	}
	defer rows.Close()
	if !rows.Next() {
		db.logger.WithFields(logrus.Fields{
			"layer":    "db",
			"function": "ChangeSongByIDDB",
		}).Errorf("error: %s", songNotFound.Error())
		return songNotFound
	}

	commandTag, err := db.conn.Exec(context.Background(),
		`UPDATE songs
	SET group_name=$1,
	    song_name=$2,
	    song_text=$3,
	    release_date=$4,
	    link=$5
	WHERE id=$6;`,
		song.Group, song.Song, song.Text, song.ReleaseDate, song.Link, id)
	if err != nil {
		db.logger.WithFields(logrus.Fields{
			"layer":       "db",
			"function":    "ChangeSongByIDDB",
			"subFunction": "Exec()",
		}).Errorf("error: %s", err.Error())
		return err
	}
	db.logger.Debugf("rows affected=%d", commandTag.RowsAffected())
	if commandTag.RowsAffected() != 1 {
		db.logger.WithFields(logrus.Fields{
			"layer":    "db",
			"function": "ChangeSongByIDDB",
		}).Errorf("error: %s", changedColumns.Error())
		return changedColumns
	}
	return nil
}

func (db *DB) AddSongDB(song models.Song) error {
	db.logger.Debugf("changing to data: song=%s group=%s release=%s link=%s text=%s", song.Song, song.Group, song.ReleaseDate, song.Link, song.Text)
	rows, err := db.conn.Query(context.Background(),
		`SELECT id FROM songs WHERE song_name=$1 AND group_name=$2;`, song.Song, song.Group)
	if err != nil {
		db.logger.WithFields(logrus.Fields{
			"layer":       "db",
			"function":    "AddSongDB",
			"subFunction": "Query()",
		}).Errorf("error: %s", err.Error())
		return err
	}
	defer rows.Close()
	if rows.Next() {
		db.logger.WithFields(logrus.Fields{
			"layer":    "db",
			"function": "AddSongDB",
		}).Errorf("error: %s", songAlreadyInDB.Error())
		return songAlreadyInDB
	}

	_, err = db.conn.Exec(context.Background(),
		`INSERT INTO songs(group_name, song_name, song_text, release_date, link)
	VALUES ($1, $2, $3, $4, $5);`,
		song.Group, song.Song, song.Text, song.ReleaseDate, song.Link)
	if err != nil {
		db.logger.WithFields(logrus.Fields{
			"layer":       "db",
			"function":    "AddSongDB",
			"subFunction": "Exec()",
		}).Errorf("error: %s", err.Error())
		return err
	}
	return nil
}
