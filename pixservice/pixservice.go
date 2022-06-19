package pixservice

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"
)

type pixabayClient struct {
	client    *http.Client
	key       string
	imageType string
}
type pixabayResp struct {
	Total     int
	TotalHits int
	Hits      []map[string]interface{}
}

func New(k, t string) *pixabayClient {
	return &pixabayClient{
		client: &http.Client{
			Timeout: 3 * time.Second,
		},
		key:       k,
		imageType: t,
	}
}

func (pix *pixabayClient) GetLinks(keyword string) ([]string, error) {
	// URL encoding
	baseUrl, err := url.Parse("https://pixabay.com/api")
	if err != nil {
		fmt.Println("Malformed URL: ", err.Error())
		return nil, err
	}

	params := url.Values{}
	params.Add("key", pix.key)
	params.Add("q", keyword)
	params.Add("image_type", pix.imageType)
	baseUrl.RawQuery = params.Encode()

	response, err := pix.client.Get(baseUrl.String())
	if err != nil {
		log.Fatal(err)
	}
	defer response.Body.Close()

	var p pixabayResp
	if err := json.NewDecoder(response.Body).Decode(&p); err != nil {
		log.Printf("Cannot parse response: %v", err.Error())
		return []string{}, err
	}

	var temp []string
	for _, e := range p.Hits {
		temp = append(temp, fmt.Sprintf("%v", e["largeImageURL"]))
	}
	return temp, nil
}
