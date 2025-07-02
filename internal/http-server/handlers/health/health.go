package health

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"

	"github.com/pvxdv/self_improver/internal/lib/api/response"
)

func New(log *zap.SugaredLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		respBody, err := json.Marshal(response.OK())
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Errorf("Failed to marshal OK response: %s", err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, err = w.Write(respBody)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Errorf("Failed to write OK response: %s", err)
		}
	}
}
