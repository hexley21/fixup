package service

import "errors"

var (
	ErrCategoryNotFound = errors.New("category not found")

	ErrSubcategoryNotFound = errors.New("subcategory not found")
	ErrSubcateogryNameTaken = errors.New("subcategory name is taken")
)