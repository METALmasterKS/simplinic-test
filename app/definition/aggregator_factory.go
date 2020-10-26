package definition

import (
	"github.com/METALmasterKS/simplinic/aggregator"
	"github.com/METALmasterKS/simplinic/bus"
	"github.com/rs/zerolog/log"
	"github.com/sarulabs/di/v2"
)

const DefAggregatorFactoryName = "agg-factory"

func DefAggregatorFactory() di.Def {
	return di.Def{
		Name: DefAggregatorFactoryName,
		Build: func(ctn di.Container) (_ interface{}, err error) {
			logger := log.With().Str(KeyComponent, DefAggregatorFactoryName).Logger()
			b := ctn.Get(DefBusName).(bus.Broker)

			return aggregator.NewAggregatorFactory(logger, b), nil
		},
	}
}
