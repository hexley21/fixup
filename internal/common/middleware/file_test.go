package middleware_test

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hexley21/fixup/internal/common/middleware"
	"github.com/hexley21/fixup/pkg/http/rest"
	"github.com/stretchr/testify/assert"
)

var (
	maxSize int64 = 10 << 20
	files         = map[string][]byte{
		"file1.txt": []byte("dummy content 1"),
		"file2.txt": []byte("dummy content 2"),
		"file3.txt": []byte("dummy content 3"),
	}
)

const (
	HeaderContentType = "Content-Type"
)

func createMultipartMultipleFiles(t *testing.T, form string, files map[string][]byte) (*bytes.Buffer, string) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	for fileName, fileContent := range files {
		part, err := writer.CreateFormFile(form, fileName)
		if err != nil {
			t.Fatal(err)
		}

		_, err = part.Write(fileContent)
		if err != nil {
			t.Fatal(err)
		}
	}

	err := writer.Close()
	if err != nil {
		t.Fatal(err)
	}

	return body, writer.FormDataContentType()
}

func TestAllowFilesAmount_ExactFiles(t *testing.T) {
	body, contentType := createMultipartMultipleFiles(t, "files", files)

	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.Header.Set(HeaderContentType, contentType)
	rec := httptest.NewRecorder()

	factory.NewAllowFilesAmount(maxSize, "files", len(files))(BasicHandler()).ServeHTTP(rec, req)

	assert.Equal(t, "ok", rec.Body.String())
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestAllowFilesAmount_TooManyFiles(t *testing.T) {
	body, contentType := createMultipartMultipleFiles(t, "files", files)

	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.Header.Set(HeaderContentType, contentType)
	rec := httptest.NewRecorder()

	factory.NewAllowFilesAmount(maxSize, "files", 1)(BasicHandler()).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, middleware.MsgTooManyFiles, errResp.Message)
		assert.Equal(t, http.StatusBadRequest, errResp.Status)
	}
}

func TestAllowFilesAmount_NotEnoughFiles(t *testing.T) {
	body, contentType := createMultipartMultipleFiles(t, "files", files)

	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.Header.Set(HeaderContentType, contentType)
	rec := httptest.NewRecorder()

	factory.NewAllowFilesAmount(maxSize, "files", 4)(BasicHandler()).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, middleware.MsgNotEnoughFiles, errResp.Message)
		assert.Equal(t, http.StatusBadRequest, errResp.Status)
	}
}

func TestAllowFilesAmount_NoFile(t *testing.T) {
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.Header.Set(HeaderContentType, writer.FormDataContentType())
	rec := httptest.NewRecorder()

	factory.NewAllowFilesAmount(maxSize, "files", 1)(BasicHandler()).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, middleware.MsgNoFile, errResp.Message)
		assert.Equal(t, http.StatusBadRequest, errResp.Status)
	}
}

func TestAllowContentType_ValidContentType(t *testing.T) {
	body, contentType := createMultipartMultipleFiles(t, "files", files)

	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.Header.Set(HeaderContentType, contentType)
	rec := httptest.NewRecorder()

	factory.NewAllowContentType(maxSize, "files", "application/octet-stream")(BasicHandler()).ServeHTTP(rec, req)

	assert.Equal(t, "ok", rec.Body.String())
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestAllowContentType_InvalidContentType(t *testing.T) {
	body, contentType := createMultipartMultipleFiles(t, "files", files)
	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.Header.Set(HeaderContentType, contentType)
	rec := httptest.NewRecorder()

	factory.NewAllowContentType(maxSize, "files", "image/jpeg")(BasicHandler()).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, http.StatusBadRequest, errResp.Status)
	}
}

func TestAllowContentType_MissingContentType(t *testing.T) {
	body, _ := createMultipartMultipleFiles(t, "files", files)

	req := httptest.NewRequest(http.MethodPost, "/", body)
	rec := httptest.NewRecorder()

	factory.NewAllowContentType(maxSize, "files", "image/jpeg")(BasicHandler()).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, rest.MsgFileReadError, errResp.Message)
		assert.Equal(t, http.StatusBadRequest, errResp.Status)
	}
}
