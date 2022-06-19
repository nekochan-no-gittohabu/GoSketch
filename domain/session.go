package domain

import (
	"errors"
	"math/rand"
	"time"

	"github.com/google/uuid"
)

type Session struct {
	ID         uuid.UUID `json:"id"`
	Keyword    string    `json:"keyword"`
	Period     int       `json:"period"`
	PhotoCount int       `json:"photoCount"`
	Photos     []Photo   `json:"photos"`
	CreatedAt  time.Time `json:"createdAt"`
	Links      []string  `json:"-"`
}

func (s *Session) SetSession() error {
	s.ID = uuid.New()
	p, err := s.NewPhoto()
	if err != nil {
		return err
	}
	s.Photos = []Photo{
		p,
	}
	s.CreatedAt = time.Now()
	return nil
}

func (s *Session) CurrentPhoto() *Photo {
	if len(s.Photos) < 1 {
		return nil
	}
	return &s.Photos[len(s.Photos)-1]
}

func (s *Session) NewPhoto() (Photo, error) {
	if len(s.Links) <= 0 {
		return Photo{}, errors.New("no links available")
	}
	rn := rand.Intn(len(s.Links))
	link := s.Links[rn] // Get a random link from the list

	return Photo{
		Link:      link,
		Duration:  0,
		CreatedAt: time.Now(),
	}, nil
}

func (s *Session) ShowNext(skip bool) (bool, error) {
	if s.CurrentPhoto() == nil {
		return true, errors.New("photo not found")
	}

	dur := s.CurrentPhoto().GetCurrentDuration()

	if skip || dur >= s.Period {
		if s.PhotoCount == len(s.Photos) {
			// End Session
			return false, errors.New("session ended")
		} else {
			return true, nil
		}
	} else if !skip {
		s.CurrentPhoto().UpdateDuration()
		return false, nil
	}
	return true, nil
}
