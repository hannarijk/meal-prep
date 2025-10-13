CREATE TABLE IF NOT EXISTS recipe_catalogue.categories
(
    id          SERIAL PRIMARY KEY,
    name        VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT categories_name_not_empty CHECK (length(trim(name)) > 0)
);

CREATE TABLE IF NOT EXISTS recipe_catalogue.recipes
(
    id          SERIAL PRIMARY KEY,
    name        VARCHAR(200) NOT NULL,
    description TEXT,
    category_id INTEGER REFERENCES recipe_catalogue.categories (id),
    created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_recipes_category ON recipe_catalogue.recipes (category_id);
CREATE INDEX IF NOT EXISTS idx_recipes_name ON recipe_catalogue.recipes (name);

INSERT INTO recipe_catalogue.categories (name, description)
VALUES ('Meat', 'Meat-based dishes'),
       ('Chicken', 'Chicken dishes'),
       ('Fish', 'Fish and seafood'),
       ('Snacks', 'Light meals and snacks'),
       ('Desserts', 'Sweet treats')
ON CONFLICT (name) DO NOTHING;

INSERT INTO recipe_catalogue.recipes (name, description, category_id)
VALUES ('Pasta with Meatballs', 'Pasta with meatballs Description', 1),
       ('Caesar Salad', 'Fresh romaine with caesar dressing', 2),
       ('Grilled Chicken', 'Juicy grilled chicken breast', 2),
       ('Grilled Salmon', 'Grilled Salmon', 3),
       ('Banana Bread', 'The best banana bread ever', 5)
ON CONFLICT DO NOTHING;