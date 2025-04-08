package main

import (
	"log/slog"
	"reflect"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v70/github"
)

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

func Test_main(t *testing.T) {
	tests := []struct {
		name string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			main()
		})
	}
}
