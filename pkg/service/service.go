package service

import (
	"errors"
	"fmt"
	"github.com/mao360/musicLib/models"
	"github.com/mao360/musicLib/pkg/postgres"
	"github.com/sirupsen/logrus"
	"net/url"
	"strconv"
	"strings"
	"time"
)

var (
	invalidDate     = errors.New("invalid date")
	invalidLink     = errors.New("invalid link")
	noText          = errors.New("no text at this page")
	invalidPage     = errors.New("invalid page")
	invalidPageSize = errors.New("invalid page size")
	invalidYear     = errors.New("invalid year")
)

type ServiceInterface interface {
	GetSongsByFilter(filters map[string]string, limit, offset string) ([]models.Song, error)
	GetTextByID(songID, limit, offset string) (string, error)
	DeleteSongByID(songID string) error
	ChangeSongByID(songID string, song models.Song) error
	AddSong(song models.Song) error
}

type Service struct {
	repo   postgres.DBInterface
	logger *logrus.Logger
}

func NewService(repo postgres.DBInterface, logger *logrus.Logger) *Service {
	return &Service{repo, logger}
}

func (s *Service) GetSongsByFilter(filters map[string]string, pageSize, page string) ([]models.Song, error) {
	s.logger.Debugf("pageSize=%s, page=%s", pageSize, page)
	offset, err := strconv.Atoi(page)
	if err != nil || offset <= 0 {
		s.logger.WithFields(logrus.Fields{
			"layer":       "service",
			"function":    "GetSongsByFilter",
			"subFunction": "Atoi() and page > 0",
		}).Errorf("error: %s", invalidPage.Error())
		return nil, invalidPage
	}
	limit, err := strconv.Atoi(pageSize)
	if err != nil || limit <= 0 {
		s.logger.WithFields(logrus.Fields{
			"layer":       "service",
			"function":    "GetSongsByFilter",
			"subFunction": "Atoi() and pageSize > 0",
		}).Errorf("error: %s", invalidPageSize.Error())
		return nil, invalidPageSize
	}
	offset = (offset - 1) * limit

	for k, v := range filters {
		if v == "" {
			delete(filters, k)
		}
	}
	s.logger.Debugf("len(filters)=%d", len(filters))

	if year, ok := filters["release_date"]; ok {
		yearInt, err := strconv.Atoi(year)
		if err != nil || yearInt > time.Now().Year() {
			s.logger.WithFields(logrus.Fields{
				"layer":    "service",
				"function": "GetSongsByFilter",
			}).Errorf("error: %s", invalidYear.Error())
			return nil, invalidYear
		}
	}

	if link, ok := filters["link"]; ok && !CheckLink(link) {
		s.logger.WithFields(logrus.Fields{
			"layer":    "service",
			"function": "GetSongsByFilter",
		}).Errorf("error: %s", invalidLink.Error())
		return nil, invalidLink
	}
	s.logger.Debugf("len(filters)=%d limit=%d offset=%d", len(filters), limit, offset)
	songs, err := s.repo.GetSongsByFilterDB(filters, limit, offset)
	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"layer":       "service",
			"function":    "GetSongsByFilter",
			"subFunction": "GetSongsByFilterDB",
		}).Errorf("error: %s", err.Error())
		return nil, nil
	}
	s.logger.Debugf("len(songs)=%d", len(songs))
	return songs, nil
}

func (s *Service) GetTextByID(songID, pageSize, page string) (string, error) {
	pageSizeInt, err := strconv.Atoi(pageSize)
	if err != nil || pageSizeInt <= 0 {
		s.logger.WithFields(logrus.Fields{
			"layer":       "service",
			"function":    "GetTextByID",
			"subFunction": "Atoi() and pageSize > 0",
		}).Errorf("error: %s", invalidPageSize.Error())
		return "", invalidPageSize
	}
	pageInt, err := strconv.Atoi(page)
	if err != nil || pageInt <= 0 {
		s.logger.WithFields(logrus.Fields{
			"layer":       "service",
			"function":    "GetTextByID",
			"subFunction": "Atoi() and page > 0",
		}).Errorf("error: %s", invalidPage.Error())
		return "", invalidPage
	}

	id, err := strconv.Atoi(songID)
	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"layer":       "service",
			"function":    "GetTextByID",
			"subFunction": "Atoi() songID",
		}).Errorf("error: %s", err.Error())
		return "", err
	}
	s.logger.Debugf("id=%d size=%d page=%d", id, pageSizeInt, pageInt)
	song, err := s.repo.SearchSongByIDDB(id)
	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"layer":       "service",
			"function":    "GetTextByID",
			"subFunction": "SearchSongByIDDB",
		}).Errorf("error: %s", err.Error())
		return "", err
	}
	s.logger.Debugf("song=%v pageSize=%d page=%d", song, pageSizeInt, pageInt)
	text := GetTextByPage(song.Text, pageSizeInt, pageInt)
	s.logger.Debugf("len(text)=%d", len(text))
	if len(text) == 0 {
		s.logger.WithFields(logrus.Fields{
			"layer":       "service",
			"function":    "GetTextByID",
			"subFunction": "GetTextByPage",
		}).Errorf("error: %s", noText.Error())
		return "", noText
	}
	return text, nil
}

func (s *Service) DeleteSongByID(songID string) error {
	id, err := strconv.Atoi(songID)
	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"layer":       "service",
			"function":    "DeleteSongByID",
			"subFunction": "Atoi() songID",
		}).Errorf("error: %s", err.Error())
		return err
	}
	s.logger.Debugf("id=%d", id)
	if err = s.repo.DeleteSongByIDDB(id); err != nil {
		s.logger.WithFields(logrus.Fields{
			"layer":       "service",
			"function":    "DeleteSongByID",
			"subFunction": "DeleteSongByIDDB",
		}).Errorf("error: %s", err.Error())
		return err
	}
	return nil
}

func (s *Service) ChangeSongByID(songID string, song models.Song) error {
	id, err := strconv.Atoi(songID)
	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"layer":       "service",
			"function":    "ChangeSongByID",
			"subFunction": "Atoi() songID",
		}).Errorf("error: %s", err.Error())
		return err
	}
	if !CheckDate(song.ReleaseDate) {
		s.logger.WithFields(logrus.Fields{
			"layer":       "service",
			"function":    "ChangeSongByID",
			"subFunction": "CheckDate",
		}).Errorf("error: %s", invalidDate.Error())
		return invalidDate
	}
	if !CheckLink(song.Link) {
		s.logger.WithFields(logrus.Fields{
			"layer":       "service",
			"function":    "ChangeSongByID",
			"subFunction": "CheckLink",
		}).Errorf("error: %s", invalidLink.Error())
		return invalidLink
	}
	s.logger.Debugf("id=%d song=%s group=%s link=%s release=%s", id, song.Song, song.Group, song.Link, song.ReleaseDate)
	if err = s.repo.ChangeSongByIDDB(id, song); err != nil {
		s.logger.WithFields(logrus.Fields{
			"layer":       "service",
			"function":    "ChangeSongByID",
			"subFunction": "ChangeSongByIDDB",
		}).Errorf("error: %s", err.Error())
		return err
	}
	return nil
}

func (s *Service) AddSong(song models.Song) error {
	if !CheckDate(song.ReleaseDate) {
		s.logger.WithFields(logrus.Fields{
			"layer":       "service",
			"function":    "AddSong",
			"subFunction": "CheckDate",
		}).Errorf("error: %s", invalidDate.Error())
		return invalidDate
	}
	if !CheckLink(song.Link) {
		s.logger.WithFields(logrus.Fields{
			"layer":       "service",
			"function":    "AddSong",
			"subFunction": "CheckLink",
		}).Errorf("error: %s", invalidLink.Error())
		return invalidLink
	}
	s.logger.Debugf("song=%s group=%s text=%s link=%s release=%s", song.Song, song.Group, song.Text, song.Link, song.ReleaseDate)
	if err := s.repo.AddSongDB(song); err != nil {
		s.logger.WithFields(logrus.Fields{
			"layer":       "service",
			"function":    "AddSong",
			"subFunction": "AddSongDB",
		}).Errorf("error: %s", err.Error())
		return err
	}
	return nil
}

func CheckDate(date string) bool {
	splitDate := strings.Split(date, ".")
	if len(splitDate) != 3 {
		return false
	}
	_, err := time.Parse(time.DateOnly, fmt.Sprintf("%s-%s-%s", splitDate[2], splitDate[1], splitDate[0]))
	if err != nil {
		return false
	}
	return true
}

func CheckLink(link string) bool {
	parseURL, err := url.Parse(link)
	if err != nil {
		return false
	}
	return parseURL.Scheme != "" && parseURL.Host != ""
}

func GetTextByPage(text string, pageSize, page int) string {
	verses := 0
	result := make([]byte, 0)
	textByte := []byte(text)
	for i := 0; i < len(textByte)-1; i++ {
		if verses >= pageSize*(page-1) && verses < page*pageSize {
			result = append(result, textByte[i])
		}
		if textByte[i] == '\n' {
			if textByte[i+1] == '\n' {
				verses++
			}
		}
	}
	return string(result)
}
