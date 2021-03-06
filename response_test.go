package typhon

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResponseWriter(t *testing.T) {
	t.Parallel()
	req := NewRequest(nil, "GET", "/abc", nil)

	// Using NewResponse, vanilla
	r := NewResponse(req)
	r.Write([]byte("boop"))
	b, _ := r.BodyBytes(true)
	assert.Equal(t, []byte("boop"), b)

	// Using NewResponse, via ResponseWriter
	r = NewResponse(req)
	r.Writer().Header().Set("abc", "def")
	r.Writer().WriteHeader(http.StatusForbidden) // Test some other fun stuff while we're here
	r.Writer().Write([]byte("boop"))
	b, _ = r.BodyBytes(true)
	assert.Equal(t, []byte("boop"), b)
	assert.Equal(t, http.StatusForbidden, r.StatusCode)
	assert.Equal(t, "def", r.Header.Get("abc"))

	// Using NewResponse, vanilla and then ResponseWriter
	r = NewResponse(req)
	r.Write([]byte("boop"))
	r.Writer().Write([]byte("woop"))
	b, _ = r.BodyBytes(true)
	assert.Equal(t, []byte("boopwoop"), b)
}

func TestResponseWriter_Error(t *testing.T) {
	t.Parallel()
	req := NewRequest(nil, "GET", "/", nil)
	rsp := NewResponse(req)
	rsp.Writer().WriteError(errors.New("abc"))
	assert.Error(t, rsp.Error)
	assert.Equal(t, "abc", rsp.Error.Error())
}

// TestResponseDecodeCloses verifies that a response body is closed after calling Decode()
func TestResponseDecodeCloses(t *testing.T) {
	t.Parallel()
	req := NewRequest(nil, "GET", "/", nil)
	rsp := NewResponse(req)
	b := []byte("{\"a\":\"b\"}\n")
	r := newDoneReader(ioutil.NopCloser(bytes.NewReader(b)), -1)
	rsp.Body = r

	bout := map[string]string{}
	rsp.Decode(&bout)
	select {
	case <-r.closed:
	default:
		assert.Fail(t, "response body was not closed after Decode()")
	}
}
