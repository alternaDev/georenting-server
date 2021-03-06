package auth

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/alternaDev/georenting-server/models"
	"github.com/dgrijalva/jwt-go"

	redis "github.com/alternaDev/georenting-server/models/redis"
)

func extractKeyIDAndLoadKeyFromDB(token *jwt.Token) (interface{}, error) {
	// Check whether the right signing algorithm was used.
	if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
		return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
	}

	// Get the user ID
	userID := (token.Header["user"].(float64))

	user, err := models.FindUserByID(int(userID))

	if err != nil {
		return nil, err
	}

	privateKey, err := StringToPrivateKey(user.PrivateKey)

	return privateKey.Public(), err
}

// GenerateJWTToken generates a JWT token for a given UserID and signs it with
// the given private key. The token will be valid for 3 days.
func GenerateJWTToken(user models.User) (string, error) {
	if user.PrivateKey == "" {
		privateKey, err := GenerateNewPrivateKey()

		if err != nil {
			return "", err
		}

		user.PrivateKey = PrivateKeyToString(privateKey)
		user.Save()
		if err != nil {
			return "", err
		}
	}

	privateKey, err := StringToPrivateKey(user.PrivateKey)

	if err != nil {
		return "", err
	}

	token := jwt.New(jwt.SigningMethodRS256)

	token.Claims["user"] = user.ID
	token.Header["user"] = user.ID
	token.Claims["exp"] = time.Now().Add(time.Hour * 72).Unix()
	token.Header["exp"] = token.Claims["exp"]

	tokenString, err := token.SignedString(privateKey)

	return tokenString, err
}

// ValidateJWTToken validates a JWT token and returns the user from the DB
func ValidateJWTToken(input string) (*models.User, error) {
	if redis.TokenIsInBlacklist(input) {
		return nil, errors.New("Token is in blacklist.")
	}

	token, err := jwt.Parse(input, extractKeyIDAndLoadKeyFromDB)

	if err != nil || !token.Valid {
		return nil, err
	}

	if token.Claims["user"] != token.Header["user"] {
		return nil, errors.New("The token has been tampered with...inside.")
	}

	userID := (token.Claims["user"].(float64))

	user, err := models.FindUserByID(int(userID))

	return &user, err
}

func getRemainingTokenValidity(input string) int {
	token, err := jwt.Parse(input, extractKeyIDAndLoadKeyFromDB)

	if err != nil {
		return 3600
	}

	timestamp := token.Claims["exp"]

	if validity, ok := timestamp.(float64); ok {
		tm := time.Unix(int64(validity), 0)
		remainer := tm.Sub(time.Now())
		if remainer > 0 {
			return int(remainer.Seconds() + 3600)
		}
	}
	return 3600
}

// ValidateSession validates a session in a HTTP Request
func ValidateSession(r *http.Request) (*models.User, error) {
	token := r.Header.Get("Authorization")

	if token == "" {
		return nil, errors.New("Auth token missing.")
	}

	return ValidateJWTToken(token)
}

// InvalidateToken makes an old Token unusable by putting it on a blacklist.
func InvalidateToken(token string) error {
	return redis.TokenInvalidate(token, time.Duration(getRemainingTokenValidity(token))*time.Second)
}
