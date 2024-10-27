package health

import (
	"encoding/json"
	"net/http"
	"runtime"
	"time"
)

type HealthResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
	Version   string `json:"version,omitempty"`
	Uptime    string `json:"uptime"`
	GoVersion string `json:"go_version"`
	Memory    struct {
		Alloc      uint64 `json:"alloc"`      // bytes allocated and not yet freed
		TotalAlloc uint64 `json:"totalAlloc"` // total bytes allocated (even if freed)
		Sys        uint64 `json:"sys"`        // bytes obtained from system
		NumGC      uint32 `json:"numGC"`      // number of garbage collections
	} `json:"memory"`
}

var startTime = time.Now()

func HealthGet() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		var memStats runtime.MemStats
		runtime.ReadMemStats(&memStats)

		health := HealthResponse{
			Status:    "healthy",
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Version:   "1.0.0", // You might want to make this configurable
			Uptime:    time.Since(startTime).String(),
			GoVersion: runtime.Version(),
		}

		health.Memory.Alloc = memStats.Alloc
		health.Memory.TotalAlloc = memStats.TotalAlloc
		health.Memory.Sys = memStats.Sys
		health.Memory.NumGC = memStats.NumGC

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if err := json.NewEncoder(w).Encode(health); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Failed to encode health check response",
			})
			return
		}
	}
}
