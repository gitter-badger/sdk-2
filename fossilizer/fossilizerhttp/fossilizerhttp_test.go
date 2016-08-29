// Copyright 2016 Stratumn SAS. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// that can be found in the LICENSE file.

package fossilizerhttp

import (
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stratumn/go/fossilizer"
	"github.com/stratumn/go/jsonhttp"
	"github.com/stratumn/go/testutil"
)

func TestRoot(t *testing.T) {
	s, a := createServer()
	a.MockGetInfo.Fn = func() (interface{}, error) { return "test", nil }

	var body map[string]interface{}
	w, err := testutil.RequestJSON(s.ServeHTTP, "GET", "/", nil, &body)
	if err != nil {
		t.Fatalf("testutil.RequestJSON(): err: %s", err)
	}

	if got, want := w.Code, http.StatusOK; want != got {
		t.Errorf("w.StatusCode = %d want %d", got, want)
	}
	if got, want := body["adapter"].(string), "test"; got != want {
		t.Errorf(`body["adapter"] = %q want %q`, got, want)
	}
	if got, want := a.MockGetInfo.CalledCount, 1; got != want {
		t.Errorf("a.MockGetInfo.CalledCount = %d want %d", got, want)
	}
}

func TestRoot_err(t *testing.T) {
	s, a := createServer()
	a.MockGetInfo.Fn = func() (interface{}, error) { return "test", errors.New("error") }

	var body map[string]interface{}
	w, err := testutil.RequestJSON(s.ServeHTTP, "GET", "/", nil, &body)
	if err != nil {
		t.Fatalf("testutil.RequestJSON(): err: %s", err)
	}

	if got, want := w.Code, jsonhttp.NewErrInternalServer("").Status(); want != got {
		t.Errorf("w.Code = %d want %d", got, want)
	}
	if got, want := body["error"].(string), jsonhttp.NewErrInternalServer("").Error(); got != want {
		t.Errorf(`body["error"] = %q want %q`, got, want)
	}
	if got, want := a.MockGetInfo.CalledCount, 1; got != want {
		t.Errorf("a.MockGetInfo.CalledCount = %d want %d", got, want)
	}
}

func TestFossilize(t *testing.T) {
	s, a := createServer()
	l, err := net.Listen("tcp", ":6666")
	if err != nil {
		t.Fatalf("net.Listen(): err: %s", err)
	}
	h := &resultHandler{t: t, listener: l, want: "\"it is known\""}

	go func() {
		defer l.Close()
		rc := a.MockAddResultChan.LastCalledWith
		a.MockFossilize.Fn = func(data []byte, meta []byte) error {
			rc <- &fossilizer.Result{
				Evidence: "it is known",
				Data:     data,
				Meta:     meta,
			}
			return nil
		}

		req := httptest.NewRequest("POST", "/fossils", nil)
		req.Form = url.Values{}
		req.Form.Set("data", "1234567890")
		req.Form.Set("callbackUrl", "http://localhost:6666")

		w := httptest.NewRecorder()
		s.ServeHTTP(w, req)

		if got, want := w.Code, http.StatusOK; want != got {
			t.Errorf("w.Code = %d want %d", got, want)
		}

		sleep := 2 * time.Second
		time.Sleep(sleep)
		t.Errorf("callback URL not called after %s", sleep)
	}()

	http.Serve(l, h)
}

func TestFossilize_noData(t *testing.T) {
	s, _ := createServer()

	req := httptest.NewRequest("POST", "/fossils", nil)
	req.Form = url.Values{}
	req.Form.Set("callbackUrl", "http://localhost:6666")

	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)

	if got, want := w.Code, newErrData("").Status(); want != got {
		t.Errorf("w.Code = %d want %d", got, want)
	}
}

func TestFossilize_dataTooShort(t *testing.T) {
	s, _ := createServer()

	req := httptest.NewRequest("POST", "/fossils", nil)
	req.Form = url.Values{}
	req.Form.Set("callbackUrl", "http://localhost:6666")
	req.Form.Set("data", "1")

	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)

	if got, want := w.Code, newErrData("").Status(); want != got {
		t.Errorf("w.Code = %d want %d", got, want)
	}
}

func TestFossilize_dataTooLong(t *testing.T) {
	s, _ := createServer()

	req := httptest.NewRequest("POST", "/fossils", nil)
	req.Form = url.Values{}
	req.Form.Set("callbackUrl", "http://localhost:6666")
	req.Form.Set("data", "12345678901234567890")

	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)

	if got, want := w.Code, newErrData("").Status(); want != got {
		t.Errorf("w.Code = %d want %d", got, want)
	}
}

func TestFossilize_dataNotHex(t *testing.T) {
	s, _ := createServer()

	req := httptest.NewRequest("POST", "/fossils", nil)
	req.Form = url.Values{}
	req.Form.Set("callbackUrl", "http://localhost:6666")
	req.Form.Set("data", "azertyuiop")

	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)

	if got, want := w.Code, newErrData("").Status(); want != got {
		t.Errorf("w.Code = %d want %d", got, want)
	}
}

func TestFossilize_noCallback(t *testing.T) {
	s, _ := createServer()

	req := httptest.NewRequest("POST", "/fossils", nil)
	req.Form = url.Values{}
	req.Form.Set("data", "1234567890")

	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)

	if got, want := w.Code, http.StatusBadRequest; want != got {
		t.Errorf("w.Code = %d want %d", got, want)
	}
}

func TestFossilize_noBody(t *testing.T) {
	s, _ := createServer()

	req := httptest.NewRequest("POST", "/fossils", nil)

	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)

	if got, want := w.Code, http.StatusBadRequest; want != got {
		t.Errorf("w.Code = %d want %d", got, want)
	}
}

func TestNotFound(t *testing.T) {
	s, _ := createServer()

	var body map[string]interface{}
	w, err := testutil.RequestJSON(s.ServeHTTP, "GET", "/azerty", nil, &body)
	if err != nil {
		t.Fatalf("testutil.RequestJSON(): err: %s", err)
	}

	if got, want := w.Code, jsonhttp.NewErrNotFound("").Status(); want != got {
		t.Errorf("w.Code = %d want %d", got, want)
	}
	if got, want := body["error"].(string), jsonhttp.NewErrNotFound("").Error(); got != want {
		t.Errorf(`body["error"] = %q want %q`, got, want)
	}
}
