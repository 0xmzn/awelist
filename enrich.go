package main

func (cmd *EnrichCmd) Run(cli *CLI) error {
	aweStore := NewAwesomeStore(cli.AwesomeFile)
	baseList, err := aweStore.LoadYAML()
	if err != nil {
		return err
	}

	awelist := NewAwesomeListManager(baseList)

	err = awelist.EnrichList()
	if err != nil {
		return err
	}

	if err = aweStore.WriteJSON(awelist.EnrichedList); err != nil {
		return err
	}

	return nil
}
