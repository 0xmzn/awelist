package cli

import (
	"errors"
	"os"
)

type EnrichCmd struct{}

func (c *EnrichCmd) Run(deps *Dependencies) error {
	enricher := deps.Enricher

	list, err := deps.Store.LoadYAML()
	if err != nil {
		return err
	}

	jsonList, err := deps.Store.LoadJson()
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			deps.Logger.Info("no lock file found, performing full enrichment")
			err = nil
		} else {
			return err
		}
	}

	err = enricher.EnrichList(list, jsonList)
	if err != nil {
		return err
	}

	if err = deps.Store.WriteJSON(list); err != nil {
		return err
	}
	return nil
}
