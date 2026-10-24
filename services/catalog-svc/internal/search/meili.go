package search

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type TitleSearchDocument struct {
	ID                string            `json:"id"`
	Slug              string            `json:"slug"`
	Type              string            `json:"type"`
	OriginalTitle     string            `json:"original_title"`
	LocalizedTitles   map[string]string `json:"localized_titles"`
	Synopsis          map[string]string `json:"synopsis"`
	ReleaseYear       int32             `json:"release_year"`
	MaturityRating    string            `json:"maturity_rating"`
	IsNetflixOriginal bool              `json:"is_netflix_original"`
}

type MeiliClient struct {
	baseURL    string
	masterKey  string
	httpClient *http.Client
}

func NewMeiliClient(baseURL, masterKey string) *MeiliClient {
	return &MeiliClient{
		baseURL:    baseURL,
		masterKey:  masterKey,
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}
}

func (m *MeiliClient) IndexTitleDocument(ctx context.Context, doc TitleSearchDocument) error {
	payload, err := json.Marshal([]TitleSearchDocument{doc})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, m.baseURL+"/indexes/titles/documents", bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+m.masterKey)
	resp, err := m.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("meilisearch index failed with status %d", resp.StatusCode)
	}
	return nil
}

type meiliSearchPayload struct {
	Q      string `json:"q"`
	Limit  int32  `json:"limit"`
	Offset int32  `json:"offset"`
}

type meiliSearchResult struct {
	Hits             []TitleSearchDocument `json:"hits"`
	EstimatedTotal   int32                 `json:"estimatedTotalHits"`
}

func (m *MeiliClient) Search(ctx context.Context, query string, limit, offset int32) ([]TitleSearchDocument, int32, error) {
	body, err := json.Marshal(meiliSearchPayload{
		Q:      query,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, 0, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, m.baseURL+"/indexes/titles/search", bytes.NewReader(body))
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+m.masterKey)
	resp, err := m.httpClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return nil, 0, fmt.Errorf("meilisearch search error status %d", resp.StatusCode)
	}
	var parsed meiliSearchResult
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, 0, err
	}
	return parsed.Hits, parsed.EstimatedTotal, nil
}

