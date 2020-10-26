package generator

import (
	"context"
	"encoding/json"
	bus2 "github.com/METALmasterKS/simplinic/bus"
	"github.com/METALmasterKS/simplinic/types"
	"github.com/rs/zerolog"
	"time"
)

// dependencies
type (
	bus interface {
		Publish(ctx context.Context, msg bus2.Message) error
	}
)

type (
	Generator struct {
		logger      zerolog.Logger
		bus         bus
		options     GeneratorOptions
		dataSources []*DataSource
	}

	GeneratorOptions struct {
		Timeout            types.Duration      `mapstructure:"timeout_s"`
		SendPeriod         types.Duration      `mapstructure:"send_period_s"`
		DataSourcesOptions []DataSoucesOptions `mapstructure:"data_sources"`
	}
)

func NewGenerator(ctx context.Context, b bus, options GeneratorOptions) (*Generator, error) {

	g := Generator{
		bus:         b,
		options:     options,
		dataSources: make([]*DataSource, len(options.DataSourcesOptions)),
	}

	for i, opts := range options.DataSourcesOptions {
		g.dataSources[i] = NewDataSource(opts.ID, opts.InitValue, opts.MaxChangesStep)
	}

	go g.run(ctx)

	return &g, nil
}

func (g *Generator) run(ctx context.Context) {
	ticker := time.NewTicker(g.options.Timeout.Duration())
	for {
		select {
		case <-ctx.Done():
			g.logger.Error().Err(ctx.Err())
		case <-ticker.C:
			g.processWithTimeout(ctx)
		}
	}
}

func (g *Generator) processWithTimeout(ctx context.Context) {
	timeoutCtx, cancel := context.WithTimeout(ctx, g.options.SendPeriod.Duration())
	defer cancel()
	g.process(timeoutCtx)
}

func (g *Generator) process(ctx context.Context) {
	for _, ds := range g.dataSources {
		ds.Increment()
		msg, err := json.Marshal(ds)
		if err != nil {
			g.logger.Error().Err(err)
		}

		if err := g.bus.Publish(ctx, bus2.Message{Body: msg}); err != nil {
			g.logger.Error().Err(err)
		}
	}
}
