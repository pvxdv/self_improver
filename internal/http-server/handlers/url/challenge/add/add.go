package add

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"go.uber.org/zap"

	"github.com/pvxdv/self_improver/internal/lib/api/response"
	"github.com/pvxdv/self_improver/internal/model"
)

type ChallengeSaver interface {
	SaveChallenge(ctx context.Context, data *model.Challenge) (int64, error)
}

func New(ctx context.Context, saver ChallengeSaver, logger *zap.SugaredLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data, err := io.ReadAll(r.Body)
		if err != nil {
			logger.Errorf("failed to read request body: %v", err)
			response.RespondWithError(w, http.StatusBadRequest, "failed to read request body", logger)
			return
		}

		var challenge model.Challenge
		if err := json.Unmarshal(data, &challenge); err != nil {
			logger.Errorf("failed to unmarshal request body: %v", err)
			response.RespondWithError(w, http.StatusBadRequest, "invalid request format", logger)
			return
		}

		//TODO validate

		id, err := saver.SaveChallenge(ctx, &challenge)
		if err != nil {
			logger.Errorf("failed to save challenge: %v", err)
			response.RespondWithError(w, http.StatusInternalServerError, "failed to save challenge", logger)
			return
		}

		logger.Infof("challenge (ID:%d) saved successfully", id)
		response.RespondWithJSON(w, http.StatusOK, response.OK(), logger)
	}
}
