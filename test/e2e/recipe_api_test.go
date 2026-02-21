package e2e

import (
	"encoding/json"
	"fmt"
	"meal-prep/services/recipe-catalogue/handlers"
	"meal-prep/services/recipe-catalogue/repository"
	"meal-prep/services/recipe-catalogue/service"
	"meal-prep/shared/middleware"
	"meal-prep/test/helpers"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type RecipeE2ETestSuite struct {
	suite.Suite
	testDB         *helpers.TestDatabase
	server         *httptest.Server
	testHttpClient *helpers.TestHttpClient
}

func (suite *RecipeE2ETestSuite) SetupSuite() {
	helpers.SuppressTestLogs()

	// Setup real database
	suite.testDB = helpers.SetupPostgresContainer(suite.T())

	// Create recipe-focused services
	recipeRepo := repository.NewRecipeRepository(suite.testDB.DB)
	categoryRepo := repository.NewCategoryRepository(suite.testDB.DB)
	ingredientRepo := repository.NewIngredientRepository(suite.testDB.DB)

	recipeService := service.NewRecipeService(recipeRepo, categoryRepo, ingredientRepo)
	recipeHandler := handlers.NewRecipeHandler(recipeService)

	// Setup HTTP server with recipe routes only
	router := mux.NewRouter()
	router.Use(middleware.LoggingMiddleware("test-recipe-service"))

	// Public recipe routes
	router.HandleFunc("/recipes", recipeHandler.GetAllRecipes).Methods("GET")
	router.HandleFunc("/recipes/{id:[0-9]+}", recipeHandler.GetRecipeByID).Methods("GET")
	router.HandleFunc("/categories", recipeHandler.GetAllCategories).Methods("GET")
	router.HandleFunc("/categories/{id:[0-9]+}/recipes", recipeHandler.GetRecipesByCategory).Methods("GET")

	// Protected recipe routes
	protected := router.PathPrefix("").Subrouter()
	protected.Use(middleware.ExtractUserFromGatewayHeaders)
	protected.HandleFunc("/recipes", recipeHandler.CreateRecipe).Methods("POST")
	protected.HandleFunc("/recipes/{id:[0-9]+}", recipeHandler.UpdateRecipe).Methods("PUT")
	protected.HandleFunc("/recipes/{id:[0-9]+}", recipeHandler.DeleteRecipe).Methods("DELETE")

	// Health check
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "healthy", "service": "recipe-catalogue"}`))
	}).Methods("GET")

	suite.server = httptest.NewServer(router)

	// Generate auth token for protected tests
	suite.testHttpClient = helpers.NewTestHttpClient(
		suite.server.Client(),
		42,                 // Test user ID
		"test@example.com", // Test email
	)
}

func (suite *RecipeE2ETestSuite) TearDownSuite() {
	if suite.server != nil {
		suite.server.Close()
	}
	suite.testDB.Cleanup(suite.T())
	helpers.RestoreTestLogs()
}

func (suite *RecipeE2ETestSuite) SetupTest() {
	suite.testDB.CleanupTestData(suite.T())
	suite.seedTestData()
}

// seedTestData creates test categories and ingredients for recipe tests
func (suite *RecipeE2ETestSuite) seedTestData() {
	// Insert test categories
	categoriesSQL := `
		INSERT INTO recipe_catalogue.categories (name, description) VALUES
			('Meat', 'Meat-based dishes'),
			('Chicken', 'Chicken dishes'), 
			('Fish', 'Fish and seafood'),
			('Vegetarian', 'Vegetarian dishes'),
			('Desserts', 'Sweet treats')
		ON CONFLICT (name) DO NOTHING;
	`

	if _, err := suite.testDB.DB.Exec(categoriesSQL); err != nil {
		suite.T().Fatalf("Failed to seed test categories: %v", err)
	}

	// Insert test recipes using subqueries to get actual category IDs
	recipesSQL := `
		INSERT INTO recipe_catalogue.recipes (name, description, category_id) VALUES
			('Test Pasta Carbonara', 'Classic Italian pasta with eggs and cheese', 
			 (SELECT id FROM recipe_catalogue.categories WHERE name = 'Meat')),
			('Test Grilled Chicken', 'Simple grilled chicken breast', 
			 (SELECT id FROM recipe_catalogue.categories WHERE name = 'Chicken')),
			('Test Salmon Teriyaki', 'Glazed salmon with teriyaki sauce', 
			 (SELECT id FROM recipe_catalogue.categories WHERE name = 'Fish')),
			('Test Veggie Stir Fry', 'Mixed vegetables stir fried', 
			 (SELECT id FROM recipe_catalogue.categories WHERE name = 'Vegetarian')),
			('Test Chocolate Cake', 'Rich chocolate layer cake', 
			 (SELECT id FROM recipe_catalogue.categories WHERE name = 'Desserts'))
		ON CONFLICT DO NOTHING;
	`

	if _, err := suite.testDB.DB.Exec(recipesSQL); err != nil {
		suite.T().Fatalf("Failed to seed test recipes: %v", err)
	}

	// Insert test ingredients
	ingredientsSQL := `
		INSERT INTO recipe_catalogue.ingredients (name, description, category) VALUES
			('Test Chicken Breast', 'Boneless skinless chicken breast', 'Meat'),
			('Test Ground Beef', 'Lean ground beef', 'Meat'),
			('Test Salmon Fillet', 'Fresh salmon fillet', 'Fish'),
			('Test Pasta', 'Dry pasta noodles', 'Grains'),
			('Test Rice', 'Long grain rice', 'Grains'),
			('Test Onion', 'Yellow onion', 'Vegetables'),
			('Test Garlic', 'Fresh garlic cloves', 'Vegetables'),
			('Test Tomato', 'Fresh tomatoes', 'Vegetables'),
			('Test Cheese', 'Grated parmesan', 'Dairy'),
			('Test Olive Oil', 'Extra virgin olive oil', 'Oils')
		ON CONFLICT (name) DO NOTHING;
	`

	if _, err := suite.testDB.DB.Exec(ingredientsSQL); err != nil {
		suite.T().Fatalf("Failed to seed test ingredients: %v", err)
	}

	// Insert test recipe-ingredient relationships
	recipeIngredientsSQL := `
		INSERT INTO recipe_catalogue.recipe_ingredients (recipe_id, ingredient_id, quantity, unit, notes)
		SELECT r.id, i.id, 100.0, 'grams', 'test ingredient'
		FROM recipe_catalogue.recipes r, recipe_catalogue.ingredients i
		WHERE r.name = 'Test Pasta Carbonara' AND i.name = 'Test Pasta'
		UNION ALL
		SELECT r.id, i.id, 200.0, 'grams', 'test ingredient'  
		FROM recipe_catalogue.recipes r, recipe_catalogue.ingredients i
		WHERE r.name = 'Test Grilled Chicken' AND i.name = 'Test Chicken Breast'
		UNION ALL
		SELECT r.id, i.id, 150.0, 'grams', 'test ingredient'
		FROM recipe_catalogue.recipes r, recipe_catalogue.ingredients i
		WHERE r.name = 'Test Salmon Teriyaki' AND i.name = 'Test Salmon Fillet'
		ON CONFLICT (recipe_id, ingredient_id) DO NOTHING;
	`

	if _, err := suite.testDB.DB.Exec(recipeIngredientsSQL); err != nil {
		suite.T().Fatalf("Failed to seed test recipe ingredients: %v", err)
	}

	// Insert test user for auth tests
	userSQL := `
		INSERT INTO auth.users (email, password_hash) VALUES
			('test@example.com', '$2a$10$test.hash.for.e2e.testing')
		ON CONFLICT (email) DO NOTHING;
	`

	if _, err := suite.testDB.DB.Exec(userSQL); err != nil {
		suite.T().Fatalf("Failed to seed test user: %v", err)
	}
}

// =============================================================================
// PUBLIC RECIPE BROWSING SCENARIOS
// =============================================================================

func (suite *RecipeE2ETestSuite) TestPublicRecipeBrowsing_CompleteUserJourney() {
	baseURL := suite.server.URL

	// 1. User browses all recipes (landing page scenario)
	resp := suite.testHttpClient.MakeRequest(suite.T(), "GET", baseURL+"/recipes", nil)
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	var recipesResp struct {
		Data []map[string]interface{} `json:"data"`
	}
	err := json.NewDecoder(resp.Body).Decode(&recipesResp)
	assert.NoError(suite.T(), err)
	recipes := recipesResp.Data
	assert.GreaterOrEqual(suite.T(), len(recipes), 1, "Should have seeded recipes")

	firstRecipe := recipes[0]
	recipeID := int(firstRecipe["id"].(float64))
	recipeName := firstRecipe["name"].(string)

	// 2. User clicks on specific recipe for details
	resp = suite.testHttpClient.MakeRequest(suite.T(), "GET", fmt.Sprintf("%s/recipes/%d", baseURL, recipeID), nil)
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	var recipeDetails map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&recipeDetails)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), recipeName, recipeDetails["name"])
	assert.Contains(suite.T(), recipeDetails, "description")
	assert.Contains(suite.T(), recipeDetails, "category_id")

	// 3. User explores available categories
	resp = suite.testHttpClient.MakeRequest(suite.T(), "GET", baseURL+"/categories", nil)
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	var categories []map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&categories)
	assert.NoError(suite.T(), err)
	assert.GreaterOrEqual(suite.T(), len(categories), 1, "Should have seeded categories")

	// 4. User filters recipes by category
	categoryID := int(categories[0]["id"].(float64))
	resp = suite.testHttpClient.MakeRequest(suite.T(), "GET", fmt.Sprintf("%s/categories/%d/recipes", baseURL, categoryID), nil)
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	var categoryRecipesResp struct {
		Data []map[string]interface{} `json:"data"`
	}
	err = json.NewDecoder(resp.Body).Decode(&categoryRecipesResp)
	assert.NoError(suite.T(), err)
	categoryRecipes := categoryRecipesResp.Data

	// Verify all recipes belong to the requested category
	for _, recipe := range categoryRecipes {
		assert.Equal(suite.T(), float64(categoryID), recipe["category_id"])
	}
}

func (suite *RecipeE2ETestSuite) TestPublicRecipeAccess_ErrorHandling() {
	baseURL := suite.server.URL

	// Test non-existent recipe
	resp := suite.testHttpClient.MakeRequest(suite.T(), "GET", fmt.Sprintf("%s/recipes/99999", baseURL), nil)
	assert.Equal(suite.T(), http.StatusNotFound, resp.StatusCode)

	var errorResponse map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&errorResponse)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), errorResponse["message"], "not found")

	// Test invalid recipe ID format
	resp = suite.testHttpClient.MakeRequest(suite.T(), "GET", fmt.Sprintf("%s/recipes/invalid", baseURL), nil)
	assert.Equal(suite.T(), http.StatusNotFound, resp.StatusCode) // Gorilla mux returns 404 for pattern mismatch

	// Test non-existent category
	resp = suite.testHttpClient.MakeRequest(suite.T(), "GET", fmt.Sprintf("%s/categories/99999/recipes", baseURL), nil)
	assert.Equal(suite.T(), http.StatusNotFound, resp.StatusCode)

	err = json.NewDecoder(resp.Body).Decode(&errorResponse)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), errorResponse["message"], "not found")
}

// =============================================================================
// AUTHENTICATED RECIPE MANAGEMENT SCENARIOS
// =============================================================================

func (suite *RecipeE2ETestSuite) TestRecipeCRUD_CompleteLifecycle() {
	baseURL := suite.server.URL

	// First, get a real category ID from the seeded data
	resp := suite.testHttpClient.MakeRequest(suite.T(), "GET", baseURL+"/categories", nil)
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	var categories []map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&categories)
	assert.NoError(suite.T(), err)
	assert.GreaterOrEqual(suite.T(), len(categories), 1, "Should have seeded categories")

	// Use the first available category ID
	categoryID := int(categories[0]["id"].(float64))

	// 1. Create new recipe
	newRecipe := map[string]interface{}{
		"name":        "E2E Test Recipe",
		"description": "Created during automated testing",
		"category_id": categoryID,
	}

	resp = suite.testHttpClient.MakeAuthenticatedRequest(suite.T(), "POST", baseURL+"/recipes", newRecipe)
	assert.Equal(suite.T(), http.StatusCreated, resp.StatusCode)

	var createdRecipe map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&createdRecipe)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), newRecipe["name"], createdRecipe["name"])
	assert.Equal(suite.T(), newRecipe["description"], createdRecipe["description"])
	assert.Contains(suite.T(), createdRecipe, "id")
	assert.Contains(suite.T(), createdRecipe, "created_at")

	recipeID := int(createdRecipe["id"].(float64))

	// 2. Read created recipe (verify via public endpoint)
	resp = suite.testHttpClient.MakeRequest(suite.T(), "GET", fmt.Sprintf("%s/recipes/%d", baseURL, recipeID), nil)
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	var retrievedRecipe map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&retrievedRecipe)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), newRecipe["name"], retrievedRecipe["name"])

	// 3. Update recipe
	updateData := map[string]interface{}{
		"name":        "Updated E2E Recipe",
		"description": "Modified during testing",
		"category_id": categoryID,
	}

	resp = suite.testHttpClient.MakeAuthenticatedRequest(suite.T(), "PUT", fmt.Sprintf("%s/recipes/%d", baseURL, recipeID), updateData)
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	var updatedRecipe map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&updatedRecipe)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), updateData["name"], updatedRecipe["name"])
	assert.Equal(suite.T(), updateData["description"], updatedRecipe["description"])

	// 4. Verify update persisted
	resp = suite.testHttpClient.MakeRequest(suite.T(), "GET", fmt.Sprintf("%s/recipes/%d", baseURL, recipeID), nil)
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	var verifyUpdatedRecipe map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&verifyUpdatedRecipe)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), updateData["name"], verifyUpdatedRecipe["name"])
	assert.Equal(suite.T(), updateData["description"], verifyUpdatedRecipe["description"])

	// 5. Delete recipe
	resp = suite.testHttpClient.MakeAuthenticatedRequest(suite.T(), "DELETE", fmt.Sprintf("%s/recipes/%d", baseURL, recipeID), nil)
	assert.Equal(suite.T(), http.StatusNoContent, resp.StatusCode)

	// 6. Verify deletion
	resp = suite.testHttpClient.MakeRequest(suite.T(), "GET", fmt.Sprintf("%s/recipes/%d", baseURL, recipeID), nil)
	assert.Equal(suite.T(), http.StatusNotFound, resp.StatusCode)

	// 7. Verify recipe no longer appears in list
	resp = suite.testHttpClient.MakeRequest(suite.T(), "GET", baseURL+"/recipes", nil)
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	var finalRecipesResp struct {
		Data []map[string]interface{} `json:"data"`
	}
	err = json.NewDecoder(resp.Body).Decode(&finalRecipesResp)
	assert.NoError(suite.T(), err)
	finalRecipes := finalRecipesResp.Data

	for _, recipe := range finalRecipes {
		assert.NotEqual(suite.T(), float64(recipeID), recipe["id"], "Deleted recipe should not appear in list")
	}
}

func (suite *RecipeE2ETestSuite) TestRecipeCreation_ValidationScenarios() {
	baseURL := suite.server.URL

	// Get a real category ID first
	resp := suite.testHttpClient.MakeRequest(suite.T(), "GET", baseURL+"/categories", nil)
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	var categories []map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&categories)
	assert.NoError(suite.T(), err)
	assert.GreaterOrEqual(suite.T(), len(categories), 1, "Should have seeded categories")

	categoryID := int(categories[0]["id"].(float64))

	testCases := []struct {
		name           string
		payload        map[string]interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name: "missing_name",
			payload: map[string]interface{}{
				"description": "Recipe without name",
				"category_id": categoryID,
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "name",
		},
		{
			name: "empty_name",
			payload: map[string]interface{}{
				"name":        "",
				"description": "Recipe with empty name",
				"category_id": categoryID,
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "name",
		},
		{
			name: "invalid_category",
			payload: map[string]interface{}{
				"name":        "Valid Recipe Name",
				"description": "Recipe with invalid category",
				"category_id": 99999,
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "category",
		},
		{
			name: "missing_category",
			payload: map[string]interface{}{
				"name":        "Valid Recipe Name",
				"description": "Recipe without category",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "category",
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			resp := suite.testHttpClient.MakeAuthenticatedRequest(suite.T(), "POST", baseURL+"/recipes", tc.payload)
			assert.Equal(t, tc.expectedStatus, resp.StatusCode)

			var errorResponse map[string]interface{}
			err := json.NewDecoder(resp.Body).Decode(&errorResponse)
			assert.NoError(t, err)
			assert.Contains(t, errorResponse["message"].(string), tc.expectedError)
		})
	}
}

func (suite *RecipeE2ETestSuite) TestRecipeUpdate_ValidationScenarios() {
	baseURL := suite.server.URL

	// First, get a real category ID from the seeded data
	resp := suite.testHttpClient.MakeRequest(suite.T(), "GET", baseURL+"/categories", nil)
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	var categories []map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&categories)
	assert.NoError(suite.T(), err)
	assert.GreaterOrEqual(suite.T(), len(categories), 1, "Should have seeded categories")

	// Use the first available category ID
	categoryID := int(categories[0]["id"].(float64))

	// First create a recipe to update
	newRecipe := map[string]interface{}{
		"name":        "Recipe to Update",
		"description": "Will be updated",
		"category_id": categoryID,
	}

	resp = suite.testHttpClient.MakeAuthenticatedRequest(suite.T(), "POST", baseURL+"/recipes", newRecipe)
	assert.Equal(suite.T(), http.StatusCreated, resp.StatusCode)

	var createdRecipe map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&createdRecipe)
	assert.NoError(suite.T(), err)
	recipeID := int(createdRecipe["id"].(float64))

	testCases := []struct {
		name           string
		payload        map[string]interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name: "empty_name_update",
			payload: map[string]interface{}{
				"name":        "",
				"description": "Some description",
				"category_id": categoryID,
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "name",
		},
		{
			name: "invalid_category_update",
			payload: map[string]interface{}{
				"name":        "Some name",
				"description": "Some description",
				"category_id": 99999,
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "category",
		},
		{
			name: "valid_partial_update",
			payload: map[string]interface{}{
				"description": "Updated description only",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "name",
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			resp := suite.testHttpClient.MakeAuthenticatedRequest(suite.T(), "PUT", fmt.Sprintf("%s/recipes/%d", baseURL, recipeID), tc.payload)
			assert.Equal(t, tc.expectedStatus, resp.StatusCode)

			if tc.expectedError != "" {
				var errorResponse map[string]interface{}
				err := json.NewDecoder(resp.Body).Decode(&errorResponse)
				assert.NoError(t, err)
				assert.Contains(t, errorResponse["message"].(string), tc.expectedError)
			}
		})
	}
}

func (suite *RecipeE2ETestSuite) TestRecipeOperations_AuthenticationRequired() {
	baseURL := suite.server.URL

	// First, get a real category ID from the seeded data
	resp := suite.testHttpClient.MakeRequest(suite.T(), "GET", baseURL+"/categories", nil)
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	var categories []map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&categories)
	assert.NoError(suite.T(), err)
	assert.GreaterOrEqual(suite.T(), len(categories), 1, "Should have seeded categories")

	// Use the first available category ID
	categoryID := int(categories[0]["id"].(float64))

	// Test operations that require authentication
	testCases := []struct {
		name   string
		method string
		url    string
		body   interface{}
	}{
		{
			name:   "create_recipe_no_auth",
			method: "POST",
			url:    "/recipes",
			body:   map[string]interface{}{"name": "Test", "category_id": categoryID},
		},
		{
			name:   "update_recipe_no_auth",
			method: "PUT",
			url:    "/recipes/1",
			body:   map[string]interface{}{"name": "Updated"},
		},
		{
			name:   "delete_recipe_no_auth",
			method: "DELETE",
			url:    "/recipes/1",
			body:   nil,
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			resp := suite.testHttpClient.MakeRequest(suite.T(), tc.method, baseURL+tc.url, tc.body)
			assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

			var errorResponse map[string]interface{}
			err := json.NewDecoder(resp.Body).Decode(&errorResponse)
			assert.NoError(t, err)
			assert.Contains(t, errorResponse["message"].(string), "Missing user context")
		})
	}
}

// Test runner
func TestRecipeE2E(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping e2e tests in short mode")
	}

	suite.Run(t, new(RecipeE2ETestSuite))
}
