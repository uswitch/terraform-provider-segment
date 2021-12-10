package utils

import (
	"encoding/json"
	"log"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/ajbosco/segment-config-go/segment"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DiagFromErrPtr(err error) *diag.Diagnostics {
	d := diag.FromErr(err)
	return &d
}

// Converts a segment resource path to its id
// E.g: workspaces/myworkspace/sources/mysource => mysource
func PathToName(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}

	return path
}

// withBackoff calls the passed function returning a result and an error and performs an exponential backoff if it fails with a 429 HTTP status code
func WithBackoff(call func() (interface{}, error), initialRetryDelay time.Duration, maxRetries int) (interface{}, error) {
	results, err := call()
	if err != nil {
		if e, ok := err.(*segment.SegmentApiError); ok && e.Code == http.StatusTooManyRequests && maxRetries > 0 {
			log.Printf("[INFO] Backoff: failed, waiting %sms before retrying, %d tries left", initialRetryDelay, maxRetries)
			time.Sleep(initialRetryDelay)
			return WithBackoff(call, initialRetryDelay*2, maxRetries-1)
		}
	}

	return results, err
}

func unmarshalGeneric(input string) interface{} {
	var decodedStr interface{}
	if err := json.Unmarshal([]byte(input), &decodedStr); err != nil {
		log.Panicln("generic unmarshal failed", err)
	}

	return decodedStr
}

// diffRulesJSONState suppresses diff if json values are equivalent, independant of whitespace or order of keys
func DiffRulesJSONState(_, old, new string, _ *schema.ResourceData) bool {
	if old == "" || new == "" {
		return old == new
	}

	encodedNew := unmarshalGeneric(new)
	encodedOld := unmarshalGeneric(old)
	return reflect.DeepEqual(encodedOld, encodedNew)
}

// Search performs a linear search in an indexed collection and returns the index of the found item, or -1
func Search(len int, eq func(i int) bool) int {
	for i := 0; i < len; i++ {
		if eq(i) {
			return i
		}
	}

	return -1
}

func CatchFirst(ops ...func() error) diag.Diagnostics {
	for _, op := range ops {
		if err := op(); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}
