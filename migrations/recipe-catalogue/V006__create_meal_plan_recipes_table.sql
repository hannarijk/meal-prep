-- Junction table linking meal plans to recipes
CREATE TABLE IF NOT EXISTS recipe_catalogue.meal_plan_recipes
(
    id           SERIAL PRIMARY KEY,
    meal_plan_id INTEGER     NOT NULL REFERENCES recipe_catalogue.meal_plans (id),
    recipe_id    INTEGER     NOT NULL REFERENCES recipe_catalogue.recipes (id),

    meal_type    VARCHAR(20) NOT NULL CHECK (meal_type IN ('breakfast', 'main')),

    created_at   TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    -- Prevent duplicate recipe in same meal type within a meal plan
    CONSTRAINT unique_meal_plan_recipe UNIQUE (meal_plan_id, recipe_id, meal_type)
);

CREATE INDEX idx_meal_plan_recipes_plan ON recipe_catalogue.meal_plan_recipes (meal_plan_id);
CREATE INDEX idx_meal_plan_recipes_recipe ON recipe_catalogue.meal_plan_recipes (recipe_id);

-- Add some samples (assumes we have meal_plan with ID 1 and recipe with IDs 1-5)
INSERT INTO recipe_catalogue.meal_plan_recipes (meal_plan_id, recipe_id, meal_type)
VALUES (1, 4, 'main'),
       (1, 1, 'main'),
       (1, 2, 'main'),
       (1, 5, 'breakfast');