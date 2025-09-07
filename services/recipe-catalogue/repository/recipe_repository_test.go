package repository

import (
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"meal-prep/shared/database"
	"meal-prep/shared/models"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// RecipeRepositoryTestSuite tests all repository data access operations
type RecipeRepositoryTestSuite struct {
	suite.Suite
	db   *database.DB
	mock sqlmock.Sqlmock
	repo RecipeRepository
}

func (suite *RecipeRepositoryTestSuite) SetupTest() {
	db, mock, err := sqlmock.New()
	require.NoError(suite.T(), err)

	suite.db = &database.DB{DB: db}
	suite.mock = mock
	suite.repo = NewRecipeRepository(suite.db)
}

func (suite *RecipeRepositoryTestSuite) TearDownTest() {
	suite.db.Close()
}

// =============================================================================
// BASIC CRUD OPERATIONS
// =============================================================================

func (suite *RecipeRepositoryTestSuite) TestGetAll_ReturnsRecipes() {
	// Arrange
	now := time.Now()
	suite.mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT d.id, d.name, d.description, d.category_id, d.created_at, d.updated_at,
		       c.id, c.name, c.description
		FROM recipe_catalogue.recipes d
		LEFT JOIN recipe_catalogue.categories c ON d.category_id = c.id
		ORDER BY d.name`)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "name", "description", "category_id", "created_at", "updated_at",
			"c_id", "c_name", "c_description",
		}).
			AddRow(1, "Pasta", "Italian dish", 1, now, now, 1, "Italian", "Italian cuisine").
			AddRow(2, "Salad", "Fresh salad", 2, now, now, 2, "Healthy", "Healthy food"))

	// Act
	recipes, err := suite.repo.GetAll()

	// Assert
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), recipes, 2)
	assert.Equal(suite.T(), "Pasta", recipes[0].Name)
	assert.Equal(suite.T(), "Italian dish", *recipes[0].Description)
	assert.Equal(suite.T(), "Italian", recipes[0].Category.Name)
}

func (suite *RecipeRepositoryTestSuite) TestGetAll_DatabaseError() {
	// Arrange
	suite.mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT d.id, d.name, d.description, d.category_id, d.created_at, d.updated_at,
		       c.id, c.name, c.description
		FROM recipe_catalogue.recipes d
		LEFT JOIN recipe_catalogue.categories c ON d.category_id = c.id
		ORDER BY d.name`)).
		WillReturnError(errors.New("connection lost"))

	// Act
	recipes, err := suite.repo.GetAll()

	// Assert
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), recipes)
	assert.Contains(suite.T(), err.Error(), "connection lost")
}

func (suite *RecipeRepositoryTestSuite) TestGetByID_ReturnsRecipe() {
	// Arrange
	now := time.Now()
	suite.mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT d.id, d.name, d.description, d.category_id, d.created_at, d.updated_at,
		       c.id, c.name, c.description
		FROM recipe_catalogue.recipes d
		LEFT JOIN recipe_catalogue.categories c ON d.category_id = c.id
		WHERE d.id = $1`)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "name", "description", "category_id", "created_at", "updated_at",
			"c_id", "c_name", "c_description",
		}).
			AddRow(1, "Carbonara", "Creamy pasta", 1, now, now, 1, "Italian", "Italian cuisine"))

	// Act
	recipe, err := suite.repo.GetByID(1)

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), recipe)
	assert.Equal(suite.T(), 1, recipe.ID)
	assert.Equal(suite.T(), "Carbonara", recipe.Name)
}

func (suite *RecipeRepositoryTestSuite) TestGetByID_NotFound() {
	// Arrange
	suite.mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT d.id, d.name, d.description, d.category_id, d.created_at, d.updated_at,
		       c.id, c.name, c.description
		FROM recipe_catalogue.recipes d
		LEFT JOIN recipe_catalogue.categories c ON d.category_id = c.id
		WHERE d.id = $1`)).
		WithArgs(999).
		WillReturnError(sql.ErrNoRows)

	// Act
	recipe, err := suite.repo.GetByID(999)

	// Assert
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), sql.ErrNoRows, err)
	assert.Nil(suite.T(), recipe)
}

func (suite *RecipeRepositoryTestSuite) TestGetByCategory_ReturnsFilteredRecipes() {
	// Arrange
	now := time.Now()
	suite.mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT d.id, d.name, d.description, d.category_id, d.created_at, d.updated_at,
		       c.id, c.name, c.description
		FROM recipe_catalogue.recipes d
		LEFT JOIN recipe_catalogue.categories c ON d.category_id = c.id
		WHERE d.category_id = $1
		ORDER BY d.name`)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "name", "description", "category_id", "created_at", "updated_at",
			"c_id", "c_name", "c_description",
		}).
			AddRow(1, "Pasta Dish", "Italian pasta", 1, now, now, 1, "Italian", "Italian cuisine"))

	// Act
	recipes, err := suite.repo.GetByCategory(1)

	// Assert
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), recipes, 1)
	assert.Equal(suite.T(), 1, *recipes[0].CategoryID)
}

func (suite *RecipeRepositoryTestSuite) TestCreate_InsertsAndReturnsRecipe() {
	// Arrange
	req := models.CreateRecipeRequest{
		Name:        "New Recipe",
		Description: "Tasty dish",
		CategoryID:  1,
	}
	now := time.Now()

	suite.mock.ExpectQuery(regexp.QuoteMeta(`
		INSERT INTO recipe_catalogue.recipes (name, description, category_id, created_at, updated_at)
		VALUES ($1, $2, $3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id, name, description, category_id, created_at, updated_at`)).
		WithArgs(req.Name, req.Description, req.CategoryID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description", "category_id", "created_at", "updated_at"}).
			AddRow(1, req.Name, req.Description, req.CategoryID, now, now))

	// Act
	recipe, err := suite.repo.Create(req)

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), recipe)
	assert.Equal(suite.T(), req.Name, recipe.Name)
	assert.Equal(suite.T(), req.Description, *recipe.Description)
	assert.Equal(suite.T(), req.CategoryID, *recipe.CategoryID)
}

func (suite *RecipeRepositoryTestSuite) TestCreate_DatabaseError() {
	// Arrange
	req := models.CreateRecipeRequest{
		Name:        "Recipe",
		Description: "Description",
		CategoryID:  1,
	}

	suite.mock.ExpectQuery(regexp.QuoteMeta(`
		INSERT INTO recipe_catalogue.recipes (name, description, category_id, created_at, updated_at)
		VALUES ($1, $2, $3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id, name, description, category_id, created_at, updated_at`)).
		WithArgs(req.Name, req.Description, req.CategoryID).
		WillReturnError(errors.New("constraint violation"))

	// Act
	recipe, err := suite.repo.Create(req)

	// Assert
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), recipe)
}

func (suite *RecipeRepositoryTestSuite) TestUpdate_UpdatesAndReturnsRecipe() {
	// Arrange
	req := models.UpdateRecipeRequest{
		Name:        "Updated Name",
		Description: "Updated desc",
		CategoryID:  2,
	}
	now := time.Now()

	suite.mock.ExpectQuery(regexp.QuoteMeta(`
		UPDATE recipe_catalogue.recipes 
        SET name = $2,
            description = $3,
            category_id = $4,
            updated_at = CURRENT_TIMESTAMP
        WHERE id = $1
		RETURNING id, name, description, category_id, created_at, updated_at`)).
		WithArgs(1, req.Name, req.Description, req.CategoryID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description", "category_id", "created_at", "updated_at"}).
			AddRow(1, req.Name, req.Description, req.CategoryID, now, now))

	// Act
	recipe, err := suite.repo.Update(1, req)

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), recipe)
	assert.Equal(suite.T(), req.Name, recipe.Name)
}

func (suite *RecipeRepositoryTestSuite) TestUpdate_NotFound() {
	// Arrange
	req := models.UpdateRecipeRequest{Name: "Updated"}

	suite.mock.ExpectQuery(regexp.QuoteMeta(`
		UPDATE recipe_catalogue.recipes 
        SET name = $2,
            description = $3,
            category_id = $4,
            updated_at = CURRENT_TIMESTAMP
        WHERE id = $1
		RETURNING id, name, description, category_id, created_at, updated_at`)).
		WithArgs(999, req.Name, req.Description, req.CategoryID).
		WillReturnError(sql.ErrNoRows)

	// Act
	recipe, err := suite.repo.Update(999, req)

	// Assert
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), sql.ErrNoRows, err)
	assert.Nil(suite.T(), recipe)
}

func (suite *RecipeRepositoryTestSuite) TestDelete_RemovesRecipe() {
	// Arrange
	suite.mock.ExpectExec(regexp.QuoteMeta("DELETE FROM recipe_catalogue.recipes WHERE id = $1")).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Act
	err := suite.repo.Delete(1)

	// Assert
	assert.NoError(suite.T(), err)
}

func (suite *RecipeRepositoryTestSuite) TestDelete_NotFound() {
	// Arrange
	suite.mock.ExpectExec(regexp.QuoteMeta("DELETE FROM recipe_catalogue.recipes WHERE id = $1")).
		WithArgs(999).
		WillReturnResult(sqlmock.NewResult(0, 0)) // No rows affected

	// Act
	err := suite.repo.Delete(999)

	// Assert
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), sql.ErrNoRows, err)
}

func (suite *RecipeRepositoryTestSuite) TestSearchRecipesByIngredients_ReturnsMatchingRecipes() {
	// Arrange
	ingredientIDs := []int{1, 2}
	now := time.Now()

	suite.mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT DISTINCT r.id, r.name, r.description, r.category_id, r.created_at, r.updated_at,
		                c.id, c.name, c.description
		FROM recipe_catalogue.recipes r
		LEFT JOIN recipe_catalogue.categories c ON r.category_id = c.id
		JOIN recipe_catalogue.recipe_ingredients ri ON r.id = ri.recipe_id
		WHERE ri.ingredient_id = ANY($1)
		ORDER BY r.name`)).
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "name", "description", "category_id", "created_at", "updated_at",
			"c_id", "c_name", "c_description",
		}).
			AddRow(1, "Recipe with Ingredients", "Uses ingredients", 1, now, now, 1, "Category", "Desc"))

	// Act
	recipes, err := suite.repo.SearchRecipesByIngredients(ingredientIDs)

	// Assert
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), recipes, 1)
	assert.Equal(suite.T(), "Recipe with Ingredients", recipes[0].Name)
}

func (suite *RecipeRepositoryTestSuite) TestSearchRecipesByIngredients_EmptySlice() {
	// Act
	recipes, err := suite.repo.SearchRecipesByIngredients([]int{})

	// Assert
	assert.NoError(suite.T(), err)
	assert.Empty(suite.T(), recipes)
}

// =============================================================================
// OPERATIONS WITH INGREDIENTS
// =============================================================================

func (suite *RecipeRepositoryTestSuite) TestGetByIDWithIngredients_ReturnsRecipeWithIngredients() {
	// Arrange
	now := time.Now()

	// Mock GetByID call
	suite.mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT d.id, d.name, d.description, d.category_id, d.created_at, d.updated_at,
		       c.id, c.name, c.description
		FROM recipe_catalogue.recipes d
		LEFT JOIN recipe_catalogue.categories c ON d.category_id = c.id
		WHERE d.id = $1`)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "name", "description", "category_id", "created_at", "updated_at",
			"c_id", "c_name", "c_description",
		}).
			AddRow(1, "Carbonara", "Italian pasta", 1, now, now, 1, "Italian", "Italian cuisine"))

	// Mock GetRecipeIngredients call
	suite.mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT ri.id, ri.recipe_id, ri.ingredient_id, ri.quantity, ri.unit, ri.notes, ri.created_at,
		       i.id, i.name, i.description, i.category, i.created_at
		FROM recipe_catalogue.recipe_ingredients ri
		JOIN recipe_catalogue.ingredients i ON ri.ingredient_id = i.id
		WHERE ri.recipe_id = $1
		ORDER BY ri.id`)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{
			"ri_id", "recipe_id", "ingredient_id", "quantity", "unit", "notes", "ri_created_at",
			"i_id", "i_name", "i_description", "i_category", "i_created_at",
		}).
			AddRow(1, 1, 1, 200.0, "grams", "Fresh pasta", now, 1, "Pasta", "Spaghetti", "Grain", now).
			AddRow(2, 1, 2, 2.0, "pieces", nil, now, 2, "Eggs", "Fresh eggs", "Protein", now))

	// Act
	result, err := suite.repo.GetByIDWithIngredients(1)

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), "Carbonara", result.Recipe.Name)
	assert.Len(suite.T(), result.Ingredients, 2)
	assert.Equal(suite.T(), 200.0, result.Ingredients[0].Quantity)
	assert.Equal(suite.T(), "grams", result.Ingredients[0].Unit)
	assert.Equal(suite.T(), "Pasta", result.Ingredients[0].Ingredient.Name)
}

func (suite *RecipeRepositoryTestSuite) TestGetAllWithIngredients_ReturnsRecipesWithIngredients() {
	// Arrange
	now := time.Now()

	// Mock GetAll call
	suite.mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT d.id, d.name, d.description, d.category_id, d.created_at, d.updated_at,
		       c.id, c.name, c.description
		FROM recipe_catalogue.recipes d
		LEFT JOIN recipe_catalogue.categories c ON d.category_id = c.id
		ORDER BY d.name`)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "name", "description", "category_id", "created_at", "updated_at",
			"c_id", "c_name", "c_description",
		}).
			AddRow(1, "Recipe 1", "Desc 1", 1, now, now, 1, "Category 1", "Cat desc").
			AddRow(2, "Recipe 2", "Desc 2", 2, now, now, 2, "Category 2", "Cat desc"))

	// Mock GetIngredientsForRecipes call
	suite.mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT ri.id, ri.recipe_id, ri.ingredient_id, ri.quantity, ri.unit, ri.notes, ri.created_at,
		       i.id, i.name, i.description, i.category, i.created_at
		FROM recipe_catalogue.recipe_ingredients ri
		JOIN recipe_catalogue.ingredients i ON ri.ingredient_id = i.id
		WHERE ri.recipe_id = ANY($1)
		ORDER BY ri.recipe_id, ri.id`)).
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{
			"ri_id", "recipe_id", "ingredient_id", "quantity", "unit", "notes", "ri_created_at",
			"i_id", "i_name", "i_description", "i_category", "i_created_at",
		}).
			AddRow(1, 1, 1, 100.0, "g", nil, now, 1, "Ingredient 1", "Desc", "Cat", now).
			AddRow(2, 2, 2, 200.0, "ml", nil, now, 2, "Ingredient 2", "Desc", "Cat", now))

	// Act
	result, err := suite.repo.GetAllWithIngredients()

	// Assert
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), result, 2)
	assert.Equal(suite.T(), "Recipe 1", result[0].Recipe.Name)
	assert.Len(suite.T(), result[0].Ingredients, 1)
	assert.Equal(suite.T(), "Ingredient 1", result[0].Ingredients[0].Ingredient.Name)
}

// =============================================================================
// TRANSACTION OPERATIONS
// =============================================================================

func (suite *RecipeRepositoryTestSuite) TestCreateWithIngredients_CommitsTransaction() {
	// Arrange
	req := models.CreateRecipeWithIngredientsRequest{
		Name:        "Recipe with Ingredients",
		Description: stringPtr("A recipe with ingredients"),
		CategoryID:  1,
		Ingredients: []models.AddRecipeIngredientRequest{
			{IngredientID: 1, Quantity: 100.0, Unit: "grams", Notes: stringPtr("Fresh")},
		},
	}
	now := time.Now()

	// Expect transaction
	suite.mock.ExpectBegin()

	// Expect recipe creation
	suite.mock.ExpectQuery(regexp.QuoteMeta(`
		INSERT INTO recipe_catalogue.recipes (name, description, category_id, created_at, updated_at)
		VALUES ($1, $2, $3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id, name, description, category_id, created_at, updated_at`)).
		WithArgs(req.Name, req.Description, req.CategoryID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description", "category_id", "created_at", "updated_at"}).
			AddRow(1, req.Name, *req.Description, req.CategoryID, now, now))

	// Expect ingredient addition
	suite.mock.ExpectExec(regexp.QuoteMeta(`
		INSERT INTO recipe_catalogue.recipe_ingredients (recipe_id, ingredient_id, quantity, unit, notes, created_at)
		VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP)`)).
		WithArgs(1, 1, 100.0, "grams", stringPtr("Fresh")).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Expect commit
	suite.mock.ExpectCommit()

	// Expect ingredient retrieval after commit
	suite.mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT ri.id, ri.recipe_id, ri.ingredient_id, ri.quantity, ri.unit, ri.notes, ri.created_at,
		       i.id, i.name, i.description, i.category, i.created_at
		FROM recipe_catalogue.recipe_ingredients ri
		JOIN recipe_catalogue.ingredients i ON ri.ingredient_id = i.id
		WHERE ri.recipe_id = $1
		ORDER BY ri.id`)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{
			"ri_id", "recipe_id", "ingredient_id", "quantity", "unit", "notes", "ri_created_at",
			"i_id", "i_name", "i_description", "i_category", "i_created_at",
		}).
			AddRow(1, 1, 1, 100.0, "grams", "Fresh", now, 1, "Ingredient", "Desc", "Cat", now))

	// Expect GetByID call for category information
	suite.mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT d.id, d.name, d.description, d.category_id, d.created_at, d.updated_at,
		       c.id, c.name, c.description
		FROM recipe_catalogue.recipes d
		LEFT JOIN recipe_catalogue.categories c ON d.category_id = c.id
		WHERE d.id = $1`)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "name", "description", "category_id", "created_at", "updated_at",
			"c_id", "c_name", "c_description",
		}).
			AddRow(1, req.Name, *req.Description, req.CategoryID, now, now, req.CategoryID, "Category", "Category desc"))

	// Act
	result, err := suite.repo.CreateWithIngredients(req)

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), req.Name, result.Recipe.Name)
	assert.Len(suite.T(), result.Ingredients, 1)
}

func (suite *RecipeRepositoryTestSuite) TestCreateWithIngredients_RollsBackOnError() {
	// Arrange
	req := models.CreateRecipeWithIngredientsRequest{
		Name:       "Recipe",
		CategoryID: 1,
		Ingredients: []models.AddRecipeIngredientRequest{
			{IngredientID: 1, Quantity: 100.0, Unit: "grams"},
		},
	}
	now := time.Now()

	// Expect transaction
	suite.mock.ExpectBegin()

	// Expect recipe creation to succeed
	suite.mock.ExpectQuery(regexp.QuoteMeta(`
		INSERT INTO recipe_catalogue.recipes (name, description, category_id, created_at, updated_at)
		VALUES ($1, $2, $3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id, name, description, category_id, created_at, updated_at`)).
		WithArgs(req.Name, req.Description, req.CategoryID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description", "category_id", "created_at", "updated_at"}).
			AddRow(1, req.Name, "", req.CategoryID, now, now))

	// Expect ingredient addition to fail
	suite.mock.ExpectExec(regexp.QuoteMeta(`
		INSERT INTO recipe_catalogue.recipe_ingredients (recipe_id, ingredient_id, quantity, unit, notes, created_at)
		VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP)`)).
		WithArgs(1, 1, 100.0, "grams", (*string)(nil)).
		WillReturnError(errors.New("constraint violation"))

	// Expect rollback
	suite.mock.ExpectRollback()

	// Act
	result, err := suite.repo.CreateWithIngredients(req)

	// Assert
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "constraint violation")
}

func (suite *RecipeRepositoryTestSuite) TestUpdateWithIngredients_UpdatesBothRecipeAndIngredients() {
	// Arrange
	req := models.UpdateRecipeWithIngredientsRequest{
		Name:        "Updated Recipe",
		Description: "Updated description",
		Ingredients: []models.AddRecipeIngredientRequest{
			{IngredientID: 1, Quantity: 150.0, Unit: "grams"},
		},
	}
	now := time.Now()

	// Expect transaction
	suite.mock.ExpectBegin()

	// Expect recipe update
	suite.mock.ExpectQuery(regexp.QuoteMeta(`
		UPDATE recipe_catalogue.recipes 
		SET name = COALESCE(NULLIF($2, ''), name),
		    description = COALESCE($3, description),
		    category_id = COALESCE(NULLIF($4, 0), category_id),
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING id, name, description, category_id, created_at, updated_at`)).
		WithArgs(1, req.Name, req.Description, req.CategoryID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description", "category_id", "created_at", "updated_at"}).
			AddRow(1, req.Name, req.Description, 1, now, now))

	// Expect ingredient deletion
	suite.mock.ExpectExec(regexp.QuoteMeta("DELETE FROM recipe_catalogue.recipe_ingredients WHERE recipe_id = $1")).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Expect ingredient addition
	suite.mock.ExpectExec(regexp.QuoteMeta(`
		INSERT INTO recipe_catalogue.recipe_ingredients (recipe_id, ingredient_id, quantity, unit, notes, created_at)
		VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP)`)).
		WithArgs(1, 1, 150.0, "grams", (*string)(nil)).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Expect commit
	suite.mock.ExpectCommit()

	// Mock GetByIDWithIngredients call after commit
	suite.mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT d.id, d.name, d.description, d.category_id, d.created_at, d.updated_at,
		       c.id, c.name, c.description
		FROM recipe_catalogue.recipes d
		LEFT JOIN recipe_catalogue.categories c ON d.category_id = c.id
		WHERE d.id = $1`)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "name", "description", "category_id", "created_at", "updated_at",
			"c_id", "c_name", "c_description",
		}).
			AddRow(1, req.Name, req.Description, 1, now, now, 1, "Category", "Desc"))

	suite.mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT ri.id, ri.recipe_id, ri.ingredient_id, ri.quantity, ri.unit, ri.notes, ri.created_at,
		       i.id, i.name, i.description, i.category, i.created_at
		FROM recipe_catalogue.recipe_ingredients ri
		JOIN recipe_catalogue.ingredients i ON ri.ingredient_id = i.id
		WHERE ri.recipe_id = $1
		ORDER BY ri.id`)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{
			"ri_id", "recipe_id", "ingredient_id", "quantity", "unit", "notes", "ri_created_at",
			"i_id", "i_name", "i_description", "i_category", "i_created_at",
		}).
			AddRow(1, 1, 1, 150.0, "grams", nil, now, 1, "Ingredient", "Desc", "Cat", now))

	// Act
	result, err := suite.repo.UpdateWithIngredients(1, req)

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), req.Name, result.Recipe.Name)
	assert.Len(suite.T(), result.Ingredients, 1)
	assert.Equal(suite.T(), 150.0, result.Ingredients[0].Quantity)
}

func (suite *RecipeRepositoryTestSuite) TestUpdateWithIngredients_SkipsIngredientsWhenNil() {
	// Arrange
	req := models.UpdateRecipeWithIngredientsRequest{
		Name:        "Updated Recipe",
		Ingredients: nil, // Should not update ingredients
	}
	now := time.Now()

	// Expect transaction
	suite.mock.ExpectBegin()

	// Expect only recipe update, no ingredient operations
	suite.mock.ExpectQuery(regexp.QuoteMeta(`
		UPDATE recipe_catalogue.recipes 
		SET name = COALESCE(NULLIF($2, ''), name),
		    description = COALESCE($3, description),
		    category_id = COALESCE(NULLIF($4, 0), category_id),
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING id, name, description, category_id, created_at, updated_at`)).
		WithArgs(1, req.Name, req.Description, req.CategoryID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description", "category_id", "created_at", "updated_at"}).
			AddRow(1, req.Name, "", 1, now, now))

	// Expect commit (no ingredient operations)
	suite.mock.ExpectCommit()

	// Mock GetByIDWithIngredients call
	suite.mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT d.id, d.name, d.description, d.category_id, d.created_at, d.updated_at,
		       c.id, c.name, c.description
		FROM recipe_catalogue.recipes d
		LEFT JOIN recipe_catalogue.categories c ON d.category_id = c.id
		WHERE d.id = $1`)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "name", "description", "category_id", "created_at", "updated_at",
			"c_id", "c_name", "c_description",
		}).
			AddRow(1, req.Name, "", 1, now, now, 1, "Category", "Desc"))

	suite.mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT ri.id, ri.recipe_id, ri.ingredient_id, ri.quantity, ri.unit, ri.notes, ri.created_at,
		       i.id, i.name, i.description, i.category, i.created_at
		FROM recipe_catalogue.recipe_ingredients ri
		JOIN recipe_catalogue.ingredients i ON ri.ingredient_id = i.id
		WHERE ri.recipe_id = $1
		ORDER BY ri.id`)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{
			"ri_id", "recipe_id", "ingredient_id", "quantity", "unit", "notes", "ri_created_at",
			"i_id", "i_name", "i_description", "i_category", "i_created_at",
		})) // Empty rows

	// Act
	result, err := suite.repo.UpdateWithIngredients(1, req)

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), req.Name, result.Recipe.Name)
	assert.Empty(suite.T(), result.Ingredients)
}

func (suite *RecipeRepositoryTestSuite) TestSearchRecipesByIngredientsWithIngredients_ReturnsRecipesWithIngredients() {
	// Arrange
	ingredientIDs := []int{1, 2}
	now := time.Now()

	// Mock SearchRecipesByIngredients call
	suite.mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT DISTINCT r.id, r.name, r.description, r.category_id, r.created_at, r.updated_at,
		                c.id, c.name, c.description
		FROM recipe_catalogue.recipes r
		LEFT JOIN recipe_catalogue.categories c ON r.category_id = c.id
		JOIN recipe_catalogue.recipe_ingredients ri ON r.id = ri.recipe_id
		WHERE ri.ingredient_id = ANY($1)
		ORDER BY r.name`)).
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "name", "description", "category_id", "created_at", "updated_at",
			"c_id", "c_name", "c_description",
		}).
			AddRow(1, "Recipe with Ingredients", "Uses ingredients", 1, now, now, 1, "Category", "Desc"))

	// Mock GetIngredientsForRecipes call
	suite.mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT ri.id, ri.recipe_id, ri.ingredient_id, ri.quantity, ri.unit, ri.notes, ri.created_at,
		       i.id, i.name, i.description, i.category, i.created_at
		FROM recipe_catalogue.recipe_ingredients ri
		JOIN recipe_catalogue.ingredients i ON ri.ingredient_id = i.id
		WHERE ri.recipe_id = ANY($1)
		ORDER BY ri.recipe_id, ri.id`)).
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{
			"ri_id", "recipe_id", "ingredient_id", "quantity", "unit", "notes", "ri_created_at",
			"i_id", "i_name", "i_description", "i_category", "i_created_at",
		}).
			AddRow(1, 1, 1, 100.0, "g", nil, now, 1, "Ingredient 1", "Desc", "Cat", now))

	// Act
	result, err := suite.repo.SearchRecipesByIngredientsWithIngredients(ingredientIDs)

	// Assert
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), result, 1)
	assert.Equal(suite.T(), "Recipe with Ingredients", result[0].Recipe.Name)
	assert.Len(suite.T(), result[0].Ingredients, 1)
	assert.Equal(suite.T(), "Ingredient 1", result[0].Ingredients[0].Ingredient.Name)
}

// =============================================================================
// RUN TEST SUITE
// =============================================================================

func TestRecipeRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(RecipeRepositoryTestSuite))
}

// =============================================================================
// HELPER FUNCTIONS
// =============================================================================

func stringPtr(s string) *string {
	return &s
}
