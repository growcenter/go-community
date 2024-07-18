package hash

import "golang.org/x/crypto/bcrypt"

func Generate(password []byte) (hashed string, err error) {
	hashedPassword, err := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hashedPassword), nil
}

func Validate(hashed string, input string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(input))
}
