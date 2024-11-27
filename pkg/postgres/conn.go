package postgres

import (
	"context"
	"database/sql"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/sirupsen/logrus"
	"time"
)

func ConnectToDB(url string, reload bool, logger *logrus.Logger) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(context.Background(), url)
	if err != nil {
		logger.Fatalf("Error creating a pool conn: %s", err)
		return nil, err
	}
	for i := 0; i < 5; i++ {
		time.Sleep(time.Second)
		err = pool.Ping(context.Background())
		if err == nil {
			break
		}
	}
	if err != nil {
		logger.Fatalf("Error connecting to database: %s", err)
		return nil, err
	}

	db, err := sql.Open("pgx", url)
	if err != nil {
		logger.Fatalf("Error to open sql connection: %s", err)
		return nil, err
	}

	if reload {
		err = goose.DownTo(db, ".", 0)
		if err != nil {
			logger.Fatalf("Error migration down %s", err)
			return nil, err
		}
	}
	err = goose.Up(db, ".")
	if err != nil {
		logger.Fatalf("Error migration up %s", err)
		return nil, err
	}
	logger.Infof("Successful connection to db with all migrations")
	return pool, nil
}
