package middlewares

import (
	"fmt"
	"main/config"
	services "main/services/auth"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4"
)

func Middleware(next echo.HandlerFunc) echo.HandlerFunc {
	cfg := config.GetConfig()
	return func(c echo.Context) error {
		if c.Request().URL.Path == "/register" || c.Request().URL.Path == "/login" {
			return next(c)
		}

		tokenHeader := c.Request().Header.Get("Authorization")
		if tokenHeader == "" {
			return c.JSON(http.StatusUnauthorized, "Missing authorization token")
		}

		tokenString := strings.Replace(tokenHeader, "Bearer ", "", 1)

		claims := &services.UserClaims{}
		t, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(cfg.SecretKey), nil
		})
		if err != nil {
			if err == jwt.ErrSignatureInvalid {
				return c.JSON(http.StatusUnauthorized, "Invalid token signature")
			}
			fmt.Println(err)
			return c.JSON(http.StatusUnauthorized, "Invalid token")
		}
		if !t.Valid {
			return c.JSON(http.StatusUnauthorized, "Invalid token")
		}

		c.Set("user", claims)

		if claims.RoleName == "Admin" {
			return next(c)
		}

		allowedRoutes := []string{"/users/:id"}
		for _, route := range allowedRoutes {
			if c.Path() == route {
				return next(c)
			}
		}

		return c.JSON(http.StatusUnauthorized, "Unauthorized access")
	}
}
