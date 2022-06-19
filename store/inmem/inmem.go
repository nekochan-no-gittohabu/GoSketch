package inmem

import (
	"context"
	"sync"

	"personal/gosketch/domain"
	"personal/gosketch/service"

	"github.com/google/uuid"
)

type InmemDatastore struct {
	sessionList []domain.Session
	mu          sync.Mutex
}

func New() *InmemDatastore {
	return &InmemDatastore{
		sessionList: []domain.Session{},
	}
}

func (ids *InmemDatastore) List(ctx context.Context) ([]domain.Session, error) {
	return ids.sessionList, nil
}

func (ids *InmemDatastore) SaveSession(ctx context.Context, s domain.Session) (domain.Session, error) {
	ids.mu.Lock()
	defer ids.mu.Unlock()

	ids.sessionList = append(ids.sessionList, s)
	return s, nil
}

func (ids *InmemDatastore) Get(ctx context.Context, id uuid.UUID) (domain.Session, error) {
	ids.mu.Lock()
	defer ids.mu.Unlock()

	for i := range ids.sessionList {
		if id == ids.sessionList[i].ID {
			return ids.sessionList[i], nil
		}
	}
	return domain.Session{}, service.ErrNoSession
}

func (ids *InmemDatastore) SavePhoto(ctx context.Context, s domain.Session, newPhoto domain.Photo) error {
	ids.mu.Lock()
	defer ids.mu.Unlock()

	for i := range ids.sessionList {
		if s.ID == ids.sessionList[i].ID {
			(&ids.sessionList[i]).CurrentPhoto().UpdateDuration()
			(&ids.sessionList[i]).Photos = append((&ids.sessionList[i]).Photos, newPhoto)
			return nil
		}
	}
	return service.ErrNoSession
}

func (ids *InmemDatastore) Delete(ctx context.Context, id uuid.UUID) error {
	ids.mu.Lock()
	defer ids.mu.Unlock()

	for i := range ids.sessionList {
		if id == ids.sessionList[i].ID {
			ids.sessionList = append(ids.sessionList[:i], ids.sessionList[i+1:]...)
			return nil
		}
	}
	return service.ErrNoSession
}
