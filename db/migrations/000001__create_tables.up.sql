--создаем типы ENUM
CREATE TYPE user_role AS ENUM (
    'admin',
    'worker'
);

CREATE TYPE reception_status AS ENUM (
    'close',
    'in_progress'
);

--Создание таблицы пользователь
CREATE TABLE "user" (
    id              UUID PRIMARY KEY,
    email           TEXT UNIQUE NOT NULL,
    password_hash   TEXT NOT NULL,
    role            user_role NOT NULL,
    created_at      TIMESTAMP NOT NULL DEFAULT now()
);

-- Создание таблицы Пунктов выдачи заказов (ПВЗ)
CREATE TABLE pickup_point (
    id                      UUID PRIMARY KEY,
    registration_date       TIMESTAMP NOT NULL DEFAULT now(),
    city                    TEXT NOT NULL
);

-- Создание таблицы Приемок товара
CREATE TABLE reception (
    id                      UUID PRIMARY KEY,
    reception_date          TIMESTAMP NOT NULL DEFAULT now(),
    pickup_point_id         UUID NOT NULL REFERENCES pickup_point(id) ON DELETE CASCADE,
    status                  reception_status NOT NULL
);

-- Создание таблицы Товаров
CREATE TABLE product (
    id                      UUID PRIMARY KEY,
    reception_id            UUID NOT NULL REFERENCES reception(id) ON DELETE CASCADE,
    product_type            TEXT NOT NULL,
    reception_date          TIMESTAMP NOT NULL DEFAULT now()
);

-- Для фильтрации по дате
CREATE INDEX IF NOT EXISTS reception_reception_date_idx ON reception(reception_date);