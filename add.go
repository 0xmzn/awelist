package main

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

func (cmd *AddLinkCmd) Run(cli *CLI) error {
	aweStore := NewAwesomeStore(cli.AwesomeFile)
	baseList, err := aweStore.Load()
	if err != nil {
		return err
	}

	awelist := NewAwesomeListManager(baseList)

	newLink := BaseLink{
		Title:       cmd.Title,
		Description: cmd.Description,
		Url:         cmd.URL,
	}

	if err = awelist.AddLink(newLink, cmd.Path); err != nil {
		return err
	}

	if err = aweStore.WriteYAML(awelist.RawList); err != nil {
		return err
	}

	return nil
}

func (cmd *AddCategoryCmd) Run(cli *CLI) error {
	aweStore := NewAwesomeStore(cli.AwesomeFile)
	baseList, err := aweStore.Load()
	if err != nil {
		return err
	}

	awelist := NewAwesomeListManager(baseList)

	newCat := BaseCategory{
		Title:       cmd.Title,
		Description: cmd.Description,
	}

	args := cmd.Path
	if len(cmd.Path) == 1 && cmd.Path[0] == "." {
		args = nil
	}

	if err = awelist.AddCategory(newCat, args); err != nil {
		return err
	}

	if err = aweStore.WriteYAML(awelist.RawList); err != nil {
		return err
	}

	return nil
}
