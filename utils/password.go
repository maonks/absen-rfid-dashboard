package utils

import "golang.org/x/crypto/bcrypt"

func HashPassword(plain string) (string, error) {

	hash, err := bcrypt.GenerateFromPassword([]byte(plain), 14)

	return string(hash), err
}

func CekPassword(hash, plain string) bool {

	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}
