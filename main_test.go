package main

import (
	"log/slog"
	"os"
	"reflect"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v70/github"
)

// 	req, _ := http.NewRequest("GET", "/", nil)
// 	w := httptest.NewRecorder()
// 	r.ServeHTTP(w, req)

func setupLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
}

func Test_setupRouter(t *testing.T) {
	type args struct {
		client *github.Client
		logger *slog.Logger
	}
	tests := []struct {
		name string
		args args
		want *gin.Engine
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := setupRouter(tt.args.client, tt.args.logger); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("setupRouter() = %v, want %v", got, tt.want)
			}
		})
	}
}

// func TestMainRoute(t *testing.T) {
// 	gin.SetMode(gin.TestMode)
// 	mockedHttpClient := mock.NewMockedHTTPClient(mock.WithRequestMatch(mock.GetOctocat))
// 	client := github.NewClient(mockedHttpClient)
// 	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
// 	r := setupRouter(client, logger)

// 	req, _ := http.NewRequest("GET", "/", nil)
// 	w := httptest.NewRecorder()
// 	r.ServeHTTP(w, req)

// 	assert.Equal(t, http.StatusOK, w.Code)
// 	assert.Contains(t, w.Body.String(), "ok")
// }
