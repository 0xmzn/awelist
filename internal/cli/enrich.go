package cli

type EnrichCmd struct{}

func (c *EnrichCmd) Run(deps *Dependencies) error {
	log := deps.Logger

	log.Info("Running generate")
	log.Debug("Debugging generate")
	return nil
}
