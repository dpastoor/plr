package config

func Read(path string) (Config, error) {
	// for now don't check on error until consider what type of
	// logging - this should be completely optional anyway

	cfg, err := read(path)
	if err != nil {
		return cfg, err
	}
	return cfg, cfg.Validate()
}
