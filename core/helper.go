package core

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (c *Core) intervalRunner(ctx context.Context, f func() error, interval time.Duration) error {
	err := f()
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			switch st.Code() {
			case codes.Unauthenticated:
				return errors.New("[☠] un-registered from Filstats server; triggering reconnect")
			case codes.Unavailable:
				c.logger.Error("[✖] could not reach Filstats server")
			default:
				c.logger.Error(err)
			}
		} else {
			c.logger.Errorf("[✖] could not send request, got: %s", err)
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
						return errors.New("[☠] un-registered from Filstats server; triggering reconnect")
					case codes.Unavailable:
						c.logger.Error("[✖] could not reach Filstats server")
						continue
					default:
						c.logger.Error(err)
					}
				} else {
					c.logger.Errorf("[✖] could not send request, got: %s", err)

					continue
				}
			}
		case <-ctx.Done():
			ticker.Stop()
			return nil
		}
	}
}
