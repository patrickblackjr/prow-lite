package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"log/slog"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v69/github"
	"github.com/migueleliasweb/go-github-mock/src/mock"
	"github.com/stretchr/testify/assert"
)

func TestFailure(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockedHttpClient := mock.NewMockedHTTPClient(mock.WithRequestMatchHandler(
		mock.GetOctocat,
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "not found", http.StatusNotFound)
		}),
	))
	client := github.NewClient(mockedHttpClient)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	r := setupRouter(client, logger)

	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "not found")
}

func TestMainRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockedHttpClient := mock.NewMockedHTTPClient(mock.WithRequestMatch(mock.GetOctocat))
	client := github.NewClient(mockedHttpClient)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	r := setupRouter(client, logger)

	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "ok")
}
