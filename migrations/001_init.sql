CREATE TABLE IF NOT EXISTS roles (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS users (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    email VARCHAR(50) NOT NULL UNIQUE CHECK(email LIKE '%_@__%.__%'),
    phone CHAR(12) UNIQUE CHECK (phone GLOB '+[0-9][0-9][0-9][0-9][0-9][0-9][0-9][0-9][0-9][0-9][0-9]'),/*Использую GLOB, тк like в sqlite поддерживает только % и _, что позволяет вводить символы, а не цифры*/
    password_hash VARCHAR(255) NOT NULL,
    role_id VARCHAR(36) NOT NULL REFERENCES roles(id),
    registration_date DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    balance REAL NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS products (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    type VARCHAR(50),
    gender CHAR(1) CHECK (gender IN ('M','W','U')),
    size VARCHAR(50),
    color VARCHAR(50),
    description VARCHAR(255),
    brand VARCHAR(50),
    season VARCHAR(50),
    price REAL NOT NULL CHECK (price >= 0),
    quantity INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS photos (
    id VARCHAR(36) PRIMARY KEY,
    filename VARCHAR(50) NOT NULL,
    product_id VARCHAR(36) NOT NULL REFERENCES products(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS carts (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    product_id VARCHAR(36) NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    count INTEGER NOT NULL DEFAULT 1 CHECK (count > 0),
    UNIQUE (user_id, product_id)
);

CREATE TABLE IF NOT EXISTS orders (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL REFERENCES users(id),
    datetime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    status VARCHAR(50) NOT NULL DEFAULT 'Принят'
        CHECK (status IN ('Принят','В работе','Собран','Готов к выдаче','Получен','Отменён')),
    refunded INTEGER NOT NULL DEFAULT 0,
    cancel_reason VARCHAR(255) NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS order_items (
    id VARCHAR(36) PRIMARY KEY,
    order_id VARCHAR(36) NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_id VARCHAR(36) NOT NULL REFERENCES products(id),
    price REAL NOT NULL CHECK (price >= 0),
    quantity INTEGER NOT NULL DEFAULT 1 CHECK (quantity > 0)
);

CREATE INDEX IF NOT EXISTS idx_products_type  ON products(type);
CREATE INDEX IF NOT EXISTS idx_products_brand ON products(brand);
CREATE INDEX IF NOT EXISTS idx_orders_status  ON orders(status);
CREATE INDEX IF NOT EXISTS idx_orders_datetime ON orders(datetime);

INSERT OR IGNORE INTO roles (id, name) VALUES
    ('00000000-0000-0000-0000-000000000001', 'Гость'),
    ('00000000-0000-0000-0000-000000000002', 'Клиент'),
    ('00000000-0000-0000-0000-000000000003', 'Менеджер'),
    ('00000000-0000-0000-0000-000000000004', 'Администратор');

INSERT OR IGNORE INTO users (id, name, email, phone, password_hash, role_id) VALUES
('550e8400-e29b-41d4-a716-446655440001', 'Admin', 'admin@shop.ru', '+70000000001','$2a$10$nuh0Hvg0E1XUeGZzpApKcOacEc/BNuHKUN6UF1mowHZ9j7.EY8H3q','00000000-0000-0000-0000-000000000004');