package main

<<<<<<< Updated upstream
// func TestFailure(t *testing.T) {
// 	gin.SetMode(gin.TestMode)
// 	mockedHttpClient := mock.NewMockedHTTPClient(mock.WithRequestMatchHandler(
// 		mock.GetOctocat,
// 		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 			http.Error(w, "not found", http.StatusNotFound)
// 		}),
// 	))
// 	client := github.NewClient(mockedHttpClient)
// 	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
// 	r := setupRouter(client, logger)
=======
import (
	"log/slog"
	"os"
	"reflect"
	"testing"
>>>>>>> Stashed changes

// 	req, _ := http.NewRequest("GET", "/", nil)
// 	w := httptest.NewRecorder()
// 	r.ServeHTTP(w, req)

<<<<<<< Updated upstream
// 	assert.Equal(t, http.StatusNotFound, w.Code)
// 	assert.Contains(t, w.Body.String(), "not found")
// }
=======
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
>>>>>>> Stashed changes

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
