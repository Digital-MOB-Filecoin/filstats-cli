package core

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// intervalRunner is a function that executes the function `f` when first called and then every `interval`
// if the function f returns an error we check to see if it's a gRPC error or something else
// if the filstats-server returned a `Unauthenticated` error, we must trigger a reconnect
func (c *Core) intervalRunner(ctx context.Context, f func() error, interval time.Duration) error {
	err := f()
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			switch st.Code() {
			case codes.Unauthenticated:
				return errors.New("[ ☠ ] un-registered from Filstats server; triggering reconnect")
			case codes.Unavailable:
				c.logger.Error("[ ✖ ] could not reach Filstats server")
			default:
				c.logger.Error(err)
			}
		} else {
			c.logger.Errorf("[ ✖ ] could not send request, got: %s", err)
			return err
		}
	}

	ticker := time.NewTicker(interval)

	for {
		select {
		case <-ticker.C:
			err := f()
			if err != nil {
				st, ok := status.FromError(err)
				if ok {
					switch st.Code() {
					case codes.Unauthenticated:
						return errors.New("[ ☠ ] un-registered from Filstats server; triggering reconnect")
					case codes.Unavailable:
						c.logger.Error("[ ✖ ] could not reach Filstats server")
						continue
					default:
						return err
					}
				} else {
					return err
				}
			}
		case <-ctx.Done():
			c.logger.Info("context was canceled, stopping")
			ticker.Stop()
			return nil
		}
	}
}
