package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInfo_ReturnsOK(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/_info", nil)

	(&Handler{}).info(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}
