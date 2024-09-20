package middleware_test

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hexley21/fixup/internal/common/middleware"
	"github.com/hexley21/fixup/internal/common/rest"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

var (
	files = map[string][]byte{
        "file1.txt": []byte("dummy content 1"),
        "file2.txt": []byte("dummy content 2"),
        "file3.txt": []byte("dummy content 3"),
    }
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
	req.Header.Set(echo.HeaderContentType, contentType)
	rec := httptest.NewRecorder()

	e := echo.New()
	c := e.NewContext(req, rec)

	mw := middleware.AllowFilesAmount("files", len(files))

	assert.NoError(t, mw(BasicHandler)(c))
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestAllowFilesAmount_TooManyFiles(t *testing.T) {
	body, contentType := createMultipartMultipleFiles(t, "files", files)

	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.Header.Set(echo.HeaderContentType, contentType)
	rec := httptest.NewRecorder()

	e := echo.New()
	c := e.NewContext(req, rec)

	mw := middleware.AllowFilesAmount("files", 1)

	var errResp *rest.ErrorResponse
	if assert.ErrorAs(t, mw(BasicHandler)(c), &errResp) {
		assert.Equal(t, rest.MsgTooManyFiles, errResp.Message)
		assert.Equal(t, http.StatusBadRequest, errResp.Status)
	}
}

func TestAllowFilesAmount_NotEnoughFiles(t *testing.T) {
	body, contentType := createMultipartMultipleFiles(t, "files", files)

	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.Header.Set(echo.HeaderContentType, contentType)
	rec := httptest.NewRecorder()
	
	e := echo.New()
	c := e.NewContext(req, rec)

	mw := middleware.AllowFilesAmount("files", 4)

	var errResp *rest.ErrorResponse
	if assert.ErrorAs(t, mw(BasicHandler)(c), &errResp) {
		assert.Equal(t, rest.MsgNotEnoughFiles, errResp.Message)
		assert.Equal(t, http.StatusBadRequest, errResp.Status)
	}
}

func TestAllowFilesAmount_NoFile(t *testing.T) {
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.Header.Set(echo.HeaderContentType, writer.FormDataContentType())
	rec := httptest.NewRecorder()

	e := echo.New()
	c := e.NewContext(req, rec)

	mw := middleware.AllowFilesAmount("files", 1)

	var errResp *rest.ErrorResponse
	if assert.ErrorAs(t, mw(BasicHandler)(c), &errResp) {
		assert.Equal(t, rest.MsgNoFile, errResp.Message)
		assert.Equal(t, http.StatusBadRequest, errResp.Status)
	}
}

func TestAllowContentType_ValidContentType(t *testing.T) {
	body, contentType := createMultipartMultipleFiles(t, "files", files)
	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.Header.Set(echo.HeaderContentType, contentType)
	rec := httptest.NewRecorder()

	e := echo.New()
	c := e.NewContext(req, rec)

	mw := middleware.AllowContentType("files", "application/octet-stream")

	assert.NoError(t, mw(BasicHandler)(c))
	assert.Equal(t, http.StatusOK, rec.Code)
}


func TestAllowContentType_InvalidContentType(t *testing.T) {
	body, contentType := createMultipartMultipleFiles(t, "files", files)
	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.Header.Set(echo.HeaderContentType, contentType)
	rec := httptest.NewRecorder()

	e := echo.New()
	c := e.NewContext(req, rec)

	mw := middleware.AllowContentType("files", "image/jpeg")
	
	var errResp *rest.ErrorResponse
	if assert.ErrorAs(t, mw(BasicHandler)(c), &errResp) {
		assert.Equal(t, http.StatusBadRequest, errResp.Status)
	}
}

func TestAllowContentType_MissingContentType(t *testing.T) {
	body, _ := createMultipartMultipleFiles(t, "files", files)
	req := httptest.NewRequest(http.MethodPost, "/", body)
	rec := httptest.NewRecorder()

	e := echo.New()
	c := e.NewContext(req, rec)

	mw := middleware.AllowContentType("files", "image/jpeg")
	
	var errResp *rest.ErrorResponse
	if assert.ErrorAs(t, mw(BasicHandler)(c), &errResp) {
		assert.Equal(t, rest.MsgFileReadError, errResp.Message)
		assert.Equal(t, http.StatusBadRequest, errResp.Status)
	}
}
