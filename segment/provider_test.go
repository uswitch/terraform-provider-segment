package segment_test

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
	provider "github.com/uswitch/terraform-provider-segment/segment"
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
	pcr.Sources = append(pcr.Sources, fmt.Sprintf("source.%s.id", name))
	return pcr
}

func (pcr PreConditionResources) addDestination(id string) PreConditionResources {
	pcr.Destinations = append(pcr.Destinations, fmt.Sprintf("destination.%s.id", id))
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

type PreConditions []PreCondition

type SourcePreCondition struct {
	name string
	PreCondition
}

func (pc PreCondition) WithSource() SourcePreCondition {
	name := acctest.RandomWithPrefix(testPrefix)
	c := SourcePreCondition{name, pc.appendResource(pc.addSource(name), `
resource "source" "%[1]s" {
	provider = segment

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
	destType := "amazon-kinesis"
	destId := strings.Join([]string{pc.name, destType}, "__")
	c := DestinationPreCondition{PreCondition: pc.PreCondition.appendResource(pc.addDestination(destId), `
resource "destination" "%s" {
	provider = segment

	source          = source.%s.id
	name            = "%s"
	enabled         = true
	connection_mode = "UNSPECIFIED"

	config = {
		stream = jsonencode({
			type  = "string"
			value = "983265867"
		})

		useMessageId = jsonencode({
			type  = "boolean"
			value = false
		})

		region = jsonencode({
			type  = "string"
			value = "eu"
		})

		roleAddress = jsonencode({
			type  = "string"
			value = "aws:eu-west-1::abcdef"
		})

		secretId = jsonencode({
			type  = "string"
			value = "abcd"
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

func NewPreConditions() PreCondition {
	return PreCondition{}
}