package cli

type AddCmd struct {
	Link     AddLinkCmd     `kong:"cmd,help='Add a new link to a category.'"`
	Category AddCategoryCmd `kong:"cmd,help='Add a new subcategory to a category.'"`
}

type AddLinkCmd struct {
	Title       string   `kong:"short='t',long='title',help='Title of the link.',required"`
	Description string   `kong:"short='d',long='description',help='Description of the link.',required"`
	URL         string   `kong:"short='u',long='url',help='URL of the link.',required"`
	Path        []string `kong:"arg,name='path',help='Path to the parent category.'"`
}

type AddCategoryCmd struct {
	Title       string   `kong:"short='t',long='title',help='Title of the new category.'"`
	Description string   `kong:"short='d',long='description',help='Description of the new category.'"`
	Path        []string `kong:"arg,name='path',help='Path to the parent category. Use a single dot (.) to add to the top level of the list.'"`
}

func (cmd *AddLinkCmd) Run(deps *Dependencies) error {
	log := deps.Logger

	log.Info("Running generate")
	log.Debug("Debugging generate")
	return nil
}

func (cmd *AddCategoryCmd) Run(deps *Dependencies) error {
	log := deps.Logger

	log.Info("Running generate")
	log.Debug("Debugging generate")
	return nil
}
