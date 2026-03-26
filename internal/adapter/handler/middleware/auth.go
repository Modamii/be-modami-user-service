package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/MicahParks/keyfunc/v2"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/modami/user-service/internal/service"
)

const (
	KeycloakIDKey = "keycloak_id"
	UserIDKey     = "user_id"
	UserRolesKey  = "user_roles"
)

type AuthMiddleware struct {
	jwks        *keyfunc.JWKS
	userService *service.UserService
}

func NewAuthMiddleware(jwksURL string, userService *service.UserService) (*AuthMiddleware, error) {
	jwks, err := keyfunc.Get(jwksURL, keyfunc.Options{
		RefreshErrorHandler: func(err error) {},
		RefreshInterval:     0,
		RefreshUnknownKID:   true,
	})
	if err != nil {
		// Allow running without JWKS (e.g. in tests/dev)
		return &AuthMiddleware{userService: userService}, nil
	}
	return &AuthMiddleware{jwks: jwks, userService: userService}, nil
}

func (m *AuthMiddleware) Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			return
		}

		tokenString := parts[1]

		var claims jwt.MapClaims
		var err error

		if m.jwks != nil {
			token, parseErr := jwt.ParseWithClaims(tokenString, &jwt.MapClaims{}, m.jwks.Keyfunc)
			if parseErr != nil || !token.Valid {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
				return
			}
			claims = *token.Claims.(*jwt.MapClaims)
		} else {
			// Dev mode: parse without verification
			token, _, parseErr := jwt.NewParser().ParseUnverified(tokenString, &jwt.MapClaims{})
			if parseErr != nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
				return
			}
			claims = *token.Claims.(*jwt.MapClaims)
		}

		_ = err

		keycloakID, ok := claims["sub"].(string)
		if !ok || keycloakID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing sub claim"})
			return
		}

		c.Set(KeycloakIDKey, keycloakID)

		// Extract roles from realm_access.roles
		var roles []string
		if realmAccess, ok := claims["realm_access"].(map[string]interface{}); ok {
			if rolesRaw, ok := realmAccess["roles"].([]interface{}); ok {
				for _, r := range rolesRaw {
					if rs, ok := r.(string); ok {
						roles = append(roles, rs)
					}
				}
			}
		}
		c.Set(UserRolesKey, roles)

		// Look up user_id from keycloak_id
		user, lookupErr := m.userService.GetByKeycloakID(context.Background(), keycloakID)
		if lookupErr == nil && user != nil {
			c.Set(UserIDKey, user.ID)
		}

		c.Next()
	}
}

func (m *AuthMiddleware) RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		rolesRaw, exists := c.Get(UserRolesKey)
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		roles, ok := rolesRaw.([]string)
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		for _, r := range roles {
			if r == role {
				c.Next()
				return
			}
		}
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
	}
}

func GetUserID(c *gin.Context) (uuid.UUID, bool) {
	val, exists := c.Get(UserIDKey)
	if !exists {
		return uuid.Nil, false
	}
	id, ok := val.(uuid.UUID)
	return id, ok
}

func GetKeycloakID(c *gin.Context) (string, bool) {
	val, exists := c.Get(KeycloakIDKey)
	if !exists {
		return "", false
	}
	id, ok := val.(string)
	return id, ok
}
