package middleware

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/modami/user-service/internal/domain"
	gokit "gitlab.com/lifegoeson-libs/pkg-gokit/apperror"
	"gitlab.com/lifegoeson-libs/pkg-gokit/response"
)

const (
	ctxKeyUserID      = "user_id"
	ctxKeyKeycloakID  = "keycloak_id"
	ctxKeyRole        = "user_role"
	ctxKeyPermissions = "user_permissions"
)

var (
	errUnauthorized = gokit.New(gokit.CodeUnauthorized, "chưa xác thực")
	errForbidden    = gokit.New(gokit.CodeForbidden, "không có quyền truy cập")
)

// userFetcher is a minimal interface to avoid importing the service package.
type userFetcher interface {
	GetByKeycloakID(ctx context.Context, keycloakID string) (*domain.User, error)
}

// AuthMiddleware validates Keycloak-issued JWTs via JWKS and loads the user into context.
type AuthMiddleware struct {
	jwksURL string
	cache   *jwksCache
	svc     userFetcher
}

// NewAuthMiddleware creates an AuthMiddleware. If jwksURL is empty, tokens are parsed
// without signature verification (dev/test mode). Returns a non-nil error when the
// initial JWKS fetch fails (the middleware is still usable — keys are refreshed lazily).
func NewAuthMiddleware(jwksURL string, svc userFetcher) (*AuthMiddleware, error) {
	a := &AuthMiddleware{
		jwksURL: jwksURL,
		cache:   &jwksCache{keys: make(map[string]*rsa.PublicKey)},
		svc:     svc,
	}
	if jwksURL != "" {
		if err := a.cache.refresh(jwksURL); err != nil {
			return a, fmt.Errorf("jwks initial fetch: %w", err)
		}
	}
	return a, nil
}

// Authenticate returns a gin middleware that enforces a valid Bearer token and loads
// the authenticated user's internal ID and role into the context.
func (a *AuthMiddleware) Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		raw := c.GetHeader("Authorization")
		if raw == "" {
			response.Err(c.Writer, errUnauthorized)
			c.Abort()
			return
		}
		parts := strings.SplitN(raw, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			response.Err(c.Writer, errUnauthorized)
			c.Abort()
			return
		}

		claims, err := a.parse(parts[1])
		if err != nil {
			response.Err(c.Writer, errUnauthorized)
			c.Abort()
			return
		}

		sub, _ := claims["sub"].(string)
		if sub == "" {
			response.Err(c.Writer, errUnauthorized)
			c.Abort()
			return
		}

		// Look up the internal user record to get the DB UUID and authoritative role.
		user, err := a.svc.GetByKeycloakID(c.Request.Context(), sub)
		if err != nil {
			response.Err(c.Writer, errUnauthorized)
			c.Abort()
			return
		}

		c.Set(ctxKeyKeycloakID, sub)
		c.Set(ctxKeyUserID, user.ID)
		c.Set(ctxKeyRole, strings.ToLower(string(user.Role)))

		// Extract resource_access permissions for the client (azp claim).
		if clientID, ok := claims["azp"].(string); ok && clientID != "" {
			if ra, ok := claims["resource_access"].(map[string]interface{}); ok {
				if client, ok := ra[clientID].(map[string]interface{}); ok {
					if rolesRaw, ok := client["roles"].([]interface{}); ok {
						perms := make([]string, 0, len(rolesRaw))
						for _, r := range rolesRaw {
							if rs, ok := r.(string); ok && rs != "" {
								perms = append(perms, rs)
							}
						}
						c.Set(ctxKeyPermissions, perms)
					}
				}
			}
		}

		c.Next()
	}
}

// RequireRole returns a middleware that allows the request only when the authenticated
// user's role matches the given role (case-insensitive).
func (a *AuthMiddleware) RequireRole(role string) gin.HandlerFunc {
	want := strings.ToLower(role)
	return func(c *gin.Context) {
		r, _ := Role(c)
		if r != want {
			response.Err(c.Writer, errForbidden)
			c.Abort()
			return
		}
		c.Next()
	}
}

// UserID returns the authenticated user's internal DB UUID from context.
func UserID(c *gin.Context) (uuid.UUID, bool) {
	val, ok := c.Get(ctxKeyUserID)
	if !ok {
		return uuid.Nil, false
	}
	id, ok := val.(uuid.UUID)
	return id, ok
}

// GetKeycloakID returns the Keycloak subject (sub claim) for the authenticated user.
func GetKeycloakID(c *gin.Context) (string, bool) {
	val, ok := c.Get(ctxKeyKeycloakID)
	if !ok {
		return "", false
	}
	id, ok := val.(string)
	return id, ok
}

// Role returns the authenticated user's role (lowercased) from context.
func Role(c *gin.Context) (string, bool) {
	val, ok := c.Get(ctxKeyRole)
	role, _ := val.(string)
	return role, ok
}

// Permissions returns the resource_access permissions for the authenticated user.
func Permissions(c *gin.Context) []string {
	val, _ := c.Get(ctxKeyPermissions)
	perms, _ := val.([]string)
	return perms
}

// HasPermission reports whether the authenticated user holds the given permission.
func HasPermission(c *gin.Context, permission string) bool {
	for _, p := range Permissions(c) {
		if p == permission {
			return true
		}
	}
	return false
}

// RequirePermission returns a middleware that allows the request only when
// the authenticated user holds the specified resource_access permission.
func RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !HasPermission(c, permission) {
			response.Err(c.Writer, errForbidden)
			c.Abort()
			return
		}
		c.Next()
	}
}

// parse validates and extracts claims from a JWT token string.
func (a *AuthMiddleware) parse(tokenStr string) (jwt.MapClaims, error) {
	if a.jwksURL == "" {
		// Dev mode: no signature verification.
		token, _, err := jwt.NewParser().ParseUnverified(tokenStr, jwt.MapClaims{})
		if err != nil {
			return nil, err
		}
		return token.Claims.(jwt.MapClaims), nil
	}

	token, err := jwt.ParseWithClaims(tokenStr, jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		kid, _ := token.Header["kid"].(string)
		return a.cache.get(kid, a.jwksURL)
	})
	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return token.Claims.(jwt.MapClaims), nil
}

// ---------------------------------------------------------------------------
// JWKS cache
// ---------------------------------------------------------------------------

type jwksCache struct {
	mu          sync.RWMutex
	keys        map[string]*rsa.PublicKey
	lastRefresh time.Time
}

func (c *jwksCache) get(kid string, url string) (*rsa.PublicKey, error) {
	c.mu.RLock()
	key, ok := c.keys[kid]
	c.mu.RUnlock()
	if ok {
		return key, nil
	}
	// Unknown kid — refresh and retry once.
	if err := c.refresh(url); err != nil {
		return nil, err
	}
	c.mu.RLock()
	key, ok = c.keys[kid]
	c.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("unknown key id: %s", kid)
	}
	return key, nil
}

func (c *jwksCache) refresh(url string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	resp, err := (&http.Client{Timeout: 10 * time.Second}).Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var jwks struct {
		Keys []struct {
			Kty string `json:"kty"`
			Kid string `json:"kid"`
			N   string `json:"n"`
			E   string `json:"e"`
		} `json:"keys"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return err
	}

	for _, k := range jwks.Keys {
		if k.Kty != "RSA" {
			continue
		}
		pub, err := parseRSAPublicKey(k.N, k.E)
		if err != nil {
			continue
		}
		c.keys[k.Kid] = pub
	}
	c.lastRefresh = time.Now()
	return nil
}

func parseRSAPublicKey(nB64, eB64 string) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(nB64)
	if err != nil {
		return nil, err
	}
	eBytes, err := base64.RawURLEncoding.DecodeString(eB64)
	if err != nil {
		return nil, err
	}
	n := new(big.Int).SetBytes(nBytes)
	e := new(big.Int).SetBytes(eBytes)
	return &rsa.PublicKey{N: n, E: int(e.Int64())}, nil
}
