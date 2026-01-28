package testdata

import (
	"meal-prep/shared/models"
	"time"
)

type RecipeBuilder struct {
	recipe models.Recipe
}

func NewRecipeBuilder() *RecipeBuilder {
	now := time.Now()
	return &RecipeBuilder{
		recipe: models.Recipe{
			ID:          1,
			Name:        "Default Recipe",
			Description: stringPtr("Default description"),
			CategoryID:  intPtr(1),
			CreatedAt:   now,
			UpdatedAt:   now,
			Category: &models.Category{
				ID:   1,
				Name: "Default Category",
			},
		},
	}
}

func (b *RecipeBuilder) WithID(id int) *RecipeBuilder {
	b.recipe.ID = id
	return b
}

func (b *RecipeBuilder) WithName(name string) *RecipeBuilder {
	b.recipe.Name = name
	return b
}

func (b *RecipeBuilder) WithDescription(description string) *RecipeBuilder {
	b.recipe.Description = &description
	return b
}

func (b *RecipeBuilder) WithNoDescription() *RecipeBuilder {
	b.recipe.Description = nil
	return b
}

func (b *RecipeBuilder) WithCategoryID(categoryID int) *RecipeBuilder {
	b.recipe.CategoryID = &categoryID
	return b
}

func (b *RecipeBuilder) WithNoCategory() *RecipeBuilder {
	b.recipe.CategoryID = nil
	b.recipe.Category = nil
	return b
}

func (b *RecipeBuilder) WithCategory(category models.Category) *RecipeBuilder {
	b.recipe.Category = &category
	if category.ID > 0 {
		b.recipe.CategoryID = &category.ID
	}
	return b
}

func (b *RecipeBuilder) Build() models.Recipe {
	return b.recipe
}

func (b *RecipeBuilder) BuildPtr() *models.Recipe {
	recipe := b.recipe
	return &recipe
}

type CategoryBuilder struct {
	category models.Category
}

func NewCategoryBuilder() *CategoryBuilder {
	return &CategoryBuilder{
		category: models.Category{
			ID:          1,
			Name:        "Default Category",
			Description: stringPtr("Default category description"),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}
}

func (b *CategoryBuilder) WithID(id int) *CategoryBuilder {
	b.category.ID = id
	return b
}

func (b *CategoryBuilder) WithName(name string) *CategoryBuilder {
	b.category.Name = name
	return b
}

func (b *CategoryBuilder) WithDescription(description string) *CategoryBuilder {
	b.category.Description = &description
	return b
}

func (b *CategoryBuilder) WithNoDescription() *CategoryBuilder {
	b.category.Description = nil
	return b
}

func (b *CategoryBuilder) Build() models.Category {
	return b.category
}

func (b *CategoryBuilder) BuildPtr() *models.Category {
	category := b.category
	return &category
}

type RecipeWithIngredientsBuilder struct {
	recipeWithIngredients models.RecipeWithIngredients
}

func NewRecipeWithIngredientsBuilder() *RecipeWithIngredientsBuilder {
	return &RecipeWithIngredientsBuilder{
		recipeWithIngredients: models.RecipeWithIngredients{
			Recipe:      NewRecipeBuilder().Build(),
			Ingredients: []models.RecipeIngredient{},
		},
	}
}

func (b *RecipeWithIngredientsBuilder) WithRecipe(recipe models.Recipe) *RecipeWithIngredientsBuilder {
	b.recipeWithIngredients.Recipe = recipe
	return b
}

func (b *RecipeWithIngredientsBuilder) WithIngredients(ingredients []models.RecipeIngredient) *RecipeWithIngredientsBuilder {
	b.recipeWithIngredients.Ingredients = ingredients
	return b
}

func (b *RecipeWithIngredientsBuilder) WithSingleIngredient(ingredient models.RecipeIngredient) *RecipeWithIngredientsBuilder {
	b.recipeWithIngredients.Ingredients = []models.RecipeIngredient{ingredient}
	return b
}

func (b *RecipeWithIngredientsBuilder) Build() models.RecipeWithIngredients {
	return b.recipeWithIngredients
}

func (b *RecipeWithIngredientsBuilder) BuildPtr() *models.RecipeWithIngredients {
	recipeWithIngredients := b.recipeWithIngredients
	return &recipeWithIngredients
}

type CreateRecipeRequestBuilder struct {
	request models.CreateRecipeRequest
}

func NewCreateRecipeRequestBuilder() *CreateRecipeRequestBuilder {
	return &CreateRecipeRequestBuilder{
		request: models.CreateRecipeRequest{
			Name:        "Default Recipe",
			Description: "Default description",
			CategoryID:  1,
		},
	}
}

func (b *CreateRecipeRequestBuilder) WithName(name string) *CreateRecipeRequestBuilder {
	b.request.Name = name
	return b
}

func (b *CreateRecipeRequestBuilder) WithDescription(description string) *CreateRecipeRequestBuilder {
	b.request.Description = description
	return b
}

func (b *CreateRecipeRequestBuilder) WithCategoryID(categoryID int) *CreateRecipeRequestBuilder {
	b.request.CategoryID = categoryID
	return b
}

func (b *CreateRecipeRequestBuilder) Build() models.CreateRecipeRequest {
	return b.request
}

type UpdateRecipeRequestBuilder struct {
	request models.UpdateRecipeRequest
}

func NewUpdateRecipeRequestBuilder() *UpdateRecipeRequestBuilder {
	return &UpdateRecipeRequestBuilder{
		request: models.UpdateRecipeRequest{
			Name:        "Updated Recipe",
			Description: "Updated description",
			CategoryID:  1,
		},
	}
}

func (b *UpdateRecipeRequestBuilder) WithName(name string) *UpdateRecipeRequestBuilder {
	b.request.Name = name
	return b
}

func (b *UpdateRecipeRequestBuilder) WithDescription(description string) *UpdateRecipeRequestBuilder {
	b.request.Description = description
	return b
}

func (b *UpdateRecipeRequestBuilder) WithCategoryID(categoryID int) *UpdateRecipeRequestBuilder {
	b.request.CategoryID = categoryID
	return b
}

func (b *UpdateRecipeRequestBuilder) Build() models.UpdateRecipeRequest {
	return b.request
}

type IngredientBuilder struct {
	ingredient models.Ingredient
}

func NewIngredientBuilder() *IngredientBuilder {
	return &IngredientBuilder{
		ingredient: models.Ingredient{
			ID:          1,
			Name:        "Default Ingredient",
			Description: stringPtr("Default ingredient description"),
			Category:    stringPtr("Default Category"),
			CreatedAt:   time.Now(),
		},
	}
}

func (b *IngredientBuilder) WithID(id int) *IngredientBuilder {
	b.ingredient.ID = id
	return b
}

func (b *IngredientBuilder) WithName(name string) *IngredientBuilder {
	b.ingredient.Name = name
	return b
}

func (b *IngredientBuilder) WithDescription(description string) *IngredientBuilder {
	b.ingredient.Description = &description
	return b
}

func (b *IngredientBuilder) WithNoDescription() *IngredientBuilder {
	b.ingredient.Description = nil
	return b
}

func (b *IngredientBuilder) WithCategory(category string) *IngredientBuilder {
	b.ingredient.Category = &category
	return b
}

func (b *IngredientBuilder) WithNoCategory() *IngredientBuilder {
	b.ingredient.Category = nil
	return b
}

func (b *IngredientBuilder) Build() models.Ingredient {
	return b.ingredient
}

func (b *IngredientBuilder) BuildPtr() *models.Ingredient {
	ingredient := b.ingredient
	return &ingredient
}

type RecipeIngredientBuilder struct {
	recipeIngredient models.RecipeIngredient
}

func NewRecipeIngredientBuilder() *RecipeIngredientBuilder {
	return &RecipeIngredientBuilder{
		recipeIngredient: models.RecipeIngredient{
			ID:           1,
			RecipeID:     1,
			IngredientID: 1,
			Ingredient:   NewIngredientBuilder().Build(),
			Quantity:     100.0,
			Unit:         "grams",
			Notes:        stringPtr("Fresh"),
			CreatedAt:    time.Now(),
		},
	}
}

func (b *RecipeIngredientBuilder) WithID(id int) *RecipeIngredientBuilder {
	b.recipeIngredient.ID = id
	return b
}

func (b *RecipeIngredientBuilder) WithRecipeID(recipeID int) *RecipeIngredientBuilder {
	b.recipeIngredient.RecipeID = recipeID
	return b
}

func (b *RecipeIngredientBuilder) WithIngredientID(ingredientID int) *RecipeIngredientBuilder {
	b.recipeIngredient.IngredientID = ingredientID
	return b
}

func (b *RecipeIngredientBuilder) WithIngredient(ingredient models.Ingredient) *RecipeIngredientBuilder {
	b.recipeIngredient.Ingredient = ingredient
	b.recipeIngredient.IngredientID = ingredient.ID
	return b
}

func (b *RecipeIngredientBuilder) WithQuantity(quantity float64) *RecipeIngredientBuilder {
	b.recipeIngredient.Quantity = quantity
	return b
}

func (b *RecipeIngredientBuilder) WithUnit(unit string) *RecipeIngredientBuilder {
	b.recipeIngredient.Unit = unit
	return b
}

func (b *RecipeIngredientBuilder) WithNotes(notes string) *RecipeIngredientBuilder {
	b.recipeIngredient.Notes = &notes
	return b
}

func (b *RecipeIngredientBuilder) WithNoNotes() *RecipeIngredientBuilder {
	b.recipeIngredient.Notes = nil
	return b
}

func (b *RecipeIngredientBuilder) Build() models.RecipeIngredient {
	return b.recipeIngredient
}

func (b *RecipeIngredientBuilder) BuildPtr() *models.RecipeIngredient {
	recipeIngredient := b.recipeIngredient
	return &recipeIngredient
}

type CreateIngredientRequestBuilder struct {
	request models.CreateIngredientRequest
}

func NewCreateIngredientRequestBuilder() *CreateIngredientRequestBuilder {
	return &CreateIngredientRequestBuilder{
		request: models.CreateIngredientRequest{
			Name:        "Default Ingredient",
			Description: stringPtr("Default description"),
			Category:    stringPtr("Default Category"),
		},
	}
}

func (b *CreateIngredientRequestBuilder) WithName(name string) *CreateIngredientRequestBuilder {
	b.request.Name = name
	return b
}

func (b *CreateIngredientRequestBuilder) WithDescription(description string) *CreateIngredientRequestBuilder {
	b.request.Description = &description
	return b
}

func (b *CreateIngredientRequestBuilder) WithNoDescription() *CreateIngredientRequestBuilder {
	b.request.Description = nil
	return b
}

func (b *CreateIngredientRequestBuilder) WithCategory(category string) *CreateIngredientRequestBuilder {
	b.request.Category = &category
	return b
}

func (b *CreateIngredientRequestBuilder) WithNoCategory() *CreateIngredientRequestBuilder {
	b.request.Category = nil
	return b
}

func (b *CreateIngredientRequestBuilder) Build() models.CreateIngredientRequest {
	return b.request
}

type UpdateIngredientRequestBuilder struct {
	request models.UpdateIngredientRequest
}

func NewUpdateIngredientRequestBuilder() *UpdateIngredientRequestBuilder {
	return &UpdateIngredientRequestBuilder{
		request: models.UpdateIngredientRequest{
			Name:        "Updated Ingredient",
			Description: "Updated description",
			Category:    "Updated Category",
		},
	}
}

func (b *UpdateIngredientRequestBuilder) WithName(name string) *UpdateIngredientRequestBuilder {
	b.request.Name = name
	return b
}

func (b *UpdateIngredientRequestBuilder) WithDescription(description string) *UpdateIngredientRequestBuilder {
	b.request.Description = description
	return b
}

func (b *UpdateIngredientRequestBuilder) WithCategory(category string) *UpdateIngredientRequestBuilder {
	b.request.Category = category
	return b
}

func (b *UpdateIngredientRequestBuilder) Build() models.UpdateIngredientRequest {
	return b.request
}

type AddRecipeIngredientRequestBuilder struct {
	request models.AddRecipeIngredientRequest
}

func NewAddRecipeIngredientRequestBuilder() *AddRecipeIngredientRequestBuilder {
	return &AddRecipeIngredientRequestBuilder{
		request: models.AddRecipeIngredientRequest{
			IngredientID: 1,
			Quantity:     100.0,
			Unit:         "grams",
			Notes:        stringPtr("Fresh"),
		},
	}
}

func (b *AddRecipeIngredientRequestBuilder) WithIngredientID(ingredientID int) *AddRecipeIngredientRequestBuilder {
	b.request.IngredientID = ingredientID
	return b
}

func (b *AddRecipeIngredientRequestBuilder) WithQuantity(quantity float64) *AddRecipeIngredientRequestBuilder {
	b.request.Quantity = quantity
	return b
}

func (b *AddRecipeIngredientRequestBuilder) WithUnit(unit string) *AddRecipeIngredientRequestBuilder {
	b.request.Unit = unit
	return b
}

func (b *AddRecipeIngredientRequestBuilder) WithNotes(notes string) *AddRecipeIngredientRequestBuilder {
	b.request.Notes = &notes
	return b
}

func (b *AddRecipeIngredientRequestBuilder) WithNoNotes() *AddRecipeIngredientRequestBuilder {
	b.request.Notes = nil
	return b
}

func (b *AddRecipeIngredientRequestBuilder) Build() models.AddRecipeIngredientRequest {
	return b.request
}

type GroceryListRequestBuilder struct {
	request models.GroceryListRequest
}

func NewGroceryListRequestBuilder() *GroceryListRequestBuilder {
	return &GroceryListRequestBuilder{
		request: models.GroceryListRequest{
			RecipeIDs: []int{1, 2},
		},
	}
}

func (b *GroceryListRequestBuilder) WithRecipeIDs(recipeIDs []int) *GroceryListRequestBuilder {
	b.request.RecipeIDs = recipeIDs
	return b
}

func (b *GroceryListRequestBuilder) WithSingleRecipe(recipeID int) *GroceryListRequestBuilder {
	b.request.RecipeIDs = []int{recipeID}
	return b
}

func (b *GroceryListRequestBuilder) WithNoRecipes() *GroceryListRequestBuilder {
	b.request.RecipeIDs = []int{}
	return b
}

func (b *GroceryListRequestBuilder) Build() models.GroceryListRequest {
	return b.request
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}

// Common test scenarios - frequently used combinations
func ValidRecipeWithCategory() models.Recipe {
	return NewRecipeBuilder().
		WithName("Pasta Carbonara").
		WithDescription("Classic Italian pasta dish").
		WithCategory(models.Category{
			ID:   1,
			Name: "Italian",
		}).
		Build()
}

func MinimalValidRecipe() models.Recipe {
	return NewRecipeBuilder().
		WithName("Simple Recipe").
		WithNoDescription().
		WithCategoryID(1).
		Build()
}

func ValidCreateRequest() models.CreateRecipeRequest {
	return NewCreateRecipeRequestBuilder().
		WithName("Test Recipe").
		WithDescription("Test description").
		WithCategoryID(1).
		Build()
}

// Ingredient test scenarios
func ValidIngredient() models.Ingredient {
	return NewIngredientBuilder().
		WithName("Tomato").
		WithDescription("Fresh red tomato").
		WithCategory("Vegetable").
		Build()
}

func ValidRecipeIngredient() models.RecipeIngredient {
	return NewRecipeIngredientBuilder().
		WithIngredient(ValidIngredient()).
		WithQuantity(200.0).
		WithUnit("grams").
		WithNotes("Diced").
		Build()
}

func ValidCreateIngredientRequest() models.CreateIngredientRequest {
	return NewCreateIngredientRequestBuilder().
		WithName("Basil").
		WithDescription("Fresh basil leaves").
		WithCategory("Herb").
		Build()
}

func ValidAddRecipeIngredientRequest() models.AddRecipeIngredientRequest {
	return NewAddRecipeIngredientRequestBuilder().
		WithIngredientID(1).
		WithQuantity(50.0).
		WithUnit("grams").
		WithNotes("Chopped").
		Build()
}
