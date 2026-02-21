INSERT INTO auth.users (email, password_hash, created_at, updated_at)
VALUES ('system@mealprep.internal', '!system-user-cannot-login', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT (email) DO NOTHING;

ALTER TABLE recipe_catalogue.recipes
    ADD COLUMN IF NOT EXISTS user_id INT;

UPDATE recipe_catalogue.recipes
SET user_id = (SELECT id FROM auth.users WHERE email = 'system@mealprep.internal')
WHERE user_id IS NULL;

ALTER TABLE recipe_catalogue.recipes
    ALTER COLUMN user_id SET NOT NULL;

CREATE INDEX IF NOT EXISTS idx_recipes_user_id ON recipe_catalogue.recipes (user_id);