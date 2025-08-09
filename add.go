package main

type AddCmd struct {
	Link     AddLinkCmd     `kong:"cmd,help='Add a new link to a category.'"`
	Category AddCategoryCmd `kong:"cmd,help='Add a new subcategory to a category.'"`
}

type AddLinkCmd struct {
	Title       string   `kong:"short='t',long='title',help='Title of the link.',required"`
	Description string   `kong:"short='d',long='description',help='Description of the link.',required"`
	URL         string   `kong:"short='u',long='url',help='URL of the link.',required"`
	Path        []string `kong:"arg,name='path',help='Path to the category where the link should be added.'"`
}

type AddCategoryCmd struct {
	Title       string   `kong:"short='t',long='title',help='Title of the new category.'"`
	Description string   `kong:"short='d',long='description',help='Description of the new category.'"`
	Path        []string `kong:"arg,name='path',help='Path to the parent category where the new category should be added.'"`
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

	err = awelist.AddLink(newLink, cmd.Path)
	if err != nil {
		return err
	}

	//TODO: write it using the store back to yaml

	return nil
}

func (cmd *AddCategoryCmd) Run(cli *CLI) error {
	panic("Not implemented yet")
}
