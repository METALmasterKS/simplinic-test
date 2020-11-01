package bus

import (
	"context"
	"strconv"
	"testing"

	"github.com/rs/zerolog"
)

func BenchmarkBus(b *testing.B) {
	ctx := context.Background()
	bus := NewBus(ctx, zerolog.Nop(), 50)
	for i := 0; i < b.N; i++ {
		ch1, err := bus.Subscribe("ch1")
		if err != nil {
			b.Error(err)
		}
		var msg = Message{
			From: "ch1",
			Body: []byte(strconv.Itoa(i)),
		}
		err = bus.Publish(ctx, msg)
		<-ch1
	}
}
