package segment

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/ajbosco/segment-config-go/segment"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

func diagFromErrPtr(err error) *diag.Diagnostics {
	d := diag.FromErr(err)
	return &d
}

// Converts a segment resource path to its id
// E.g: workspaces/myworkspace/sources/mysource => mysource
func pathToName(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}

	return path
}

// withBackoff calls the passed function returning a result and an error and performs an exponential backoff if it fails with a 429 HTTP status code
func withBackoff(call func() (interface{}, error), initialRetryDelay time.Duration, maxRetries int) (interface{}, error) {
	results, err := call()
	if err != nil {
		if e, ok := err.(*segment.SegmentApiError); ok && e.Code == http.StatusTooManyRequests && maxRetries > 0 {
			log.Printf("[INFO] Backoff: failed, waiting %sms before retrying, %d tries left", initialRetryDelay, maxRetries)
			time.Sleep(initialRetryDelay)
			return withBackoff(call, initialRetryDelay*2, maxRetries-1)
		}
	}

	return results, err
}
