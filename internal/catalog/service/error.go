package service

import "errors"

var (
	ErrCategoryNotFound = errors.New("category not found")

	ErrSubcategoryNotFound = errors.New("subcategory not found")
	ErrSubcategoryNotDeleted = errors.New("subcategory not deleted")
	ErrSubcateogryNameTaken = errors.New("subcategory name is taken")
)