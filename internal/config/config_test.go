package config

import (
	"crypto/rsa"
	"encoding/json"
	"os"
	"path"
	"testing"

	"dario.cat/mergo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	pathToPriv       = "../testdata/keys/priv.pem"
	pathToPub        = "../testdata/keys/pub.pem"
	pathToInvalidPEM = "../testdata/invalid.pem"
)

type commonTestArgs struct {
	name              string
	jsonStr           string
	flags             []string
	env               []string
	json              map[string]any
	want              CommonConfig
	jsonToEnv         bool
	jsonToFlagsC      bool
	jsonToFlagsConfig bool
}

func runCommonConfigTests(t *testing.T, functest func(flags []string) CommonConfig) {
	isDevOnlyCfg := CommonConfig{
		IsDev: true,
	}

	isDevWithAddr := CommonConfig{
		// Addr:  "localhost:8080",
		IsDev: true,
	}

	tests := []commonTestArgs{
		{
			name: "Env vars only",
			env:  []string{"IS_DEV", "true"},
			want: isDevOnlyCfg,
		},
		{
			name:  "Flags only",
			flags: []string{"-dev"},
			want:  isDevOnlyCfg,
		},
		{
			name:      "JSON from env",
			json:      map[string]any{"dev": true},
			want:      isDevOnlyCfg,
			jsonToEnv: true,
		},
		{
			name:              "JSON from config flag",
			json:              map[string]any{"dev": true},
			want:              isDevOnlyCfg,
			jsonToFlagsConfig: true,
		},
		{
			name:         "JSON from c flag",
			json:         map[string]any{"dev": true},
			want:         isDevOnlyCfg,
			jsonToFlagsC: true,
		},
		{
			name: "Flags && JSON",
			json: map[string]any{
				"dev":     true,
				"address": "localhost:8000",
			},
			flags:        []string{"-a", "localhost:8080"},
			want:         isDevWithAddr,
			jsonToFlagsC: true,
		},
		{
			name: "JSON && Env vars",
			json: map[string]any{
				"dev":     true,
				"address": "localhost:8000",
			},
			env:          []string{"ADDRESS", "localhost:8080"},
			want:         isDevWithAddr,
			jsonToFlagsC: true,
		},
		{
			name: "JSON && Flags",
			flags: []string{
				"-a", "localhost:8080", "-dev",
			},
			env:  []string{"ADDRESS", "localhost:8080"},
			want: isDevWithAddr,
		},
		{
			name: "JSON && Flags && Env",
			json: map[string]any{
				"log_lvl": "error",
				"address": "localhost:8888",
			},
			flags: []string{
				"-a", "localhost:8000", "-dev",
			},
			env: []string{"ADDRESS", "localhost:8080"},
			want: CommonConfig{
				// Addr:   "localhost:8080",
				IsDev:  true,
				LogLvl: "error",
			},
			jsonToFlagsC: true,
		},
		{
			name:         "Invalid json",
			jsonStr:      `{"dev":true}}}`,
			jsonToFlagsC: true,
		},
		{
			name: "Invalid json path",
			flags: []string{
				"-c", "wrong path",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				f     *os.File
				err   error
				fPath string
			)
			if tt.jsonToEnv || tt.jsonToFlagsC || tt.jsonToFlagsConfig {
				dir := os.TempDir()
				fPath = path.Join(dir, "config.json")
				f, err = os.OpenFile(fPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
				require.NoError(t, err)
				defer f.Close()
				defer os.Remove(fPath)

				switch {
				case len(tt.json) > 0:
					var jsonBytes []byte
					jsonBytes, err = json.Marshal(tt.json)
					require.NoError(t, err)
					_, err = f.Write(jsonBytes)
					require.NoError(t, err)
				case tt.jsonStr != "":
					_, err = f.Write([]byte(tt.jsonStr))
					require.NoError(t, err)
				default:
					_, err = f.Write([]byte("{}"))
					require.NoError(t, err)
				}
			}

			if tt.jsonToEnv {
				tt.env = append(tt.env, "CONFIG", fPath)
			}
			if tt.jsonToFlagsC {
				tt.flags = append(tt.flags, "-c", fPath)
			}
			if tt.jsonToFlagsConfig {
				tt.flags = append(tt.flags, "-config", fPath)
			}

			for i := 0; i < len(tt.env)-1; i += 2 {
				t.Setenv(tt.env[i], tt.env[i+1])
			}

			mergo.Merge(&tt.want, defaultCommon)
			cfg := functest(tt.flags)

			cfg.Config = ""
			tt.want.Config = ""

			assert.Equal(t, tt.want, cfg)
		})
	}
}

func TestLoadCommonServer(t *testing.T) {
	runCommonConfigTests(t, func(flags []string) CommonConfig {
		cfg := LoadServerCfg(flags...)
		return cfg.CommonConfig
	})
}

func TestLoadCommonAgent(t *testing.T) {
	runCommonConfigTests(t, func(flags []string) CommonConfig {
		cfg := LoadAgentCfg(flags...)
		return cfg.CommonConfig
	})
}

func TestLoadServerCfg(t *testing.T) {
	tests := []struct {
		name string
		env  []string
		want ServerConfig
	}{
		{
			name: "DatabaseDSN disables restore",
			env:  []string{"DATABASE_DSN", "postgresql://localhost"},
			want: ServerConfig{
				DatabaseDSN: "postgresql://localhost",
				Restore:     false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for i := 0; i < len(tt.env)-1; i += 2 {
				t.Setenv(tt.env[i], tt.env[i+1])
			}

			mergo.Merge(&tt.want, defaultServer)
			cfg := LoadServerCfg()

			cfg.Config = ""
			tt.want.Config = ""

			assert.Equal(t, tt.want, cfg)
		})
	}
}

func TestParsePrivateKey(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		wantPanic bool
	}{
		{
			name: "valid priv",
			path: pathToPriv,
		},
		{
			name:      "Invalid path",
			path:      "invalid",
			wantPanic: true,
		},
		{
			name:      "Invalid PEM",
			path:      pathToInvalidPEM,
			wantPanic: true,
		},
		{
			name:      "Public Key Instead of Private",
			path:      pathToPub,
			wantPanic: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var key *rsa.PrivateKey
			testF := func() {
				key = parsePrivateKey(tt.path)
			}
			if tt.wantPanic {
				assert.Panics(t, testF)
				assert.Nil(t, key)
			} else {
				assert.NotPanics(t, testF)
				assert.NotNil(t, key)
			}
		})
	}
}

func TestParsePublicKey(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		wantPanic bool
	}{
		{
			name: "valid pub",
			path: pathToPub,
		},
		{
			name:      "Invalid path",
			path:      "invalid",
			wantPanic: true,
		},
		{
			name:      "Invalid PEM",
			path:      pathToInvalidPEM,
			wantPanic: true,
		},
		{
			name:      "Private Key Instead of Public",
			path:      pathToPriv,
			wantPanic: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var key *rsa.PublicKey
			testF := func() {
				key = parsePublicKey(tt.path)
			}
			if tt.wantPanic {
				assert.Panics(t, testF)
				assert.Nil(t, key)
			} else {
				assert.NotPanics(t, testF)
				assert.NotNil(t, key)
			}
		})
	}
}
