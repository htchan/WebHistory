package shutdown

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"
)

type ShutdownFunc struct {
	name string
	f    func() error
}

func (f ShutdownFunc) run() error {
	return f.f()
}

type ShutdownHandler struct {
	funcs []ShutdownFunc
}

func New() *ShutdownHandler {
	return &ShutdownHandler{}
}

func (handler *ShutdownHandler) Register(name string, f func() error) {
	handler.funcs = append(handler.funcs, ShutdownFunc{name: name, f: f})
}

func (handler *ShutdownHandler) Listen(timeout time.Duration) error {
	// TODO: catch kill signal
	kill := make(chan os.Signal, 1)
	signal.Notify(kill, syscall.SIGINT, syscall.SIGTERM)
	<-kill

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// run ShutdownFunc one by one
	for _, fn := range handler.funcs {
		fn := fn
		fnComplete := make(chan interface{})

		go func() {
			if err := fn.run(); err != nil {
				log.Error().Err(err).Str("name", fn.name).Msg("shutdown error")
			} else {
				log.Info().Str("name", fn.name).Msg("shutdown complete")
			}

			fnComplete <- nil
		}()

		select {
		case <-ctx.Done():
			log.Info().Str("name", fn.name).Msg("shutdown timeout")
		case <-fnComplete:
		}

		close(fnComplete)
	}

	return ctx.Err()
}
