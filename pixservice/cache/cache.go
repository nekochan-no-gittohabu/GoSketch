package cache

import (
	"errors"
	"log"
	"time"

	"personal/gosketch/service"
)

type cache struct {
	timeout time.Duration
	links   map[string][]string
	touts   map[string]time.Time
	service.ImageService
}

func New(t time.Duration, i service.ImageService) *cache {
	return &cache{
		timeout:      t,
		links:        make(map[string][]string),
		touts:        make(map[string]time.Time),
		ImageService: i,
	}
}

func (c *cache) GetLinks(keyword string) ([]string, error) {
	if c.timeout < time.Since(c.touts[keyword]) {
		delete(c.touts, keyword)
		delete(c.links, keyword)
	}
	links, err := c.GetCache(keyword)
	if err != nil {
		log.Println(err)
		slinks, err := c.SetCache(keyword)
		if err != nil {
			return nil, err
		}
		return slinks, nil
	}
	return links, nil
}

func (c *cache) GetCache(keyword string) ([]string, error) {
	if links := c.links[keyword]; links != nil {
		log.Println("getting images from the cache")
		return links, nil
	}
	return nil, errors.New("not found in the cache")
}

func (c *cache) SetCache(keyword string) ([]string, error) {
	var err error
	c.links[keyword], err = c.ImageService.GetLinks(keyword)
	c.touts[keyword] = time.Now()
	log.Println("getting images from the pixservice")
	return c.links[keyword], err
}
