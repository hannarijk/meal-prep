-- Create databases for each service (we'll keep them in same DB but separate schemas)
CREATE SCHEMA IF NOT EXISTS auth;
CREATE SCHEMA IF NOT EXISTS dish_catalogue;
CREATE SCHEMA IF NOT EXISTS recommendations;

CREATE TABLE IF NOT EXISTS auth.users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_users_email ON auth.users(email);

------------------------------------------------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS dish_catalogue.categories (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS dish_catalogue.dishes (
    id SERIAL PRIMARY KEY,
    name VARCHAR(200) NOT NULL,
    description TEXT,
    category_id INTEGER REFERENCES dish_catalogue.categories(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_dishes_category ON dish_catalogue.dishes(category_id);
CREATE INDEX IF NOT EXISTS idx_dishes_name ON dish_catalogue.dishes(name);

INSERT INTO dish_catalogue.categories (name, description) VALUES
    ('Meat', 'Morning meals and beverages'),
    ('Chicken', 'Midday meals and snacks'),
    ('Fish', 'Evening meals and hearty dishes'),
    ('Snacks', 'Light meals and quick bites'),
    ('Desserts', 'Sweet treats and after-meal delights')
ON CONFLICT (name) DO NOTHING;

INSERT INTO dish_catalogue.dishes (name, description, category_id) VALUES
   ('Pasta with Meatballs', 'Pasta with meatballs Description', 1),
   ('Caesar Salad', 'Fresh romaine with caesar dressing', 2),
   ('Grilled Chicken', 'Juicy grilled chicken breast', 2),
   ('Grilled Salmon', 'Grilled Salmon', 3),
   ('Banana Bread', 'The best banana bread ever', 5)
ON CONFLICT DO NOTHING;

------------------------------------------------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS recommendations.user_preferences (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    preferred_categories INTEGER[],
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id)
);

-- Track when users actually cook dishes
CREATE TABLE IF NOT EXISTS recommendations.cooking_history (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    dish_id INTEGER NOT NULL,
    cooked_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    rating INTEGER CHECK (rating BETWEEN 1 AND 5) -- optional user rating
);

-- Recommendation logs for analytics
CREATE TABLE IF NOT EXISTS recommendations.recommendation_history (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    dish_id INTEGER NOT NULL,
    recommended_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    clicked BOOLEAN DEFAULT FALSE,
    algorithm_used VARCHAR(50) -- 'preference', 'time_decay', 'hybrid'
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_user_preferences_user_id ON recommendations.user_preferences(user_id);
CREATE INDEX IF NOT EXISTS idx_cooking_history_user_id ON recommendations.cooking_history(user_id);
CREATE INDEX IF NOT EXISTS idx_cooking_history_dish_id ON recommendations.cooking_history(dish_id);
CREATE INDEX IF NOT EXISTS idx_cooking_history_user_dish ON recommendations.cooking_history(user_id, dish_id);
CREATE INDEX IF NOT EXISTS idx_cooking_history_cooked_at ON recommendations.cooking_history(cooked_at);
CREATE INDEX IF NOT EXISTS idx_recommendation_history_user_id ON recommendations.recommendation_history(user_id);