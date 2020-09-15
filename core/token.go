package core

import (
	"io/ioutil"
	"os"
	"path"

	"github.com/pkg/errors"
)

// searchToken verifies if there's any existing auth token stored on disk
// the token is used to identify a client that is already registered with the Filstats server
func (c *Core) searchToken() error {
	c.logger.Debug("looking for existing token")

	fileName := path.Join(c.config.DataFolder, TokenFile)
	dat, err := ioutil.ReadFile(fileName)
	if os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return errors.Wrap(err, "could not read token file from disk")
	}

	c.logger.Debug("found token")

	c.token = string(dat)

	return nil
}

// writeToken stores the auth token to disk
func (c *Core) writeToken(token string) error {
	c.logger.Debug("persisting token to disk")
	defer func() {
		c.logger.Debug("done persisting token")
	}()

	fileName := path.Join(c.config.DataFolder, TokenFile)

	// make sure the data folder exists
	_ = os.MkdirAll(c.config.DataFolder, os.ModePerm)

	err := ioutil.WriteFile(fileName, []byte(token), 0644)
	if err != nil {
		return errors.Wrap(err, "could not write token file to disk")
	}

	return nil
}
