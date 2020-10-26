package generator

import "math/rand"

type (
	//easyjson:json
	DataSource struct {
		ID             string `json:"id"`
		Value          int    `json:"value"`
		MaxChangesStep int    `json:"-"`
	}

	DataSoucesOptions struct {
		ID             string `mapstructure:"id"`
		InitValue      int    `mapstructure:"init_value"`
		MaxChangesStep int    `mapstructure:"max_change_step"`
	}
)

func NewDataSource(id string, initVal, maxChange int) *DataSource {
	return &DataSource{
		ID:             id,
		Value:          initVal,
		MaxChangesStep: maxChange,
	}
}

func (d *DataSource) Increment() {
	d.Value += rand.Intn(d.MaxChangesStep + 1)
}
