package shutdown

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"
)

// TODO: move this to separate package
type shutdownFunc struct {
	name string
	f    func() error
}

func (f *shutdownFunc) run() error {
	if f.f == nil {
		return ErrNilFunc
	}

	return f.f()
}

type ShutdownHandler struct {
	funcs []shutdownFunc
}

func New() *ShutdownHandler {
	return &ShutdownHandler{}
}

func (handler *ShutdownHandler) Register(name string, f func() error) {
	handler.funcs = append(handler.funcs, shutdownFunc{name: name, f: f})
}

func (handler *ShutdownHandler) Listen(timeout time.Duration) error {
	kill := make(chan os.Signal, 1)
	signal.Notify(kill, syscall.SIGINT, syscall.SIGTERM)
	<-kill

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// run ShutdownFunc one by one
	for _, fn := range handler.funcs {
		f := func() <-chan struct{} {
			fnComplete := make(chan struct{}, 1)

			go func() {
				defer close(fnComplete)

				if err := fn.run(); err != nil {
					log.Error().Err(err).Str("name", fn.name).Msg("shutdown error")
				} else {
					log.Info().Str("name", fn.name).Msg("shutdown complete")
				}

				fnComplete <- struct{}{}
			}()

			return fnComplete
		}

		select {
		case <-f():
		case <-ctx.Done():
			log.Info().Str("name", fn.name).Msg("shutdown timeout")
		}
	}

	return ctx.Err()
}
