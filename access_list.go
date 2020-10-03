package jwt

import (
	"strings"
)

// AccessList Errors
const (
	ErrEmptyACLAction strError = "empty access list action"
	ErrEmptyACLClaim  strError = "empty access list claim"
	ErrEmptyMethod    strError = "empty http method"
	ErrEmptyPath      strError = "empty http path"
	ErrEmptyClaim     strError = "empty claim value"
	ErrEmptyValue     strError = "empty value"
	ErrNoValues       strError = "no acl.Values"

	ErrUnsupportedACLAction strError = "unsupported access list action: %s"
	ErrUnsupportedClaim     strError = "access list does not support %s claim, only roles"
	ErrUnsupportedMethod    strError = "unsupported http method: %s"
)

// AccessListEntry represent an access list entry.
type AccessListEntry struct {
	Action  string   `json:"action,omitempty"`
	Values  []string `json:"values,omitempty"`
	Claim   string   `json:"claim,omitempty"`
	Methods []string `json:"method,omitempty"`
	Path    string   `json:"path,omitempty"`
}

// NewAccessListEntry return an instance of AccessListEntry.
func NewAccessListEntry() *AccessListEntry {
	return &AccessListEntry{}
}

// Validate checks access list entry compliance
func (acl *AccessListEntry) Validate() error {
	if acl.Action == "" {
		return ErrEmptyACLAction
	}
	if acl.Action != "allow" && acl.Action != "deny" {
		return ErrUnsupportedACLAction.WithArgs(acl.Action)
	}
	if acl.Claim == "" {
		return ErrEmptyACLClaim
	}
	if len(acl.Values) == 0 {
		return ErrNoValues
	}
	return nil
}

// Allow sets action to allow in an access list entry.
func (acl *AccessListEntry) Allow() {
	acl.Action = "allow"
	return
}

// Deny sets action to deny in an access list entry.
func (acl *AccessListEntry) Deny() {
	acl.Action = "deny"
	return
}

// SetAction sets action in an access list entry.
func (acl *AccessListEntry) SetAction(s string) error {
	if s == "" {
		return ErrEmptyACLAction
	}
	if s != "allow" && s != "deny" {
		return ErrUnsupportedACLAction.WithArgs(s)
	}
	acl.Action = s
	return nil
}

// SetClaim sets claim value of an access list entry.
func (acl *AccessListEntry) SetClaim(s string) error {
	if s == "" {
		return ErrEmptyClaim
	}
	if s != "roles" {
		return ErrUnsupportedClaim.WithArgs(s)
	}
	acl.Claim = s
	return nil
}

// AddMethod adds http method to an access list entry.
func (acl *AccessListEntry) AddMethod(s string) error {
	if s == "" {
		return ErrEmptyMethod
	}
	s = strings.ToUpper(s)
	switch s {
	case "GET", "POST", "PUT", "PATCH", "DELETE":
	default:
		return ErrUnsupportedMethod.WithArgs(s)
	}
	acl.Methods = append(acl.Methods, s)
	return nil
}

// SetPath sets http path substring to an access list entry.
func (acl *AccessListEntry) SetPath(s string) error {
	if s == "" {
		return ErrEmptyPath
	}
	acl.Path = s
	return nil
}

// AddValue adds value to an access list entry.
func (acl *AccessListEntry) AddValue(s string) error {
	if s == "" {
		return ErrEmptyValue
	}
	acl.Values = append(acl.Values, s)
	return nil
}

// SetValue sets value to an access list entry.
func (acl *AccessListEntry) SetValue(arr []string) error {
	if len(arr) == 0 {
		return ErrEmptyValue
	}
	acl.Values = arr
	return nil
}

// GetAction returns access list entry action.
func (acl *AccessListEntry) GetAction() string {
	return acl.Action
}

// GetClaim returns access list entry claim name.
func (acl *AccessListEntry) GetClaim() string {
	return acl.Claim
}

// GetValues returns access list entry claim values.
func (acl *AccessListEntry) GetValues() string {
	return strings.Join(acl.Values, " ")
}

// IsClaimAllowed checks whether access list entry allows the claims.
func (acl *AccessListEntry) IsClaimAllowed(claims *UserClaims, opts *TokenValidatorOptions) (bool, bool) {
	claimMatches := false
	methodMatches := false
	pathMatches := false
	switch acl.Claim {
	case "roles":
		if len(claims.Roles) == 0 {
			return false, false
		}
		for _, role := range claims.Roles {
			if claimMatches {
				break
			}
			for _, value := range acl.Values {
				if value == role || value == "*" || value == "any" {
					claimMatches = true
					break
				}
			}
		}
	default:
		return false, false
	}

	if opts != nil {
		if opts.ValidateMethodPath && opts.Metadata != nil {
			// The opts.Metadata shoud contain method and path keys
			if len(acl.Methods) < 1 {
				methodMatches = true
			} else {
				// Match HTTP Request Method
				if reqMethod, exists := opts.Metadata["method"]; exists {
					for _, method := range acl.Methods {
						if reqMethod.(string) == method {
							methodMatches = true
							break
						}
					}
				} else {
					methodMatches = true
				}
			}

			if acl.Path == "" {
				pathMatches = true
			} else {
				// Match HTTP Request URI
				if reqPath, exists := opts.Metadata["path"]; exists {
					if strings.Contains(reqPath.(string), acl.Path) {
						pathMatches = true
					}
				} else {
					pathMatches = true
				}
			}
		} else {
			methodMatches = true
			pathMatches = true
		}
	} else {
		methodMatches = true
		pathMatches = true
	}

	if claimMatches && methodMatches && pathMatches {
		if acl.Action == "allow" {
			return true, false
		}
		return false, true
	}
	return false, false
}
