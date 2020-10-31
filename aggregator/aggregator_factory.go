package aggregator

import (
	"context"
	"io"
	"io/ioutil"
	"os"

	"github.com/rs/zerolog"
)

// AggregatorFactory interface
type AggregatorFactory interface {
	CreateAggregator(ctx context.Context, options Options) (c *Aggregator, err error)
}

type aggregatorFactory struct {
	bus    bus
	logger zerolog.Logger
	writer *LockedWriter
}

// NewAggregatorFactory constructor
func NewAggregatorFactory(logger zerolog.Logger, bus bus, st StorageType) (AggregatorFactory *aggregatorFactory, err error) {

	var writer = LockedWriter{
		Writer: ioutil.Discard,
	}
	switch st {
	case LogStorageType:
		writer.Writer = os.Stdout

	case FileStorageType:
		pwd, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		writer.Writer, err = os.OpenFile(pwd+"/"+storageFileName, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			return nil, err
		}
	}
	return &aggregatorFactory{
		bus:    bus,
		logger: logger,
		writer: &writer,
	}, nil
}

// Connect as service
func (f *aggregatorFactory) CreateAggregator(ctx context.Context, options Options) (c *Aggregator, err error) {

	if c, err = NewAggregator(
		ctx,
		f.logger.With().Strs("data_ids", options.DataIDs).Logger(),
		f.bus,
		f.writer,
		options,
	); err != nil {
		return nil, err
	}

	return c, nil
}

func (f *aggregatorFactory) Close() (err error) {
	closer, ok := f.writer.Writer.(io.Closer)
	if !ok {
		f.logger.Error().Msg("interface convertion failed")
		return nil
	}
	if err = closer.Close(); err != nil {
		f.logger.Error().Err(err).Msg("file close")
	}
	return err
}
