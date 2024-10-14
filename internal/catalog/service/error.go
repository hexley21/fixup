package service

import "errors"

var (
	ErrSubcategoryNotFound = errors.New("subcategory not found")
	ErrSubcategoryNotUpdated = errors.New("subcategory not updated")
	ErrSubcategoryNotDeleted = errors.New("subcategory not deleted")
	ErrSubcateogryNameTaken = errors.New("subcategory name is taken")
)