package postgresdb

import (
	"context"
	"fmt"
	"personal/gosketch/domain"
	"personal/gosketch/service"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type PsqlDatastore struct {
	conpool *pgxpool.Pool
}

func New(s string) (*PsqlDatastore, error) {
	config, err := pgxpool.ParseConfig(s)
	if err != nil {
		return nil, err
	}
	pool, err := pgxpool.ConnectConfig(context.Background(), config)
	if err != nil {
		return nil, err
	}
	return &PsqlDatastore{
		conpool: pool,
	}, nil
}

func (p *PsqlDatastore) Close() {
	p.conpool.Close()
}

func (p *PsqlDatastore) List(ctx context.Context) ([]domain.Session, error) {
	rows, err := p.conpool.Query(ctx, "SELECT * FROM sessions")
	if err != nil {
		err = fmt.Errorf("query failed: %w", err)
		return nil, err
	}
	defer rows.Close()

	var sessionList []domain.Session
	for rows.Next() {
		tempSession, err := scanSession(rows)
		if err != nil {
			return nil, err
		}
		photoList, err := p.listPhotos(ctx, tempSession.ID)
		if err != nil {
			return nil, err
		}
		tempSession.Photos = photoList
		sessionList = append(sessionList, tempSession)
	}

	if rows.Err() != nil {
		return sessionList, rows.Err()
	}

	return sessionList, nil
}

func (p *PsqlDatastore) SaveSession(ctx context.Context, s domain.Session) (domain.Session, error) {
	conn, err := p.conpool.Begin(ctx)
	if err != nil {
		return domain.Session{}, err
	}
	defer func() {
		if err != nil {
			conn.Rollback(ctx)
		}
		err = conn.Commit(ctx)
		if err != nil {
			conn.Rollback(ctx)
		}
	}()

	sqlStatement := `INSERT INTO sessions (session_id, keyword, per, photo_count, created_at, links)
	VALUES ($1, $2, $3, $4, $5, $6)`
	_, err = conn.Exec(ctx, sqlStatement, s.ID, s.Keyword, s.Period, s.PhotoCount, s.CreatedAt, s.Links)
	if err != nil {
		err = fmt.Errorf("queryrow failed: %w", err)
		return domain.Session{}, err
	}

	sqlStatement = `INSERT INTO photos (session_id, link, dur , created_at) 
	VALUES  ($1, $2, $3, $4)`

	_, err = conn.Exec(ctx, sqlStatement, s.ID, s.CurrentPhoto().Link, s.CurrentPhoto().Duration, s.CurrentPhoto().CreatedAt)
	if err != nil {
		err = fmt.Errorf("savesession queryrow failed: %w", err)
		return domain.Session{}, err
	}

	return s, err
}

func (p *PsqlDatastore) Get(ctx context.Context, id uuid.UUID) (domain.Session, error) {
	sqlStatement := `SELECT * FROM sessions
	WHERE session_id = $1`

	r := p.conpool.QueryRow(ctx, sqlStatement, id)
	tempSession, err := scanSession(r)
	if err != nil {
		err = fmt.Errorf("get queryrow failed: %w", err)
		return domain.Session{}, err
	}
	photoList, err := p.listPhotos(ctx, id)
	if err != nil {
		return domain.Session{}, err
	}
	tempSession.Photos = photoList
	return tempSession, nil
}

func (p *PsqlDatastore) Delete(ctx context.Context, id uuid.UUID) error {
	sqlStatement := `DELETE FROM sessions
	WHERE session_id = $1`

	cTag, err := p.conpool.Exec(ctx, sqlStatement, id)
	if err != nil {
		err = fmt.Errorf("delete query failed: %w", err)
		return err
	}
	if cTag.RowsAffected() != 1 {
		return service.ErrNoSession
	}
	return nil
}

func (p *PsqlDatastore) SavePhoto(ctx context.Context, s domain.Session, newPhoto domain.Photo) error {
	conn, err := p.conpool.Begin(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			conn.Rollback(ctx)
		}
		err = conn.Commit(ctx)
		if err != nil {
			conn.Rollback(ctx)
		}
	}()

	sqlStatement := `UPDATE photos
		SET dur = $2
		WHERE session_id = $1
		AND created_at = $3`

	_, err = conn.Exec(ctx, sqlStatement, s.ID, s.CurrentPhoto().GetCurrentDuration(), s.CurrentPhoto().CreatedAt)
	if err != nil {
		err = fmt.Errorf("savephoto update query failed: %w", err)
		return err
	}

	sqlStatement = `INSERT INTO photos (session_id, link, dur , created_at) 
	VALUES  ($1, $2, $3, $4)`

	_, err = conn.Exec(ctx, sqlStatement, s.ID, newPhoto.Link, newPhoto.Duration, newPhoto.CreatedAt)
	if err != nil {
		err = fmt.Errorf("savephoto insert query failed: %w", err)
		return err
	}

	return err
}

func (p *PsqlDatastore) listPhotos(ctx context.Context, id uuid.UUID) ([]domain.Photo, error) {
	sqlStatement := `SELECT * FROM photos
		WHERE session_id = $1`

	rows, err := p.conpool.Query(ctx, sqlStatement, id)
	if err != nil {
		err = fmt.Errorf("listphotos query failed: %w", err)
		return nil, err
	}
	defer rows.Close()

	var photoList []domain.Photo

	for rows.Next() {
		tempPhoto, err := scanPhoto(rows)
		if err != nil {
			return nil, err
		}
		photoList = append(photoList, tempPhoto)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return photoList, nil
}

func scanSession(r pgx.Row) (domain.Session, error) {
	var (
		tempPeriod, tempPhotoCount int
		tempKeyword                string
		tempId                     uuid.UUID
		tempCreatedAt              time.Time
		tempLinks                  []string
	)

	err := r.Scan(&tempId, &tempKeyword, &tempPeriod, &tempPhotoCount, &tempCreatedAt, &tempLinks)

	return domain.Session{
		ID:         tempId,
		Keyword:    tempKeyword,
		Period:     tempPeriod,
		PhotoCount: tempPhotoCount,
		CreatedAt:  tempCreatedAt,
		Links:      tempLinks,
	}, err
}

func scanPhoto(r pgx.Row) (domain.Photo, error) {
	var (
		tempId        uuid.UUID
		tempDur       int
		tempLink      string
		tempCreatedAt time.Time
	)
	err := r.Scan(&tempLink, &tempDur, &tempCreatedAt, &tempId)
	return domain.Photo{
		Link:      tempLink,
		Duration:  tempDur,
		CreatedAt: tempCreatedAt,
	}, err
}
