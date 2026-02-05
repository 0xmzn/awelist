package cli

import "github.com/0xmzn/awelist/internal/types"

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
	mngr := deps.ListManager
	store := deps.Store

	list, err := store.LoadYAML()
	if err != nil {
		return err
	}

	newLink := types.Link{
		Title:       cmd.Title,
		Description: cmd.Description,
		URL:         cmd.URL,
	}

	if err = mngr.AddLink(list, &newLink, cmd.Path); err != nil {
		return err
	}

	if err = store.WriteYAML(list); err != nil {
		return err
	}

	return nil
}

func (cmd *AddCategoryCmd) Run(deps *Dependencies) error {
	store := deps.Store
	mngr := deps.ListManager

	list, err := store.LoadYAML()
	if err != nil {
		return err
	}

	newCat := types.Category{
		Title:       cmd.Title,
		Description: cmd.Description,
	}

	if err = mngr.AddCategory(&list, &newCat, cmd.Path); err != nil {
		return err
	}

	if err = store.WriteYAML(list); err != nil {
		return err
	}

	return nil
}
