package main

import (
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	_ "github.com/mao360/musicLib/migrations"
	"github.com/mao360/musicLib/pkg/delivery"
	"github.com/mao360/musicLib/pkg/postgres"
	"github.com/mao360/musicLib/pkg/service"
	"github.com/sirupsen/logrus"
	"os"
	"strconv"
)

// @title Music Lib App API
// @version 1.0
// @description API server for Music Lib App
// @BasePath /
func main() {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	logger.SetOutput(os.Stdout)

	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	err := godotenv.Load()
	if err != nil {
		logger.Fatal("error loading .env file")
	}
	connURL := os.Getenv("CONN_URL")
	servicePort := os.Getenv("SERVICE_PORT")
	externalServiceDomain := os.Getenv("EXTERNAL_SERVICE_DOMAIN")
	reloadMigration, err := strconv.ParseBool(os.Getenv("RELOAD_MIGRATION"))
	if err != nil {
		logger.Fatal("error parsing RELOAD_MIGRATION")
	}

	conn, err := postgres.ConnectToDB(connURL, reloadMigration, logger)
	defer conn.Close()
	if err != nil {
		logger.Fatalf("can`t connect to database: %v", err)
	}
	db := postgres.NewDB(conn, logger)
	s := service.NewService(db, logger)
	h := delivery.NewHandler(s, logger, externalServiceDomain)
	e := echo.New()

	e.GET("/songs", h.GetSongsByFilter)
	e.GET("/song/:id", h.GetText)
	e.DELETE("/song/:id", h.DeleteSong)
	e.PUT("/song/:id", h.ChangeSong)
	e.POST("/song", h.AddSong)

	err = e.Start(servicePort)
	if err != nil {
		logger.Fatalf("failed to sarat server %v", err)
	}
}
