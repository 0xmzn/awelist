package cli

type EnrichCmd struct{}

func (c *EnrichCmd) Run(deps *Dependencies) error {
	log := deps.Logger
	enricher := deps.Enricher

	list, _ := deps.Store.LoadYAML()

	log.Info("loaded", "count", list.TotalCount())

	err := enricher.EnrichList(list)
	if err != nil {
		return err
	}

	log.Info("Running generate")
	log.Debug("Debugging generate")
	return nil
}
