package config

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParseServerConfig(t *testing.T) {
	var testCases = map[string]struct {
		filename string
		isErr    bool
		config   ServerConfig
	}{
		"doNotExist": {
			filename: "shouldNotExist",
			isErr:    true,
		},
		"notYAML": {
			filename: "testdata/bad.notYAML",
			isErr:    true,
		},
		"good": {
			filename: "testdata/good.server.yml",
			config: ServerConfig{
				Postgres: PGConfig{
					ConnStr: "postgres://@HOST:PORT?user=USER&password=PASSWORD&dbname=DBNAME&sslmode=enabled",
					MaxIdle: 2,
					MaxOpen: 4,
				},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			parsed, err := ParseServerConfig(tc.filename)
			switch {
			case tc.isErr && err == nil:
				t.Fatalf("want error, got none")
			case !tc.isErr && err != nil:
				t.Fatalf("want no error, got %v", err)
			case !tc.isErr && err == nil:
				// good, NOP
			}

			if want, got := tc.config, parsed; !cmp.Equal(want, got) {
				t.Fatalf("parsed config:\n want= %v,\n  got= %v,\n diff= %v\n", want, got, cmp.Diff(want, got))
			}
		})
	}
}
