package config

import (
	"log/slog"
	"reflect"
	"testing"
)

func TestGetProwLiteConfig(t *testing.T) {
	type args struct {
		logger *slog.Logger
	}
	tests := []struct {
		name string
		args args
		want *ProwLiteConfig
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetProwLiteConfig(tt.args.logger); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetProwLiteConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}
