package provider_test

import (
	"log"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/uswitch/segment-config-go/segment"
	"github.com/uswitch/terraform-provider-segment/internal/utils"
)

func TestMain(m *testing.M) {
	resource.TestMain(m)
}

// sourceSweeper allows to wipe out any lingering sources remaining after test in case the tests fail with an unexpected state.
// It will delete any resource starting with the value of the `testPrefix` constant.
// Use it in test init:
// 	func init() {
//		resource.AddTestSweepers("unique-name", sourceSweeper("unique-name"))
//	}
func sourceSweeper(name string) *resource.Sweeper {
	var token, workspace string
	ok := false
	if token, ok = os.LookupEnv("SEGMENT_ACCESS_TOKEN"); !ok {
		panic("SEGMENT_ACCESS_TOKEN must be set for acceptance tests")
	}
	if workspace, ok = os.LookupEnv("SEGMENT_WORKSPACE"); !ok {
		panic("SEGMENT_WORKSPACE must be set for acceptance tests")
	}

	client := segment.NewClient(token, workspace)

	sweep := func(_ string) error {
		sources, err := client.ListSources()

		log.Printf("[INFO] Sweeping through %d sources", len(sources.Sources))

		var errs error
		deleted := 0
		for _, source := range sources.Sources {
			src := utils.PathToName(source.Name)
			log.Printf("[INFO] Checking source %s", src)
			if strings.HasPrefix(src, testPrefix) {
				log.Printf("[INFO] Deleting source %s", source.Name)
				if multierror.Append(errs, client.DeleteSource(source.Name)) != nil {
					deleted += 1
				}
			}
		}

		log.Printf("[INFO] Deleted %d sources", deleted)

		return err
	}

	return &resource.Sweeper{
		Name: name,
		F:    sweep,
	}
}
