-- Category Types Table
CREATE TABLE category_types (
    id SERIAL PRIMARY KEY,
    name VARCHAR(30) NOT NULL UNIQUE CHECK (LENGTH(name) > 1)
);

-- Categories Table
CREATE TABLE categories (
    id SERIAL PRIMARY KEY,
    type_id INT NOT NULL REFERENCES category_types(id),
    name VARCHAR(100) NOT NULL CHECK (LENGTH(name) > 1)
);

-- Subcategories Table
CREATE TABLE subcategories (
    id SERIAL PRIMARY KEY,
    category_id INT NOT NULL REFERENCES categories(id),
    name VARCHAR(100) NOT NULL CHECK (LENGTH(name) > 1)
);

-- Services Table
CREATE TABLE services (
    id SERIAL PRIMARY KEY,
    subcategory_id INT NOT NULL REFERENCES subcategories(id),
    name VARCHAR(100) NOT NULL,
    description TEXT
);

-- Provider Services Table
CREATE TABLE provider_services (
    provider_id BIGINT NOT NULL,
    service_id INT NOT NULL REFERENCES services(id),
    PRIMARY KEY(provider_id, service_id)
);


CREATE OR REPLACE FUNCTION prevent_duplicate_type_name()
RETURNS TRIGGER AS $$
BEGIN
    IF EXISTS (
        SELECT 1 
        FROM categories 
        WHERE type_id = NEW.type_id 
        AND name = NEW.name
        AND id <> NEW.id
    ) THEN
        RAISE EXCEPTION 'A row with the same type_id and name already exists.';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER prevent_duplicate_trigger
BEFORE INSERT OR UPDATE ON categories
FOR EACH ROW
EXECUTE FUNCTION prevent_duplicate_type_name();

CREATE OR REPLACE FUNCTION prevent_duplicate_subcategory()
RETURNS TRIGGER AS $$
BEGIN
    IF EXISTS (
        SELECT 1 
        FROM subcategories 
        WHERE name = NEW.name AND category_id = NEW.category_id
    ) THEN
        RAISE EXCEPTION 'Duplicate subcategory name and category id combination';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER prevent_duplicate_subcategory_trigger
BEFORE INSERT OR UPDATE ON subcategories
FOR EACH ROW EXECUTE FUNCTION prevent_duplicate_subcategory();
