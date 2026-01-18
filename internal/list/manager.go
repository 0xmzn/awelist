package list

import (
	"fmt"

	"github.com/0xmzn/awelist/internal/types"
	"github.com/gosimple/slug"
)

type Manager struct{}

func NewManager() *Manager {
	return &Manager{}
}

func (m *Manager) AddLink(list types.AwesomeList, link *types.Link, path []string) error {
	if len(path) == 0 {
		return fmt.Errorf("path cannot be empty when adding a link")
	}

	cat, err := m.findCategory(list, path)
	if err != nil {
		return err
	}

	for _, l := range cat.Links {
		if slug.Make(l.Title) == slug.Make(link.Title) {
			return fmt.Errorf("link with title %q already exists in %q", link.Title, cat.Title)
		}
		if l.URL == link.URL {
			return fmt.Errorf("link with url %q already exists in %q", link.URL, cat.Title)
		}
	}

	cat.Links = append(cat.Links, link)
	return nil
}

func (m *Manager) AddCategory(list *types.AwesomeList, newCat *types.Category, path []string) error {
	if len(path) == 0 || (len(path) == 1 && path[0] == ".") {
		for _, c := range *list {
			if slug.Make(c.Title) == slug.Make(newCat.Title) {
				return fmt.Errorf("category %q already exists at root", newCat.Title)
			}
		}
		*list = append(*list, newCat)
		return nil
	}

	parent, err := m.findCategory(*list, path)
	if err != nil {
		return err
	}

	for _, sub := range parent.Subcategories {
		if slug.Make(sub.Title) == slug.Make(newCat.Title) {
			return fmt.Errorf("subcategory %q already exists under %q", newCat.Title, parent.Title)
		}
	}

	parent.Subcategories = append(parent.Subcategories, newCat)
	return nil
}

func (m *Manager) findCategory(list types.AwesomeList, path []string) (*types.Category, error) {
	if len(path) == 0 {
		return nil, fmt.Errorf("path is empty")
	}

	targetSlug := slug.Make(path[0])
	var found *types.Category

	for _, c := range list {
		if slug.Make(c.Title) == targetSlug {
			found = c
			break
		}
	}

	if found == nil {
		return nil, fmt.Errorf("category %q not found", path[0])
	}

	if len(path) == 1 {
		return found, nil
	}

	return m.findCategory(types.AwesomeList(found.Subcategories), path[1:])
}
