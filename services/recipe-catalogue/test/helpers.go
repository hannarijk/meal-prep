package test

import (
	"context"
	"meal-prep/shared/middleware"
	"net/http"
)

func AddAuthContext(req *http.Request, userID int, email string) *http.Request {
	userCtx := &middleware.UserContext{
		UserID: userID,
		Email:  email,
	}
	ctx := context.WithValue(req.Context(), middleware.UserCtxKey, userCtx)
	return req.WithContext(ctx)
}
