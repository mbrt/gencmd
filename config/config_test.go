package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigLoad(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
		want    Config
	}{
		{
			name:    "empty",
			path:    "testdata/empty.yaml",
			wantErr: false,
			want:    DefaultConfig(),
		},
		{
			name:    "empty",
			path:    "testdata/override.yaml",
			wantErr: false,
			want: Config{
				LLM: LLMConfig{
					ModelName:      "gemini-2.5-flash",
					PromptTemplate: "custom template",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := LoadFrom(tt.path)
			if tt.wantErr {
				assert.Error(t, err)
			}
			if got.LLM.AutoFromEnv != tt.want.LLM.AutoFromEnv || got.LLM.ModelName != tt.want.LLM.ModelName {
				t.Errorf("LoadConfig() = %v, want %v", got, tt.want)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
