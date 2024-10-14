DROP TABLE IF EXISTS category_types CASCADE;
DROP TABLE IF EXISTS categories CASCADE;
DROP TABLE IF EXISTS subcategories CASCADE;
DROP TABLE IF EXISTS services CASCADE;
DROP TABLE IF EXISTS provider_services CASCADE;

DROP TRIGGER IF EXISTS prevent_duplicate_trigger ON categories;
DROP FUNCTION IF EXISTS prevent_duplicate_type_name;

DROP TRIGGER IF EXISTS prevent_duplicate_subcategory_trigger ON subcategories;
DROP FUNCTION IF EXISTS prevent_duplicate_subcategory;