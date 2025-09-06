-- User preferences table
CREATE TABLE IF NOT EXISTS recommendations.user_preferences (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    preferred_categories INTEGER[],
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id)
);

-- Cooking history table
CREATE TABLE IF NOT EXISTS recommendations.cooking_history (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    recipe_id INTEGER NOT NULL,
    cooked_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    rating INTEGER CHECK (rating BETWEEN 1 AND 5)
);

-- Recommendation history table
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