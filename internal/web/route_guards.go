package web

import "net/http"

// RoleMiddleware is the server-side authorization boundary for protected browser routes.
// It expects a session/user lookup to be attached to request context by the auth layer.
// The auth implementation should provide the concrete user role and reject missing sessions.
type UserRoleProvider interface {
	RoleFromRequest(*http.Request) (string, bool)
}

func RequireRoles(provider UserRoleProvider, roles ...string) func(http.Handler) http.Handler {
	allowed := make(map[string]struct{}, len(roles))
	for _, role := range roles { allowed[role] = struct{}{} }
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, ok := provider.RoleFromRequest(r)
			if !ok { http.Error(w, "Authentication required", http.StatusUnauthorized); return }
			if _, ok := allowed[role]; !ok { http.Error(w, "Forbidden", http.StatusForbidden); return }
			next.ServeHTTP(w, r)
		})
	}
}
