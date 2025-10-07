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

// RecipeIngredientsE2ETestSuite tests scenarios involving both recipe and ingredient APIs
type RecipeIngredientsE2ETestSuite struct {
	suite.Suite
	testDB         *helpers.TestDatabase
	server         *httptest.Server
	testHttpClient *helpers.TestHttpClient
}

func (suite *RecipeIngredientsE2ETestSuite) SetupSuite() {
	helpers.SuppressTestLogs()

	// Setup real database with testcontainers
	suite.testDB = helpers.SetupPostgresContainer(suite.T())

	// Initialize repositories and services
	recipeRepo := repository.NewRecipeRepository(suite.testDB.DB)
	categoryRepo := repository.NewCategoryRepository(suite.testDB.DB)
	ingredientRepo := repository.NewIngredientRepository(suite.testDB.DB)

	recipeService := service.NewRecipeService(recipeRepo, categoryRepo, ingredientRepo)
	ingredientService := service.NewIngredientService(ingredientRepo, recipeRepo)

	recipeHandler := handlers.NewRecipeHandler(recipeService)
	ingredientHandler := handlers.NewIngredientHandler(ingredientService)

	// Setup HTTP server with both recipe and ingredient routes
	router := mux.NewRouter()
	router.Use(middleware.LoggingMiddleware("test-recipe-ingredients-service"))

	// Public routes
	router.HandleFunc("/recipes", recipeHandler.GetAllRecipes).Methods("GET")
	router.HandleFunc("/recipes/{id:[0-9]+}", recipeHandler.GetRecipeByID).Methods("GET")
	router.HandleFunc("/categories", recipeHandler.GetAllCategories).Methods("GET")
	router.HandleFunc("/categories/{id:[0-9]+}/recipes", recipeHandler.GetRecipesByCategory).Methods("GET")
	router.HandleFunc("/ingredients", ingredientHandler.GetAllIngredients).Methods("GET")
	router.HandleFunc("/ingredients/{id:[0-9]+}", ingredientHandler.GetIngredientByID).Methods("GET")
	router.HandleFunc("/ingredients/{id:[0-9]+}/recipes", ingredientHandler.GetRecipesUsingIngredient).Methods("GET")
	router.HandleFunc("/recipes/{id:[0-9]+}/ingredients", ingredientHandler.GetRecipeIngredients).Methods("GET")

	// Protected routes
	protected := router.PathPrefix("").Subrouter()
	protected.Use(middleware.ExtractUserFromGatewayHeaders)
	protected.HandleFunc("/recipes", recipeHandler.CreateRecipe).Methods("POST")
	protected.HandleFunc("/recipes/{id:[0-9]+}", recipeHandler.UpdateRecipe).Methods("PUT")
	protected.HandleFunc("/recipes/{id:[0-9]+}", recipeHandler.DeleteRecipe).Methods("DELETE")
	protected.HandleFunc("/ingredients", ingredientHandler.CreateIngredient).Methods("POST")
	protected.HandleFunc("/ingredients/{id:[0-9]+}", ingredientHandler.UpdateIngredient).Methods("PUT")
	protected.HandleFunc("/ingredients/{id:[0-9]+}", ingredientHandler.DeleteIngredient).Methods("DELETE")
	protected.HandleFunc("/recipes/{id:[0-9]+}/ingredients", ingredientHandler.AddRecipeIngredient).Methods("POST")
	protected.HandleFunc("/recipes/{id:[0-9]+}/ingredients", ingredientHandler.SetRecipeIngredients).Methods("PUT")
	protected.HandleFunc("/recipes/{recipeId:[0-9]+}/ingredients/{ingredientId:[0-9]+}", ingredientHandler.UpdateRecipeIngredient).Methods("PUT")
	protected.HandleFunc("/recipes/{recipeId:[0-9]+}/ingredients/{ingredientId:[0-9]+}", ingredientHandler.RemoveRecipeIngredient).Methods("DELETE")
	protected.HandleFunc("/shopping-list", ingredientHandler.GenerateShoppingList).Methods("POST")

	// Health check
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "healthy", "service": "recipe-ingredients-catalogue"}`))
	}).Methods("GET")

	suite.server = httptest.NewServer(router)

	// Setup authenticated HTTP client
	suite.testHttpClient = helpers.NewTestHttpClient(
		suite.server.Client(),
		42,                 // Test user ID
		"test@example.com", // Test email
	)
}

func (suite *RecipeIngredientsE2ETestSuite) TearDownSuite() {
	if suite.server != nil {
		suite.server.Close()
	}
	suite.testDB.Cleanup(suite.T())
	helpers.RestoreTestLogs()
}

func (suite *RecipeIngredientsE2ETestSuite) SetupTest() {
	suite.testDB.CleanupTestData(suite.T())
	suite.seedTestData()
}

func (suite *RecipeIngredientsE2ETestSuite) seedTestData() {
	// Insert test categories
	categoriesSQL := `
		INSERT INTO recipe_catalogue.categories (name, description) VALUES
			('Meat', 'Meat-based dishes'),
			('Chicken', 'Chicken dishes'),
			('Fish', 'Fish and seafood'),
			('Desserts', 'Sweet treats')
		ON CONFLICT (name) DO NOTHING;
	`
	if _, err := suite.testDB.DB.Exec(categoriesSQL); err != nil {
		suite.T().Fatalf("Failed to seed categories: %v", err)
	}

	// Insert test recipes using subqueries to get actual category IDs
	recipesSQL := `
		INSERT INTO recipe_catalogue.recipes (name, description, category_id) VALUES
			('Pasta Carbonara', 'Classic Italian pasta with eggs and cheese', 
			 (SELECT id FROM recipe_catalogue.categories WHERE name = 'Meat')),
			('Grilled Chicken', 'Simple grilled chicken breast', 
			 (SELECT id FROM recipe_catalogue.categories WHERE name = 'Chicken')),
			('Salmon Teriyaki', 'Glazed salmon with teriyaki sauce', 
			 (SELECT id FROM recipe_catalogue.categories WHERE name = 'Fish')),
			('Chocolate Cake', 'Rich chocolate layer cake', 
			 (SELECT id FROM recipe_catalogue.categories WHERE name = 'Desserts'))
		ON CONFLICT DO NOTHING;
	`

	if _, err := suite.testDB.DB.Exec(recipesSQL); err != nil {
		suite.T().Fatalf("Failed to seed test recipes: %v", err)
	}

	// Insert test ingredients
	ingredientsSQL := `
		INSERT INTO recipe_catalogue.ingredients (name, description, category) VALUES
			('Chicken Breast', 'Boneless skinless chicken breast', 'Meat'),
			('Ground Beef', 'Lean ground beef', 'Meat'),
			('Salmon Fillet', 'Fresh salmon fillet', 'Fish'),
			('Pasta', 'Dry pasta noodles', 'Grains'),
			('Rice', 'Long grain rice', 'Grains'),
			('Onion', 'Yellow onion', 'Vegetables'),
			('Garlic', 'Fresh garlic cloves', 'Vegetables'),
			('Tomato', 'Fresh tomatoes', 'Vegetables'),
			('Cheese', 'Grated parmesan', 'Dairy'),
			('Olive Oil', 'Extra virgin olive oil', 'Oils')
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
		WHERE r.name = 'Pasta Carbonara' AND i.name = 'Pasta'
		UNION ALL
		SELECT r.id, i.id, 200.0, 'grams', 'test ingredient'  
		FROM recipe_catalogue.recipes r, recipe_catalogue.ingredients i
		WHERE r.name = 'Grilled Chicken' AND i.name = 'Chicken Breast'
		UNION ALL
		SELECT r.id, i.id, 150.0, 'grams', 'test ingredient'
		FROM recipe_catalogue.recipes r, recipe_catalogue.ingredients i
		WHERE r.name = 'Salmon Teriyaki' AND i.name = 'Salmon Fillet'
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
// RECIPE CREATION WITH INGREDIENT SEARCH SCENARIO
// =============================================================================

// 1. User creates recipe "Honey Garlic Chicken"
// 2. User searches for "chicken" → finds Chicken Breast
// 3. User searches for "garlic" → finds Garlic
// 4. User searches for "oil" → finds Olive Oil
// 5. User adds all 3 ingredients with quantities
// 6. User verifies recipe has all ingredients
// 7. User tests search edge cases (partial, case-insensitive)
// 8. User checks that recipe appears in ingredient usage lists
func (suite *RecipeIngredientsE2ETestSuite) TestCreateRecipeWithIngredientSearch_CompleteUserJourney() {
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

	// Step 1: User creates a new recipe
	newRecipe := map[string]interface{}{
		"name":        "Honey Garlic Chicken",
		"description": "Sweet and savory chicken dish",
		"category_id": categoryID,
	}

	resp = suite.testHttpClient.MakeAuthenticatedRequest(suite.T(), "POST", baseURL+"/recipes", newRecipe)
	assert.Equal(suite.T(), http.StatusCreated, resp.StatusCode)

	var createdRecipe map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&createdRecipe)
	assert.NoError(suite.T(), err)
	newRecipeID := int(createdRecipe["id"].(float64))

	// Step 2: User searches for ingredients by name to build the recipe

	// Search for "chicken" ingredients
	resp = suite.testHttpClient.MakeRequest(suite.T(), "GET", baseURL+"/ingredients?search=chicken", nil)
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	var chickenSearchResults []map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&chickenSearchResults)
	assert.NoError(suite.T(), err)
	assert.GreaterOrEqual(suite.T(), len(chickenSearchResults), 1, "Should find chicken ingredients")

	// Find and pick chicken breast from search results
	var chickenBreastID int
	chickenFound := false
	for _, ingredient := range chickenSearchResults {
		if ingredient["name"].(string) == "Chicken Breast" {
			chickenBreastID = int(ingredient["id"].(float64))
			chickenFound = true
			break
		}
	}
	assert.True(suite.T(), chickenFound, "Should find Chicken Breast in search results")

	// Search for "garlic" ingredients
	resp = suite.testHttpClient.MakeRequest(suite.T(), "GET", baseURL+"/ingredients?search=garlic", nil)
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	var garlicSearchResults []map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&garlicSearchResults)
	assert.NoError(suite.T(), err)
	assert.GreaterOrEqual(suite.T(), len(garlicSearchResults), 1, "Should find garlic ingredients")

	// Find and pick garlic from search results
	var garlicID int
	garlicFound := false
	for _, ingredient := range garlicSearchResults {
		if ingredient["name"].(string) == "Garlic" {
			garlicID = int(ingredient["id"].(float64))
			garlicFound = true
			break
		}
	}
	assert.True(suite.T(), garlicFound, "Should find Garlic in search results")

	// Search for oil ingredients for cooking
	resp = suite.testHttpClient.MakeRequest(suite.T(), "GET", baseURL+"/ingredients?search=oil", nil)
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	var oilSearchResults []map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&oilSearchResults)
	assert.NoError(suite.T(), err)
	assert.GreaterOrEqual(suite.T(), len(oilSearchResults), 1, "Should find oil ingredients")

	// Find and pick olive oil from search results
	var oliveOilID int
	oilFound := false
	for _, ingredient := range oilSearchResults {
		if ingredient["name"].(string) == "Olive Oil" {
			oliveOilID = int(ingredient["id"].(float64))
			oilFound = true
			break
		}
	}
	assert.True(suite.T(), oilFound, "Should find Olive Oil in search results")

	// Step 3: User adds each selected ingredient to the recipe

	// Add chicken breast to recipe
	addChicken := map[string]interface{}{
		"ingredient_id": chickenBreastID,
		"quantity":      500.0,
		"unit":          "grams",
		"notes":         "boneless, skinless",
	}

	resp = suite.testHttpClient.MakeAuthenticatedRequest(suite.T(), "POST",
		fmt.Sprintf("%s/recipes/%d/ingredients", baseURL, newRecipeID), addChicken)
	assert.Equal(suite.T(), http.StatusCreated, resp.StatusCode)

	var addedChicken map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&addedChicken)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 500.0, addedChicken["quantity"])
	assert.Equal(suite.T(), "boneless, skinless", addedChicken["notes"])

	// Add garlic to recipe
	addGarlic := map[string]interface{}{
		"ingredient_id": garlicID,
		"quantity":      4.0,
		"unit":          "cloves",
		"notes":         "minced",
	}

	resp = suite.testHttpClient.MakeAuthenticatedRequest(suite.T(), "POST",
		fmt.Sprintf("%s/recipes/%d/ingredients", baseURL, newRecipeID), addGarlic)
	assert.Equal(suite.T(), http.StatusCreated, resp.StatusCode)

	var addedGarlic map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&addedGarlic)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 4.0, addedGarlic["quantity"])
	assert.Equal(suite.T(), "cloves", addedGarlic["unit"])

	// Add olive oil to recipe
	addOil := map[string]interface{}{
		"ingredient_id": oliveOilID,
		"quantity":      2.0,
		"unit":          "tablespoons",
		"notes":         "for cooking",
	}

	resp = suite.testHttpClient.MakeAuthenticatedRequest(suite.T(), "POST",
		fmt.Sprintf("%s/recipes/%d/ingredients", baseURL, newRecipeID), addOil)
	assert.Equal(suite.T(), http.StatusCreated, resp.StatusCode)

	var addedOil map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&addedOil)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 2.0, addedOil["quantity"])
	assert.Equal(suite.T(), "for cooking", addedOil["notes"])

	// Step 4: Verify the complete recipe with all searched and added ingredients
	resp = suite.testHttpClient.MakeRequest(suite.T(), "GET",
		fmt.Sprintf("%s/recipes/%d/ingredients", baseURL, newRecipeID), nil)
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	var finalRecipeIngredients []map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&finalRecipeIngredients)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 3, len(finalRecipeIngredients), "Recipe should have 3 ingredients")

	// Verify all ingredients were added correctly with their search-selected details
	ingredientNames := make(map[string]map[string]interface{})
	for _, ingredient := range finalRecipeIngredients {
		ingredientInfo := ingredient["ingredient"].(map[string]interface{})
		name := ingredientInfo["name"].(string)
		ingredientNames[name] = ingredient
	}

	// Verify chicken breast
	assert.Contains(suite.T(), ingredientNames, "Chicken Breast")
	chickenIngredient := ingredientNames["Chicken Breast"]
	assert.Equal(suite.T(), 500.0, chickenIngredient["quantity"])
	assert.Equal(suite.T(), "grams", chickenIngredient["unit"])
	assert.Equal(suite.T(), "boneless, skinless", chickenIngredient["notes"])

	// Verify garlic
	assert.Contains(suite.T(), ingredientNames, "Garlic")
	garlicIngredient := ingredientNames["Garlic"]
	assert.Equal(suite.T(), 4.0, garlicIngredient["quantity"])
	assert.Equal(suite.T(), "cloves", garlicIngredient["unit"])
	assert.Equal(suite.T(), "minced", garlicIngredient["notes"])

	// Verify olive oil
	assert.Contains(suite.T(), ingredientNames, "Olive Oil")
	oilIngredient := ingredientNames["Olive Oil"]
	assert.Equal(suite.T(), 2.0, oilIngredient["quantity"])
	assert.Equal(suite.T(), "tablespoons", oilIngredient["unit"])
	assert.Equal(suite.T(), "for cooking", oilIngredient["notes"])

	// Step 5: Test search functionality edge cases during recipe creation

	// Test partial name search
	resp = suite.testHttpClient.MakeRequest(suite.T(), "GET", baseURL+"/ingredients?search=chick", nil)
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	var partialSearchResults []map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&partialSearchResults)
	assert.NoError(suite.T(), err)

	// Should still find chicken breast with partial search
	partialChickenFound := false
	for _, ingredient := range partialSearchResults {
		if ingredient["name"].(string) == "Chicken Breast" {
			partialChickenFound = true
			break
		}
	}
	assert.True(suite.T(), partialChickenFound, "Partial search should find Chicken Breast")

	// Test case-insensitive search
	resp = suite.testHttpClient.MakeRequest(suite.T(), "GET", baseURL+"/ingredients?search=OLIVE", nil)
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	var caseInsensitiveResults []map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&caseInsensitiveResults)
	assert.NoError(suite.T(), err)

	// Should find olive oil with uppercase search
	upperCaseOilFound := false
	for _, ingredient := range caseInsensitiveResults {
		if ingredient["name"].(string) == "Olive Oil" {
			upperCaseOilFound = true
			break
		}
	}
	assert.True(suite.T(), upperCaseOilFound, "Case-insensitive search should find Olive Oil")

	// Test empty search results
	resp = suite.testHttpClient.MakeRequest(suite.T(), "GET", baseURL+"/ingredients?search=nonexistent", nil)
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	var emptyResults []map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&emptyResults)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 0, len(emptyResults), "Search for non-existent ingredient should return empty results")

	// Step 6: Verify recipe appears in ingredient usage lists
	resp = suite.testHttpClient.MakeRequest(suite.T(), "GET",
		fmt.Sprintf("%s/ingredients/%d/recipes", baseURL, chickenBreastID), nil)
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	var recipesUsingChicken []map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&recipesUsingChicken)
	assert.NoError(suite.T(), err)

	// Should find our new recipe in the list
	honeyGarlicFound := false
	for _, recipe := range recipesUsingChicken {
		if recipe["name"].(string) == "Honey Garlic Chicken" {
			honeyGarlicFound = true
			break
		}
	}
	assert.True(suite.T(), honeyGarlicFound, "New recipe should appear in chicken breast usage list")
}

// =============================================================================
// RECIPE MODIFICATION WORKFLOW SCENARIO
// =============================================================================

func (suite *RecipeIngredientsE2ETestSuite) TestModifyExistingRecipeIngredients_UserWorkflow() {
	baseURL := suite.server.URL

	// Get an existing recipe to modify
	var existingRecipeID int
	err := suite.testDB.DB.QueryRow("SELECT id FROM recipe_catalogue.recipes WHERE name = 'Pasta Carbonara' LIMIT 1").Scan(&existingRecipeID)
	assert.NoError(suite.T(), err)

	// User wants to enhance Pasta Carbonara by adding olive oil
	// Search for oil options
	resp := suite.testHttpClient.MakeRequest(suite.T(), "GET", baseURL+"/ingredients?search=oil", nil)
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	var oilOptions []map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&oilOptions)
	assert.NoError(suite.T(), err)

	// Pick olive oil
	var oliveOilID int
	for _, ingredient := range oilOptions {
		if ingredient["name"].(string) == "Olive Oil" {
			oliveOilID = int(ingredient["id"].(float64))
			break
		}
	}

	// Add olive oil to Pasta Carbonara
	addOliveOil := map[string]interface{}{
		"ingredient_id": oliveOilID,
		"quantity":      5.0,
		"unit":          "grams",
	}

	resp = suite.testHttpClient.MakeAuthenticatedRequest(suite.T(), "POST",
		fmt.Sprintf("%s/recipes/%d/ingredients", baseURL, existingRecipeID), addOliveOil)
	assert.Equal(suite.T(), http.StatusCreated, resp.StatusCode)

	// Verify the recipe now has the additional ingredient
	resp = suite.testHttpClient.MakeRequest(suite.T(), "GET",
		fmt.Sprintf("%s/recipes/%d/ingredients", baseURL, existingRecipeID), nil)
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	var updatedIngredients []map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&updatedIngredients)
	assert.NoError(suite.T(), err)

	// Should have original ingredients plus olive oil
	assert.GreaterOrEqual(suite.T(), len(updatedIngredients), 2, "Should have at least the added chicken")

	// Verify chicken was added correctly
	oliveOilFound := false
	for _, ingredient := range updatedIngredients {
		if ingredient["ingredient"].(map[string]interface{})["name"].(string) == "Olive Oil" {
			oliveOilFound = true
			assert.Equal(suite.T(), 5.0, ingredient["quantity"])
			assert.Equal(suite.T(), "grams", ingredient["unit"])
			break
		}
	}
	assert.True(suite.T(), oliveOilFound, "Should find added olive oil in recipe")
}

// =============================================================================
// CROSS-RECIPE INGREDIENT USAGE SCENARIO
// =============================================================================

func (suite *RecipeIngredientsE2ETestSuite) TestCrossRecipeIngredientUsage_AnalyticsWorkflow() {
	baseURL := suite.server.URL

	// User wants to see which recipes use olive oil
	var oliveOilID int
	err := suite.testDB.DB.QueryRow("SELECT id FROM recipe_catalogue.ingredients WHERE name = 'Olive Oil' LIMIT 1").Scan(&oliveOilID)
	assert.NoError(suite.T(), err)

	// Get all recipes using olive oil
	resp := suite.testHttpClient.MakeRequest(suite.T(), "GET",
		fmt.Sprintf("%s/ingredients/%d/recipes", baseURL, oliveOilID), nil)
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	var recipesUsingOliveOil []map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&recipesUsingOliveOil)
	assert.NoError(suite.T(), err)

	// Store recipe names for verification
	recipeNames := make([]string, 0)
	for _, recipe := range recipesUsingOliveOil {
		recipeNames = append(recipeNames, recipe["name"].(string))
	}

	// Now create a new recipe that also uses olive oil

	// First, get a real category ID from the seeded data
	resp = suite.testHttpClient.MakeRequest(suite.T(), "GET", baseURL+"/categories", nil)
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	var categories []map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&categories)
	assert.NoError(suite.T(), err)
	assert.GreaterOrEqual(suite.T(), len(categories), 1, "Should have seeded categories")

	// Use the first available category ID
	categoryID := int(categories[0]["id"].(float64))

	// Create a new recipe
	newRecipe := map[string]interface{}{
		"name":        "Mediterranean Pasta",
		"description": "Pasta with olive oil and herbs",
		"category_id": categoryID,
	}

	resp = suite.testHttpClient.MakeAuthenticatedRequest(suite.T(), "POST", baseURL+"/recipes", newRecipe)
	assert.Equal(suite.T(), http.StatusCreated, resp.StatusCode)

	var createdRecipe map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&createdRecipe)
	assert.NoError(suite.T(), err)
	newRecipeID := int(createdRecipe["id"].(float64))

	// Add olive oil to the new recipe
	addOil := map[string]interface{}{
		"ingredient_id": oliveOilID,
		"quantity":      3.0,
		"unit":          "tablespoons",
		"notes":         "extra virgin",
	}

	resp = suite.testHttpClient.MakeAuthenticatedRequest(suite.T(), "POST",
		fmt.Sprintf("%s/recipes/%d/ingredients", baseURL, newRecipeID), addOil)
	assert.Equal(suite.T(), http.StatusCreated, resp.StatusCode)

	// Verify the new recipe now appears in olive oil usage
	resp = suite.testHttpClient.MakeRequest(suite.T(), "GET",
		fmt.Sprintf("%s/ingredients/%d/recipes", baseURL, oliveOilID), nil)
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	var updatedRecipesUsingOliveOil []map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&updatedRecipesUsingOliveOil)
	assert.NoError(suite.T(), err)

	// Should have more recipes than before
	assert.GreaterOrEqual(suite.T(), len(updatedRecipesUsingOliveOil), len(recipesUsingOliveOil),
		"Should have at least as many recipes as before")

	// Verify new recipe appears in the list
	newRecipeFound := false
	for _, recipe := range updatedRecipesUsingOliveOil {
		if recipe["name"].(string) == "Mediterranean Pasta" {
			newRecipeFound = true
			break
		}
	}
	assert.True(suite.T(), newRecipeFound, "New recipe should appear in olive oil usage list")
}

// Run the test suite
func TestRecipeIngredientsE2ETestSuite(t *testing.T) {
	suite.Run(t, new(RecipeIngredientsE2ETestSuite))
}
