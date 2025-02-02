package reports

import (
	"encoding/json"
	"fmt"
	"net/http"
)

const baseUrl = "https://botw-compendium.herokuapp.com/api/v3/compendium"

type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type LozClient struct {
	httpClient HttpClient
}

func NewClient(httpClient HttpClient) *LozClient {
	return &LozClient{
		httpClient: httpClient,
	}
}

type Monster struct {
	Name            string   `json:"name"`
	Id              int      `json:"id"`
	Category        string   `json:"category"`
	Description     string   `json:"description"`
	Image           string   `json:"image"`
	CommonLocations []string `json:"common_locations"`
	Drops           []string `json:"drops"`
	Dlc             bool     `json:"dlc"`
}

type GetMonstersResponse struct {
	Data []Monster `json:"data"`
}

func (c *LozClient) GetMonsters() (*GetMonstersResponse, error) {
	req, err := http.NewRequest(http.MethodGet, baseUrl+"/category/monsters", nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	reqUrl := req.URL
	queryParams := req.URL.Query()
	queryParams.Set("game", "totk")
	reqUrl.RawQuery = queryParams.Encode()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error getting monsters: %w", err)
	}

	var responseBody *GetMonstersResponse
	if err := json.NewDecoder(resp.Body).Decode(&responseBody); err != nil {
		return nil, fmt.Errorf("error parsing monsters: %w", err)
	}

	return responseBody, nil
}
