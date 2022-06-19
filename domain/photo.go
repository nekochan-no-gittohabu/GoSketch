package domain

import (
	"time"
)

type Photo struct {
	Link      string    `json:"link"`
	Duration  int       `json:"dur"`
	CreatedAt time.Time `json:"createdAt"`
}

func (photo *Photo) UpdateDuration() {
	photo.Duration = photo.GetCurrentDuration()
}

func (photo *Photo) GetCurrentDuration() int {
	return int(time.Since(photo.CreatedAt).Seconds())
}
