package segment_test

import (
	"log"
	"strings"
	"testing"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/uswitch/terraform-provider-segment/segment"
)

func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func sourceSweeper(name string) *resource.Sweeper {
	sweep := func(_ string) error {
		meta := testAccProvider.Meta().(segment.ProviderMetadata)
		client := meta.Client

		sources, err := client.ListSources()

		log.Printf("[INFO] Sweeping through %d sources", len(sources.Sources))

		var errs error
		deleted := 0
		for _, source := range sources.Sources {
			if strings.HasPrefix(source.Name, "test-acc-") {
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
