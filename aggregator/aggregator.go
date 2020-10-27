package aggregator

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	bus2 "github.com/METALmasterKS/simplinic/bus"
	"github.com/METALmasterKS/simplinic/types"
	"github.com/rs/zerolog"
)

// dependencies
type (
	bus interface {
		Subscribe(name string) (<-chan bus2.Message, error)
		Unsubscribe(name string) error
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
	timer := time.NewTimer(g.options.Period.Duration())
	for {
		select {
		case <-ctx.Done():
			g.logger.Error().Err(ctx.Err())
			return
		case <-timer.C:
			g.stop()
			return
		default:
		}
		g.process(ctx)
	}
}

func (g *Aggregator) stop() {
	for _, dataID := range g.options.DataIDs {
		if err := g.bus.Unsubscribe(dataID); err != nil {
			g.logger.Error().Err(err)
		}
	}
}

func (g *Aggregator) process(ctx context.Context) {
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 10)
	for name := range g.dataSources {
		if len(g.dataSources[name]) == 0 {
			continue
		}
		wg.Add(1)
		semaphore <- struct{}{}
		go func(msgCh <-chan bus2.Message) {
			defer func() {
				<-semaphore
				wg.Done()
			}()
			for {
				select {
				case msg := <-msgCh:
					var d data
					err := json.Unmarshal(msg.Body, &d)
					if err != nil {
						g.logger.Error().Err(err)
					}

					g.logger.Info().Int("len", len(msgCh)).Str("id", d.ID).Msg("data")
				default:
				}
				return
			}
		}(g.dataSources[name])

	}
	wg.Wait()
}
