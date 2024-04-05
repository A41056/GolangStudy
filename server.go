package main

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"main/common"
	"main/common/asyncjob"
	"main/common/db"
	"main/config"
	"main/middlewares"
	services "main/services/auth"
	users "main/services/user"

	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/scrypt"
)

func main() {
	cfg := config.GetConfig()
	db, err := db.NewDB(cfg.MongoDB.ConnectionString)

	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	defer db.Close()

	e := echo.New()

	e.Use(middlewares.Middleware)

	e.POST("/register", func(c echo.Context) error {
		var req users.RegisterRequest
		if err := c.Bind(&req); err != nil {
			return err
		}

		salt := make([]byte, 32)
		_, err = rand.Read(salt)
		if err != nil {
			return err
		}
		user := users.User{
			Username: req.Username,
			FullName: req.FullName,
			RoleName: req.RoleName,
		}

		hashedPassword, err := scrypt.Key([]byte(req.Password), salt, 16384, 8, 1, 32)
		if err != nil {
			return err
		}
		user.PasswordHash = hashedPassword
		user.PasswordSalt = salt

		if err := users.CreateUser(db, user); err != nil {
			return err
		}
		return c.JSON(http.StatusCreated, "User created successfully")
	})

	e.POST("/login", func(c echo.Context) error {
		var req struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		if err := c.Bind(&req); err != nil {
			return err
		}
		user, err := services.Authenticate(db, req.Username, req.Password)
		if err != nil {
			return echo.ErrUnauthorized
		}
		token, err := services.GenerateToken(user)
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, map[string]string{
			"token": token,
		})
	})

	e.GET("/users/:id", func(c echo.Context) error {
		id := c.Param("id")
		user, err := users.GetUserByID(db, id)
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, user)
	})

	e.PUT("/users/:id", func(c echo.Context) error {
		id := c.Param("id")

		var updateUser users.User
		if err := c.Bind(&updateUser); err != nil {
			return err
		}

		if err := users.UpdateUser(db, id, updateUser); err != nil {
			return err
		}

		return c.JSON(http.StatusOK, "User updated successfully")
	})

	e.DELETE("/users/:id", func(c echo.Context) error {
		id := c.Param("id")
		claims := c.Get("user").(*services.UserClaims)
		if claims.RoleName != "Admin" {
			return c.JSON(http.StatusForbidden, "Only admins can delete users")
		}
		if err := users.DeleteUser(db, id); err != nil {
			return err
		}
		return c.JSON(http.StatusOK, "User deleted successfully")
	})

	e.GET("/users", func(c echo.Context) error {
		pageIndexStr := c.QueryParam("pageindex")
		pageSizeStr := c.QueryParam("pagesize")

		pageIndex, err := strconv.Atoi(pageIndexStr)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid pageindex"})
		}

		pageSize, err := strconv.Atoi(pageSizeStr)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid pagesize"})
		}

		paging := common.Paging{
			PageIndex: pageIndex,
			PageSize:  pageSize,
		}
		paging.Process()

		usersList, err := users.GetUserList(db, paging)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		return c.JSON(http.StatusOK, usersList)
	})

	j1 := asyncjob.NewJob(func(ctx context.Context) error {
		fmt.Println("Job 1 is running")
		return errors.New("Job 1 failed")

	})

	if err := j1.Execute(context.Background()); err != nil {
		log.Println(err)

		for {
			err := j1.Retry(context.Background())
			if err == nil || j1.State() == asyncjob.StateRetryFailed {
				break
			}
		}
	}

	e.Logger.Fatal(e.Start(":8081"))
}
