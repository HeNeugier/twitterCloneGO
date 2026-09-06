package auth

import (
	"testing"
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

//func TestValidateJWT(t *testing.T) {
//	// Create some hashed passwords
//	password1 := "somepassword1"
//	password2 := "someotherpassword2"
//	hashed1, _ := HashPassword(password1)
//	hashed2, _ := HashPassword(password2)
//
//	tests := []struct {
//		name          string
//		password      string
//		hash          string
//		wantErr       bool
//		matchPassword bool
//	}{
//		{
//			name:          "Correct password",
//			password:      password1,
//			hash:          hashed1,
//			wantErr:       false,
//			matchPassword: true,
//		},
//		{
//			name:          "Incorrect password",
//			password:      "wrongo",
//			hash:          hashed1,
//			wantErr:       false,
//			matchPassword: false,
//		},
//		{
//			name:          "Password doesn't match",
//			password:      password1,
//			hash:          hashed2,
//			wantErr:       false,
//			matchPassword: false,
//		},
//		{
//			name:          "Empty password",
//			password:      "",
//			hash:          hashed1,
//			wantErr:       false,
//			matchPassword: false,
//		},
//		{
//			name:          "Invalid hash",
//			password:      password1,
//			hash:          "invalidhash",
//			wantErr:       true,
//			matchPassword: false,
//		},
//	}
//
//	for _, test := range tests {
//		t.Run(test.name, func(t *testing.T) {
//			match, err := CheckPasswordHash(test.password, test.hash)
//			if (err != nil) != test.wantErr {
//				t.Errorf("CheckPasswordHash() error = %v, wantErr %v", err, test.wantErr)
//			}
//			if !test.wantErr && match != test.matchPassword {
//				t.Errorf("CheckPasswordHash() expects %v, got %v", test.matchPassword, match)
//			}
//		})
//	}
//}
