package aggregator

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	bus2 "github.com/METALmasterKS/simplinic/bus"
	"github.com/METALmasterKS/simplinic/types"
	"github.com/rs/zerolog"
)

// dependencies
type (
	bus interface {
		Subscribe(name string) (<-chan bus2.Message, error)
	}
)

type (
	Aggregator struct {
		logger      zerolog.Logger
		bus         bus
		options     Options
		dataSources map[string]<-chan bus2.Message
	}

	Options struct {
		Period  types.Duration `mapstructure:"aggregate_period_s"`
		DataIDs []string       `mapstructure:"sub_ids"`
	}

	data struct {
		ID    string `json:"id"`
		Value int    `json:"value"`
	}
)

func NewAggregator(ctx context.Context, logger zerolog.Logger, b bus, options Options) (g *Aggregator, err error) {

	g = &Aggregator{
		bus:         b,
		options:     options,
		logger:      logger,
		dataSources: make(map[string]<-chan bus2.Message, len(options.DataIDs)),
	}

	for _, dataID := range options.DataIDs {

		g.dataSources[dataID], err = g.bus.Subscribe(dataID)
	}

	go g.run(ctx)

	return g, nil
}

func (g *Aggregator) run(ctx context.Context) {
	ticker := time.NewTicker(g.options.Period.Duration())
	for {
		select {
		case <-ctx.Done():
			g.logger.Error().Err(ctx.Err())
		case <-ticker.C:
			g.process(ctx)
		}
	}
}

func (g *Aggregator) process(ctx context.Context) {
	for name := range g.dataSources {
		g.logger.Info().Str("source", name).Msg("started")

		for {
			select {
			case msg := <-g.dataSources[name]:
				var d data
				err := json.Unmarshal(msg.Body, &d)
				if err != nil {
					g.logger.Error().Err(err)
				}

				g.logger.Info().Str("data", fmt.Sprintf("%v", d)).Msg("data")
				continue
			default:
			}
			break
		}
		g.logger.Info().Str("source", name).Msg("finished")
	}
}
