package types

import "time"

type Duration int64

func (d Duration) Duration() time.Duration {
	return time.Duration(d) * time.Second
}
