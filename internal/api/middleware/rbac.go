package middleware

import (
	"net/http"
	"strings"
)

// RequirePermission returns middleware that checks if the authenticated user
// has the required permission.
func RequirePermission(required string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := GetAuthUser(r.Context())
			if user == nil {
				jsonError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
				return
			}

			if !hasPermission(user.Permissions, required) {
				jsonError(w, http.StatusForbidden, "FORBIDDEN", "you do not have permission to perform this action")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireAnyPermission checks that the user has at least one of the listed permissions.
func RequireAnyPermission(perms ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := GetAuthUser(r.Context())
			if user == nil {
				jsonError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
				return
			}

			for _, p := range perms {
				if hasPermission(user.Permissions, p) {
					next.ServeHTTP(w, r)
					return
				}
			}

			jsonError(w, http.StatusForbidden, "FORBIDDEN", "you do not have permission to perform this action")
		})
	}
}

func hasPermission(userPerms []string, required string) bool {
	for _, p := range userPerms {
		if p == required || p == "*" {
			return true
		}
		// Wildcard resource match: "documents:*" matches "documents:read"
		parts := strings.SplitN(required, ":", 2)
		if len(parts) == 2 && p == parts[0]+":*" {
			return true
		}
	}
	return false
}
