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
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// IngredientRepositoryTestSuite tests all ingredient repository operations
type IngredientRepositoryTestSuite struct {
	suite.Suite
	db   *database.DB
	mock sqlmock.Sqlmock
	repo IngredientRepository
}

func (suite *IngredientRepositoryTestSuite) SetupTest() {
	db, mock, err := sqlmock.New()
	require.NoError(suite.T(), err)

	suite.db = &database.DB{DB: db}
	suite.mock = mock
	suite.repo = NewIngredientRepository(suite.db)
}

func (suite *IngredientRepositoryTestSuite) TearDownTest() {
	suite.db.Close()
}

// =============================================================================
// BASIC INGREDIENT CRUD OPERATIONS
// =============================================================================

func (suite *IngredientRepositoryTestSuite) TestGetAllIngredients_ReturnsAllIngredients() {
	// Arrange
	now := time.Now()
	suite.mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, name, description, category, created_at 
		FROM recipe_catalogue.ingredients 
		ORDER BY category, name`)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description", "category", "created_at"}).
			AddRow(1, "Tomato", "Fresh tomato", "Vegetable", now).
			AddRow(2, "Chicken", "Free range chicken", "Protein", now).
			AddRow(3, "Rice", nil, nil, now))

	// Act
	ingredients, err := suite.repo.GetAllIngredients()

	// Assert
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), ingredients, 3)

	assert.Equal(suite.T(), "Tomato", ingredients[0].Name)
	assert.Equal(suite.T(), "Fresh tomato", *ingredients[0].Description)
	assert.Equal(suite.T(), "Vegetable", *ingredients[0].Category)

	assert.Equal(suite.T(), "Rice", ingredients[2].Name)
	assert.Nil(suite.T(), ingredients[2].Description)
	assert.Nil(suite.T(), ingredients[2].Category)
}

func (suite *IngredientRepositoryTestSuite) TestGetAllIngredients_DatabaseError() {
	// Arrange
	suite.mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, name, description, category, created_at 
		FROM recipe_catalogue.ingredients 
		ORDER BY category, name`)).
		WillReturnError(errors.New("database connection lost"))

	// Act
	ingredients, err := suite.repo.GetAllIngredients()

	// Assert
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), ingredients)
	assert.Contains(suite.T(), err.Error(), "database connection lost")
}

func (suite *IngredientRepositoryTestSuite) TestGetIngredientByID_ReturnsIngredient() {
	// Arrange
	now := time.Now()
	suite.mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, name, description, category, created_at 
		FROM recipe_catalogue.ingredients 
		WHERE id = $1`)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description", "category", "created_at"}).
			AddRow(1, "Basil", "Fresh basil leaves", "Herb", now))

	// Act
	ingredient, err := suite.repo.GetIngredientByID(1)

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), ingredient)
	assert.Equal(suite.T(), 1, ingredient.ID)
	assert.Equal(suite.T(), "Basil", ingredient.Name)
	assert.Equal(suite.T(), "Fresh basil leaves", *ingredient.Description)
	assert.Equal(suite.T(), "Herb", *ingredient.Category)
}

func (suite *IngredientRepositoryTestSuite) TestGetIngredientByID_NotFound() {
	// Arrange
	suite.mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, name, description, category, created_at 
		FROM recipe_catalogue.ingredients 
		WHERE id = $1`)).
		WithArgs(999).
		WillReturnError(sql.ErrNoRows)

	// Act
	ingredient, err := suite.repo.GetIngredientByID(999)

	// Assert
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), sql.ErrNoRows, err)
	assert.Nil(suite.T(), ingredient)
}

func (suite *IngredientRepositoryTestSuite) TestGetIngredientsByCategory_ReturnsFilteredIngredients() {
	// Arrange
	now := time.Now()
	suite.mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, name, description, category, created_at 
		FROM recipe_catalogue.ingredients 
		WHERE category = $1 
		ORDER BY name`)).
		WithArgs("Vegetable").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description", "category", "created_at"}).
			AddRow(1, "Carrot", "Orange carrot", "Vegetable", now).
			AddRow(2, "Tomato", "Red tomato", "Vegetable", now))

	// Act
	ingredients, err := suite.repo.GetIngredientsByCategory("Vegetable")

	// Assert
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), ingredients, 2)
	assert.Equal(suite.T(), "Carrot", ingredients[0].Name)
	assert.Equal(suite.T(), "Tomato", ingredients[1].Name)
}

func (suite *IngredientRepositoryTestSuite) TestSearchIngredients_ReturnsMatchingIngredients() {
	// Arrange
	now := time.Now()
	suite.mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, name, description, category, created_at 
		FROM recipe_catalogue.ingredients 
		WHERE name ILIKE $1 OR description ILIKE $1 
		ORDER BY name`)).
		WithArgs("%tomato%").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description", "category", "created_at"}).
			AddRow(1, "Tomato", "Fresh tomato", "Vegetable", now).
			AddRow(2, "Cherry Tomato", "Small tomatoes", "Vegetable", now))

	// Act
	ingredients, err := suite.repo.SearchIngredients("tomato")

	// Assert
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), ingredients, 2)
	assert.Equal(suite.T(), "Tomato", ingredients[0].Name)
	assert.Equal(suite.T(), "Cherry Tomato", ingredients[1].Name)
}

func (suite *IngredientRepositoryTestSuite) TestCreateIngredient_InsertsAndReturnsIngredient() {
	// Arrange
	req := models.CreateIngredientRequest{
		Name:        "New Ingredient",
		Description: stringPtr("Fresh ingredient"),
		Category:    stringPtr("Vegetable"),
	}
	now := time.Now()

	suite.mock.ExpectQuery(regexp.QuoteMeta(`
		INSERT INTO recipe_catalogue.ingredients (name, description, category, created_at)
		VALUES ($1, $2, $3, CURRENT_TIMESTAMP)
		RETURNING id, name, description, category, created_at`)).
		WithArgs(req.Name, req.Description, req.Category).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description", "category", "created_at"}).
			AddRow(1, req.Name, *req.Description, *req.Category, now))

	// Act
	ingredient, err := suite.repo.CreateIngredient(req)

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), ingredient)
	assert.Equal(suite.T(), req.Name, ingredient.Name)
	assert.Equal(suite.T(), *req.Description, *ingredient.Description)
	assert.Equal(suite.T(), *req.Category, *ingredient.Category)
}

func (suite *IngredientRepositoryTestSuite) TestCreateIngredient_WithNullFields() {
	// Arrange
	req := models.CreateIngredientRequest{
		Name:        "Simple Ingredient",
		Description: nil,
		Category:    nil,
	}
	now := time.Now()

	suite.mock.ExpectQuery(regexp.QuoteMeta(`
		INSERT INTO recipe_catalogue.ingredients (name, description, category, created_at)
		VALUES ($1, $2, $3, CURRENT_TIMESTAMP)
		RETURNING id, name, description, category, created_at`)).
		WithArgs(req.Name, req.Description, req.Category).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description", "category", "created_at"}).
			AddRow(1, req.Name, nil, nil, now))

	// Act
	ingredient, err := suite.repo.CreateIngredient(req)

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), ingredient)
	assert.Equal(suite.T(), req.Name, ingredient.Name)
	assert.Nil(suite.T(), ingredient.Description)
	assert.Nil(suite.T(), ingredient.Category)
}

func (suite *IngredientRepositoryTestSuite) TestUpdateIngredient_UpdatesAndReturnsIngredient() {
	// Arrange
	req := models.UpdateIngredientRequest{
		Name:        "Updated Name",
		Description: "Updated description",
		Category:    "Updated Category",
	}
	now := time.Now()

	suite.mock.ExpectQuery(regexp.QuoteMeta(`
		UPDATE recipe_catalogue.ingredients 
        SET name = $2,
            description = $3,
            category = $4,
            updated_at = CURRENT_TIMESTAMP
        WHERE id = $1
		RETURNING id, name, description, category, created_at`)).
		WithArgs(1, req.Name, req.Description, req.Category).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description", "category", "created_at"}).
			AddRow(1, req.Name, req.Description, req.Category, now))

	// Act
	ingredient, err := suite.repo.UpdateIngredient(1, req)

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), ingredient)
	assert.Equal(suite.T(), req.Name, ingredient.Name)
	assert.Equal(suite.T(), req.Description, *ingredient.Description)
	assert.Equal(suite.T(), req.Category, *ingredient.Category)
}

func (suite *IngredientRepositoryTestSuite) TestUpdateIngredient_NotFound() {
	// Arrange
	req := models.UpdateIngredientRequest{Name: "Updated"}

	suite.mock.ExpectQuery(regexp.QuoteMeta(`
		UPDATE recipe_catalogue.ingredients 
        SET name = $2,
            description = $3,
            category = $4,
            updated_at = CURRENT_TIMESTAMP
        WHERE id = $1
		RETURNING id, name, description, category, created_at`)).
		WithArgs(999, req.Name, req.Description, req.Category).
		WillReturnError(sql.ErrNoRows)

	// Act
	ingredient, err := suite.repo.UpdateIngredient(999, req)

	// Assert
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), sql.ErrNoRows, err)
	assert.Nil(suite.T(), ingredient)
}

func (suite *IngredientRepositoryTestSuite) TestDeleteIngredient_RemovesIngredient() {
	// Arrange
	suite.mock.ExpectExec(regexp.QuoteMeta("DELETE FROM recipe_catalogue.ingredients WHERE id = $1")).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Act
	err := suite.repo.DeleteIngredient(1)

	// Assert
	assert.NoError(suite.T(), err)
}

func (suite *IngredientRepositoryTestSuite) TestDeleteIngredient_NotFound() {
	// Arrange
	suite.mock.ExpectExec(regexp.QuoteMeta("DELETE FROM recipe_catalogue.ingredients WHERE id = $1")).
		WithArgs(999).
		WillReturnResult(sqlmock.NewResult(0, 0)) // No rows affected

	// Act
	err := suite.repo.DeleteIngredient(999)

	// Assert
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), sql.ErrNoRows, err)
}

func (suite *IngredientRepositoryTestSuite) TestIngredientExists_ReturnsTrue() {
	// Arrange
	suite.mock.ExpectQuery(regexp.QuoteMeta("SELECT EXISTS(SELECT 1 FROM recipe_catalogue.ingredients WHERE id = $1)")).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	// Act
	exists, err := suite.repo.IngredientExists(1)

	// Assert
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), exists)
}

func (suite *IngredientRepositoryTestSuite) TestIngredientExists_ReturnsFalse() {
	// Arrange
	suite.mock.ExpectQuery(regexp.QuoteMeta("SELECT EXISTS(SELECT 1 FROM recipe_catalogue.ingredients WHERE id = $1)")).
		WithArgs(999).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	// Act
	exists, err := suite.repo.IngredientExists(999)

	// Assert
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), exists)
}

// =============================================================================
// RECIPE-INGREDIENT RELATIONSHIP OPERATIONS
// =============================================================================

func (suite *IngredientRepositoryTestSuite) TestGetRecipeIngredients_ReturnsRecipeIngredients() {
	// Arrange
	now := time.Now()
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
			AddRow(1, 1, 1, 200.0, "grams", "Fresh", now, 1, "Tomato", "Red tomato", "Vegetable", now).
			AddRow(2, 1, 2, 1.0, "piece", nil, now, 2, "Onion", "Yellow onion", "Vegetable", now))

	// Act
	recipeIngredients, err := suite.repo.GetRecipeIngredients(1)

	// Assert
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), recipeIngredients, 2)

	assert.Equal(suite.T(), 1, recipeIngredients[0].RecipeID)
	assert.Equal(suite.T(), 1, recipeIngredients[0].IngredientID)
	assert.Equal(suite.T(), 200.0, recipeIngredients[0].Quantity)
	assert.Equal(suite.T(), "grams", recipeIngredients[0].Unit)
	assert.Equal(suite.T(), "Fresh", *recipeIngredients[0].Notes)
	assert.Equal(suite.T(), "Tomato", recipeIngredients[0].Ingredient.Name)

	assert.Nil(suite.T(), recipeIngredients[1].Notes)
}

func (suite *IngredientRepositoryTestSuite) TestAddRecipeIngredient_AddsAndReturnsRecipeIngredient() {
	// Arrange
	req := models.AddRecipeIngredientRequest{
		IngredientID: 1,
		Quantity:     250.0,
		Unit:         "grams",
		Notes:        stringPtr("Fresh ingredient"),
	}
	now := time.Now()

	// Expect ingredient addition
	suite.mock.ExpectQuery(regexp.QuoteMeta(`
		INSERT INTO recipe_catalogue.recipe_ingredients (recipe_id, ingredient_id, quantity, unit, notes, created_at)
		VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP)
		RETURNING id, recipe_id, ingredient_id, quantity, unit, notes, created_at`)).
		WithArgs(1, req.IngredientID, req.Quantity, req.Unit, req.Notes).
		WillReturnRows(sqlmock.NewRows([]string{"id", "recipe_id", "ingredient_id", "quantity", "unit", "notes", "created_at"}).
			AddRow(1, 1, req.IngredientID, req.Quantity, req.Unit, *req.Notes, now))

	// Expect ingredient details retrieval
	suite.mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, name, description, category, created_at 
		FROM recipe_catalogue.ingredients 
		WHERE id = $1`)).
		WithArgs(req.IngredientID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description", "category", "created_at"}).
			AddRow(1, "Tomato", "Red tomato", "Vegetable", now))

	// Act
	recipeIngredient, err := suite.repo.AddRecipeIngredient(1, req)

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), recipeIngredient)
	assert.Equal(suite.T(), 1, recipeIngredient.RecipeID)
	assert.Equal(suite.T(), req.IngredientID, recipeIngredient.IngredientID)
	assert.Equal(suite.T(), req.Quantity, recipeIngredient.Quantity)
	assert.Equal(suite.T(), req.Unit, recipeIngredient.Unit)
	assert.Equal(suite.T(), "Tomato", recipeIngredient.Ingredient.Name)
}

func (suite *IngredientRepositoryTestSuite) TestUpdateRecipeIngredient_UpdatesAndReturnsRecipeIngredient() {
	// Arrange
	req := models.AddRecipeIngredientRequest{
		IngredientID: 1,
		Quantity:     300.0,
		Unit:         "grams",
		Notes:        stringPtr("Updated notes"),
	}
	now := time.Now()

	// Expect recipe ingredient update
	suite.mock.ExpectQuery(regexp.QuoteMeta(`
		UPDATE recipe_catalogue.recipe_ingredients 
		SET quantity = $3, unit = $4, notes = $5
		WHERE recipe_id = $1 AND ingredient_id = $2
		RETURNING id, recipe_id, ingredient_id, quantity, unit, notes, created_at`)).
		WithArgs(1, 1, req.Quantity, req.Unit, req.Notes).
		WillReturnRows(sqlmock.NewRows([]string{"id", "recipe_id", "ingredient_id", "quantity", "unit", "notes", "created_at"}).
			AddRow(1, 1, 1, req.Quantity, req.Unit, *req.Notes, now))

	// Expect ingredient details retrieval
	suite.mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, name, description, category, created_at 
		FROM recipe_catalogue.ingredients 
		WHERE id = $1`)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description", "category", "created_at"}).
			AddRow(1, "Tomato", "Red tomato", "Vegetable", now))

	// Act
	recipeIngredient, err := suite.repo.UpdateRecipeIngredient(1, 1, req)

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), recipeIngredient)
	assert.Equal(suite.T(), req.Quantity, recipeIngredient.Quantity)
	assert.Equal(suite.T(), req.Unit, recipeIngredient.Unit)
	assert.Equal(suite.T(), *req.Notes, *recipeIngredient.Notes)
}

func (suite *IngredientRepositoryTestSuite) TestRemoveRecipeIngredient_RemovesIngredient() {
	// Arrange
	suite.mock.ExpectExec(regexp.QuoteMeta("DELETE FROM recipe_catalogue.recipe_ingredients WHERE recipe_id = $1 AND ingredient_id = $2")).
		WithArgs(1, 1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Act
	err := suite.repo.RemoveRecipeIngredient(1, 1)

	// Assert
	assert.NoError(suite.T(), err)
}

func (suite *IngredientRepositoryTestSuite) TestRemoveRecipeIngredient_NotFound() {
	// Arrange
	suite.mock.ExpectExec(regexp.QuoteMeta("DELETE FROM recipe_catalogue.recipe_ingredients WHERE recipe_id = $1 AND ingredient_id = $2")).
		WithArgs(1, 999).
		WillReturnResult(sqlmock.NewResult(0, 0)) // No rows affected

	// Act
	err := suite.repo.RemoveRecipeIngredient(1, 999)

	// Assert
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), sql.ErrNoRows, err)
}

// =============================================================================
// TRANSACTION OPERATIONS
// =============================================================================

func (suite *IngredientRepositoryTestSuite) TestSetRecipeIngredients_CommitsTransaction() {
	// Arrange
	ingredients := []models.AddRecipeIngredientRequest{
		{IngredientID: 1, Quantity: 100.0, Unit: "grams", Notes: stringPtr("Fresh")},
		{IngredientID: 2, Quantity: 2.0, Unit: "pieces", Notes: nil},
	}

	// Expect transaction
	suite.mock.ExpectBegin()

	// Expect deletion of existing ingredients
	suite.mock.ExpectExec(regexp.QuoteMeta("DELETE FROM recipe_catalogue.recipe_ingredients WHERE recipe_id = $1")).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 2))

	// Expect ingredient insertions
	for _, ingredient := range ingredients {
		suite.mock.ExpectExec(regexp.QuoteMeta(`
			INSERT INTO recipe_catalogue.recipe_ingredients (recipe_id, ingredient_id, quantity, unit, notes, created_at)
			VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP)`)).
			WithArgs(1, ingredient.IngredientID, ingredient.Quantity, ingredient.Unit, ingredient.Notes).
			WillReturnResult(sqlmock.NewResult(1, 1))
	}

	// Expect commit
	suite.mock.ExpectCommit()

	// Act
	err := suite.repo.SetRecipeIngredients(1, ingredients)

	// Assert
	assert.NoError(suite.T(), err)
}

func (suite *IngredientRepositoryTestSuite) TestSetRecipeIngredients_RollsBackOnError() {
	// Arrange
	ingredients := []models.AddRecipeIngredientRequest{
		{IngredientID: 1, Quantity: 100.0, Unit: "grams"},
		{IngredientID: 2, Quantity: 2.0, Unit: "pieces"},
	}

	// Expect transaction
	suite.mock.ExpectBegin()

	// Expect deletion to succeed
	suite.mock.ExpectExec(regexp.QuoteMeta("DELETE FROM recipe_catalogue.recipe_ingredients WHERE recipe_id = $1")).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 2))

	// Expect first insertion to succeed
	suite.mock.ExpectExec(regexp.QuoteMeta(`
		INSERT INTO recipe_catalogue.recipe_ingredients (recipe_id, ingredient_id, quantity, unit, notes, created_at)
		VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP)`)).
		WithArgs(1, ingredients[0].IngredientID, ingredients[0].Quantity, ingredients[0].Unit, ingredients[0].Notes).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Expect second insertion to fail
	suite.mock.ExpectExec(regexp.QuoteMeta(`
		INSERT INTO recipe_catalogue.recipe_ingredients (recipe_id, ingredient_id, quantity, unit, notes, created_at)
		VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP)`)).
		WithArgs(1, ingredients[1].IngredientID, ingredients[1].Quantity, ingredients[1].Unit, ingredients[1].Notes).
		WillReturnError(errors.New("constraint violation"))

	// Expect rollback
	suite.mock.ExpectRollback()

	// Act
	err := suite.repo.SetRecipeIngredients(1, ingredients)

	// Assert
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "constraint violation")
}

func (suite *IngredientRepositoryTestSuite) TestSetRecipeIngredients_EmptyIngredients() {
	// Arrange - Empty ingredients list should clear all ingredients

	// Expect transaction
	suite.mock.ExpectBegin()

	// Expect deletion of existing ingredients
	suite.mock.ExpectExec(regexp.QuoteMeta("DELETE FROM recipe_catalogue.recipe_ingredients WHERE recipe_id = $1")).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 3))

	// Expect commit (no insertions)
	suite.mock.ExpectCommit()

	// Act
	err := suite.repo.SetRecipeIngredients(1, []models.AddRecipeIngredientRequest{})

	// Assert
	assert.NoError(suite.T(), err)
}

// =============================================================================
// COMPLEX QUERIES
// =============================================================================

func (suite *IngredientRepositoryTestSuite) TestGetIngredientsForRecipes_ReturnsIngredientsMap() {
	// Arrange
	recipeIDs := []int{1, 2}
	now := time.Now()

	suite.mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT ri.id, ri.recipe_id, ri.ingredient_id, ri.quantity, ri.unit, ri.notes, ri.created_at,
		       i.id, i.name, i.description, i.category, i.created_at
		FROM recipe_catalogue.recipe_ingredients ri
		JOIN recipe_catalogue.ingredients i ON ri.ingredient_id = i.id
		WHERE ri.recipe_id = ANY($1)
		ORDER BY ri.recipe_id, ri.id`)).
		WithArgs(pq.Array(recipeIDs)).
		WillReturnRows(sqlmock.NewRows([]string{
			"ri_id", "recipe_id", "ingredient_id", "quantity", "unit", "notes", "ri_created_at",
			"i_id", "i_name", "i_description", "i_category", "i_created_at",
		}).
			AddRow(1, 1, 1, 100.0, "g", nil, now, 1, "Ingredient 1", "Desc 1", "Cat 1", now).
			AddRow(2, 1, 2, 200.0, "ml", "Fresh", now, 2, "Ingredient 2", "Desc 2", "Cat 2", now).
			AddRow(3, 2, 1, 150.0, "g", nil, now, 1, "Ingredient 1", "Desc 1", "Cat 1", now))

	// Act
	result, err := suite.repo.GetIngredientsForRecipes(recipeIDs)

	// Assert
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), result, 2)

	// Recipe 1 should have 2 ingredients
	assert.Len(suite.T(), result[1], 2)
	assert.Equal(suite.T(), "Ingredient 1", result[1][0].Ingredient.Name)
	assert.Equal(suite.T(), "Ingredient 2", result[1][1].Ingredient.Name)

	// Recipe 2 should have 1 ingredient
	assert.Len(suite.T(), result[2], 1)
	assert.Equal(suite.T(), "Ingredient 1", result[2][0].Ingredient.Name)
}

func (suite *IngredientRepositoryTestSuite) TestGetIngredientsForRecipes_EmptySlice() {
	// Act
	result, err := suite.repo.GetIngredientsForRecipes([]int{})

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Empty(suite.T(), result)
}

func (suite *IngredientRepositoryTestSuite) TestGetRecipesUsingIngredient_ReturnsRecipes() {
	// Arrange
	now := time.Now()
	suite.mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT DISTINCT r.id, r.name, r.description, r.category_id, r.created_at, r.updated_at,
		                c.id, c.name, c.description
		FROM recipe_catalogue.recipes r
		LEFT JOIN recipe_catalogue.categories c ON r.category_id = c.id
		JOIN recipe_catalogue.recipe_ingredients ri ON r.id = ri.recipe_id
		WHERE ri.ingredient_id = $1
		ORDER BY r.name`)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "name", "description", "category_id", "created_at", "updated_at",
			"c_id", "c_name", "c_description",
		}).
			AddRow(1, "Pasta Dish", "Italian pasta", 1, now, now, 1, "Italian", "Italian cuisine").
			AddRow(2, "Tomato Soup", "Fresh soup", 2, now, now, 2, "Soup", "Soup category"))

	// Act
	recipes, err := suite.repo.GetRecipesUsingIngredient(1)

	// Assert
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), recipes, 2)
	assert.Equal(suite.T(), "Pasta Dish", recipes[0].Name)
	assert.Equal(suite.T(), "Italian", recipes[0].Category.Name)
	assert.Equal(suite.T(), "Tomato Soup", recipes[1].Name)
}

// =============================================================================
// ERROR HANDLING AND EDGE CASES
// =============================================================================

func (suite *IngredientRepositoryTestSuite) TestScanIngredient_HandlesNullFields() {
	// This is implicitly tested in other methods, but we can test the scanning directly
	// by observing the behavior in GetIngredientByID with null fields
	now := time.Now()
	suite.mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, name, description, category, created_at 
		FROM recipe_catalogue.ingredients 
		WHERE id = $1`)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description", "category", "created_at"}).
			AddRow(1, "Plain Ingredient", nil, nil, now))

	// Act
	ingredient, err := suite.repo.GetIngredientByID(1)

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), ingredient)
	assert.Equal(suite.T(), "Plain Ingredient", ingredient.Name)
	assert.Nil(suite.T(), ingredient.Description)
	assert.Nil(suite.T(), ingredient.Category)
}

func (suite *IngredientRepositoryTestSuite) TestScanRecipeIngredient_HandlesNullNotes() {
	// Test through GetRecipeIngredients with null notes
	now := time.Now()
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
			AddRow(1, 1, 1, 200.0, "grams", nil, now, 1, "Tomato", nil, nil, now))

	// Act
	recipeIngredients, err := suite.repo.GetRecipeIngredients(1)

	// Assert
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), recipeIngredients, 1)
	assert.Nil(suite.T(), recipeIngredients[0].Notes)
	assert.Nil(suite.T(), recipeIngredients[0].Ingredient.Description)
	assert.Nil(suite.T(), recipeIngredients[0].Ingredient.Category)
}

// =============================================================================
// RUN TEST SUITE
// =============================================================================

func TestIngredientRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(IngredientRepositoryTestSuite))
}

// =============================================================================
// TABLE-DRIVEN TESTS FOR EDGE CASES
// =============================================================================

func TestIngredientRepository_EdgeCases(t *testing.T) {
	tests := []struct {
		name            string
		searchQuery     string
		expectedQueries int
		expectError     bool
		description     string
	}{
		{
			name:            "normal_search",
			searchQuery:     "tomato",
			expectedQueries: 1,
			expectError:     false,
			description:     "Normal search should execute one query",
		},
		{
			name:            "empty_search",
			searchQuery:     "",
			expectedQueries: 1,
			expectError:     false,
			description:     "Empty search should still execute query",
		},
		{
			name:            "sql_injection_attempt",
			searchQuery:     "'; DROP TABLE ingredients; --",
			expectedQueries: 1,
			expectError:     false,
			description:     "SQL injection attempts should be safely parameterized",
		},
		{
			name:            "unicode_search",
			searchQuery:     "томат",
			expectedQueries: 1,
			expectError:     false,
			description:     "Unicode characters should be handled safely",
		},
		{
			name:            "special_characters",
			searchQuery:     "%_\\",
			expectedQueries: 1,
			expectError:     false,
			description:     "Special SQL characters should be escaped",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fresh mock for each test
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			repo := NewIngredientRepository(&database.DB{DB: db})

			// Setup expectation
			if tt.expectedQueries > 0 {
				mock.ExpectQuery(regexp.QuoteMeta(`
					SELECT id, name, description, category, created_at 
					FROM recipe_catalogue.ingredients 
					WHERE name ILIKE $1 OR description ILIKE $1 
					ORDER BY name`)).
					WithArgs("%" + tt.searchQuery + "%").
					WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description", "category", "created_at"}))
			}

			// Act
			ingredients, err := repo.SearchIngredients(tt.searchQuery)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// For empty results, expect empty slice (nil slice with len 0 is fine in Go)
				assert.Equal(t, 0, len(ingredients), "Expected no ingredients for empty search results")
			}

			// Verify expectations
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// =============================================================================
// BENCHMARK TESTS
// =============================================================================

func BenchmarkIngredientRepository_GetAllIngredients(b *testing.B) {
	db, mock, err := sqlmock.New()
	require.NoError(b, err)
	defer db.Close()

	repo := NewIngredientRepository(&database.DB{DB: db})
	now := time.Now()

	// Setup mock expectation for benchmarking
	for i := 0; i < b.N; i++ {
		mock.ExpectQuery(regexp.QuoteMeta(`
			SELECT id, name, description, category, created_at 
			FROM recipe_catalogue.ingredients 
			ORDER BY category, name`)).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description", "category", "created_at"}).
				AddRow(1, "Benchmark Ingredient", "Desc", "Cat", now))
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := repo.GetAllIngredients()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkIngredientRepository_IngredientExists(b *testing.B) {
	db, mock, err := sqlmock.New()
	require.NoError(b, err)
	defer db.Close()

	repo := NewIngredientRepository(&database.DB{DB: db})

	// Setup mock expectation for benchmarking
	for i := 0; i < b.N; i++ {
		mock.ExpectQuery(regexp.QuoteMeta("SELECT EXISTS(SELECT 1 FROM recipe_catalogue.ingredients WHERE id = $1)")).
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := repo.IngredientExists(1)
		if err != nil {
			b.Fatal(err)
		}
	}
}
