// Copyright 2017 Stratumn SAS. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package storetesting

import (
	"github.com/stratumn/sdk/cs"
	"github.com/stratumn/sdk/store"
	"github.com/stratumn/sdk/types"
)

// MockAdapter is used to mock a store.
//
// It implements github.com/stratumn/sdk/store.Adapter.
type MockAdapter struct {
	// The mock for the GetInfo function.
	MockGetInfo MockGetInfo

	// The mock for the AddDidSaveChannel function.
	MockAddDidSaveChannel MockAddDidSaveChannel

	// The mock for the SaveSegment function.
	MockSaveSegment MockSaveSegment

	// The mock for the GetSegment function.
	MockGetSegment MockGetSegment

	// The mock for the DeleteSegment function.
	MockDeleteSegment MockDeleteSegment

	// The mock for the FindSegments function.
	MockFindSegments MockFindSegments

	// The mock for the GetMapIDs function.
	MockGetMapIDs MockGetMapIDs
}

// MockGetInfo mocks the GetInfo function.
type MockGetInfo struct {
	// The number of times the function was called.
	CalledCount int

	// An optional implementation of the function.
	Fn func() (interface{}, error)
}

// MockAddDidSaveChannel mocks the SaveSegment function.
type MockAddDidSaveChannel struct {
	// The number of times the function was called.
	CalledCount int

	// The segment that was passed to each call.
	CalledWith []chan *cs.Segment

	// The last segment that was passed.
	LastCalledWith chan *cs.Segment

	// An optional implementation of the function.
	Fn func(chan *cs.Segment)
}

// MockSaveSegment mocks the SaveSegment function.
type MockSaveSegment struct {
	// The number of times the function was called.
	CalledCount int

	// The segment that was passed to each call.
	CalledWith []*cs.Segment

	// The last segment that was passed.
	LastCalledWith *cs.Segment

	// An optional implementation of the function.
	Fn func(*cs.Segment) error
}

// MockGetSegment mocks the GetSegment function.
type MockGetSegment struct {
	// The number of times the function was called.
	CalledCount int

	// The link hash that was passed to each call.
	CalledWith []*types.Bytes32

	// The last link hash that was passed.
	LastCalledWith *types.Bytes32

	// An optional implementation of the function.
	Fn func(*types.Bytes32) (*cs.Segment, error)
}

// MockDeleteSegment mocks the DeleteSegment function.
type MockDeleteSegment struct {
	// The number of times the function was called.
	CalledCount int

	// The link hash that was passed to each call.
	CalledWith []*types.Bytes32

	// The last link hash that was passed.
	LastCalledWith *types.Bytes32

	// An optional implementation of the function.
	Fn func(*types.Bytes32) (*cs.Segment, error)
}

// MockFindSegments mocks the FindSegments function.
type MockFindSegments struct {
	// The number of times the function was called.
	CalledCount int

	// The filter that was passed to each call.
	CalledWith []*store.Filter

	// The last filter that was passed.
	LastCalledWith *store.Filter

	// An optional implementation of the function.
	Fn func(*store.Filter) (cs.SegmentSlice, error)
}

// MockGetMapIDs mocks the GetMapIDs function.
type MockGetMapIDs struct {
	// The number of times the function was called.
	CalledCount int

	// The pagination that was passed to each call.
	CalledWith []*store.Pagination

	// The last pagination that was passed.
	LastCalledWith *store.Pagination

	// An optional implementation of the function.
	Fn func(*store.Pagination) ([]string, error)
}

// GetInfo implements github.com/stratumn/sdk/store.Adapter.GetInfo.
func (a *MockAdapter) GetInfo() (interface{}, error) {
	a.MockGetInfo.CalledCount++

	if a.MockGetInfo.Fn != nil {
		return a.MockGetInfo.Fn()
	}

	return nil, nil
}

// AddDidSaveChannel implements
// github.com/stratumn/sdk/store.Adapter.AddDidSaveChannel.
func (a *MockAdapter) AddDidSaveChannel(saveChan chan *cs.Segment) {
	a.MockAddDidSaveChannel.CalledCount++
	a.MockAddDidSaveChannel.CalledWith = append(a.MockAddDidSaveChannel.CalledWith, saveChan)
	a.MockAddDidSaveChannel.LastCalledWith = saveChan

	if a.MockAddDidSaveChannel.Fn != nil {
		a.MockAddDidSaveChannel.Fn(saveChan)
	}
}

// SaveSegment implements github.com/stratumn/sdk/store.Adapter.SaveSegment.
func (a *MockAdapter) SaveSegment(segment *cs.Segment) error {
	a.MockSaveSegment.CalledCount++
	a.MockSaveSegment.CalledWith = append(a.MockSaveSegment.CalledWith, segment)
	a.MockSaveSegment.LastCalledWith = segment

	if a.MockSaveSegment.Fn != nil {
		return a.MockSaveSegment.Fn(segment)
	}

	return nil
}

// GetSegment implements github.com/stratumn/sdk/store.Adapter.GetSegment.
func (a *MockAdapter) GetSegment(linkHash *types.Bytes32) (*cs.Segment, error) {
	a.MockGetSegment.CalledCount++
	a.MockGetSegment.CalledWith = append(a.MockGetSegment.CalledWith, linkHash)
	a.MockGetSegment.LastCalledWith = linkHash

	if a.MockGetSegment.Fn != nil {
		return a.MockGetSegment.Fn(linkHash)
	}

	return nil, nil
}

// DeleteSegment implements github.com/stratumn/sdk/store.Adapter.DeleteSegment.
func (a *MockAdapter) DeleteSegment(linkHash *types.Bytes32) (*cs.Segment, error) {
	a.MockDeleteSegment.CalledCount++
	a.MockDeleteSegment.CalledWith = append(a.MockDeleteSegment.CalledWith, linkHash)
	a.MockDeleteSegment.LastCalledWith = linkHash

	if a.MockDeleteSegment.Fn != nil {
		return a.MockDeleteSegment.Fn(linkHash)
	}

	return nil, nil
}

// FindSegments implements github.com/stratumn/sdk/store.Adapter.FindSegments.
func (a *MockAdapter) FindSegments(filter *store.Filter) (cs.SegmentSlice, error) {
	a.MockFindSegments.CalledCount++
	a.MockFindSegments.CalledWith = append(a.MockFindSegments.CalledWith, filter)
	a.MockFindSegments.LastCalledWith = filter

	if a.MockFindSegments.Fn != nil {
		return a.MockFindSegments.Fn(filter)
	}

	return nil, nil
}

// GetMapIDs implements github.com/stratumn/sdk/store.Adapter.GetMapIDs.
func (a *MockAdapter) GetMapIDs(pagination *store.Pagination) ([]string, error) {
	a.MockGetMapIDs.CalledCount++
	a.MockGetMapIDs.CalledWith = append(a.MockGetMapIDs.CalledWith, pagination)
	a.MockGetMapIDs.LastCalledWith = pagination

	if a.MockGetMapIDs.Fn != nil {
		return a.MockGetMapIDs.Fn(pagination)
	}

	return nil, nil
}
