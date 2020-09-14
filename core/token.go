package core

import (
	"io/ioutil"
	"os"
	"path"

	"github.com/pkg/errors"
)

func (c *Core) searchToken() error {
	log.Debug("looking for existing token")

	fileName := path.Join(c.config.DataFolder, TokenFile)
	dat, err := ioutil.ReadFile(fileName)
	if os.IsNotExist(err) {
		log.Warn("could not find any token on disk")

		return nil
	} else if err != nil {
		return errors.Wrap(err, "could not read token file from disk")
	}

	log.Debug("found token")

	c.token = string(dat)

	return nil
}

func (c *Core) writeToken(token string) error {
	log.Debug("persisting token to disk")
	defer func() {
		log.Debug("done persisting token")
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
