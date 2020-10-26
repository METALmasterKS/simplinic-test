package aggregator

import (
	"context"

	"github.com/rs/zerolog"
)

// AggregatorFactory interface
type AggregatorFactory interface {
	CreateAggregator(ctx context.Context, options Options) (c *Aggregator, err error)
}

type aggregatorFactory struct {
	bus    bus
	logger zerolog.Logger
}

// NewAggregatorFactory constructor
func NewAggregatorFactory(logger zerolog.Logger, bus bus) AggregatorFactory {
	return &aggregatorFactory{
		bus:    bus,
		logger: logger,
	}
}

// Connect as service
func (f *aggregatorFactory) CreateAggregator(ctx context.Context, options Options) (c *Aggregator, err error) {

	if c, err = NewAggregator(ctx, f.logger, f.bus, options); err != nil {
		return nil, err
	}

	return c, nil
}
