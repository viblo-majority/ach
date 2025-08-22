// Licensed to The Moov Authors under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. The Moov Authors licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/moov-io/base/log"
	"github.com/stretchr/testify/require"
	kitlog "github.com/go-kit/log"
)

func TestFiles__CreateFileV2EndpointSuccess(t *testing.T) {
	logger := log.NewNopLogger()
	repo := NewRepositoryInMemory(testTTLDuration, logger)
	svc := NewService(repo)
	router := MakeHTTPHandler(svc, repo, kitlog.NewNopLogger())

	// Use a valid ACH file
	fd, err := os.Open(filepath.Join("..", "test", "testdata", "ppd-valid.json"))
	require.NoError(t, err)
	defer fd.Close()

	// test status code
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/files/v2/create", fd)
	req.Header.Set("Origin", "https://moov.io")
	req.Header.Set("X-Request-Id", "test123")
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)
	w.Flush()

	require.Equal(t, http.StatusOK, w.Code)

	var resp createFileV2Response
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)

	require.NotEmpty(t, resp.ID)
	require.NotNil(t, resp.File)
	require.Empty(t, resp.Errors)
}

func TestFiles__CreateFileV2EndpointJSONErr(t *testing.T) {
	logger := log.NewNopLogger()
	repo := NewRepositoryInMemory(testTTLDuration, logger)
	svc := NewService(repo)
	router := MakeHTTPHandler(svc, repo, kitlog.NewNopLogger())

	// Use an invalid ACH file (no batches)
	fd, err := os.Open(filepath.Join("..", "test", "testdata", "ppd-noBatches.json"))
	require.NoError(t, err)
	defer fd.Close()

	// test status code
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/files/v2/create", fd)
	req.Header.Set("Origin", "https://moov.io")
	req.Header.Set("X-Request-Id", "test123")
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)
	w.Flush()

	require.Equal(t, http.StatusBadRequest, w.Code)

	var resp createFileV2Response
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)

	require.NotEmpty(t, resp.ID)
	require.NotEmpty(t, resp.Errors)
	
	// Check that we have structured errors
	require.Len(t, resp.Errors, 1)
	require.NotEmpty(t, resp.Errors[0].ErrorType)
	require.NotEmpty(t, resp.Errors[0].Message)
}

func TestFiles__CreateFileV2EndpointTextErr(t *testing.T) {
	logger := log.NewNopLogger()
	repo := NewRepositoryInMemory(testTTLDuration, logger)
	svc := NewService(repo)
	router := MakeHTTPHandler(svc, repo, kitlog.NewNopLogger())

	// Use an invalid ACH file
	fd, err := os.Open(filepath.Join("..", "test", "issues", "testdata", "issue702.ach"))
	require.NoError(t, err)
	defer fd.Close()

	// test status code
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/files/v2/create", fd)
	req.Header.Set("Origin", "https://moov.io")
	req.Header.Set("X-Request-Id", "test123")
	req.Header.Set("Content-Type", "text/plain")

	router.ServeHTTP(w, req)
	w.Flush()

	require.Equal(t, http.StatusBadRequest, w.Code)

	var resp createFileV2Response
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)

	require.NotEmpty(t, resp.ID)
	require.NotEmpty(t, resp.Errors)
	
	// Check that we have structured errors
	require.NotEmpty(t, resp.Errors[0].ErrorType)
	require.NotEmpty(t, resp.Errors[0].Message)
}

func TestFiles__CreateFileV2EndpointWithInvalidJSON(t *testing.T) {
	logger := log.NewNopLogger()
	repo := NewRepositoryInMemory(testTTLDuration, logger)
	svc := NewService(repo)
	router := MakeHTTPHandler(svc, repo, kitlog.NewNopLogger())

	// Test with completely invalid JSON
	body := bytes.NewReader([]byte(`{"invalid": json}`))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/files/v2/create", body)
	req.Header.Set("Origin", "https://moov.io")
	req.Header.Set("X-Request-Id", "test123")
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)
	w.Flush()

	require.Equal(t, http.StatusBadRequest, w.Code)

	var resp createFileV2Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)

	require.NotEmpty(t, resp.ID)
	require.NotEmpty(t, resp.Errors)
	
	// Verify we have structured error information
	require.NotEmpty(t, resp.Errors[0].ErrorType)
	require.NotEmpty(t, resp.Errors[0].Message)
}