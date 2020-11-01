package aggregator

import (
	"context"
	"encoding/json"
	"io"
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

// StorageType
type StorageType uint8

// storage types
const (
	LogStorageType StorageType = iota
	FileStorageType
)

const storageFileName = "simplinic.log"

type (
	Aggregator struct {
		logger      zerolog.Logger
		writer      io.Writer
		bus         bus
		options     Options
		dataSources map[string]<-chan bus2.Message

		m      sync.Mutex
		buffer map[string][]int
	}

	Options struct {
		Period  types.Duration `mapstructure:"aggregate_period_s"`
		DataIDs []string       `mapstructure:"sub_ids"`
	}

	//easyjson:json
	data struct {
		ID    string `json:"id"`
		Value int    `json:"value"`
	}
	//easyjson:json
	avgData struct {
		ID    string  `json:"id"`
		Value float64 `json:"value"`
	}
)

func NewAggregator(ctx context.Context, logger zerolog.Logger, b bus, writer io.Writer, options Options) (g *Aggregator, err error) {

	g = &Aggregator{
		bus:         b,
		options:     options,
		logger:      logger,
		writer:      writer,
		dataSources: make(map[string]<-chan bus2.Message, len(options.DataIDs)),
		buffer:      make(map[string][]int, 0),
	}

	for _, dataID := range options.DataIDs {
		g.dataSources[dataID], err = g.bus.Subscribe(dataID)
	}

	go g.run(ctx)

	return g, nil
}

func (g *Aggregator) run(ctx context.Context) {
	ticker := time.NewTicker(g.options.Period.Duration())
	g.logger.Info().Msg("start")
	for {
		select {
		case <-ctx.Done():
			g.stop()
			g.logger.Error().Err(ctx.Err()).Msg("stop")
			return
		case <-ticker.C:
			g.flush()
		default:
			g.process(ctx)
		}
	}
}

func (g *Aggregator) stop() {
	for _, dataID := range g.options.DataIDs {
		if err := g.bus.Unsubscribe(dataID); err != nil {
			g.logger.Error().Err(err).Msg("stop")
		}
	}
}

func (g *Aggregator) process(_ context.Context) {
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
						g.logger.Error().Err(err).Msg("unmarshal")
					}

					g.m.Lock()
					if _, ok := g.buffer[d.ID]; !ok {
						g.buffer[d.ID] = make([]int, 0, 100)
					}
					g.buffer[d.ID] = append(g.buffer[d.ID], d.Value)
					g.m.Unlock()

				default:
				}
				return
			}
		}(g.dataSources[name])

	}
	wg.Wait()
}

func (g *Aggregator) flush() {
	var (
		buffer = make(map[string][]int, 0)
	)

	g.m.Lock()
	buffer, g.buffer = g.buffer, buffer
	g.m.Unlock()
	g.logger.Debug().Msg("flushed")

	for name, ints := range buffer {
		var (
			dataJson []byte
			err      error
		)
		dataJson, err = json.Marshal(avgData{
			ID:    name,
			Value: avg(ints),
		})
		if err != nil {
			g.logger.Error().Err(err).Msg("marshal")
		}
		_, err = g.writer.Write(append(dataJson, []byte("\n")...))
		if err != nil {
			g.logger.Error().Err(err).Msg("write")
		}
	}
}

func avg(ints []int) (avg float64) {
	for _, v := range ints {
		avg += float64(v)
	}
	return avg / float64(len(ints))
}
