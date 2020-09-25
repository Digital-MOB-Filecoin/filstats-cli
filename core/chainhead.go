package core

import (
	"context"
	"time"

	"github.com/bep/debounce"
	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/digital-mob-filecoin/filstats-cli/node"
)

func (c *Core) filstatsChainHead(ctx context.Context) error {
	// send initial chain head
	head, err := c.node.GetChainHead()
	if err != nil {
		return err
	}

	if head == nil {
		return errors.New("got nil chain head")
	}

	_, err = c.filstatsServer.ChainHead(c.contextWithToken(ctx), head.ToChainHeadRequest())
	if err != nil {
		return errors.Wrap(err, "could not call Filstats.ChainHead")
	}

	return nil
}

func (c *Core) watchNewHeads(ctx context.Context) error {
	ch, err := c.node.SubscribeNewHeads(ctx)
	if err != nil {
		return errors.Wrap(err, "could not subscribe to new heads")
	}

	var latestHead node.ChainHead

	d := debounce.New(time.Second)

	var globalErrors []error
	log := c.logger.WithField("_req", "ChainHead")

	for head := range ch {
		c.logger.Trace("processing head")
		latestHead = head

		if len(globalErrors) > 0 {
			return globalErrors[0]
		}

		d(func() {
			log.Debug("[⇢] outgoing request")
			start := time.Now()
			defer func() {
				log.WithField("duration", time.Since(start)).Debug("[⇠] finalized request")
			}()

			_, err = c.filstatsServer.ChainHead(c.contextWithToken(ctx), latestHead.ToChainHeadRequest())
			if err != nil {
				st, ok := status.FromError(err)
				if ok {
					switch st.Code() {
					case codes.Unauthenticated:
						globalErrors = append(globalErrors, errors.New("un-registered from Filstats server; triggering reconnect"))
					case codes.Unavailable:
						log.Error("could not reach Filstats server")
					default:
						log.Error(err)
					}
				} else {
					log.Error(err)
				}
			}
		})
	}

	if ctx.Err() == context.Canceled {
		return nil
	}

	return errors.New("unexpected close of new heads subscription")
}
