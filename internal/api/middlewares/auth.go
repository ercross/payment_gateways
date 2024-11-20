package middlewares

import (
	"context"
	"errors"
	"net/http"
)

type ContextKey string

const userIDKey ContextKey = "user_id"

func Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), userIDKey, 1)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getUserIDFromContext(ctx context.Context) (string, error) {
	userID, ok := ctx.Value(userIDKey).(string)
	if !ok {
		return "", errors.New("user_id not found in context")
	}
	return userID, nil
}
