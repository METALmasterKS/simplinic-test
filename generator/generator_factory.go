package generator

import (
	"context"

	"github.com/rs/zerolog"
)

// GeneratorFactory interface
type GeneratorFactory interface {
	CreateGenerator(ctx context.Context, options Options) (c *Generator, err error)
}

type generatorFactory struct {
	bus    bus
	logger zerolog.Logger
}

// NewGeneratorFactory constructor
func NewGeneratorFactory(logger zerolog.Logger, bus bus) GeneratorFactory {
	return &generatorFactory{
		bus:    bus,
		logger: logger,
	}
}

// Connect as service
func (f *generatorFactory) CreateGenerator(ctx context.Context, options Options) (c *Generator, err error) {

	var ids = make([]string, len(options.DataSourcesOptions))
	for i := range options.DataSourcesOptions {
		ids[i] = options.DataSourcesOptions[i].ID
	}

	if c, err = NewGenerator(
		ctx,
		f.logger.With().Strs("data_ids", ids).Logger(),
		f.bus,
		options,
	); err != nil {
		return nil, err
	}

	return c, nil
}
