package cli

import (
	"errors"
	"os"
	"time"

	"github.com/0xmzn/awelist/internal/types"
)

type EnrichCmd struct{}

func (c *EnrichCmd) Run(deps *Dependencies) error {
	enricher := deps.Enricher

	list, err := deps.Store.LoadAwesomeFile()
	if err != nil {
		return err
	}

	var jsonList types.AwesomeList
	lock, err := deps.Store.LoadLockFile()
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			deps.Logger.Info("no lock file found, performing full enrichment")
		} else {
			return err
		}
	} else {
		jsonList = lock.List
	}

	failedLinks, err := enricher.EnrichList(list, jsonList)
	if err != nil {
		return err
	}

	newLock := &types.LockFile{
		Metadata: types.LockMetadata{
			UpdatedAt:   time.Now(),
			FailedLinks: failedLinks,
		},
		List: list,
	}

	if err = deps.Store.WriteLockFile(newLock); err != nil {
		return err
	}
	return nil
}
