package resetter

import (
	"context"
	"time"

	"github.com/inconshreveable/log15"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/store"
)

type UploadResetter struct {
	Store         store.Store
	ResetInterval time.Duration
	Metrics       ResetterMetrics
}

// Run periodically moves all uploads that have been in the PROCESSING state for a
// while back to QUEUED. For each updated upload record, the conversion process that
// was responsible for handling the upload did not hold a row lock, indicating that
// it has died.
func (ur *UploadResetter) Run() {
	for {
		resetIDs, erroredIDs, err := ur.Store.ResetStalled(context.Background(), time.Now())
		if err != nil {
			ur.Metrics.Errors.Inc()
			log15.Error("Failed to reset stalled uploads", "error", err)
		}
		for _, id := range resetIDs {
			log15.Debug("Reset stalled upload", "uploadID", id)
		}
		for _, id := range erroredIDs {
			log15.Debug("Failed stalled upload", "uploadID", id)
		}

		ur.Metrics.UploadResets.Add(float64(len(resetIDs)))
		ur.Metrics.UploadResetFailures.Add(float64(len(erroredIDs)))
		time.Sleep(ur.ResetInterval)
	}
}
