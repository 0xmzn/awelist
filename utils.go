package main

import "github.com/gosimple/slug"

func slugifiy(str string) string {
	return slug.Make(str)
}