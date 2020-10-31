package definition

import (
	"context"
	"fmt"

	"github.com/METALmasterKS/simplinic/bus"
	"github.com/rs/zerolog/log"
	"github.com/sarulabs/di/v2"
	"github.com/spf13/viper"
)

const DefBusName = "bus"

func DefBus() di.Def {
	return di.Def{
		Name: DefBusName,
		Build: func(ctn di.Container) (_ interface{}, err error) {
			ctx := ctn.Get(DefContextName).(context.Context)
			logger := log.With().Str(KeyComponent, DefBusName).Logger()

			queueCfg := viper.GetViper().Sub("queue")
			if queueCfg == nil {
				return nil, fmt.Errorf("no config for queue")
			}
			capacity := queueCfg.GetInt("size")
			if capacity == 0 {
				return nil, fmt.Errorf("size must be greather than 0")
			}

			return bus.NewBus(ctx, logger, capacity), nil
		},
	}
}
