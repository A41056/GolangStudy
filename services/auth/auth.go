package services

import (
	"bytes"
	"context"
	"fmt"
	"main/common/db"
	"main/config"
	users "main/services/user"

	"time"

	"github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/scrypt"
)

type UserClaims struct {
	Username string `json:"username"`
	RoleName string `json:"roleName"`
	jwt.StandardClaims
}

func Authenticate(db *db.DB, username, password string) (users.User, error) {
	var user users.User
	collection := db.Client().Database("PracticeDb").Collection("Users")
	err := collection.FindOne(context.Background(), bson.M{"username": username}).Decode(&user)
	if err != nil {
		return user, err
	}

	providedPasswordHash, err := scrypt.Key([]byte(password), user.PasswordSalt, 16384, 8, 1, 32)

	if err != nil {
		return user, err
	}

	if !bytes.Equal(providedPasswordHash, user.PasswordHash) {
		fmt.Println("Invalid password")
		return user, fmt.Errorf("invalid password")
	}

	return user, nil
}

func GenerateToken(user users.User) (string, error) {
	cfg := config.GetConfig()
	claims := &UserClaims{
		Username: user.Username,
		RoleName: user.RoleName,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 24).Unix(),
			Issuer:    cfg.Issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.SecretKey))
}
