package segment_test

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"testing"

	"github.com/ajbosco/segment-config-go/segment"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
	provider "github.com/uswitch/terraform-provider-segment/segment"
)

func init() {
	resource.AddTestSweepers("destination-filters", sourceSweeper("destination-filters"))
}

func TestAccDestinationFilter_basic(t *testing.T) {
	name := acctest.RandomWithPrefix(testPrefix)
	rName := "segment_destination_filter." + name
	var filter segment.DestinationFilter
	c := withDestination(name)
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		PreCheck:          func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccDestinationFilterExists(t, rName, &filter),
					testAccDestinationFilterBasicFields(t, rName, &filter),
					resource.TestCheckResourceAttr(rName, "title", "Foo"),
					resource.TestCheckResourceAttr(rName, "description", "Bar"),
					resource.TestCheckResourceAttr(rName, "condition", "context.castPermissions.marketing = false"),
					resource.TestCheckResourceAttr(rName, "enabled", "true"),
				),
				Config: c(testAccDestinationFilterConfigBasic),
			},
			{
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccDestinationFilterExists(t, rName, &filter),
					testAccDestinationFilterBasicFields(t, rName, &filter),
					resource.TestCheckResourceAttr(rName, "title", "Foo1"),
					resource.TestCheckResourceAttr(rName, "description", "Bar1"),
					resource.TestCheckResourceAttr(rName, "condition", "context.castPermissions.marketing = true"),
					resource.TestCheckResourceAttr(rName, "enabled", "false"),
				),
				Config: c(testAccDestinationFilterConfigBasicUpdated),
			},
			{
				ResourceName:      rName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDestinationFilter_actions(t *testing.T) {
	name := acctest.RandomWithPrefix(testPrefix)
	rName := "segment_destination_filter." + name
	c := withDestination(name)
	var filter segment.DestinationFilter

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		PreCheck:          func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccDestinationFilterExists(t, rName, &filter),
					testAccDestinationFilterActions(t, rName, &filter),
					resource.TestCheckResourceAttr(rName, "actions.0.drop.#", "1"),
				),
				Config: c(testAccDestinationFilterConfigBasic),
			},
			{
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccDestinationFilterExists(t, rName, &filter),
					testAccDestinationFilterActions(t, rName, &filter),
					resource.TestCheckResourceAttr(rName, "actions.0.allow_fields.#", "1"),
					resource.TestCheckResourceAttr(rName, "actions.0.allow_fields.0.context.0", "foo"),
					resource.TestCheckResourceAttr(rName, "actions.0.allow_fields.0.properties.0", "bar"),
					resource.TestCheckResourceAttr(rName, "actions.0.allow_fields.0.traits.0", "baz"),
				),
				Config: c(testAccDestinationFilterConfigActionsAllowEvents),
			},
			{
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccDestinationFilterExists(t, rName, &filter),
					testAccDestinationFilterActions(t, rName, &filter),
					resource.TestCheckResourceAttr(rName, "actions.0.allow_fields.#", "1"),
					resource.TestCheckResourceAttr(rName, "actions.0.allow_fields.0.context.0", "foo"),
					resource.TestCheckResourceAttr(rName, "actions.0.allow_fields.0.properties.0", "bar"),
					resource.TestCheckResourceAttr(rName, "actions.0.allow_fields.0.traits.0", "baz"),
					resource.TestCheckResourceAttr(rName, "actions.0.block_fields.0.context.0", "one"),
					resource.TestCheckResourceAttr(rName, "actions.0.block_fields.0.properties.0", "two"),
					resource.TestCheckResourceAttr(rName, "actions.0.block_fields.0.traits.0", "three"),
				),
				Config: c(testAccDestinationFilterConfigActionsBlockEvents),
			},
			{
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccDestinationFilterExists(t, rName, &filter),
					testAccDestinationFilterActions(t, rName, &filter),
					resource.TestCheckResourceAttr(rName, "actions.0.allow_fields.#", "1"),
					resource.TestCheckResourceAttr(rName, "actions.0.allow_fields.0.context.0", "foo"),
					resource.TestCheckResourceAttr(rName, "actions.0.allow_fields.0.properties.0", "bar"),
					resource.TestCheckResourceAttr(rName, "actions.0.allow_fields.0.traits.0", "baz"),
					resource.TestCheckResourceAttr(rName, "actions.0.block_fields.0.context.0", "one"),
					resource.TestCheckResourceAttr(rName, "actions.0.block_fields.0.properties.0", "two"),
					resource.TestCheckResourceAttr(rName, "actions.0.block_fields.0.traits.0", "three"),
					resource.TestCheckResourceAttr(rName, "actions.0.sample.0.percent", "0.5"),
					resource.TestCheckResourceAttr(rName, "actions.0.sample.0.path", "userId"),
				),
				Config: c(testAccDestinationFilterConfigActionsOneSampling),
			},
			{
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccDestinationFilterExists(t, rName, &filter),
					testAccDestinationFilterActions(t, rName, &filter),
					resource.TestCheckResourceAttr(rName, "actions.0.allow_fields.#", "1"),
					resource.TestCheckResourceAttr(rName, "actions.0.allow_fields.#", "1"),
					resource.TestCheckResourceAttr(rName, "actions.0.allow_fields.0.context.0", "foo"),
					resource.TestCheckResourceAttr(rName, "actions.0.allow_fields.0.properties.0", "bar"),
					resource.TestCheckResourceAttr(rName, "actions.0.allow_fields.0.traits.0", "baz"),
					resource.TestCheckResourceAttr(rName, "actions.0.block_fields.0.context.0", "one"),
					resource.TestCheckResourceAttr(rName, "actions.0.block_fields.0.properties.0", "two"),
					resource.TestCheckResourceAttr(rName, "actions.0.block_fields.0.traits.0", "three"),
					resource.TestCheckResourceAttr(rName, "actions.0.sample.0.percent", "0.8"),
					resource.TestCheckResourceAttr(rName, "actions.0.sample.0.path", "properties.price"),
					resource.TestCheckResourceAttr(rName, "actions.0.sample.1.percent", "0.5"),
					resource.TestCheckResourceAttr(rName, "actions.0.sample.1.path", "userId"),
				),
				Config: c(testAccDestinationFilterConfigActionsTwoSamplings),
			},
		},
	})
}

// Assertions

func testAccDestinationFilterExists(t *testing.T, filterResName string, filter *segment.DestinationFilter) func(s *terraform.State) error {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(provider.ProviderMetadata).Client
		filterState := s.RootModule().Resources[filterResName]
		sourceName, destinationName, filterId := provider.SplitDestinationFilterId(filterState.Primary.ID)
		log.Printf("[INFO] Getting filter %s/%s/%s", sourceName, destinationName, filterId)

		f, err := client.GetDestinationFilter(sourceName, destinationName, filterId)
		if err != nil {
			return err
		}

		if f == nil {
			return errors.New("Expected filter " + filterState.Primary.ID + " not to be nil")
		}

		*filter = *f

		return nil
	}
}

func testAccDestinationFilterBasicFields(t *testing.T, filterResName string, f *segment.DestinationFilter) func(s *terraform.State) error {
	return func(s *terraform.State) error {
		filterState := s.RootModule().Resources[filterResName]
		log.Printf("[INFO] Checking fields for %s", filterResName)

		assert.Equal(t, f.Title, filterState.Primary.Attributes["title"])
		assert.Equal(t, f.Description, filterState.Primary.Attributes["description"])
		assert.Equal(t, f.Conditions, filterState.Primary.Attributes["condition"])
		assert.Equal(t, strconv.FormatBool(f.IsEnabled), filterState.Primary.Attributes["enabled"])

		return nil
	}
}

func testAccSampledAction(t *testing.T, filterState map[string]string, action segment.SamplingEventAction) {
	sPath := "actions.0.sample."
	sampleCount, err := strconv.ParseInt(filterState[sPath+"#"], 10, 32)
	assert.NoError(t, err)

	log.Printf("[INFO] Checking: %s", fmt.Sprint(action.Percent)+action.Path)

	has := map[string]bool{}
	for i := 0; i < int(sampleCount); i++ {
		path := fmt.Sprintf("%s%d.", sPath, i)
		percent := filterState[path+"percent"]
		fieldPath := filterState[path+"path"]

		has[percent+fieldPath] = true
	}

	log.Printf("[INFO] Built samples config set: %+v", has)

	expectedAttrs := fmt.Sprint(action.Percent) + action.Path
	assert.True(t, has[expectedAttrs])
}

func testAccEventFields(t *testing.T, filterState map[string]string, action segment.FieldsListEventAction, allowOrBlock string) {
	sPath := "actions.0." + allowOrBlock + "_fields.0"

	if action.Fields.Properties != nil {
		log.Printf("[INFO] Checking properties")
		for i, p := range action.Fields.Properties.Fields {
			path := sPath + ".properties." + fmt.Sprint(i)
			sp := filterState[path]

			log.Printf("[INFO] Checking %s at %s", p, path)
			assert.Equal(t, p, sp)
		}
	}

	if action.Fields.Context != nil {
		log.Printf("[INFO] Checking context")
		for i, p := range action.Fields.Context.Fields {
			path := sPath + ".context." + fmt.Sprint(i)
			sp := filterState[path]

			log.Printf("[INFO] Checking %s at %s", p, path)
			assert.Equal(t, p, sp)
		}
	}

	if action.Fields.Traits != nil {
		log.Printf("[INFO] Checking traits")
		for i, p := range action.Fields.Traits.Fields {
			path := sPath + ".traits." + fmt.Sprint(i)
			sp := filterState[path]

			log.Printf("[INFO] Checking %s at %s", p, path)
			assert.Equal(t, p, sp)
		}
	}
}

func testAccDestinationFilterActions(t *testing.T, filterResName string, f *segment.DestinationFilter) func(s *terraform.State) error {
	return func(s *terraform.State) error {
		filterState := s.RootModule().Resources[filterResName].Primary.Attributes
		log.Printf("[INFO] %+v", filterState)
		hasDrop := filterState["actions.0.drop.#"] == "1"

		for _, a := range f.Actions {
			switch a.ActionType() {
			case segment.DestinationFilterActionTypeDropEvent:
				assert.True(t, hasDrop)
			case segment.DestinationFilterActionTypeAllowList:
				testAccEventFields(t, filterState, a.(segment.FieldsListEventAction), "allow")
			case segment.DestinationFilterActionTypeBlockList:
				testAccEventFields(t, filterState, a.(segment.FieldsListEventAction), "block")
			case segment.DestinationFilterActionTypeSampling:
				testAccSampledAction(t, filterState, a.(segment.SamplingEventAction))
			default:
				return errors.New("Unrecognized action")
			}
		}

		return nil
	}
}

// Config

func withDestination(filterName string) func(string) string {
	precond := PreCondition{}.WithSource().WithDestination()
	return func(config string) string {
		return precond.Build(func(r PreConditionResources) string {
			return fmt.Sprintf(config, filterName, r.Destinations[0])
		})
	}
}

const testAccDestinationFilterConfigBasic = `
resource "segment_destination_filter" "%s" {
	destination = %s
	title       = "Foo"
	description = "Bar"
	condition   = "context.castPermissions.marketing = false"
	enabled     = true
	actions {
		drop {}
	}
}
`

const testAccDestinationFilterConfigBasicUpdated = `
resource "segment_destination_filter" "%s" {
	destination = %s
	title       = "Foo1"
	description = "Bar1"
	condition   = "context.castPermissions.marketing = true"
	enabled     = false
	actions {
		drop {}
	}
}
`

const testAccDestinationFilterConfigActionsAllowEvents = `
resource "segment_destination_filter" "%s" {
	destination = %s
	title       = "Foo"
	description = "Bar"
	condition   = "context.castPermissions.marketing = false"
	enabled     = true
	actions {
		allow_fields {
			context = ["foo"]
			properties = ["bar"]
			traits = ["baz"]
		}
	}
}
`
const testAccDestinationFilterConfigActionsBlockEvents = `
resource "segment_destination_filter" "%s" {
	destination = %s
	title       = "Foo"
	description = "Bar"
	condition   = "context.castPermissions.marketing = false"
	enabled     = true
	actions {
		allow_fields {
			context = ["foo"]
			properties = ["bar"]
			traits = ["baz"]
		}
		block_fields {
			context = ["one"]
			properties = ["two"]
			traits = ["three"]
		}
	}
}
`

const testAccDestinationFilterConfigActionsOneSampling = `
resource "segment_destination_filter" "%s" {
	destination = %s
	title       = "Foo"
	description = "Bar"
	condition   = "context.castPermissions.marketing = false"
	enabled     = true
	actions {
		allow_fields {
			context = ["foo"]
			properties = ["bar"]
			traits = ["baz"]
		}
		block_fields {
			context = ["one"]
			properties = ["two"]
			traits = ["three"]
		}
		sample {
			percent = 0.5
			path = "userId"
		}
	}
}
`
const testAccDestinationFilterConfigActionsTwoSamplings = `
resource "segment_destination_filter" "%s" {
	destination = %s
	title       = "Foo"
	description = "Bar"
	condition   = "context.castPermissions.marketing = false"
	enabled     = true
	actions {
		allow_fields {
			context = ["foo"]
			properties = ["bar"]
			traits = ["baz"]
		}
		sample {
			percent = 0.8
			path = "properties.price"
		}
		block_fields {
			context = ["one"]
			properties = ["two"]
			traits = ["three"]
		}
		sample {
			percent = 0.5
			path = "userId"
		}
	}
}
`
