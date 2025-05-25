package ping

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPing(t *testing.T) {
	tests := []struct {
		name         string
		mockErr      error
		expectedCode int
		expectedBody string
	}{
		{
			name:         "success ping",
			mockErr:      nil,
			expectedCode: http.StatusOK,
		},
		{
			name:         "error ping",
			mockErr:      errors.New("db down"),
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewMockPingService(t)
			s.EXPECT().Ping(context.Background()).Return(tt.mockErr)

			handler := Ping(s)
			req := httptest.NewRequest(http.MethodGet, "/ping", nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedCode, resp.StatusCode)
		})
	}
}
