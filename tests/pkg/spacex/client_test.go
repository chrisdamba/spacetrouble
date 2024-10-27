package spacex_test

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/chrisdamba/spacetrouble/pkg/spacex"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

type mockHTTPClient struct {
	doFunc func(*http.Request) (*http.Response, error)
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.doFunc(req)
}

func newTestClient(doFunc func(*http.Request) (*http.Response, error)) *spacex.Client {
	return spacex.NewClient(
		spacex.WithHTTPClient(&mockHTTPClient{doFunc: doFunc}),
		spacex.WithBaseURL("https://test.spacex.com/v4"),
	)
}
func TestCheckLaunchConflict(t *testing.T) {
	futureDate := time.Now().Add(24 * time.Hour)
	baseTime := time.Date(2025, 10, 1, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		name          string
		launchpadID   string
		launchTime    time.Time
		setupResponse func(*http.Request) (*http.Response, error)
		want          bool
		wantErr       string
	}{
		{
			name:        "empty launchpad ID",
			launchpadID: "",
			launchTime:  futureDate,
			want:        false,
			wantErr:     "launchpad ID cannot be empty",
		},
		{
			name:        "past launch date",
			launchpadID: "pad1",
			launchTime:  time.Now().Add(-24 * time.Hour),
			want:        false,
			wantErr:     "launch date must be in the future",
		},
		{
			name:        "launchpad not found",
			launchpadID: "invalid_pad",
			launchTime:  futureDate,
			setupResponse: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusNotFound,
					Body:       io.NopCloser(strings.NewReader("")),
				}, nil
			},
			want:    false,
			wantErr: "checking launchpad: fetching launchpad: launchpad not found",
		},
		{
			name:        "inactive launchpad",
			launchpadID: "pad1",
			launchTime:  futureDate,
			setupResponse: func(req *http.Request) (*http.Response, error) {
				if strings.Contains(req.URL.Path, "launchpads") {
					launchpad := spacex.LaunchPad{
						Id:     "pad1",
						Status: "inactive",
					}
					body, _ := json.Marshal(launchpad)
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(bytes.NewReader(body)),
					}, nil
				}
				return nil, nil
			},
			want:    false,
			wantErr: "",
		},
		{
			name:        "available active launchpad",
			launchpadID: "pad1",
			launchTime:  futureDate,
			setupResponse: func(req *http.Request) (*http.Response, error) {
				if strings.Contains(req.URL.Path, "launchpads") {
					launchpad := spacex.LaunchPad{
						Id:     "pad1",
						Status: "active",
					}
					body, _ := json.Marshal(launchpad)
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(bytes.NewReader(body)),
					}, nil
				}
				if strings.Contains(req.URL.Path, "launches/query") {
					resp := spacex.LaunchesResponse{
						Launches: []spacex.Launch{},
					}
					body, _ := json.Marshal(resp)
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(bytes.NewReader(body)),
					}, nil
				}
				return nil, nil
			},
			want:    true,
			wantErr: "",
		},
		{
			name:        "conflict with existing launch",
			launchpadID: "pad1",
			launchTime:  baseTime,
			setupResponse: func(req *http.Request) (*http.Response, error) {
				if strings.Contains(req.URL.Path, "launchpads") {
					launchpad := spacex.LaunchPad{
						Id:     "pad1",
						Status: "active",
					}
					body, _ := json.Marshal(launchpad)
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(bytes.NewReader(body)),
					}, nil
				}
				if strings.Contains(req.URL.Path, "launches/query") {
					// Create a launch on the same day
					resp := spacex.LaunchesResponse{
						Launches: []spacex.Launch{
							{
								LaunchPadID:   "pad1",
								Date:          baseTime.Unix(),
								DatePrecision: "day",
							},
						},
					}
					body, _ := json.Marshal(resp)
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(bytes.NewReader(body)),
					}, nil
				}
				return nil, nil
			},
			want:    false, // Should be unavailable due to same-day conflict
			wantErr: "",
		},

		{
			name:        "no conflict with previous day launch",
			launchpadID: "pad1",
			launchTime:  baseTime,
			setupResponse: func(req *http.Request) (*http.Response, error) {
				if strings.Contains(req.URL.Path, "launchpads") {
					launchpad := spacex.LaunchPad{
						Id:     "pad1",
						Status: "active",
					}
					body, _ := json.Marshal(launchpad)
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(bytes.NewReader(body)),
					}, nil
				}
				if strings.Contains(req.URL.Path, "launches/query") {
					// create a launch on the previous day
					resp := spacex.LaunchesResponse{
						Launches: []spacex.Launch{
							{
								LaunchPadID:   "pad1",
								Date:          baseTime.AddDate(0, 0, -1).Unix(),
								DatePrecision: "day",
							},
						},
					}
					body, _ := json.Marshal(resp)
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(bytes.NewReader(body)),
					}, nil
				}
				return nil, nil
			},
			want:    true, // should be available as launch is on previous day
			wantErr: "",
		},
		{
			name:        "launch after existing launch day",
			launchpadID: "pad1",
			launchTime:  futureDate,
			setupResponse: func(req *http.Request) (*http.Response, error) {
				if strings.Contains(req.URL.Path, "launchpads") {
					launchpad := spacex.LaunchPad{
						Id:     "pad1",
						Status: "active",
					}
					body, _ := json.Marshal(launchpad)
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(bytes.NewReader(body)),
					}, nil
				}
				if strings.Contains(req.URL.Path, "launches/query") {
					// Create a launch that's before the requested date
					launchTime := futureDate.Add(-48 * time.Hour) // Two days before
					resp := spacex.LaunchesResponse{
						Launches: []spacex.Launch{
							{
								LaunchPadID:   "pad1",
								Date:          launchTime.Unix(),
								DatePrecision: "day",
							},
						},
					}
					body, _ := json.Marshal(resp)
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(bytes.NewReader(body)),
					}, nil
				}
				return nil, nil
			},
			want:    true, // Should be available as launch is on a different day
			wantErr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := newTestClient(tt.setupResponse)
			got, err := client.CheckLaunchConflict(context.Background(), tt.launchpadID, tt.launchTime)

			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got, "CheckLaunchConflict result mismatch")
		})
	}
}

func TestLaunch_IsDayAvailable(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name      string
		launch    spacex.Launch
		checkDate time.Time
		want      bool
		wantErr   bool
	}{
		{
			name: "day precision - date after launch",
			launch: spacex.Launch{
				Date:          now.Unix(),
				DatePrecision: "day",
			},
			checkDate: now.Add(24 * time.Hour),
			want:      true,
			wantErr:   false,
		},
		{
			name: "day precision - same day",
			launch: spacex.Launch{
				Date:          now.Unix(),
				DatePrecision: "day",
			},
			checkDate: now,
			want:      false,
			wantErr:   false,
		},
		{
			name: "month precision - within month",
			launch: spacex.Launch{
				Date:          now.Unix(),
				DatePrecision: "month",
			},
			checkDate: now.AddDate(0, 0, 15),
			want:      false,
			wantErr:   false,
		},
		{
			name: "month precision - next month",
			launch: spacex.Launch{
				Date:          now.Unix(),
				DatePrecision: "month",
			},
			checkDate: now.AddDate(0, 1, 1),
			want:      true,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.launch.IsDayAvailable(tt.checkDate)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestClient_GetLaunchPadById(t *testing.T) {
	tests := []struct {
		name          string
		launchpadID   string
		setupResponse func(*http.Request) (*http.Response, error)
		want          spacex.LaunchPad
		wantErr       string
	}{
		{
			name:        "successful retrieval",
			launchpadID: "pad1",
			setupResponse: func(req *http.Request) (*http.Response, error) {
				launchpad := spacex.LaunchPad{
					Id:     "pad1",
					Status: "active",
				}
				body, _ := json.Marshal(launchpad)
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewReader(body)),
				}, nil
			},
			want: spacex.LaunchPad{
				Id:     "pad1",
				Status: "active",
			},
			wantErr: "",
		},
		{
			name:        "not found",
			launchpadID: "invalid_pad",
			setupResponse: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusNotFound,
					Body:       io.NopCloser(strings.NewReader("")),
				}, nil
			},
			want:    spacex.LaunchPad{},
			wantErr: "launchpad not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := newTestClient(tt.setupResponse)
			got, err := client.GetLaunchPadById(context.Background(), tt.launchpadID)

			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
