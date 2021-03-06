package agenttestcases

import (
	"testing"

	cj "github.com/gibson042/canonicaljson-go"
	"github.com/stratumn/sdk/agent/client"
	"github.com/stratumn/sdk/testutil"
	"github.com/stretchr/testify/assert"
)

// TestCreateMapOK tests the client's ability to handle a CreateMap request.
func (f Factory) TestCreateMapOK(t *testing.T) {
	process := "test"
	segment, err := f.Client.CreateMap(process, nil, "test")
	assert.NoError(t, err)
	assert.NotNil(t, segment)
}

// TestCreateMapWithRefs tests the client's ability to handle a CreateMap request
// when one or multiple references are passed.
func (f Factory) TestCreateMapWithRefs(t *testing.T) {
	process := "test"
	refs := []client.SegmentRef{{Process: "other", LinkHash: testutil.RandomHash()}}

	segment, err := f.Client.CreateMap(process, refs, "test")
	assert.NoError(t, err)
	assert.NotNil(t, segment)
	assert.NotNil(t, segment.Link.Meta["refs"])
	want, _ := cj.Marshal(refs)
	got, _ := cj.Marshal(segment.Link.Meta["refs"])
	assert.Equal(t, want, got)
}

// TestCreateMapWithBadRefs tests the client's ability to handle a CreateMap request
// when the provided reference is ill formatted.
func (f Factory) TestCreateMapWithBadRefs(t *testing.T) {
	process, arg := "test", "wrongref"
	refs := []client.SegmentRef{{Process: "wrong"}}

	segment, err := f.Client.CreateMap(process, refs, arg)
	assert.EqualError(t, err, "missing segment or (process and linkHash)")
	assert.Nil(t, segment)
}

// TestCreateMapHandlesWrongInitArgs tests the client's ability to handle a CreateMap request
// when the provided arguments do not match those of the 'init' function.
func (f Factory) TestCreateMapHandlesWrongInitArgs(t *testing.T) {
	process := "test"
	parent, err := f.Client.CreateMap(process, nil)

	assert.EqualError(t, err, "a title is required")
	assert.Nil(t, parent)
}
