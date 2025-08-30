# How to run
```bash
  make build
````
#### Start services
```bash
  make up
````
#### Change log level
```bash
  LOG_LEVEL=debug make up
```

# Docker
```bash
  docker-compose down -v
```
```bash
  docker-compose up -d
```
#### Optional: Remove all unused volumes (BE CAREFUL!)
```bash
  docker volume prune
````

#### Optional: Remove all unused volumes without confirmation
```bash
  docker volume prune -f
```

## Terminal 1 - Auth with logging
```bash
  go run ./services/auth 2>&1 | tee auth.log
```

## Terminal 2 - Dish-catalogue with logging
```bash
  go run ./services/dish-catalogue 2>&1 | tee dish-catalogue.log
```

## Terminal 3 - Recommendations with logging
```bash
  go run ./services/recommendations 2>&1 | tee recommendations.log
```

# Testing

### Auth
```bash
  curl http://localhost:8001/health
```

### Dish-catalogue
```bash
  curl http://localhost:8002/health
```

### Recommendations
```bash 
  curl http://localhost:8003/health
```

## Scenarios

### Step 1: User Registration & Authentication
```bash 
    curl -X POST http://localhost:8001/register \
      -H "Content-Type: application/json" \
      -d '{
        "email": "foodie@example.com",
        "password": "password123"
      }'
```

### Step 2: Set up token from response
```bash 
  export TOKEN="your-jwt-token-here"
```

### Step 2: Explore Available Dishes & Categories
#### Get all categories
```bash 
  curl http://localhost:8002/categories
```
#### Get all dishes
```bash 
  curl http://localhost:8002/dishes
```
#### Get dishes by category
```bash 
  curl http://localhost:8002/categories/1/dishes
```

### Step 3: Set User Preferences
#### Set preferences for Meat (1) and Fish (3) categories
```bash 
    curl -X PUT http://localhost:8003/preferences \
      -H "Content-Type: application/json" \
      -H "Authorization: Bearer $TOKEN" \
      -d '{
        "preferred_categories": [1, 3]
      }'
```

### Step 4: Get Recommendations (Different Algorithms)
#### Get hybrid recommendations (default)
```bash
  curl -H "Authorization: Bearer $TOKEN" \
"http://localhost:8003/recommendations?limit=5"
```

#### Get preference-based recommendations
```bash
  curl -H "Authorization: Bearer $TOKEN" \
"http://localhost:8003/recommendations?algorithm=preference&limit=5"
```

#### Get time-decay recommendations
```bash
  curl -H "Authorization: Bearer $TOKEN" \
"http://localhost:8003/recommendations?algorithm=time_decay&limit=5"
```

### Step 5: Simulate Cooking & Rating
#### Log that you cooked dish ID 1 (Pasta with Meatballs) with rating 5
```bash
  curl -X POST http://localhost:8003/cooking \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "dish_id": 1,
    "rating": 5
  }'
```

#### Log cooking dish ID 3 (Grilled Chicken) with rating 4
```bash
  curl -X POST http://localhost:8003/cooking \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "dish_id": 3,
    "rating": 4
  }'
```

### Step 6: Check Cooking History
#### Get cooking history
```bash
  curl -H "Authorization: Bearer $TOKEN" \
"http://localhost:8003/cooking/history?limit=10"
```

### Step 7: Get Updated Recommendations
#### Get new recommendations (should now consider cooking history)
```bash
  curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8003/recommendations?algorithm=time_decay&limit=8"
```
Notice how:
* Recently cooked dishes (same day) get lower scores
* Never-cooked dishes get medium priority
* The reason field explains why each dish was recommended

### Step 8: Test Dish Management (Protected)
#### Create a new dish
```bash
  curl -X POST http://localhost:8002/dishes \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "name": "Protein Cookies",
    "description": "Healthy dessert",
    "category_id": 5
  }'
```
#### Update the dish
```bash
  curl -X PUT http://localhost:8002/dishes/6 \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "name": "Protein Cookies with Coconut",
    "description": "Healthy dessert with Coconut"
  }'
```
### Step 9: Error Testing
#### Test without authentication
```bash
  curl http://localhost:8003/recommendations
```
#### Test invalid algorithm
```bash
  curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8003/recommendations?algorithm=invalid"
```
Expected Result: ✅ Success (200) with hybrid recommendations
```
{
  "dishes": [...],
  "algorithm": "hybrid",
  "generated_at": "2025-08-28T...",
  "total_scored": 5
}
```
Why: The service has a validateAlgorithm() method that defaults to "hybrid" for invalid algorithms instead of throwing an error. This is actually good UX - graceful degradation.

#### Test invalid dish ID
```bash
  curl -X POST http://localhost:8003/cooking \
-H "Content-Type: application/json" \
-H "Authorization: Bearer $TOKEN" \
-d '{"dish_id": 999, "rating": 5}'
```
Expected Result: Status 400 Bad Request
```
{
  "error": "recommendations_error",
  "code": 400,
  "message": "dish not found"
}
```