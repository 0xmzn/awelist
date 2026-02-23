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

	if len(lock.Metadata.FailedLinks) == 0 {
		fmt.Printf("Last enriched at: %s\n", lock.Metadata.UpdatedAt.Format("2006-01-02 15:04:05"))
		fmt.Println("No failed links found.")
		return nil
	}

	fmt.Printf("Last enriched at: %s\n", lock.Metadata.UpdatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("Found %d failed links during the last enrichment:\n\n", len(lock.Metadata.FailedLinks))

	for url, reason := range lock.Metadata.FailedLinks {
		fmt.Printf("- %s\n  Reason: %s\n", url, reason)
	}
	return nil
}
