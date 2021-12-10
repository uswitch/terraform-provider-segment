package provider_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/uswitch/terraform-provider-segment/internal/provider"
)

const testPrefix = "test-acc"

var testAccProvider *schema.Provider
var testAccProviders map[string]func() (*schema.Provider, error)
var testAccProviderConfigure sync.Once

func init() {
	testAccProvider = provider.Provider()
	testAccProviders = map[string]func() (*schema.Provider, error){
		"segment": func() (*schema.Provider, error) { return testAccProvider, nil },
	}
}

// testAccPreCheck verifies that the test runs in a properly set environment. It should be added to any
func testAccPreCheck(t *testing.T) {
	testAccProviderConfigure.Do(func() {
		if os.Getenv("SEGMENT_ACCESS_TOKEN") == "" {
			t.Fatal("SEGMENT_ACCESS_TOKEN must be set for acceptance tests")
		}
		if os.Getenv("SEGMENT_WORKSPACE") == "" {
			t.Fatal("SEGMENT_WORKSPACE must be set for acceptance tests")
		}

		testAccProvider.Configure(context.Background(), terraform.NewResourceConfigRaw(nil))
	})
}

type PreConditionResources struct {
	Sources      []string
	Destinations []string
}

func (pcr PreConditionResources) addSource(name string) PreConditionResources {
	pcr.Sources = append(pcr.Sources, fmt.Sprintf("segment_source.%s.id", name))
	return pcr
}

func (pcr PreConditionResources) addDestination(id string) PreConditionResources {
	pcr.Destinations = append(pcr.Destinations, fmt.Sprintf("segment_destination.%s.id", id))
	return pcr
}

type PreCondition struct {
	PreConditionResources
	config string
}

func (pc PreCondition) appendResource(r PreConditionResources, config string, args ...interface{}) PreCondition {
	return PreCondition{
		PreConditionResources: r,
		config:                pc.config + fmt.Sprintf(config, args...),
	}
}

type SourcePreCondition struct {
	name string
	PreCondition
}

func (pc PreCondition) WithSource() SourcePreCondition {
	name := acctest.RandomWithPrefix(testPrefix)
	c := SourcePreCondition{name, pc.appendResource(pc.addSource(name), `
resource "segment_source" "%[1]s" {

	# name
	catalog_name = "catalog/sources/javascript"
	source_name  = "%[1]s"
}`, name),
	}

	return c
}

type DestinationPreCondition struct {
	PreCondition
}

func (pc SourcePreCondition) WithDestination() DestinationPreCondition {
	destType := "rtb-house"
	destId := strings.Join([]string{pc.name, destType}, "__")
	// Segment enforces workspace id as a secret for Kinesis
	c := DestinationPreCondition{PreCondition: pc.PreCondition.appendResource(pc.addDestination(destId), `
resource "segment_destination" "%s" {

	source          = segment_source.%s.id
	name            = "%s"
	enabled         = true
	connection_mode = "UNSPECIFIED"

  config = {
    #API Key
    apiKey = jsonencode({
      type  = "string"
      value = "abcdefgh"
    })
  }
}`, destId, pc.name, destType),
	}

	return c
}

func (pc PreCondition) Build(createConfig func(res PreConditionResources) string) string {
	finalConfig := pc.config + createConfig(pc.PreConditionResources)
	log.Println("[INFO] Building config:")
	log.Printf("[INFO] %s", finalConfig)
	return finalConfig
}

// PreCondition provides a way to pre-create configuration that is required by a test, but not being tested.
// It can be created using NewPreCondition().
// Example:
// 	NewPreCondition().
//		WithSource().
//		WithDestination().
//		Build(func (r PreConditionResources) string {
//			return "<Config under test>"
//		})
// The PreConditionResources parameter provides a way to access the resource ids of the created precondition resources, so they can be referenced by the tested code.
// Example:
// 	func (r PreConditionResources) string {
//			source := r.Sources[0]
//			return fmt.Sprintf(`
//				resource "my_resource" {
//					source = %s
//				}
//			`, source)
//	}
// Note: PreConditionResources contains terraform resource ids, so they don't need to be wrapped in "" when being used.
func NewPreCondition() PreCondition {
	return PreCondition{}
}
