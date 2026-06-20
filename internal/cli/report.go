package cli

import (
	"fmt"
	"sort"
)

type ReportCmd struct{}

func (cmd *ReportCmd) Run(deps *Dependencies) error {
	lock, err := deps.Store.LoadLockFile()
	if err != nil {
		return fmt.Errorf("could not load lock file. run 'awelist enrich'")
	}

	fmt.Printf("Last enriched at: %s\n", lock.Metadata.UpdatedAt.Format("2006-01-02 15:04:05"))

	m := lock.Metadata
	printSummary(m.ProviderMetrics, len(m.UnhandledLinks))

	if len(m.UnhandledLinks) > 0 {
		fmt.Println("\nUnsupported links:")
		for _, url := range m.UnhandledLinks {
			fmt.Printf("  - %s\n", url)
		}
	}

	if len(m.FailedLinks) > 0 {
		fmt.Println("\nFailed links:")
		keys := make([]string, 0, len(m.FailedLinks))
		for url := range m.FailedLinks {
			keys = append(keys, url)
		}
		sort.Strings(keys)
		for _, url := range keys {
			fmt.Printf("  - %s\n    Reason: %s\n", url, m.FailedLinks[url])
		}
	}

	return nil
}
