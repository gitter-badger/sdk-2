// Copyright 2016 Stratumn SAS. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// that can be found in the LICENSE file.

package storehttp

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/stratumn/go/cs"
	"github.com/stratumn/go/cs/cstesting"
	"github.com/stratumn/go/jsonhttp"
	"github.com/stratumn/go/store"
	"github.com/stratumn/go/store/storetesting"
)

func TestRootOK(t *testing.T) {
	s, a := createServer()
	defer s.Close()

	a.MockGetInfo.Fn = func() (interface{}, error) { return "test", nil }

	var dict map[string]interface{}
	res, err := getJSON(s.URL, &dict)

	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != 200 {
		t.Fatal("unexpected status code")
	}
	if dict["adapter"].(string) != "test" {
		t.Fatal("unexpected adapter dict")
	}
	if a.MockGetInfo.CalledCount != 1 {
		t.Fatal("unexpected number of calls to GetInfo()")
	}
}

func TestRootErr(t *testing.T) {
	s, a := createServer()
	defer s.Close()

	a.MockGetInfo.Fn = func() (interface{}, error) { return "test", errors.New("error") }

	var dict map[string]interface{}
	res, err := getJSON(s.URL, &dict)

	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != jsonhttp.ErrInternalServer.Status {
		t.Fatal("unexpected status code")
	}
	if dict["error"].(string) != jsonhttp.ErrInternalServer.Msg {
		t.Fatal("unexpected error message")
	}
	if a.MockGetInfo.CalledCount != 1 {
		t.Fatal("unexpected number of calls to GetInfo()")
	}
}

func TestSaveSegmentOK(t *testing.T) {
	s, a := createServer()
	defer s.Close()

	a.MockSaveSegment.Fn = func(*cs.Segment) error { return nil }

	s1 := cstesting.RandomSegment()
	var s2 cs.Segment
	res, err := postJSON(s.URL+"/segments", &s2, s1)

	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(a.MockSaveSegment.LastCalledWith, s1) {
		t.Fatal("unexpected argument passed to SaveSegment()")
	}
	if res.StatusCode != 200 {
		t.Fatal("unexpected status code")
	}
	if !reflect.DeepEqual(s1, &s2) {
		t.Fatal("expected segments to be equal")
	}
	if a.MockSaveSegment.CalledCount != 1 {
		t.Fatal("unexpected number of calls to SaveSegment()")
	}
}

func TestSaveSegmentErr(t *testing.T) {
	s, a := createServer()
	defer s.Close()

	a.MockSaveSegment.Fn = func(*cs.Segment) error { return errors.New("test") }

	var dict map[string]interface{}
	res, err := postJSON(s.URL+"/segments", &dict, cstesting.RandomSegment())

	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != jsonhttp.ErrInternalServer.Status {
		t.Fatal("unexpected status code")
	}
	if dict["error"].(string) != jsonhttp.ErrInternalServer.Msg {
		t.Fatal("unexpected error message")
	}
	if a.MockSaveSegment.CalledCount != 1 {
		t.Fatal("unexpected number of calls to SaveSegment()")
	}
}

func TestSaveSegmentInvalidSegment(t *testing.T) {
	s, a := createServer()
	defer s.Close()

	s1 := cstesting.RandomSegment()
	s1.Meta["linkHash"] = true

	var dict map[string]interface{}
	res, err := postJSON(s.URL+"/segments", &dict, s1)

	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != jsonhttp.ErrBadRequest.Status {
		t.Fatal("unexpected status code")
	}
	if dict["error"].(string) != "meta.linkHash should be a non empty string" {
		t.Fatal("unexpected error message")
	}
	if a.MockSaveSegment.CalledCount != 0 {
		t.Fatal("unexpected number of calls to SaveSegment()")
	}
}

func TestSaveSegmentInvalidJSON(t *testing.T) {
	s, a := createServer()
	defer s.Close()

	var dict map[string]interface{}
	res, err := postJSON(s.URL+"/segments", &dict, "1234567890azertyui")

	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != jsonhttp.ErrBadRequest.Status {
		t.Fatal("unexpected status code")
	}
	if dict["error"].(string) != jsonhttp.ErrBadRequest.Msg {
		t.Log(dict["error"].(string))
		t.Log(jsonhttp.ErrBadRequest.Msg)
		t.Fatal("unexpected error message")
	}
	if a.MockSaveSegment.CalledCount != 0 {
		t.Fatal("unexpected number of calls to SaveSegment()")
	}
}

func TestGetSegmentFound(t *testing.T) {
	s, a := createServer()
	defer s.Close()

	s1 := cstesting.RandomSegment()
	a.MockGetSegment.Fn = func(string) (*cs.Segment, error) { return s1, nil }

	var s2 cs.Segment
	res, err := getJSON(s.URL+"/segments/abcde", &s2)

	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(a.MockGetSegment.LastCalledWith, "abcde") {
		t.Fatal("unexpected argument passed to GetSegment()")
	}
	if res.StatusCode != 200 {
		t.Fatal("unexpected status code")
	}
	if !reflect.DeepEqual(s1, &s2) {
		t.Fatal("expected segments to be equal")
	}
	if a.MockGetSegment.CalledCount != 1 {
		t.Fatal("unexpected number of calls to GetSegment()")
	}
}

func TestGetSegmentNotFound(t *testing.T) {
	s, a := createServer()
	defer s.Close()

	var dict map[string]interface{}
	res, err := getJSON(s.URL+"/segments/abcde", &dict)

	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(a.MockGetSegment.LastCalledWith, "abcde") {
		t.Fatal("unexpected argument passed to GetSegment()")
	}
	if res.StatusCode != jsonhttp.ErrNotFound.Status {
		t.Fatal("unexpected status code")
	}
	if dict["error"].(string) != jsonhttp.ErrNotFound.Msg {
		t.Fatal("unexpected error message")
	}
	if a.MockGetSegment.CalledCount != 1 {
		t.Fatal("unexpected number of calls to GetSegment()")
	}
}

func TestGetSegmentErr(t *testing.T) {
	s, a := createServer()
	defer s.Close()

	a.MockGetSegment.Fn = func(string) (*cs.Segment, error) { return nil, errors.New("error") }

	var dict map[string]interface{}
	res, err := getJSON(s.URL+"/segments/abcde", &dict)

	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(a.MockGetSegment.LastCalledWith, "abcde") {
		t.Fatal("unexpected argument passed to GetSegment()")
	}
	if res.StatusCode != jsonhttp.ErrInternalServer.Status {
		t.Fatal("unexpected status code")
	}
	if dict["error"].(string) != jsonhttp.ErrInternalServer.Msg {
		t.Fatal("unexpected error message")
	}
	if a.MockGetSegment.CalledCount != 1 {
		t.Fatal("unexpected number of calls to GetSegment()")
	}
}

func TestDeleteSegmentFound(t *testing.T) {
	s, a := createServer()
	defer s.Close()

	s1 := cstesting.RandomSegment()
	a.MockDeleteSegment.Fn = func(string) (*cs.Segment, error) { return s1, nil }

	var s2 cs.Segment
	res, err := deleteJSON(s.URL+"/segments/abcde", &s2)

	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(a.MockDeleteSegment.LastCalledWith, "abcde") {
		t.Fatal("unexpected argument passed to DeleteSegment()")
	}
	if res.StatusCode != 200 {
		t.Fatal("unexpected status code")
	}
	if !reflect.DeepEqual(s1, &s2) {
		t.Fatal("expected segments to be equal")
	}
	if a.MockDeleteSegment.CalledCount != 1 {
		t.Fatal("unexpected number of calls to DeleteSegment()")
	}
}

func TestDeleteSegmentNotFound(t *testing.T) {
	s, a := createServer()
	defer s.Close()

	var dict map[string]interface{}
	res, err := deleteJSON(s.URL+"/segments/abcde", &dict)

	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(a.MockDeleteSegment.LastCalledWith, "abcde") {
		t.Fatal("unexpected argument passed to DeleteSegment()")
	}
	if res.StatusCode != jsonhttp.ErrNotFound.Status {
		t.Fatal("unexpected status code")
	}
	if dict["error"].(string) != jsonhttp.ErrNotFound.Msg {
		t.Fatal("unexpected error message")
	}
	if a.MockDeleteSegment.CalledCount != 1 {
		t.Fatal("unexpected number of calls to DeleteSegment()")
	}
}

func TestDeleteSegmentErr(t *testing.T) {
	s, a := createServer()
	defer s.Close()

	a.MockDeleteSegment.Fn = func(string) (*cs.Segment, error) { return nil, errors.New("error") }

	var dict map[string]interface{}
	res, err := deleteJSON(s.URL+"/segments/abcde", &dict)

	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(a.MockDeleteSegment.LastCalledWith, "abcde") {
		t.Fatal("unexpected argument passed to DeleteSegment()")
	}
	if res.StatusCode != jsonhttp.ErrInternalServer.Status {
		t.Fatal("unexpected status code")
	}
	if dict["error"].(string) != jsonhttp.ErrInternalServer.Msg {
		t.Fatal("unexpected error message")
	}
	if a.MockDeleteSegment.CalledCount != 1 {
		t.Fatal("unexpected number of calls to DeleteSegment()")
	}
}

func TestFindSegmentsOK(t *testing.T) {
	s, a := createServer()
	defer s.Close()

	var s1 cs.SegmentSlice
	for i := 0; i < 10; i++ {
		s1 = append(s1, cstesting.RandomSegment())
	}
	a.MockFindSegments.Fn = func(*store.Filter) (cs.SegmentSlice, error) { return s1, nil }

	var s2 cs.SegmentSlice
	res, err := getJSON(s.URL+"/segments?offset=1&limit=2&mapId=123&prevLinkHash=abc&tags=one+two", &s2)

	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != 200 {
		t.Fatal("unexpected status code")
	}
	if !reflect.DeepEqual(s1, s2) {
		t.Fatal("expected segment slices to be equal")
	}
	if a.MockFindSegments.CalledCount != 1 {
		t.Fatal("unexpected number of calls to FindSegments()")
	}

	f := a.MockFindSegments.LastCalledWith
	if f.Offset != 1 {
		t.Fatal("unexpected offset")
	}
	if f.Limit != 2 {
		t.Fatal("unexpected limit")
	}
	if f.MapID != "123" {
		t.Fatal("unexpected map ID")
	}
	if f.PrevLinkHash != "abc" {
		t.Fatal("unexpected previous link hash")
	}
	if !reflect.DeepEqual(f.Tags, []string{"one", "two"}) {
		t.Fatal("unexpected tags")
	}
}

func TestFindSegmentsErr(t *testing.T) {
	s, a := createServer()
	defer s.Close()

	a.MockFindSegments.Fn = func(*store.Filter) (cs.SegmentSlice, error) { return nil, errors.New("test") }

	var dict map[string]interface{}
	res, err := getJSON(s.URL+"/segments?offset=1&limit=2&mapId=123&prevLinkHash=abc&tags=one,two", &dict)

	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != jsonhttp.ErrInternalServer.Status {
		t.Fatal("unexpected status code")
	}
	if dict["error"].(string) != jsonhttp.ErrInternalServer.Msg {
		t.Fatal("unexpected error message")
	}
	if a.MockFindSegments.CalledCount != 1 {
		t.Fatal("unexpected number of calls to FindSegments()")
	}
}

func TestFindSegmentsValidation(t *testing.T) {
	s, a := createServer()
	defer s.Close()

	var dict map[string]interface{}
	res, err := getJSON(s.URL+"/segments?offset=hello", &dict)

	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != ErrOffset.Status {
		t.Fatal("unexpected status code")
	}
	if dict["error"].(string) != ErrOffset.Msg {
		t.Fatal("unexpected error message")
	}
	if a.MockFindSegments.CalledCount != 0 {
		t.Fatal("unexpected number of calls to FindSegments()")
	}
}

func TestGetMapIDsOK(t *testing.T) {
	s, a := createServer()
	defer s.Close()

	slice1 := []string{"one", "two", "three"}
	a.MockGetMapIDs.Fn = func(*store.Pagination) ([]string, error) { return slice1, nil }

	var slice2 []string
	res, err := getJSON(s.URL+"/maps?offset=20&limit=10", &slice2)

	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != 200 {
		t.Fatal("unexpected status code")
	}
	if len(slice1) != len(slice2) {
		t.Fatal("expected map ID slices to be have same length")
	}
	for i := 0; i < len(slice1); i++ {
		if !cstesting.ContainsString(slice2, slice1[i]) {
			t.Fatal("expected map ID slices to have same elements")
		}
	}
	if a.MockGetMapIDs.CalledCount != 1 {
		t.Fatal("unexpected number of calls to FindSegments()")
	}

	pagination := a.MockGetMapIDs.LastCalledWith
	if pagination.Offset != 20 {
		t.Fatal("unexpected offset")
	}
	if pagination.Limit != 10 {
		t.Fatal("unexpected limit")
	}
}

func TestGetMapIDsErr(t *testing.T) {
	s, a := createServer()
	defer s.Close()

	a.MockGetMapIDs.Fn = func(*store.Pagination) ([]string, error) { return nil, errors.New("test") }

	var dict map[string]interface{}
	res, err := getJSON(s.URL+"/maps", &dict)

	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != jsonhttp.ErrInternalServer.Status {
		t.Fatal("unexpected status code")
	}
	if dict["error"].(string) != jsonhttp.ErrInternalServer.Msg {
		t.Fatal("unexpected error message")
	}
	if a.MockGetMapIDs.CalledCount != 1 {
		t.Fatal("unexpected number of calls to FindSegments()")
	}
}

func TestGetMapIDsValidation(t *testing.T) {
	s, a := createServer()
	defer s.Close()

	var dict map[string]interface{}
	res, err := getJSON(s.URL+"/maps?limit=-1", &dict)

	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != ErrLimit.Status {
		t.Fatal("unexpected status code")
	}
	if dict["error"].(string) != ErrLimit.Msg {
		t.Fatal("unexpected error message")
	}
	if a.MockGetMapIDs.CalledCount != 0 {
		t.Fatal("unexpected number of calls to FindSegments()")
	}
}

func TestRootNotFound(t *testing.T) {
	s, _ := createServer()
	defer s.Close()

	var dict map[string]interface{}
	res, err := getJSON(s.URL+"/dsfsdf", &dict)

	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != jsonhttp.ErrNotFound.Status {
		t.Fatal("unexpected status code")
	}
	if dict["error"].(string) != jsonhttp.ErrNotFound.Msg {
		t.Fatal("unexpected error message")
	}
}

func createServer() (*httptest.Server, *storetesting.MockAdapter) {
	a := &storetesting.MockAdapter{}
	s := httptest.NewServer(New(a, &jsonhttp.Config{}))

	return s, a
}

func getJSON(url string, target interface{}) (*http.Response, error) {
	return requestJSON(http.MethodGet, url, target, nil)
}

func postJSON(url string, target interface{}, payload interface{}) (*http.Response, error) {
	return requestJSON(http.MethodPost, url, target, payload)
}

func deleteJSON(url string, target interface{}) (*http.Response, error) {
	return requestJSON(http.MethodDelete, url, target, nil)
}

func requestJSON(method, url string, target, payload interface{}) (*http.Response, error) {
	var req *http.Request
	var err error
	var body []byte

	if payload != nil {
		body, err = json.Marshal(payload)
		if err != nil {
			return nil, err
		}

		req, err = http.NewRequest(method, url, bytes.NewReader(body))
	} else {
		req, err = http.NewRequest(method, url, nil)
	}

	if err != nil {
		return nil, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	if err = json.NewDecoder(res.Body).Decode(&target); err != nil {
		return nil, err
	}

	return res, nil
}