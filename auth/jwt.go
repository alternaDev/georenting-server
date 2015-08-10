package auth

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/alternaDev/georenting-server/models"
	"github.com/dgrijalva/jwt-go"
)

// GenerateJWTToken generates a JWT token for a given UserID and signs it with
// the given private key. The token will be valid for 3 days.
func GenerateJWTToken(user models.User) (string, error) {

	if user.PrivateKey == "" {
		privateKey, err := GenerateNewPrivateKey()

		if err != nil {
			return "", err
		}

		user.PrivateKey = PrivateKeyToString(privateKey)
		models.DB.Save(&user)
	}

	privateKey, err := StringToPrivateKey(user.PrivateKey)

	if err != nil {
		return "", err
	}

	token := jwt.New(jwt.SigningMethodRS512)

	token.Claims["user"] = user.ID
	token.Claims["exp"] = time.Now().Add(time.Hour * 72).Unix()

	tokenString, err := token.SignedString(privateKey)

	return tokenString, err
}

// ValidateJWTToken validates a JWT token and returns the user from the DB
// TODO: Implement Logout via Redis and a blacklist
func ValidateJWTToken(input string) (models.User, error) {
	var user models.User

	token, err := jwt.Parse(input, func(token *jwt.Token) (interface{}, error) {

		// Check whether the right signing algorithm was used.
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		// Check the expiration time.
		expiry, err := strconv.ParseInt(token.Claims["exp"].(string), 10, 64)

		if err != nil {
			return nil, err
		}

		if expiry > time.Now().Unix() {
			return nil, err
		}

		// Get the user ID
		userID, err := strconv.ParseInt(token.Claims["user"].(string), 10, 64)

		if err != nil {
			return nil, err
		}

		models.DB.First(&user, userID)

		privateKey, err := StringToPrivateKey(user.PrivateKey)

		return privateKey.PublicKey, err
	})

	if err == nil && token.Valid {
		return user, nil

	}
	return models.User{}, errors.New("The given token is invalid.")
}

// ValidateSession validates a session in a HTTP Request
func ValidateSession(r *http.Request) (models.User, error) {
	token := r.Header.Get("Authorization")

	if token == "" {
		return models.User{}, errors.New("Auth token missing.")
	}

	return ValidateJWTToken(token)
}
