package delivery

import (
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/mao360/musicLib/models"
	"github.com/mao360/musicLib/pkg/service"
	"github.com/sirupsen/logrus"
	"net/http"
)

type HandlerInterface interface {
	GetSongsByFilter(c echo.Context) error
	GetText(c echo.Context) error
	DeleteSong(c echo.Context) error
	ChangeSong(c echo.Context) error
	AddSong(c echo.Context) error
}

type Handler struct {
	service               service.ServiceInterface
	externalServiceDomain string
	logger                *logrus.Logger
}

func NewHandler(service service.ServiceInterface, logger *logrus.Logger, domain string) *Handler {
	return &Handler{service, domain, logger}
}

// @Summary Get all song
// @Description Get all songs, use filters
// @Produce json
// @Param page query string true "The page query parameter (required)"
// @Param pageSize query string true "The pageSize query parameter (required)"
// @Param group_name query string false "The group_name query parameter (optional)"
// @Param song_name query string false "The song_name query parameter (optional)"
// @Param song_text query string false "The song_text query parameter (optional)"
// @Param release_date query string false "The release_date query parameter (optional)"
// @Param link query string false "The link query parameter (optional)"
// @Sucess 200 {object} []models.Song
// @Failure 500 {object} error
// @Router /songs [get]
func (h *Handler) GetSongsByFilter(c echo.Context) error {
	h.logger.WithFields(logrus.Fields{
		"handler": "GetSongsByFilter",
	}).Infof("started")

	pageNum := c.QueryParam("page")
	pageSize := c.QueryParam("pageSize")

	group := c.QueryParam("group_name")
	song := c.QueryParam("song_name")
	text := c.QueryParam("song_text")
	releaseDate := c.QueryParam("release_date")
	link := c.QueryParam("link")

	filters := map[string]string{
		"group_name":   group,
		"song_name":    song,
		"song_text":    text,
		"release_date": releaseDate,
		"link":         link,
	}
	h.logger.Debugf("filters=%v, pageSize=%s, pageNum=%s", filters, pageSize, pageNum)
	songs, err := h.service.GetSongsByFilter(filters, pageSize, pageNum)
	if err != nil {
		h.logger.WithFields(logrus.Fields{
			"layer":    "delivery",
			"handler":  "GetSongsByFilter",
			"function": "service.GetSongsByFilter",
		}).Errorf("err: %v", err)
		return c.String(500, err.Error())
	}
	h.logger.Debugf("songs: %v", songs)
	if err = c.JSON(200, songs); err != nil {
		h.logger.WithFields(logrus.Fields{
			"layer":    "delivery",
			"handler":  "GetSongsByFilter",
			"function": "c.JSON",
		}).Errorf("err: %v", err)
		return c.String(500, err.Error())
	}
	h.logger.WithFields(logrus.Fields{
		"handler": "GetSongsByFilter",
	}).Infof("finished")
	return c.String(200, "ok")
}

// @Summary Get text
// @Description Get song text with pagination
// @Produce json
// @Param page query string true "The page query parameter (required)"
// @Param pageSize query string true "The pageSize query parameter (required)"
// @Param id path string true "ID"
// @Sucess 200 {object} string
// @Failure 500 {object} error
// @Router /song/{id} [get]
func (h *Handler) GetText(c echo.Context) error {
	h.logger.WithFields(logrus.Fields{
		"handler": "GetText",
	}).Infof("started")

	pageNum := c.QueryParam("page")
	pageSize := c.QueryParam("pageSize")
	songID := c.Param("id")

	h.logger.Debugf("songID=%s pageSize=%s pageNum=%s", songID, pageSize, pageNum)
	text, err := h.service.GetTextByID(songID, pageSize, pageNum)
	if err != nil {
		h.logger.WithFields(logrus.Fields{
			"layer":    "delivery",
			"handler":  "GetText",
			"function": "service.SearchSongByID",
		}).Errorf("err: %v", err)
		return c.String(500, err.Error())
	}
	h.logger.Debugf("text: %s", text)
	if err = c.JSON(200, text); err != nil {
		h.logger.WithFields(logrus.Fields{
			"layer":    "delivery",
			"handler":  "GetText",
			"function": "c.JSON",
		}).Errorf("err: %v", err)
		return c.String(500, err.Error())
	}
	h.logger.WithFields(logrus.Fields{
		"handler": "GetText",
	}).Infof("finished")
	return c.String(200, "ok")
}

// @Summary Delete Song
// @Description Delete song from db
// @Param id path string true "ID"
// @Sucess 200 {object} string
// @Failure 500 {object} error
// @Router /song/{id} [delete]
func (h *Handler) DeleteSong(c echo.Context) error {
	h.logger.WithFields(logrus.Fields{
		"handler": "DeleteSong",
	}).Infof("started")
	songId := c.Param("id")
	h.logger.Debugf("songId=%s", songId)
	if err := h.service.DeleteSongByID(songId); err != nil {
		h.logger.WithFields(logrus.Fields{
			"layer":    "delivery",
			"handler":  "DeleteSong",
			"function": "service.DeleteSongByID",
		}).Errorf("err: %v", err)
		return c.String(500, err.Error())
	}
	h.logger.WithFields(logrus.Fields{
		"handler": "DeleteSong",
	}).Infof("finished")
	return c.String(200, "ok")
}

// @Summary Delete Song
// @Description Delete song from db
// @Param requestBody body models.Song true "JSON payload for creating a resource"
// @Accept json
// @Param id path string true "ID"
// @Sucess 200 string
// @Failure 500 {object} error
// @Router /song/{id} [delete]
func (h *Handler) ChangeSong(c echo.Context) error {
	h.logger.WithFields(logrus.Fields{
		"handler": "ChangeSong",
	}).Infof("started")
	song := models.Song{}
	if err := c.Bind(&song); err != nil {
		h.logger.WithFields(logrus.Fields{
			"layer":    "delivery",
			"handler":  "ChangeSong",
			"function": "c.Bind",
		}).Errorf("err: %v", err)
		return c.String(500, err.Error())
	}
	songID := c.Param("id")
	h.logger.Debugf("songID=%s song=%s group=%s link=%s release=%s text=%s", songID, song.Song, song.Group, song.Link, song.ReleaseDate, song.Text)
	if err := h.service.ChangeSongByID(songID, song); err != nil {
		h.logger.WithFields(logrus.Fields{
			"layer":    "delivery",
			"handler":  "ChangeSong",
			"function": "service.ChangeSongByID",
		}).Errorf("err: %v", err)
		return c.String(500, err.Error())
	}
	h.logger.WithFields(logrus.Fields{
		"handler": "ChangeSong",
	}).Infof("finished")
	return c.String(200, "ok")
}

// @Summary Add Song
// @Description Add Song to db
// @Accept json
// @Param requestBody body models.Song true "JSON payload for creating a resource"
// @Sucess 200 {object} []models.Song
// @Failure 400 {object} error
// @Failure 500 {object} error
// @Router /song [post]
func (h *Handler) AddSong(c echo.Context) error {
	h.logger.WithFields(logrus.Fields{
		"handler": "AddSong",
	}).Infof("started")
	h.logger.Infof("handler AddSong started")
	song := models.Song{}
	if err := c.Bind(&song); err != nil {
		h.logger.WithFields(logrus.Fields{
			"layer":    "delivery",
			"handler":  "AddSong",
			"function": "c.Bind",
		}).Errorf("err: %v", err)
		return c.String(500, err.Error())
	}
	h.logger.Debugf("addSong data: song=%s text=%s\n", song.Song, song.Text)
	outputSong, statusCode, err := SendRequestToExternalService(song, h.externalServiceDomain)
	if statusCode != 200 {
		if statusCode == 400 {
			h.logger.WithFields(logrus.Fields{
				"layer":    "delivery",
				"handler":  "AddSong",
				"function": "SendRequestToExternalService",
			}).Errorf("400")
			return c.String(statusCode, "external service 400")
		} else {
			h.logger.WithFields(logrus.Fields{
				"layer":    "delivery",
				"handler":  "AddSong",
				"function": "SendRequestToExternalService",
			}).Errorf("500")
			return c.String(statusCode, "external service 500")
		}
	}
	h.logger.Debugf("outputSong: song=%s group=%s release=%s link=%s text=%s", outputSong.Song, outputSong.Group, outputSong.ReleaseDate, outputSong.Link, outputSong.Text)
	if err = h.service.AddSong(outputSong); err != nil {
		h.logger.WithFields(logrus.Fields{
			"layer":    "delivery",
			"handler":  "AddSong",
			"function": "service.AddSong",
		}).Errorf("err: %v", err)
		return c.String(500, err.Error())
	}
	h.logger.WithFields(logrus.Fields{
		"handler": "AddSong",
	}).Infof("finished")
	return c.String(200, "ok")
}

func SendRequestToExternalService(input models.Song, externalServiceDomain string) (models.Song, int, error) {
	url := fmt.Sprintf("%s/info?group=%s&song=%s", externalServiceDomain, input.Group, input.Song)
	resp, err := http.Get(url)
	defer resp.Body.Close()
	song := models.Song{}
	err = json.NewDecoder(resp.Body).Decode(&song)
	if err != nil {
		return models.Song{}, 0, err
	}
	return song, resp.StatusCode, nil
}
