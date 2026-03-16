package mocks

import (
	"github.com/mkailbowdy/internal/models"
	"time"
)

var mockSite = models.Site{
	ID:        1,
	Url:       "www.example.com",
	Created:   time.Now(),
	Urlhash:   "iiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiii",
	Pagehash:  "iiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiii",
	Selector:  "body",
	Changed:   false,
	UpdatedAt: time.Now(),
}

type SiteModel struct{}

func (m *SiteModel) Insert(url, urlhash, pagehash, selector string) (int, error) {
	return 2, nil
}

func (m *SiteModel) Get(url string) (models.Site, error) {
	switch url {
	case "www.example.com":
		return mockSite, nil
	default:
		return models.Site{}, models.ErrNoRecord
	}
}

func (m *SiteModel) GetAll() ([]models.Site, error) {
	return []models.Site{}, nil
}

func (m *SiteModel) MarkAsChanged(urlhash string) error {
	return nil
}
func (m *SiteModel) Update(urlhash, pagehash string) error {
	return nil
}
