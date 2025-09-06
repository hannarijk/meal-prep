CREATE TABLE IF NOT EXISTS recipe_catalogue.recipe_ingredients (
    id SERIAL PRIMARY KEY,
    recipe_id INTEGER NOT NULL REFERENCES recipe_catalogue.recipes(id) ON DELETE CASCADE,
    ingredient_id INTEGER NOT NULL REFERENCES recipe_catalogue.ingredients(id) ON DELETE RESTRICT,
    quantity DECIMAL(8,2) NOT NULL,
    unit VARCHAR(20) NOT NULL,
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    -- Business constraints
    CONSTRAINT recipe_ingredients_quantity_positive CHECK (quantity > 0),
    CONSTRAINT recipe_ingredients_unit_not_empty CHECK (length(trim(unit)) > 0),
    CONSTRAINT recipe_ingredients_unique_per_recipe UNIQUE (recipe_id, ingredient_id)
);

CREATE INDEX IF NOT EXISTS idx_recipe_ingredients_recipe_id ON recipe_catalogue.recipe_ingredients (recipe_id);
CREATE INDEX IF NOT EXISTS idx_recipe_ingredients_ingredient_id ON recipe_catalogue.recipe_ingredients (ingredient_id);

-- Add some sample recipe ingredients (assumes we have recipes with IDs 1-4)
INSERT INTO recipe_catalogue.recipe_ingredients (recipe_id, ingredient_id, quantity, unit, notes)
SELECT * FROM (VALUES
    -- Pasta with Meatballs (recipe_id = 1)
    (1, 4, 200.00, 'grams', 'any pasta shape'),
    (1, 2, 300.00, 'grams', 'for meatballs'),
    (1, 8, 2.00, 'pieces', 'diced'),
    (1, 6, 1.00, 'piece', 'diced'),
    (1, 7, 3.00, 'cloves', 'minced'),

    -- Caesar Salad (recipe_id = 2)
    (2, 9, 1.00, 'head', 'chopped'),
    (2, 10, 50.00, 'grams', 'grated'),
    (2, 11, 3.00, 'tablespoons', 'for dressing'),

    -- Grilled Chicken (recipe_id = 3)
    (3, 1, 400.00, 'grams', 'boneless'),
    (3, 11, 2.00, 'tablespoons', 'for marinade'),
    (3, 12, 1.00, 'teaspoon', 'to taste'),

    -- Grilled Salmon (recipe_id = 4)
    (4, 3, 300.00, 'grams', 'skin-on fillet'),
    (4, 11, 1.00, 'tablespoon', ''),
    (4, 12, 0.50, 'teaspoon', 'to taste')
) AS sample_data(recipe_id, ingredient_id, quantity, unit, notes)
WHERE EXISTS (SELECT 1 FROM recipe_catalogue.recipes WHERE id = sample_data.recipe_id)
  AND EXISTS (SELECT 1 FROM recipe_catalogue.ingredients WHERE id = sample_data.ingredient_id)
ON CONFLICT (recipe_id, ingredient_id) DO NOTHING;