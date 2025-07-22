package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigLoad(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr string
		want    Config
		env     map[string]string
	}{
		{
			name: "empty",
			path: "testdata/empty.yaml",
			want: Default(),
		},
		{
			name:    "bad",
			path:    "testdata/_not_found.yaml",
			wantErr: "no such file or directory",
			want:    Default(),
		},
		{
			name: "override",
			path: "testdata/override.yaml",
			want: Config{
				LLM: LLMConfig{
					ModelName:      "gemini-2.5-flash",
					PromptTemplate: "custom template",
				},
			},
		},
		{
			name: "gemini env",
			path: "testdata/empty.yaml",
			env: map[string]string{
				"GOOGLE_GENAI_API_KEY": "xyz-123",
			},
			want: Config{
				LLM: LLMConfig{
					Provider:       "googleai",
					ModelName:      "gemini-2.0-flash-lite",
					PromptTemplate: defaultPromptTemplate,
				},
			},
		},
		{
			name: "openai env",
			path: "testdata/openai.yaml",
			env: map[string]string{
				"OPENAI_API_KEY": "xyz-123",
			},
			want: Config{
				LLM: LLMConfig{
					Provider:       "openai",
					ModelName:      "gpt-4.1-mini",
					PromptTemplate: defaultPromptTemplate,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.env {
				t.Setenv(k, v)
			}

			got, err := LoadFrom(tt.path)
			if tt.wantErr != "" {
				assert.ErrorContains(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSetUnsetOption(t *testing.T) {
	t.Run("set new", func(t *testing.T) {
		lines := []string{"FOO=bar"}
		lines = setOption(lines, "BAZ", "qux")
		assert.Equal(t, []string{"FOO=bar", `BAZ="qux"`}, lines)
	})

	t.Run("set existing", func(t *testing.T) {
		lines := []string{"FOO=bar"}
		lines = setOption(lines, "FOO", "qux")
		assert.Equal(t, []string{`FOO="qux"`}, lines)
	})

	t.Run("set commented", func(t *testing.T) {
		lines := []string{"# FOO=bar"}
		lines = setOption(lines, "FOO", "qux")
		assert.Equal(t, []string{`FOO="qux"`}, lines)
	})

	t.Run("set with spaces", func(t *testing.T) {
		lines := []string{"  FOO = bar  "}
		lines = setOption(lines, "FOO", "qux")
		assert.Equal(t, []string{`FOO="qux"`}, lines)
	})

	t.Run("unset", func(t *testing.T) {
		lines := []string{"FOO=bar", "BAZ=qux"}
		lines = unsetOption(lines, "FOO")
		assert.Equal(t, []string{"#FOO=bar", "BAZ=qux"}, lines)
	})

	t.Run("unset commented", func(t *testing.T) {
		lines := []string{"# FOO=bar"}
		lines = unsetOption(lines, "FOO")
		assert.Equal(t, []string{"# FOO=bar"}, lines)
	})

	t.Run("unset non-existing", func(t *testing.T) {
		lines := []string{"FOO=bar"}
		lines = unsetOption(lines, "BAZ")
		assert.Equal(t, []string{"FOO=bar"}, lines)
	})
}
