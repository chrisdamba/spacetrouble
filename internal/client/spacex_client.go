package spacex

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	httpClient HTTPClient
	baseURL    string
}

type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}

type Option func(*Client)

type Launch struct {
	LaunchPadID   string `json:"launchpad"`
	Date          int64  `json:"date_unix"`
	DatePrecision string `json:"date_precision"`
}

type LaunchPad struct {
	Id     string `json:"id"`
	Status string `json:"status"`
}

type LaunchesResponse struct {
	Launches []Launch `json:"docs"`
}

type SearchQuery struct {
	Query   map[string]interface{} `json:"query"`
	Options map[string]interface{} `json:"options"`
}

var (
	ErrNotFound      error = errors.New("launchpad not found")
	ErrBadStatusCode error = errors.New("invalid status code from spacex")
)

func WithBaseURL(url string) Option {
	return func(c *Client) {
		c.baseURL = url
	}
}

func WithHTTPClient(httpClient HTTPClient) Option {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

func NewClient(opts ...Option) *Client {
	client := &Client{
		httpClient: &http.Client{Timeout: 15 * time.Second},
		baseURL:    "https://api.spacexdata.com/v4",
	}

	for _, opt := range opts {
		opt(client)
	}

	return client
}

func (o *LaunchPad) IsActive() bool {
	return o.Status == "active"
}

func (c *Client) CheckLaunchConflict(ctx context.Context, launchpadID string, ts time.Time) (bool, error) {
	if err := validateInputs(launchpadID, ts); err != nil {
		return false, fmt.Errorf("invalid input: %w", err)
	}

	launchpad, err := c.getLaunchpad(ctx, launchpadID)
	if err != nil {
		return false, fmt.Errorf("checking launchpad: %w", err)
	}

	if !launchpad.IsActive() {
		return false, nil
	}

	launches, err := c.getUpcomingLaunches(ctx, launchpadID)
	if err != nil {
		return false, fmt.Errorf("checking upcoming launches: %w", err)
	}

	return c.isDateAvailable(launches, ts)
}

func (c *Client) GetLaunchPadById(ctx context.Context, launchpadID string) (LaunchPad, error) {
	var ans LaunchPad
	u := fmt.Sprintf("%s/%s/%s", c.baseURL, "launchpads", launchpadID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return ans, err
	}
	req.Header.Add("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return ans, err
	}
	defer func() {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}()

	if resp.StatusCode == http.StatusNotFound {
		return ans, ErrNotFound
	}

	if resp.StatusCode != http.StatusOK {
		return ans, ErrBadStatusCode
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ans, err
	}
	return ans, json.Unmarshal(body, &ans)
}

func (c *Client) GetUpcomingLaunchesLaunchPad(ctx context.Context, launchpadID string) ([]Launch, error) {
	u := fmt.Sprintf("%s/%s", c.baseURL, "launches/query")
	q := c.generateUpcomingSearchQuery(launchpadID)
	jsonBytes, err := json.Marshal(q)

	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewBuffer(jsonBytes))

	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)

	if err != nil {
		return nil, err
	}
	defer func() {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}()

	if resp.StatusCode != 200 {
		fmt.Println(resp.StatusCode)
		return nil, ErrBadStatusCode
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var launchesResponse LaunchesResponse

	if err := json.Unmarshal(body, &launchesResponse); err != nil {
		return nil, err
	}
	return launchesResponse.Launches, nil
}

func (c *Client) getLaunchpad(ctx context.Context, id string) (LaunchPad, error) {
	launchpad, err := c.GetLaunchPadById(ctx, id)
	if err != nil {
		return LaunchPad{}, fmt.Errorf("fetching launchpad: %w", err)
	}
	return launchpad, nil
}

func (c *Client) getUpcomingLaunches(ctx context.Context, launchpadID string) ([]Launch, error) {
	launches, err := c.GetUpcomingLaunchesLaunchPad(ctx, launchpadID)
	if err != nil {
		return nil, fmt.Errorf("fetching upcoming launches: %w", err)
	}
	return launches, nil
}

func (c *Client) isDateAvailable(launches []Launch, ts time.Time) (bool, error) {
	for _, launch := range launches {
		available, err := launch.IsDayAvailable(ts)
		if err != nil {
			return false, fmt.Errorf("checking date availability: %w", err)
		}
		if !available {
			return false, nil
		}
	}
	return true, nil
}

func (c *Client) generateUpcomingSearchQuery(launchpadID string) SearchQuery {
	searchQuery := SearchQuery{
		Query:   make(map[string]interface{}),
		Options: make(map[string]interface{}),
	}

	searchQuery.Query["upcoming"] = true
	searchQuery.Query["launchpad"] = launchpadID
	searchQuery.Options["select"] = []string{"launchpad", "date_unix", "date_precision"}
	searchQuery.Options["sort"] = map[string]string{"date_unix": "asc"}
	searchQuery.Options["limit"] = 10000

	return searchQuery
}

func (l *Launch) IsDayAvailable(t time.Time) (bool, error) {
	// get the launch date from Unix timestamp
	launchDate := time.Unix(l.Date, 0).UTC()
	requestDate := t.UTC()

	// normalise both dates to start of day
	launchStart := time.Date(launchDate.Year(), launchDate.Month(), launchDate.Day(), 0, 0, 0, 0, time.UTC)
	requestStart := time.Date(requestDate.Year(), requestDate.Month(), requestDate.Day(), 0, 0, 0, 0, time.UTC)

	// calculate end based on precision
	var end time.Time
	switch l.DatePrecision {
	case "quarter":
		end = launchStart.AddDate(0, 3, 0).Add(-time.Second)
	case "half":
		end = launchStart.AddDate(0, 6, 0).Add(-time.Second)
	case "year":
		end = launchStart.AddDate(1, 0, 0).Add(-time.Second)
	case "month":
		end = launchStart.AddDate(0, 1, 0).Add(-time.Second)
	case "hour":
		end = launchDate.Add(time.Hour)
	case "day":
		end = launchStart.AddDate(0, 0, 1).Add(-time.Second)
	default:
		return false, fmt.Errorf("invalid date precision: %s", l.DatePrecision)
	}

	// date is unavailable if it falls within the launch window
	return !(requestStart.Equal(launchStart) || (requestStart.After(launchStart) && requestStart.Before(end))), nil
}

func validateInputs(launchpadID string, ts time.Time) error {
	if launchpadID == "" {
		return fmt.Errorf("launchpad ID cannot be empty")
	}

	if ts.Before(time.Now()) {
		return fmt.Errorf("launch date must be in the future")
	}

	return nil
}
