// Copyright 2020 Paul Greenberg greenpau@outlook.com
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package auth

import (
	"github.com/greenpau/caddy-auth-jwt/pkg/acl"
	"github.com/greenpau/caddy-auth-jwt/pkg/claims"
	"github.com/greenpau/caddy-auth-jwt/pkg/grantor"
	"github.com/greenpau/caddy-auth-jwt/pkg/kms"
	"github.com/greenpau/caddy-auth-jwt/pkg/validator"
	"testing"
	"time"
)

func TestGrantValidate(t *testing.T) {
	secret := "75f03764-147c-4d87-b2f0-4fda89e331c8"

	claims := &claims.UserClaims{}
	claims.ExpiresAt = time.Now().Add(time.Duration(900) * time.Second).Unix()
	claims.Name = "Smith, John"
	claims.Email = "jsmith@gmail.com"
	claims.Origin = "localhost"
	claims.Subject = "jsmith"
	claims.Roles = append(claims.Roles, "anonymous")

	grantor := grantor.NewTokenGrantor()
	if err := grantor.Validate(); err == nil {
		t.Fatalf("grantor validation expected to fail, but succeeded")
	}

	if _, err := grantor.GrantTokenWithMethod("DUMMY", claims); err == nil {
		t.Fatalf("grantor signing with dummy method expected to fail, but succeeded")
	}

	if _, err := grantor.GrantTokenWithMethod("HS512", claims); err == nil {
		t.Fatalf("grantor signing with misconfiguration expected to fail, but succeeded")
	}

	if _, err := grantor.GrantTokenWithMethod("HS512", nil); err == nil {
		t.Fatalf("grantor signing of nil claims expected to fail, but succeeded")
	}

	km := kms.NewKeyManager()
	if err := km.AddKey("0", secret); err != nil {
		t.Fatal(err)
	}
	grantor.AddKeyManager(km)

	if err := grantor.Validate(); err != nil {
		t.Fatalf("grantor validation expected to succeeded, but failed: %s", err)
	}

	token, err := grantor.GrantTokenWithMethod("HS512", claims)
	if err != nil {
		t.Fatalf("grantor signing of user claims failed, but expected to succeed: %s", err)
	}
	t.Logf("Granted Token: %s", token)

	validator := validator.NewTokenValidator()
	validator.KeyManagers = []*kms.KeyManager{km}
	if err := validator.ConfigureTokenBackends(); err != nil {
		t.Fatalf("validator backend configuration failed: %s", err)
	}

	entry := acl.NewAccessListEntry()
	entry.Allow()
	if err := entry.SetClaim("roles"); err != nil {
		t.Fatalf("default access list configuration error: %s", err)
	}

	for _, v := range []string{"anonymous", "guest"} {
		if err := entry.AddValue(v); err != nil {
			t.Fatalf("default access list configuration error: %s", err)
		}
	}
	validator.AccessList = append(validator.AccessList, entry)

	userClaims, valid, err := validator.ValidateToken(token, nil)
	if err != nil {
		t.Fatalf("token validation error: %s, valid: %t, claims: %v", err, valid, userClaims)
	}
	if !valid {
		t.Fatalf("token validation error: not valid, claims: %v", userClaims)
	}
	if userClaims == nil {
		t.Fatalf("token validation error: user claims is nil")
	}
	t.Logf("Token claims: %v", userClaims)
}
