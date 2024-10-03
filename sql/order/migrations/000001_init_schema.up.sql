-- Enums
CREATE TYPE CAPITAL_TYPE AS ENUM ('PRIMARY', 'ADMIN', 'MINOR');
CREATE TYPE ORDER_STATUS AS ENUM ('PENDING', 'PAUSED', 'CANCELLED', 'COMPLETED');

-- Currencies table
CREATE TABLE currencies (
    id SERIAL PRIMARY KEY,
    currency VARCHAR(100) NOT NULL,
    code VARCHAR(10) NOT NULL,
    symbol VARCHAR(10) NOT NULL
);

-- Cities Table
CREATE TABLE cities (
    city_id SERIAL PRIMARY KEY,
    name VARCHAR(120) NOT NULL,
    location GEOGRAPHY(POINT, 4326),
    country VARCHAR(120) NOT NULL,
    iso2 VARCHAR(2) NOT NULL,
    admin_name VARCHAR(120),
    capital CAPITAL_TYPE
);

-- Orders table
CREATE TABLE orders (
    id BIGINT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    service_id INTEGER NOT NULL,
    location GEOGRAPHY(POINT, 4326),
    time_start TIMESTAMP NOT NULL,
    time_end TIMESTAMP NOT NULL,
    description TEXT NOT NULL
);

-- Provider Locations Table
CREATE TABLE provider_locations (
    provider_id BIGINT PRIMARY KEY NOT NULL, 
    city_id INT NOT NULL REFERENCES cities(city_id),
    location GEOGRAPHY(POINT, 4326),
    address VARCHAR(255) NOT NULL
);

-- Offers Table
CREATE TABLE offers (
    id BIGINT PRIMARY KEY NOT NULL,
    provider_id BIGINT NOT NULL,
    order_id BIGINT REFERENCES orders(id) NOT NULL,
    price NUMERIC(10, 4) NOT NULL,
    currency_id INT REFERENCES currencies(id) NOT NULL,
    book_time TIMESTAMP NOT NULL
);

-- jobs table
CREATE TABLE jobs (
    id BIGINT PRIMARY KEY ,
    offer_id BIGINT NOT NULL REFERENCES offers(id),
    order_id BIGINT NOT NULL REFERENCES orders(id),
    user_id BIGINT NOT NULL,
    provider_id BIGINT NOT NULL,
    service_id INT NOT NULL,
    status ORDER_STATUS DEFAULT 'PENDING',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Reviews Table
CREATE TABLE reviews (
    id BIGINT PRIMARY KEY NOT NULL,
    job_id BIGINT NOT NULL REFERENCES jobs(id),
    user_id BIGINT NOT NULL,
    provider_id BIGINT NOT NULL,
    rating NUMERIC(3,2) CHECK (rating >= 0 AND rating <= 5 AND (rating * 10) % 5 = 0) NOT NULL,
    comment TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
