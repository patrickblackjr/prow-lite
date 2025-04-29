package logging

import (
	"log/slog"
	"reflect"
	"testing"
)

func TestSetupLogging(t *testing.T) {
	tests := []struct {
		name string
		want *slog.Logger
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SetupLogging(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SetupLogging() = %v, want %v", got, tt.want)
			}
		})
	}
}
