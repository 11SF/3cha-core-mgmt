package config

import env "github.com/caarlos0/env/v11"

func parseEnv[T any](opts env.Options) (T, error) {
	var cfg T
	if err := env.ParseWithOptions(&cfg, opts); err != nil {
		return cfg, err
	}
	return cfg, nil
}
