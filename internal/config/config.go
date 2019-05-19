package config

import (
	"os"

	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v3"
)

func ParseServerConfig(filename string) (ServerConfig, error) {
	var c ServerConfig

	f, err := os.Open(filename)
	if err != nil {
		return ServerConfig{}, errors.Wrapf(err, "cannot open config file")
	}
	d := yaml.NewDecoder(f)
	d.KnownFields(true)
	if err := d.Decode(&c); err != nil {
		return ServerConfig{}, errors.Wrapf(err, "cannot unmarshal")
	}

	return c, nil
}

type ServerConfig struct {
	Postgres PGConfig `yaml:"postgres"`
}

func (c ServerConfig) Check() error {
	if err := c.Postgres.Check(); err != nil {
		return errors.Wrapf(err, "invalid 'postgres'")
	}

	return nil
}

type PGConfig struct {
	ConnStr string `yaml:"url"`
	MaxIdle int    `yaml:"maxIdleConnection"`
	MaxOpen int    `yaml:"maxOpenConnection"`
}

func (c PGConfig) Check() error {
	if !(len(c.ConnStr) > 0) {
		return errors.New("no 'url'")
	}
	if !(c.MaxIdle >= 1) {
		return errors.New("'maxIdle' must >= 1")
	}
	if !(c.MaxOpen >= 1) {
		return errors.New("'maxOpen' must >= 1")
	}
	if !(c.MaxIdle <= c.MaxOpen) {
		return errors.New("'maxIdle' must <= 'maxOpen'")
	}

	return nil
}
