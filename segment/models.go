package segment

import "github.com/uswitch/segment-config-go/segment"

type SegmentMetadata struct {
	client    *segment.Client
	workspace string
}
