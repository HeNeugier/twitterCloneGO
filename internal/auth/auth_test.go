package auth

import (
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestCheckPasswordHash(t *testing.T) {
	// Create some hashed passwords
	password1 := "somepassword1"
	password2 := "someotherpassword2"
	hashed1, _ := HashPassword(password1)
	hashed2, _ := HashPassword(password2)

	tests := []struct {
		name          string
		password      string
		hash          string
		wantErr       bool
		matchPassword bool
	}{
		{
			name:          "Correct password",
			password:      password1,
			hash:          hashed1,
			wantErr:       false,
			matchPassword: true,
		},
		{
			name:          "Incorrect password",
			password:      "wrongo",
			hash:          hashed1,
			wantErr:       false,
			matchPassword: false,
		},
		{
			name:          "Password doesn't match",
			password:      password1,
			hash:          hashed2,
			wantErr:       false,
			matchPassword: false,
		},
		{
			name:          "Empty password",
			password:      "",
			hash:          hashed1,
			wantErr:       false,
			matchPassword: false,
		},
		{
			name:          "Invalid hash",
			password:      password1,
			hash:          "invalidhash",
			wantErr:       true,
			matchPassword: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			match, err := CheckPasswordHash(test.password, test.hash)
			if (err != nil) != test.wantErr {
				t.Errorf("CheckPasswordHash() error = %v, wantErr %v", err, test.wantErr)
			}
			if !test.wantErr && match != test.matchPassword {
				t.Errorf("CheckPasswordHash() expects %v, got %v", test.matchPassword, match)
			}
		})
	}
}

func TestValidateJWT(t *testing.T) {
	// Create some JWTs for users
	user1, _ := uuid.NewUUID()
	user2, _ := uuid.NewUUID()
	secret1 := "supersecret"
	secret2 := "moresecretsecret"
	tok1, _ := MakeJWT(user1, secret1, time.Duration(time.Second*360))
	tok2, _ := MakeJWT(user2, secret2, time.Duration(time.Second*360))

	tests := []struct {
		name         string
		tokenString  string
		secretString string
		user         uuid.UUID
		wantErr      bool
		matchUUID    bool
	}{
		{
			name:         "Correct token",
			tokenString:  tok1,
			secretString: secret1,
			user:         user1,
			wantErr:      false,
			matchUUID:    true,
		},
		{
			name:         "Incorrect user for token",
			tokenString:  tok1,
			secretString: secret1,
			user:         user2,
			wantErr:      false,
			matchUUID:    false,
		},
		{
			name:         "Incorrect token for user",
			tokenString:  tok2,
			secretString: secret1,
			user:         user1,
			wantErr:      true,
			matchUUID:    false,
		},
		{
			name:         "Incorrect secret",
			tokenString:  tok1,
			secretString: secret2,
			user:         user1,
			wantErr:      true,
			matchUUID:    true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			uuid, err := ValidateJWT(test.tokenString, test.secretString)
			match := uuid == test.user
			if (err != nil) != test.wantErr {
				t.Errorf("ValidateJWT() error = %v, wantErr %v", err, test.wantErr)
			}
			if !test.wantErr && match != test.matchUUID {
				t.Errorf("ValidateJWT() gotUserID = %v, want %v", uuid, test.user)
			}
		})
	}
}

func TestGetBearerToken(t *testing.T) {
	// Create some new tokens
	tokky1, _ := MakeJWT(uuid.New(), "verysecretstring", time.Duration(time.Second*360))
	tokky2, _ := MakeJWT(uuid.New(), "notsosecret", time.Duration(time.Second*360))
	shorty, _ := MakeJWT(uuid.New(), "short", time.Duration(1))

	tests := []struct {
		name    string
		token   string
		header  http.Header
		wantErr bool
	}{
		{
			name:    "Correct token and Bearer format",
			token:   tokky1,
			header:  http.Header{"Authorization": []string{"Bearer " + tokky1}},
			wantErr: false,
		},
		{
			name:    "Correct token but incorrect 'Bearer'",
			token:   tokky2,
			header:  http.Header{"Authorization": []string{"Tokky " + tokky2}},
			wantErr: true,
		},
		{
			name:    "Incorrect token",
			token:   tokky2,
			header:  http.Header{"Authorization": []string{"Tokky " + tokky1}},
			wantErr: true,
		},
		{
			name:    "Expired token",
			token:   shorty,
			header:  http.Header{"Authorization": []string{"Tokky " + shorty}},
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			bearerToken, err := GetBearerToken(test.header)
			if (err != nil) != test.wantErr {
				t.Errorf("GetBearerToken() error = %v, wantErr %v", err, test.wantErr)
			}
			if !test.wantErr && bearerToken != test.token {
				t.Errorf("GetBearerToken() expects %v, got %v", test.token, bearerToken)
			}
		})
	}
}
