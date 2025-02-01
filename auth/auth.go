package auth

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

const userFile = "users.txt"

var users = make(map[string]string)

type Auth struct {
	CallbackURL string
}

func NewAuth(url string) *Auth {
	auth := &Auth{CallbackURL: url}
	auth.loadUsersFromFile()
	return auth
}

func (a *Auth) loadUsersFromFile() {
	file, err := os.Open(userFile)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("User file does not exist, starting fresh.")
			return
		}
		fmt.Println("Error opening user file:", err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, ":")
		if len(parts) == 2 {
			users[parts[0]] = parts[1]
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading user file:", err)
	}
}

func (a *Auth) saveUserToFile(username, hashedPassword string) error {
	file, err := os.OpenFile(userFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(fmt.Sprintf("%s:%s\n", username, hashedPassword))
	return err
}

func (a *Auth) LoginPage(c *gin.Context) {
	session := sessions.Default(c)
	user := session.Get("user")
	if user != nil {
		c.Redirect(http.StatusFound, "/")
		return
	}
	c.HTML(http.StatusOK, "login.html", gin.H{
		"IsLoggedIn":  false,
		"CurrentPage": "login",
	})
}

func (a *Auth) Login(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")

	hashedPassword, exists := users[username]
	if !exists {
		a.renderLoginError(c, "Invalid username or password")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)); err != nil {
		a.renderLoginError(c, "Invalid username or password")
		return
	}

	session := sessions.Default(c)
	session.Set("user", username)
	session.Save()
	c.Redirect(http.StatusFound, "/")
}

func (a *Auth) renderLoginError(c *gin.Context, errorMsg string) {
	c.HTML(http.StatusUnauthorized, "login.html", gin.H{
		"Error":       errorMsg,
		"IsLoggedIn":  false,
		"CurrentPage": "login",
	})
}

func (a *Auth) RegisterPage(c *gin.Context) {
	session := sessions.Default(c)
	user := session.Get("user")
	if user != nil {
		c.Redirect(http.StatusFound, "/")
		return
	}
	c.HTML(http.StatusOK, "register.html", gin.H{
		"IsLoggedIn":  false,
		"CurrentPage": "register",
	})
}

func (a *Auth) Register(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")

	if username == "" || password == "" {
		c.HTML(http.StatusBadRequest, "register.html", gin.H{
			"Error":       "Username and password required",
			"IsLoggedIn":  false,
			"CurrentPage": "register",
		})
		return
	}

	if _, exists := users[username]; exists {
		c.HTML(http.StatusConflict, "register.html", gin.H{
			"Error":       "User already exists",
			"IsLoggedIn":  false,
			"CurrentPage": "register",
		})
		return
	}

	hashedPassword, err := hashPassword(password)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "register.html", gin.H{
			"Error":       "Error creating user, please try again",
			"IsLoggedIn":  false,
			"CurrentPage": "register",
		})
		return
	}

	// Save user in memory and file
	users[username] = hashedPassword
	if err := a.saveUserToFile(username, hashedPassword); err != nil {
		c.HTML(http.StatusInternalServerError, "register.html", gin.H{
			"Error":       "Error saving user data, please try again",
			"IsLoggedIn":  false,
			"CurrentPage": "register",
		})
		return
	}

	session := sessions.Default(c)
	session.Set("user", username)
	session.Save()
	c.Redirect(http.StatusFound, "/")
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func (a *Auth) Logout(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	session.Save()
	c.HTML(http.StatusOK, "logout.html", gin.H{
		"IsLoggedIn":  false,
		"CurrentPage": "logout",
	})
}

func (a *Auth) ProfilePage(c *gin.Context) {
	session := sessions.Default(c)
	user := session.Get("user")
	if user == nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}
	c.HTML(http.StatusOK, "profile.html", gin.H{
		"IsLoggedIn":  true,
		"CurrentPage": "profile",
		"Username":    user,
	})
}
