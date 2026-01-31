package cli

type EnrichCmd struct{}

func (c *EnrichCmd) Run(deps *Dependencies) error {
	enricher := deps.Enricher

	list, err := deps.Store.LoadYAML()
	if err != nil {
		return err
	}

	err = enricher.EnrichList(list)
	if err != nil {
		return err
	}

	if err = deps.Store.WriteJSON(list); err != nil {
		return err
	}
	return nil
}
