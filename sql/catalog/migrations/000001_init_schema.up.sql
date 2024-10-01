-- Category Types Table
CREATE TABLE category_types (
    id SERIAL PRIMARY KEY,
    name VARCHAR(30) NOT NULL UNIQUE
);

-- Categories Table
CREATE TABLE categories (
    id SERIAL PRIMARY KEY,
    type_id INT NOT NULL REFERENCES category_types(id),
    name VARCHAR(100) NOT NULL
);

-- Subcategories Table
CREATE TABLE subcategories (
    id SERIAL PRIMARY KEY,
    category_id INT NOT NULL REFERENCES categories(id),
    name VARCHAR(100) NOT NULL
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
