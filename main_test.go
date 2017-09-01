package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	log "github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestRawOutput(t *testing.T) {
	log, status := run(t, "text/plain", "a-b-c")
	assert.Equal(t, http.StatusOK, status)

	assert.Contains(t, log, "\"msg\":\"a-b-c\"")
	assert.Contains(t, log, "test-ns")
	assert.Contains(t, log, "test-ot")
	assert.Contains(t, log, "test-on")
}

func TestJSONOutput(t *testing.T) {
	log, status := run(t, "application/json", "{\"a\":1}")
	assert.Equal(t, http.StatusOK, status)

	assert.Contains(t, log, "{\"a\":1,")
	assert.Contains(t, log, "test-ns")
	assert.Contains(t, log, "test-ot")
	assert.Contains(t, log, "test-on")
}

func TestJSONOutputMsgFromOriginal(t *testing.T) {
	log, status := run(t, "application/json", "{\"msg\":\"a\"}")
	assert.Equal(t, http.StatusOK, status)

	assert.Contains(t, log, "\"msg\":\"a\",")
	assert.NotContains(t, log, "\"fields.msg\"")
	assert.Contains(t, log, "test-ns")
	assert.Contains(t, log, "test-ot")
	assert.Contains(t, log, "test-on")
}

func TestBadJSONOutput(t *testing.T) {
	log, status := run(t, "application/json", "{\"a:1}")
	assert.Equal(t, http.StatusBadRequest, status)

	assert.Contains(t, log, "\"msg\":\"{\\\"a:1}")
	assert.Contains(t, log, "test-ns")
	assert.Contains(t, log, "test-ot")
	assert.Contains(t, log, "test-on")
}

func run(t *testing.T, contenttype, content string) (string, int) {
	buf := bytes.NewBuffer([]byte{})

	format := new(log.JSONFormatter)
	format.TimestampFormat = "2006-01-02 15:04:05"

	logger := &log.Logger{
		Out:       buf,
		Formatter: format,
		Hooks:     make(log.LevelHooks),
		Level:     log.InfoLevel,
	}

	r := CreateRouter(logger)
	ts := httptest.NewServer(r)
	defer ts.Close()

	url := ts.URL + "/log/test-ns/test-ot/test-on"
	resp, err := http.Post(url, contenttype, bytes.NewBufferString(content))
	if err != nil {
		t.Fatal(err)
	}

	return buf.String(), resp.StatusCode
}
