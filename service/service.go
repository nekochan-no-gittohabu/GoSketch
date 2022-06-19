package service

import (
	"context"
	"errors"

	"personal/gosketch/api"
	"personal/gosketch/domain"

	"github.com/google/uuid"
)

type Service struct {
	datastore Datastore
	image     ImageService
}

var (
	ErrNoSession      = errors.New("session not found")
	ErrSessionExpired = errors.New("session ended")
)

type Datastore interface {
	List(ctx context.Context) ([]domain.Session, error)
	SaveSession(ctx context.Context, s domain.Session) (domain.Session, error)
	Get(ctx context.Context, id uuid.UUID) (domain.Session, error)
	Delete(ctx context.Context, id uuid.UUID) error
	SavePhoto(ctx context.Context, s domain.Session, newPhoto domain.Photo) error
}

type ImageService interface {
	GetLinks(keyword string) ([]string, error)
}

func NewService(d Datastore, i ImageService) api.Service {
	return Service{
		datastore: d,
		image:     i,
	}
}

func (svc Service) ListSessions(ctx context.Context) ([]domain.Session, error) {
	return svc.datastore.List(ctx)
}

func (svc Service) CreateSession(ctx context.Context, s domain.Session) (domain.Session, error) {
	var err error
	s.Links, err = svc.image.GetLinks(s.Keyword) // Save links to the slice in the session
	if err != nil {
		return domain.Session{}, err
	}
	err = s.SetSession()
	if err != nil {
		return domain.Session{}, err
	}
	return svc.datastore.SaveSession(ctx, s)
}

func (svc Service) GetSession(ctx context.Context, id uuid.UUID) (domain.Session, error) {
	return svc.datastore.Get(ctx, id)
}

func (svc Service) SkipPhoto(ctx context.Context, id uuid.UUID, skip bool) (domain.Photo, error) {
	s, err := svc.datastore.Get(ctx, id)
	if err != nil {
		return domain.Photo{}, err
	}

	ok, err := s.ShowNext(skip)
	if ok {
		p, err := s.NewPhoto()
		if err != nil {
			return domain.Photo{}, err
		}
		return p, svc.datastore.SavePhoto(ctx, s, p)
	}
	return *s.CurrentPhoto(), err
}

func (svc Service) DeleteSession(ctx context.Context, id uuid.UUID) error {
	return svc.datastore.Delete(ctx, id)
}
