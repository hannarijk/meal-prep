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

// IngredientE2ETestSuite tests complete ingredient workflows
type IngredientE2ETestSuite struct {
	suite.Suite
	testDB         *helpers.TestDatabase
	server         *httptest.Server
	testHttpClient *helpers.TestHttpClient
}

func (suite *IngredientE2ETestSuite) SetupSuite() {
	helpers.SuppressTestLogs()

	// Setup real database with testcontainers
	suite.testDB = helpers.SetupPostgresContainer(suite.T())

	// Initialize repositories and services
	recipeRepo := repository.NewRecipeRepository(suite.testDB.DB)
	ingredientRepo := repository.NewIngredientRepository(suite.testDB.DB)

	ingredientService := service.NewIngredientService(ingredientRepo, recipeRepo)
	groceryService := service.NewGroceryService(ingredientRepo, recipeRepo)

	ingredientHandler := handlers.NewIngredientHandler(ingredientService)
	groceryHandler := handlers.NewGroceryHandler(groceryService)

	// Setup HTTP server with ingredient routes
	router := mux.NewRouter()
	router.Use(middleware.LoggingMiddleware("test-ingredient-service"))

	// Public ingredient routes
	router.HandleFunc("/ingredients", ingredientHandler.GetAllIngredients).Methods("GET")
	router.HandleFunc("/ingredients/{id:[0-9]+}", ingredientHandler.GetIngredientByID).Methods("GET")
	router.HandleFunc("/ingredients/{id:[0-9]+}/recipes", ingredientHandler.GetRecipesUsingIngredient).Methods("GET")
	router.HandleFunc("/recipes/{id:[0-9]+}/ingredients", ingredientHandler.GetRecipeIngredients).Methods("GET")

	// Protected ingredient routes
	protected := router.PathPrefix("").Subrouter()
	protected.Use(middleware.ExtractUserFromGatewayHeaders)
	protected.HandleFunc("/ingredients", ingredientHandler.CreateIngredient).Methods("POST")
	protected.HandleFunc("/ingredients/{id:[0-9]+}", ingredientHandler.UpdateIngredient).Methods("PUT")
	protected.HandleFunc("/ingredients/{id:[0-9]+}", ingredientHandler.DeleteIngredient).Methods("DELETE")
	protected.HandleFunc("/recipes/{id:[0-9]+}/ingredients", ingredientHandler.AddRecipeIngredient).Methods("POST")
	protected.HandleFunc("/recipes/{id:[0-9]+}/ingredients", ingredientHandler.SetRecipeIngredients).Methods("PUT")
	protected.HandleFunc("/recipes/{recipeId:[0-9]+}/ingredients/{ingredientId:[0-9]+}", ingredientHandler.UpdateRecipeIngredient).Methods("PUT")
	protected.HandleFunc("/recipes/{recipeId:[0-9]+}/ingredients/{ingredientId:[0-9]+}", ingredientHandler.RemoveRecipeIngredient).Methods("DELETE")
	protected.HandleFunc("/grocery-list", groceryHandler.GenerateGroceryList).Methods("POST")

	// Health check
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "healthy", "service": "ingredient-catalogue"}`))
	}).Methods("GET")

	suite.server = httptest.NewServer(router)

	// Setup authenticated HTTP client
	suite.testHttpClient = helpers.NewTestHttpClient(
		suite.server.Client(),
		42,                 // Test user ID
		"test@example.com", // Test email
	)
}

func (suite *IngredientE2ETestSuite) TearDownSuite() {
	if suite.server != nil {
		suite.server.Close()
	}
	suite.testDB.Cleanup(suite.T())
	helpers.RestoreTestLogs()
}

func (suite *IngredientE2ETestSuite) SetupTest() {
	suite.testDB.CleanupTestData(suite.T())
	suite.seedTestData()
}

func (suite *IngredientE2ETestSuite) seedTestData() {
	// Insert test categories
	categoriesSQL := `
		INSERT INTO recipe_catalogue.categories (name, description) VALUES
			('Vegetables', 'Fresh vegetables'),
			('Meat', 'Various meats'),
			('Fish', 'Fish and seafood'),
			('Dairy', 'Dairy products')
		ON CONFLICT (name) DO NOTHING;
	`
	if _, err := suite.testDB.DB.Exec(categoriesSQL); err != nil {
		suite.T().Fatalf("Failed to seed categories: %v", err)
	}

	// Insert test ingredients
	ingredientsSQL := `
		INSERT INTO recipe_catalogue.ingredients (name, category) VALUES
			('Tomato', 'Vegetables'),
			('Chicken Breast', 'Meat'),
			('Mozzarella', 'Dairy'),
			('Basil', 'Vegetables'),
			('Onion', 'Vegetables'),
			('Lettuce', 'Vegetables')
		ON CONFLICT (name) DO NOTHING;
	`
	if _, err := suite.testDB.DB.Exec(ingredientsSQL); err != nil {
		suite.T().Fatalf("Failed to seed ingredients: %v", err)
	}

	// Insert test recipes
	recipesSQL := `
		INSERT INTO recipe_catalogue.recipes (name, description, category_id)
		SELECT 'Caprese Salad', 'Fresh tomato and mozzarella salad', c.id
		FROM recipe_catalogue.categories c WHERE c.name = 'Vegetables'
		UNION ALL
		SELECT 'Grilled Chicken', 'Simple grilled chicken breast', c.id
		FROM recipe_catalogue.categories c WHERE c.name = 'Meat'
		UNION ALL
		SELECT 'Grilled Salmon', 'Simple grilled salmon fillet', c.id
		FROM recipe_catalogue.categories c WHERE c.name = 'Fish';
	`
	if _, err := suite.testDB.DB.Exec(recipesSQL); err != nil {
		suite.T().Fatalf("Failed to seed recipes: %v", err)
	}

	// Insert recipe-ingredient relationships
	recipeIngredientsSQL := `
		INSERT INTO recipe_catalogue.recipe_ingredients (recipe_id, ingredient_id, quantity, unit, notes)
		SELECT r.id, i.id, 2.0, 'piece', 'large tomatoes'
		FROM recipe_catalogue.recipes r, recipe_catalogue.ingredients i
		WHERE r.name = 'Caprese Salad' AND i.name = 'Tomato'
		UNION ALL
		SELECT r.id, i.id, 150.0, 'grams', 'fresh mozzarella'
		FROM recipe_catalogue.recipes r, recipe_catalogue.ingredients i
		WHERE r.name = 'Caprese Salad' AND i.name = 'Mozzarella'
		UNION ALL
		SELECT r.id, i.id, 200.0, 'grams', 'boneless skinless'
		FROM recipe_catalogue.recipes r, recipe_catalogue.ingredients i
		WHERE r.name = 'Grilled Chicken' AND i.name = 'Chicken Breast'
		ON CONFLICT (recipe_id, ingredient_id) DO NOTHING;
	`
	if _, err := suite.testDB.DB.Exec(recipeIngredientsSQL); err != nil {
		suite.T().Fatalf("Failed to seed recipe ingredients: %v", err)
	}
}

// =============================================================================
// PUBLIC INGREDIENT BROWSING SCENARIOS
// =============================================================================

func (suite *IngredientE2ETestSuite) TestPublicIngredientBrowsing_CompleteUserJourney() {
	baseURL := suite.server.URL

	// 1. User browses all available ingredients
	resp := suite.testHttpClient.MakeRequest(suite.T(), "GET", baseURL+"/ingredients", nil)
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	var ingredients []map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&ingredients)
	assert.NoError(suite.T(), err)
	assert.GreaterOrEqual(suite.T(), len(ingredients), 4, "Should have seeded ingredients")

	// Find tomato ingredient for detailed testing
	var tomatoID int
	var tomatoFound bool
	for _, ingredient := range ingredients {
		if ingredient["name"].(string) == "Tomato" {
			tomatoID = int(ingredient["id"].(float64))
			tomatoFound = true
			assert.Equal(suite.T(), "Vegetables", ingredient["category"])
			break
		}
	}
	assert.True(suite.T(), tomatoFound, "Tomato ingredient should be found")

	// 2. User clicks on specific ingredient for details
	resp = suite.testHttpClient.MakeRequest(suite.T(), "GET", fmt.Sprintf("%s/ingredients/%d", baseURL, tomatoID), nil)
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	var ingredientDetails map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&ingredientDetails)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Tomato", ingredientDetails["name"])
	assert.Equal(suite.T(), "Vegetables", ingredientDetails["category"])

	// 3. User wants to see what recipes use this ingredient
	resp = suite.testHttpClient.MakeRequest(suite.T(), "GET", fmt.Sprintf("%s/ingredients/%d/recipes", baseURL, tomatoID), nil)
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	var recipesUsingTomato []map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&recipesUsingTomato)
	assert.NoError(suite.T(), err)
	assert.GreaterOrEqual(suite.T(), len(recipesUsingTomato), 1, "Should find recipes using tomato")

	// Verify the recipe contains expected information
	capreseFound := false
	for _, recipe := range recipesUsingTomato {
		if recipe["name"].(string) == "Caprese Salad" {
			capreseFound = true
			assert.Contains(suite.T(), recipe, "description")
			assert.Contains(suite.T(), recipe, "category_id")
			break
		}
	}
	assert.True(suite.T(), capreseFound, "Should find Caprese Salad using tomato")
}

func (suite *IngredientE2ETestSuite) TestRecipeIngredientsBrowsing_UserJourney() {
	baseURL := suite.server.URL

	// Get a recipe ID to test with
	var recipeID int
	recipesSQL := "SELECT id FROM recipe_catalogue.recipes WHERE name = 'Caprese Salad' LIMIT 1"
	err := suite.testDB.DB.QueryRow(recipesSQL).Scan(&recipeID)
	assert.NoError(suite.T(), err)

	// User views recipe ingredients
	resp := suite.testHttpClient.MakeRequest(suite.T(), "GET", fmt.Sprintf("%s/recipes/%d/ingredients", baseURL, recipeID), nil)
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	var recipeIngredients []map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&recipeIngredients)
	assert.NoError(suite.T(), err)
	assert.GreaterOrEqual(suite.T(), len(recipeIngredients), 2, "Caprese should have multiple ingredients")

	// Verify ingredient details in recipe context
	tomatoFound := false
	mozzarellaFound := false
	for _, ingredient := range recipeIngredients {
		switch ingredient["ingredient"].(map[string]interface{})["name"].(string) {
		case "Tomato":
			tomatoFound = true
			assert.Equal(suite.T(), 2.0, ingredient["quantity"].(float64))
			assert.Equal(suite.T(), "piece", ingredient["unit"])
			assert.Equal(suite.T(), "large tomatoes", ingredient["notes"])
		case "Mozzarella":
			mozzarellaFound = true
			assert.Equal(suite.T(), 150.0, ingredient["quantity"].(float64))
			assert.Equal(suite.T(), "grams", ingredient["unit"])
		}
	}
	assert.True(suite.T(), tomatoFound && mozzarellaFound, "Should find both tomato and mozzarella")
}

// =============================================================================
// INGREDIENT MANAGEMENT SCENARIOS (AUTHENTICATED)
// =============================================================================

func (suite *IngredientE2ETestSuite) TestIngredientManagement_CompleteWorkflow() {
	baseURL := suite.server.URL

	// 1. Create new ingredient
	newIngredient := map[string]interface{}{
		"name":     "Red Onion",
		"category": "Vegetables",
	}

	resp := suite.testHttpClient.MakeAuthenticatedRequest(suite.T(), "POST", baseURL+"/ingredients", newIngredient)
	assert.Equal(suite.T(), http.StatusCreated, resp.StatusCode)

	var createdIngredient map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&createdIngredient)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Red Onion", createdIngredient["name"])
	newIngredientID := int(createdIngredient["id"].(float64))

	// 2. Update the ingredient
	update := map[string]interface{}{
		"name":     "Sweet Red Onion",
		"category": "Vegetables",
	}

	resp = suite.testHttpClient.MakeAuthenticatedRequest(suite.T(), "PUT", fmt.Sprintf("%s/ingredients/%d", baseURL, newIngredientID), update)
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	var updatedIngredient map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&updatedIngredient)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Sweet Red Onion", updatedIngredient["name"])

	// 3. Verify update by fetching ingredient
	resp = suite.testHttpClient.MakeRequest(suite.T(), "GET", fmt.Sprintf("%s/ingredients/%d", baseURL, newIngredientID), nil)
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	var fetchedIngredient map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&fetchedIngredient)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Sweet Red Onion", fetchedIngredient["name"])

	// 4. Delete the ingredient
	resp = suite.testHttpClient.MakeAuthenticatedRequest(suite.T(), "DELETE", fmt.Sprintf("%s/ingredients/%d", baseURL, newIngredientID), nil)
	assert.Equal(suite.T(), http.StatusNoContent, resp.StatusCode)

	// 5. Verify deletion
	resp = suite.testHttpClient.MakeRequest(suite.T(), "GET", fmt.Sprintf("%s/ingredients/%d", baseURL, newIngredientID), nil)
	assert.Equal(suite.T(), http.StatusNotFound, resp.StatusCode)
}

// =============================================================================
// RECIPE-INGREDIENT RELATIONSHIP SCENARIOS
// =============================================================================

func (suite *IngredientE2ETestSuite) TestRecipeIngredientManagement_CompleteWorkflow() {
	baseURL := suite.server.URL

	// Get test recipe and ingredient IDs
	var recipeID, basilID int
	err := suite.testDB.DB.QueryRow("SELECT id FROM recipe_catalogue.recipes WHERE name = 'Caprese Salad' LIMIT 1").Scan(&recipeID)
	assert.NoError(suite.T(), err)
	err = suite.testDB.DB.QueryRow("SELECT id FROM recipe_catalogue.ingredients WHERE name = 'Basil' LIMIT 1").Scan(&basilID)
	assert.NoError(suite.T(), err)

	// 1. Add ingredient to recipe
	addIngredient := map[string]interface{}{
		"ingredient_id": basilID,
		"quantity":      15.0,
		"unit":          "grams",
		"notes":         "fresh basil leaves",
	}

	resp := suite.testHttpClient.MakeAuthenticatedRequest(suite.T(), "POST", fmt.Sprintf("%s/recipes/%d/ingredients", baseURL, recipeID), addIngredient)
	assert.Equal(suite.T(), http.StatusCreated, resp.StatusCode)

	var addedIngredient map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&addedIngredient)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 15.0, addedIngredient["quantity"])
	assert.Equal(suite.T(), "fresh basil leaves", addedIngredient["notes"])

	// 2. Update recipe ingredient quantity
	updatePayload := map[string]interface{}{
		"ingredient_id": basilID,
		"quantity":      20.0,
		"unit":          "grams",
		"notes":         "extra fresh basil",
	}

	resp = suite.testHttpClient.MakeAuthenticatedRequest(suite.T(), "PUT", fmt.Sprintf("%s/recipes/%d/ingredients/%d", baseURL, recipeID, basilID), updatePayload)
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	var updatedIngredient map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&updatedIngredient)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 20.0, updatedIngredient["quantity"])
	assert.Equal(suite.T(), "extra fresh basil", updatedIngredient["notes"])

	// 3. Verify recipe now has 3 ingredients
	resp = suite.testHttpClient.MakeRequest(suite.T(), "GET", fmt.Sprintf("%s/recipes/%d/ingredients", baseURL, recipeID), nil)
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	var recipeIngredients []map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&recipeIngredients)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 3, len(recipeIngredients), "Recipe should have 3 ingredients")

	// 4. Remove basil from recipe
	resp = suite.testHttpClient.MakeAuthenticatedRequest(suite.T(), "DELETE", fmt.Sprintf("%s/recipes/%d/ingredients/%d", baseURL, recipeID, basilID), nil)
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	// 5. Verify basil removed
	resp = suite.testHttpClient.MakeRequest(suite.T(), "GET", fmt.Sprintf("%s/recipes/%d/ingredients", baseURL, recipeID), nil)
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	err = json.NewDecoder(resp.Body).Decode(&recipeIngredients)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 2, len(recipeIngredients), "Recipe should have 2 ingredients after removal")
}

// =============================================================================
// GROCERY LIST GENERATION SCENARIO
// =============================================================================

func (suite *IngredientE2ETestSuite) TestGroceryListGeneration_UserWorkflow() {
	baseURL := suite.server.URL

	// Get recipe IDs for grocery list
	var capreseID, chickenID int
	err := suite.testDB.DB.QueryRow("SELECT id FROM recipe_catalogue.recipes WHERE name = 'Caprese Salad' LIMIT 1").Scan(&capreseID)
	assert.NoError(suite.T(), err)
	err = suite.testDB.DB.QueryRow("SELECT id FROM recipe_catalogue.recipes WHERE name = 'Grilled Chicken' LIMIT 1").Scan(&chickenID)
	assert.NoError(suite.T(), err)

	// Generate grocery list for multiple recipes
	groceryListRequest := map[string]interface{}{
		"recipe_ids": []int{capreseID, chickenID},
	}

	resp := suite.testHttpClient.MakeAuthenticatedRequest(suite.T(), "POST", baseURL+"/grocery-list", groceryListRequest)
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	var groceryList []map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&groceryList)
	assert.NoError(suite.T(), err)
	assert.GreaterOrEqual(suite.T(), len(groceryList), 3, "Should have items for both recipes")

	// Verify grocery list contains expected ingredients
	ingredientNames := make(map[string]bool)
	for _, item := range groceryList {
		ingredient := item["ingredient"].(map[string]interface{})
		ingredientNames[ingredient["name"].(string)] = true

		// Verify grocery list item structure
		assert.Contains(suite.T(), item, "ingredient_id")
		assert.Contains(suite.T(), item, "total_quantity")
		assert.Contains(suite.T(), item, "unit")
		assert.Contains(suite.T(), item, "recipes")
	}

	assert.True(suite.T(), ingredientNames["Tomato"], "Grocery list should contain tomato")
	assert.True(suite.T(), ingredientNames["Mozzarella"], "Grocery list should contain mozzarella")
	assert.True(suite.T(), ingredientNames["Chicken Breast"], "Grocery list should contain chicken")
}

// =============================================================================
// ERROR HANDLING SCENARIOS
// =============================================================================

func (suite *IngredientE2ETestSuite) TestIngredientErrorHandling_EdgeCases() {
	baseURL := suite.server.URL

	// Test creating duplicate ingredient
	duplicateIngredient := map[string]interface{}{
		"name":     "Tomato", // Already exists
		"category": "Vegetables",
	}

	resp := suite.testHttpClient.MakeAuthenticatedRequest(suite.T(), "POST", baseURL+"/ingredients", duplicateIngredient)
	assert.Equal(suite.T(), http.StatusConflict, resp.StatusCode)

	// Test accessing non-existent ingredient
	resp = suite.testHttpClient.MakeRequest(suite.T(), "GET", baseURL+"/ingredients/99999", nil)
	assert.Equal(suite.T(), http.StatusNotFound, resp.StatusCode)

	// Test invalid ingredient ID format
	resp = suite.testHttpClient.MakeRequest(suite.T(), "GET", baseURL+"/ingredients/invalid", nil)
	assert.Equal(suite.T(), http.StatusNotFound, resp.StatusCode)

	// Test unauthorized access to protected endpoints
	resp = suite.testHttpClient.MakeRequest(suite.T(), "POST", baseURL+"/ingredients", duplicateIngredient)
	assert.Equal(suite.T(), http.StatusUnauthorized, resp.StatusCode)
}

// =============================================================================
// BULK OPERATIONS SCENARIO
// =============================================================================

func (suite *IngredientE2ETestSuite) TestBulkRecipeIngredientOperations_SetAllIngredients() {
	baseURL := suite.server.URL

	// Get test recipe and ingredient IDs from seeded data
	var recipeID, lettuceID, onionID int
	err := suite.testDB.DB.QueryRow("SELECT id FROM recipe_catalogue.recipes WHERE name = 'Grilled Salmon' LIMIT 1").Scan(&recipeID)
	assert.NoError(suite.T(), err)
	err = suite.testDB.DB.QueryRow("SELECT id FROM recipe_catalogue.ingredients WHERE name = 'Lettuce' LIMIT 1").Scan(&lettuceID)
	assert.NoError(suite.T(), err)
	err = suite.testDB.DB.QueryRow("SELECT id FROM recipe_catalogue.ingredients WHERE name = 'Onion' LIMIT 1").Scan(&onionID)
	assert.NoError(suite.T(), err)

	// Replace all recipe ingredients in one operation
	setIngredients := []map[string]interface{}{
		{
			"ingredient_id": lettuceID,
			"quantity":      1.0,
			"unit":          "head",
			"notes":         "fresh lettuce",
		},
		{
			"ingredient_id": onionID,
			"quantity":      0.5,
			"unit":          "piece",
			"notes":         "diced onion",
		},
	}

	resp := suite.testHttpClient.MakeAuthenticatedRequest(suite.T(), "PUT", fmt.Sprintf("%s/recipes/%d/ingredients", baseURL, recipeID), setIngredients)
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	// Verify the recipe now has exactly these ingredients
	resp = suite.testHttpClient.MakeRequest(suite.T(), "GET", fmt.Sprintf("%s/recipes/%d/ingredients", baseURL, recipeID), nil)
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	var recipeIngredients []map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&recipeIngredients)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 2, len(recipeIngredients), "Recipe should have exactly 2 ingredients")

	// Verify specific ingredients
	foundLettuce := false
	foundOnion := false
	for _, ingredient := range recipeIngredients {
		ingredientName := ingredient["ingredient"].(map[string]interface{})["name"].(string)
		switch ingredientName {
		case "Lettuce":
			foundLettuce = true
			assert.Equal(suite.T(), 1.0, ingredient["quantity"])
			assert.Equal(suite.T(), "fresh lettuce", ingredient["notes"])
		case "Onion":
			foundOnion = true
			assert.Equal(suite.T(), 0.5, ingredient["quantity"])
			assert.Equal(suite.T(), "diced onion", ingredient["notes"])
		}
	}
	assert.True(suite.T(), foundLettuce && foundOnion, "Should find both lettuce and onion")
}

// Run the test suite
func TestIngredientE2ETestSuite(t *testing.T) {
	suite.Run(t, new(IngredientE2ETestSuite))
}
