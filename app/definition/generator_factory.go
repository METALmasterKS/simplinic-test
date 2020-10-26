package definition

import (
	"github.com/METALmasterKS/simplinic/bus"
	"github.com/METALmasterKS/simplinic/generator"
	"github.com/rs/zerolog/log"
	"github.com/sarulabs/di/v2"
)

const DefGeneratorFactoryName = "gen-factory"

func DefGeneratorFactory() di.Def {
	return di.Def{
		Name: DefGeneratorFactoryName,
		Build: func(ctn di.Container) (_ interface{}, err error) {
			logger := log.With().Str(KeyComponent, DefGeneratorFactoryName).Logger()
			b := ctn.Get(DefBusName).(bus.Broker)

			return generator.NewGeneratorFactory(logger, b), nil
		},
	}
}
