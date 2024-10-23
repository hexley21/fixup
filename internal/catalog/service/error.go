package service

import "errors"

var (
	ErrCategoryTypeNotFound = errors.New("category type not found")
	ErrCategoryTypeNameTaken = errors.New("category type name is taken")

	ErrCategoryNotFound = errors.New("category not found")
	ErrCategoryNameTaken = errors.New("category name is taken")

	ErrSubcategoryNotFound = errors.New("subcategory not found")
	ErrSubcategoryNameTaken = errors.New("subcategory name is taken")
)
