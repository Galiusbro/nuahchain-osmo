-- Migration: User profile extensions
-- Created: 2025-11-14
-- Description: Добавляет поле image_url для аватарки пользователя

ALTER TABLE users ADD COLUMN IF NOT EXISTS image_url VARCHAR(500);

CREATE INDEX IF NOT EXISTS idx_users_image_url ON users(image_url) WHERE image_url IS NOT NULL;

