package generator

import (
	"context"
	"encoding/json"
	"time"

	bus2 "github.com/METALmasterKS/simplinic/bus"
	"github.com/METALmasterKS/simplinic/types"
	"github.com/rs/zerolog"
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
		options     Options
		dataSources []*DataSource
	}

	Options struct {
		Timeout            types.Duration      `mapstructure:"timeout_s"`
		SendPeriod         types.Duration      `mapstructure:"send_period_s"`
		DataSourcesOptions []DataSoucesOptions `mapstructure:"data_sources"`
	}
)

func NewGenerator(ctx context.Context, logger zerolog.Logger, b bus, options Options) (*Generator, error) {

	g := Generator{
		logger:      logger,
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
	ticker := time.NewTicker(g.options.SendPeriod.Duration())
	g.logger.Info().Msg("start")
	for {
		select {
		case <-ctx.Done():
			g.logger.Error().Err(ctx.Err()).Msg("stop")
			return
		case <-ticker.C:
			g.processWithTimeout(ctx)
		}
	}
}

func (g *Generator) processWithTimeout(ctx context.Context) {
	timeoutCtx, cancel := context.WithTimeout(ctx, g.options.Timeout.Duration())
	defer cancel()
	g.process(timeoutCtx)
}

func (g *Generator) process(ctx context.Context) {
	for _, ds := range g.dataSources {
		ds.Increment()
		msg, err := json.Marshal(ds)
		if err != nil {
			g.logger.Error().Err(err).Msg("marshal")
		}

		if err := g.bus.Publish(ctx, bus2.Message{From: ds.ID, Body: msg}); err != nil {
			g.logger.Error().Err(err).Msg("publish")
		}
	}
}
