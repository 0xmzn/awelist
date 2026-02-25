package cli

import (
	"fmt"
)

type ReportCmd struct{}

func (cmd *ReportCmd) Run(deps *Dependencies) error {
	lock, err := deps.Store.LoadLockFile()
	if err != nil {
		return fmt.Errorf("could not load lock file. run 'awelist enrich'")
	}

	fmt.Printf("Last enriched at: %s\n", lock.Metadata.UpdatedAt.Format("2006-01-02 15:04:05"))

	if len(lock.Metadata.ProviderMetrics) > 0 {
		fmt.Println("\nProvider Summary:")
		for _, m := range lock.Metadata.ProviderMetrics {
			fmt.Printf("  • %-16s | Attempted: %d | Success: %d | Failed: %d\n",
				m.Provider, m.Attempted, m.Successful, m.Failed)
		}
	}

	if len(lock.Metadata.FailedLinks) == 0 {
		fmt.Println("\nNo failed links found.")
		return nil
	}

	fmt.Printf("\nFound %d failed links during the last enrichment:\n", len(lock.Metadata.FailedLinks))
	for url, reason := range lock.Metadata.FailedLinks {
		fmt.Printf("  - %s\n    Reason: %s\n", url, reason)
	}

	return nil
}
