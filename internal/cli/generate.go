package cli

type GenerateCmd struct {
	HTML         bool   `kong:"short='H',help='output HTML.'"`
	TemplateFile string `kong:"arg,required,help='path to template file.'"`
}

func (cmd *GenerateCmd) Run(deps *Dependencies) error {
	log := deps.Logger

	log.Info("Running generate")
	log.Debug("Debugging generate")
	return nil
}
