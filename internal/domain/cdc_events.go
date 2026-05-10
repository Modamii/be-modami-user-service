package domain

type KeycloakCDCOp string

const (
	KeycloakCDCOpCreate   KeycloakCDCOp = "c" // INSERT
	KeycloakCDCOpUpdate   KeycloakCDCOp = "u" // UPDATE
	KeycloakCDCOpDelete   KeycloakCDCOp = "d" // DELETE
	KeycloakCDCOpSnapshot KeycloakCDCOp = "r" // Initial snapshot read
)

// KeycloakUserEntity maps the columns of the Keycloak user_entity table
type KeycloakUserEntity struct {
	ID                       string  `json:"id"`
	Email                    *string `json:"email"`
	EmailConstraint          *string `json:"email_constraint"`
	EmailVerified            bool    `json:"email_verified"`
	Enabled                  bool    `json:"enabled"`
	FederationLink           *string `json:"federation_link"`
	FirstName                *string `json:"first_name"`
	LastName                 *string `json:"last_name"`
	RealmID                  string  `json:"realm_id"`
	Username                 *string `json:"username"`
	CreatedTimestamp         int64   `json:"created_timestamp"`
	ServiceAccountClientLink *string `json:"service_account_client_link"`
	NotBefore                int     `json:"not_before"`
}

// IsServiceAccount returns true when the entity is a Keycloak service account.
func (e *KeycloakUserEntity) IsServiceAccount() bool {
	return e.ServiceAccountClientLink != nil
}

// KeycloakCDCEvent is the top-level Debezium envelope for a row change on the
// Keycloak user_entity table.
type KeycloakCDCEvent struct {
	Before *KeycloakUserEntity `json:"before"`
	After  *KeycloakUserEntity `json:"after"`
	Op     KeycloakCDCOp       `json:"op"`
	TsMs   int64               `json:"ts_ms"`
}

// KeycloakSyncFields holds the subset of user fields that are owned by Keycloak
// and must be updated atomically when a CDC event arrives.
type KeycloakSyncFields struct {
	Email         string
	Username      string
	EmailVerified bool
	Status        UserStatus
}

// KeycloakChangedFields records which fields differ between Before and After.
type KeycloakChangedFields struct {
	Email         bool
	Username      bool
	FirstName     bool
	LastName      bool
	EmailVerified bool
	Enabled       bool
}

// HasAny returns true if at least one field changed.
func (f KeycloakChangedFields) HasAny() bool {
	return f.Email || f.Username || f.FirstName || f.LastName || f.EmailVerified || f.Enabled
}

// DetectChanges compares Before and After and returns which fields changed.
// When Before is nil every field present in After is considered changed.
func (e *KeycloakCDCEvent) DetectChanges() KeycloakChangedFields {
	a := e.After
	if a == nil {
		return KeycloakChangedFields{}
	}
	b := e.Before
	if b == nil {
		// No prior state — treat all non-empty After fields as changed.
		return KeycloakChangedFields{
			Email:         a.Email != nil && *a.Email != "",
			Username:      a.Username != nil && *a.Username != "",
			FirstName:     a.FirstName != nil,
			LastName:      a.LastName != nil,
			EmailVerified: a.EmailVerified,
			Enabled:       !a.Enabled,
		}
	}

	strVal := func(p *string) string {
		if p == nil {
			return ""
		}
		return *p
	}

	return KeycloakChangedFields{
		Email:         strVal(b.Email) != strVal(a.Email),
		Username:      strVal(b.Username) != strVal(a.Username),
		FirstName:     strVal(b.FirstName) != strVal(a.FirstName),
		LastName:      strVal(b.LastName) != strVal(a.LastName),
		EmailVerified: b.EmailVerified != a.EmailVerified,
		Enabled:       b.Enabled != a.Enabled,
	}
}
