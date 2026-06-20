package cli

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/0xmzn/awelist/internal/types"
)

type EnrichCmd struct {
	TTL time.Duration `kong:"long='ttl',help='How long before a cached enrichment is considered stale.',default='24h'"`
}

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
			fmt.Println("no lock file found, performing full enrichment")
		} else {
			return err
		}
	} else {
		jsonList = lock.List
	}

	metrics, failedLinks, unhandled, err := enricher.EnrichList(list, jsonList, c.TTL)
	if err != nil {
		return err
	}

	newLock := &types.LockFile{
		Metadata: types.LockMetadata{
			UpdatedAt:       time.Now(),
			ProviderMetrics: metrics,
			FailedLinks:     failedLinks,
			UnhandledLinks:  unhandled,
		},
		List: list,
	}

	if err = deps.Store.WriteLockFile(newLock); err != nil {
		return err
	}

	printSummary(metrics, len(unhandled))
	if len(failedLinks) > 0 {
		fmt.Println("\nSome links failed. Run 'awelist report' for details.")
	}
	return nil
}

func printSummary(metrics []types.ProviderMetrics, unhandled int) {
	var total, succeeded, failed int
	for _, m := range metrics {
		total += m.Attempted
		succeeded += m.Successful
		failed += m.Failed
	}
	total += unhandled

	fmt.Printf("\nSummary: %d total | %d succeeded | %d failed | %d skipped\n",
		total, succeeded, failed, unhandled)

	for _, m := range metrics {
		fmt.Printf("  - %s | Attempted: %d | Succeeded: %d | Failed: %d\n",
			m.Provider, m.Attempted, m.Successful, m.Failed)
	}
	if unhandled > 0 {
		fmt.Printf("  - Other | %d skipped (no supported source)\n", unhandled)
	}
}
