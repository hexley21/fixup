package service

import "errors"

var (
	ErrCategoryTypeNotFound = errors.New("category type not found")

	ErrCategoryNotFound = errors.New("category not found")

	ErrSubcategoryNotFound = errors.New("subcategory not found")
	ErrSubcateogryNameTaken = errors.New("subcategory name is taken")
)