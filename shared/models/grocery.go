package models

type GroceryListItem struct {
	IngredientID  int        `json:"ingredient_id"`
	Ingredient    Ingredient `json:"ingredient"`
	TotalQuantity float64    `json:"total_quantity"`
	Unit          string     `json:"unit"`
	Recipes       []string   `json:"recipes"` // List of recipe names using this ingredient
}

type GroceryListRequest struct {
	RecipeIDs []int `json:"recipe_ids"`
}
