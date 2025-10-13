CREATE TABLE IF NOT EXISTS recipe_catalogue.ingredients
(
    id          SERIAL PRIMARY KEY,
    name        VARCHAR(100)                        NOT NULL UNIQUE,
    description TEXT,
    category    VARCHAR(50)                         NOT NULL,
    created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    CONSTRAINT ingredients_name_not_empty CHECK (length(trim(name)) > 0),
    CONSTRAINT ingredients_category_valid CHECK (category IN
                                                 ('Meat', 'Vegetables', 'Dairy', 'Grains', 'Spices', 'Oils', 'Fish',
                                                  'Fruits'))
);

CREATE INDEX IF NOT EXISTS idx_ingredients_category ON recipe_catalogue.ingredients (category);
CREATE INDEX IF NOT EXISTS idx_ingredients_name ON recipe_catalogue.ingredients (name);

-- Add some samples
INSERT INTO recipe_catalogue.ingredients (name, description, category)
VALUES ('Chicken Breast', 'Boneless skinless chicken breast', 'Meat'),
       ('Ground Beef', 'Lean ground beef', 'Meat'),
       ('Salmon Fillet', 'Fresh salmon fillet', 'Fish'),
       --('Pasta', 'Dry pasta noodles', 'Grains'),
       ('Rice', 'Long grain rice', 'Grains'),
       ('Onion', 'Yellow onion', 'Vegetables'),
       ('Garlic', 'Fresh garlic cloves', 'Vegetables'),
       --('Tomato', 'Fresh tomatoes', 'Vegetables'),
       ('Lettuce', 'Romaine lettuce', 'Vegetables'),
       ('Parmesan Cheese', 'Grated parmesan', 'Dairy'),
       ('Olive Oil', 'Extra virgin olive oil', 'Oils'),
       ('Salt', 'Table salt', 'Spices'),
       ('Black Pepper', 'Ground black pepper', 'Spices'),
       ('Eggs', 'Large eggs', 'Dairy')
ON CONFLICT (name) DO NOTHING;