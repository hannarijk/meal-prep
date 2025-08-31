-- Create databases for each service (we'll keep them in same DB but separate schemas)
CREATE SCHEMA IF NOT EXISTS auth;
CREATE SCHEMA IF NOT EXISTS recipe_catalogue;
CREATE SCHEMA IF NOT EXISTS recommendations;

-- Users table (auth service)
CREATE TABLE IF NOT EXISTS auth.users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_users_email ON auth.users(email);

------------------------------------------------------------------------------------------------------------------------

-- Recipe catalogue tables
CREATE TABLE IF NOT EXISTS recipe_catalogue.categories (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS recipe_catalogue.recipes (
    id SERIAL PRIMARY KEY,
    name VARCHAR(200) NOT NULL,
    description TEXT,
    category_id INTEGER REFERENCES recipe_catalogue.categories(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_recipes_category ON recipe_catalogue.recipes(category_id);
CREATE INDEX IF NOT EXISTS idx_recipes_name ON recipe_catalogue.recipes(name);

------------------------------------------------------------------------------------------------------------------------

-- Recommendations tables
CREATE TABLE IF NOT EXISTS recommendations.user_preferences (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    preferred_categories INTEGER[],
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id)
);

CREATE TABLE IF NOT EXISTS recommendations.cooking_history (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    recipe_id INTEGER NOT NULL,
    cooked_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    rating INTEGER CHECK (rating BETWEEN 1 AND 5)
);

CREATE TABLE IF NOT EXISTS recommendations.recommendation_history (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    recipe_id INTEGER NOT NULL,
    recommended_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    clicked BOOLEAN DEFAULT FALSE,
    algorithm_used VARCHAR(50)
);

CREATE INDEX IF NOT EXISTS idx_user_preferences_user_id ON recommendations.user_preferences(user_id);
CREATE INDEX IF NOT EXISTS idx_cooking_history_user_id ON recommendations.cooking_history(user_id);
CREATE INDEX IF NOT EXISTS idx_cooking_history_recipe_id ON recommendations.cooking_history(recipe_id);
CREATE INDEX IF NOT EXISTS idx_cooking_history_user_recipe ON recommendations.cooking_history(user_id, recipe_id);
CREATE INDEX IF NOT EXISTS idx_cooking_history_cooked_at ON recommendations.cooking_history(cooked_at);
CREATE INDEX IF NOT EXISTS idx_recommendation_history_user_id ON recommendations.recommendation_history(user_id);

------------------------------------------------------------------------------------------------------------------------

-- Insert sample categories
INSERT INTO recipe_catalogue.categories (name, description) VALUES
    ('Meat', 'Morning meals and beverages'),
    ('Chicken', 'Midday meals and snacks'),
    ('Fish', 'Evening meals and hearty recipes'),
    ('Snacks', 'Light meals and quick bites'),
    ('Desserts', 'Sweet treats and after-meal delights')
ON CONFLICT (name) DO NOTHING;

-- Insert sample recipes
INSERT INTO recipe_catalogue.recipes (name, description, category_id) VALUES
    ('Pasta with Meatballs', 'Pasta with meatballs Description', 1),
    ('Caesar Salad', 'Fresh romaine with caesar dressing', 2),
    ('Grilled Chicken', 'Juicy grilled chicken breast', 2),
    ('Grilled Salmon', 'Grilled Salmon', 3),
    ('Banana Bread', 'The best banana bread ever', 5)
ON CONFLICT DO NOTHING;