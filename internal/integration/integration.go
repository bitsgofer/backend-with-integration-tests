package integration

import (
	"bytes"
	"database/sql"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"

	"github.com/bitsgofer/backend-with-integration-tests/internal/config"
)

type Setup struct {
	t      *testing.T
	logger log.Logger // logs events from integration test setup

	Logger       log.Logger // for packages to use
	ServerConfig config.ServerConfig
}

type Option func(*Setup)

func PrintDebugLog(s *Setup) {
	s.logger = log.NewJSONLogger(os.Stderr)
	s.logger = log.With(s.logger,
		"from", "integration.Setup",
		"ts", log.DefaultTimestampUTC,
		"caller", log.DefaultCaller,
	)
	s.logger = level.NewFilter(s.logger, level.AllowDebug())

	s.Logger = log.NewJSONLogger(os.Stderr)
	s.Logger = log.With(s.Logger,
		"ts", log.DefaultTimestampUTC,
		"caller", log.DefaultCaller,
	)
	s.Logger = level.NewFilter(s.Logger, level.AllowDebug())
}

func New(t *testing.T, opts ...Option) *Setup {
	s := &Setup{
		t: t,
	}
	s.logger = log.NewJSONLogger(os.Stderr)
	s.logger = log.With(s.logger,
		"from", "integration.Setup",
		"ts", log.DefaultTimestampUTC,
		"caller", log.DefaultCaller,
	)
	s.logger = level.NewFilter(s.logger, level.AllowError())

	s.Logger = log.NewJSONLogger(os.Stderr)
	s.Logger = log.With(s.Logger,
		"ts", log.DefaultTimestampUTC,
		"caller", log.DefaultCaller,
	)
	s.Logger = level.NewFilter(s.Logger, level.AllowError())

	// apply options
	for _, opt := range opts {
		opt(s)
	}

	// generate config from templates using new, clean spaces for data
	rngPrefix := randNamespace(t.Name())
	s.setupFromTemplate("../../config.d/test.server.yml", rngPrefix)

	return s
}

func (s *Setup) setupFromTemplate(templateFile, rngPrefix string) {
	c, err := config.ParseServerConfig(templateFile)
	if err != nil {
		s.t.Fatalf("cannot parse server config; err= %v", err)
	}

	pgConf, err := newDBAndUserFromTemplate(s.logger, c.Postgres, rngPrefix)
	if err != nil {
		s.t.Fatalf("cannot create new PG database and role; err= %v", err)
	}
	c.Postgres = pgConf

	s.ServerConfig = c
}

func newDBAndUserFromTemplate(logger log.Logger, template config.PGConfig, rngPrefix string) (config.PGConfig, error) {
	const createRole = "CREATE ROLE %s WITH LOGIN PASSWORD '%s';"
	const createDB = "CREATE DATABASE %s WITH OWNER %s CONNECTION LIMIT 10;"

	// connect as admin, create user and database
	// REF: compose/infra.yml
	adminStr := template.ConnStr
	adminStr = strings.Replace(adminStr, "HOST:PORT", "postgres.test:5432", -1)
	adminStr = strings.Replace(adminStr, "USER", "pgadmin", -1)
	adminStr = strings.Replace(adminStr, "PASSWORD", "pgpassword", -1)
	adminStr = strings.Replace(adminStr, "&dbname=DBNAME", "", -1)

	db, err := sql.Open("postgres", adminStr)
	if err != nil {
		return config.PGConfig{}, errors.Wrapf(err, "cannot connect to PG as admin")
	}
	defer db.Close()

	query := fmt.Sprintf(createRole, rngPrefix, rngPrefix)
	if _, err := db.Exec(query); err != nil {
		return config.PGConfig{}, errors.Wrapf(err, "cannot create role")
	}
	query = fmt.Sprintf(createDB, rngPrefix, rngPrefix)
	if _, err := db.Exec(query); err != nil {
		return config.PGConfig{}, errors.Wrapf(err, "cannot create database")
	}
	level.Debug(logger).Log("created database", rngPrefix)

	// new user config
	userStr := template.ConnStr
	userStr = strings.Replace(userStr, "HOST:PORT", "postgres.test:5432", -1)
	userStr = strings.Replace(userStr, "USER", rngPrefix, -1)
	userStr = strings.Replace(userStr, "PASSWORD", rngPrefix, -1)
	userStr = strings.Replace(userStr, "DBNAME", rngPrefix, -1)
	template.ConnStr = userStr

	return template, nil
}

var rng = rand.New(rand.NewSource(time.Now().Unix()))

// randNamespace prefix a string (typically a test name) with 4 random characters.
//
// The result will be transformed to suitable length and format to use in infra components.
func randNamespace(s string) string {
	// transform to _XXXX_STR
	// - X is a digit
	// - STR is s with '/' replaced with '_'
	var b bytes.Buffer

	b.WriteByte('_')
	for i := 0; i < 4; i++ {
		digit := rng.Intn(10)
		b.WriteByte('0' + byte(digit))
	}
	b.WriteByte('_')

	s = strings.Replace(s, "/", "_", -1)
	b.WriteString(strings.ToLower(s))

	const lengthLimit = 40
	res := b.String()
	if len(res) > lengthLimit {
		return res[0:lengthLimit]
	}
	return res
}
