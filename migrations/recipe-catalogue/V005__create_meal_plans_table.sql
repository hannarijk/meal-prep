-- Create meal_plans table
-- Supports flexible date ranges (week/biweek/month/custom)
CREATE TABLE IF NOT EXISTS recipe_catalogue.meal_plans
(
    id           SERIAL PRIMARY KEY,
    user_id      INTEGER             NOT NULL,

    -- Flexible period (can be any date range)
    title        VARCHAR(200)        NOT NULL,
    period_start DATE                NOT NULL,
    period_end   DATE                NOT NULL,
    period_type  VARCHAR(20) DEFAULT 'week' CHECK (period_type IN ('week', 'biweek', 'month', 'custom')),

    -- Public sharing
    is_public    BOOLEAN     DEFAULT false,
    slug         VARCHAR(255) UNIQUE NOT NULL,

    -- Timestamps
    created_at   TIMESTAMP   DEFAULT CURRENT_TIMESTAMP,
    updated_at   TIMESTAMP   DEFAULT CURRENT_TIMESTAMP,
    deleted_at   TIMESTAMP
);

CREATE INDEX idx_meal_plans_user_id ON recipe_catalogue.meal_plans (user_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_meal_plans_period ON recipe_catalogue.meal_plans (user_id, period_start DESC) WHERE deleted_at IS NULL;
CREATE INDEX idx_meal_plans_public_slug ON recipe_catalogue.meal_plans (slug) WHERE is_public = true AND deleted_at IS NULL;

-- Add some samples (assumes we have user with ID 1)
-- Week plan
INSERT INTO recipe_catalogue.meal_plans (user_id, title, period_start, period_end, period_type, slug)
VALUES (1, 'Week of Oct 6-12', '2025-10-06', '2025-10-12', 'week', 'week-of-oct-6-12');

-- Biweekly plan
INSERT INTO recipe_catalogue.meal_plans (user_id, title, period_start, period_end, period_type, slug)
VALUES (1, 'Weeks of Oct 6-19', '2025-10-06', '2025-10-19', 'biweek', 'weeks-of-6-19');

-- Monthly plan
INSERT INTO recipe_catalogue.meal_plans (user_id, title, period_start, period_end, period_type, slug)
VALUES (1, 'October 2025', '2025-10-01', '2025-10-31', 'month', 'month-oct-2025');

-- Multiple plans per period
INSERT INTO recipe_catalogue.meal_plans (user_id, title, period_start, period_end, period_type, slug)
VALUES (1, 'Week Oct 6-12, Vegetarian', '2025-10-06', '2025-10-12', 'week', 'week-of-oct-6-12-vegetarian');
INSERT INTO recipe_catalogue.meal_plans (user_id, title, period_start, period_end, period_type, slug)
VALUES (1, 'Week Oct 6-12, Family', '2025-10-06', '2025-10-12', 'week', 'week-of-oct-6-12-family');