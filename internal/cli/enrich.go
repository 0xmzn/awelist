package cli

type EnrichCmd struct{}

func (c *EnrichCmd) Run(deps *Dependencies) error {
	log := deps.Logger

	list, _ := deps.Store.LoadYAML()

	log.Info("loaded", "count", list.TotalCount())

	log.Info("Running generate")
	log.Debug("Debugging generate")
	return nil
}
