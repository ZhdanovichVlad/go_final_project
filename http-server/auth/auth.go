package authorization

import (
	"bytes"
	"encoding/json"
	"fmt"
	http_server "github.com/ZhdanovichVlad/go_final_project/http-server"
	"github.com/ZhdanovichVlad/go_final_project/storage/users"
	"github.com/dgrijalva/jwt-go"
	"net/http"
	"os"
	"time"
)

type Claims struct {
	Authorized bool `json:"authorized"`
	Sum        int  `json:"sum"`
	jwt.StandardClaims
}

// Authorization function for user authorization.
// Authorization функция для авторизации пользователя.
func Authorization(w http.ResponseWriter, req *http.Request) {
	var user users.User
	var buf bytes.Buffer
	_, err := buf.ReadFrom(req.Body)
	if err != nil {
		msg, errInt := http_server.JsonErrorMarshal(http_server.TaskResponseError{"Error when reading from req.Body"}, true)
		w.WriteHeader(errInt)
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.Write(msg)
		return
	}

	if err = json.Unmarshal(buf.Bytes(), &user); err != nil {
		msg, errInt := http_server.JsonErrorMarshal(http_server.TaskResponseError{"dicerelization error"}, false)
		w.WriteHeader(errInt)
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.Write(msg)
		return
	}
	envPassword, err := user.GetPasswordFromEnv()
	if err != nil {
		msg, errInt := http_server.JsonErrorMarshal(http_server.TaskResponseError{".env read error"}, false)
		w.WriteHeader(errInt)
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.Write(msg)
		return
	}

	if user.GetPassword() != envPassword {

		msg, errInt := http_server.JsonErrorMarshal(http_server.TaskResponseError{"wrong password"}, true)
		w.WriteHeader(errInt)
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.Write(msg)
		return
	}

	// проверяем наличие JWT-токена в cookie
	cookie, err := req.Cookie("token")
	if err == nil {
		tokenStr := cookie.Value
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(envPassword), nil
		})

		if err == nil && token.Valid && claims.ExpiresAt > time.Now().Unix() {
			response := map[string]string{"token": tokenStr}
			answear, err := json.Marshal(response)
			if err != nil {
				msg, errInt := http_server.JsonErrorMarshal(http_server.TaskResponseError{"sirelization error"}, false)
				w.WriteHeader(errInt)
				w.Header().Set("Content-Type", "application/json; charset=UTF-8")
				w.Write(msg)
				return
			}
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.Write(answear)
			return
		}
	}
	// генерация нового JWT-токена
	expirationTime := time.Now().Add(8 * time.Hour)
	claims := &Claims{
		Authorized: true,
		Sum:        500,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(envPassword))
	if err != nil {
		msg, errInt := http_server.JsonErrorMarshal(http_server.TaskResponseError{"error when creating a new token"}, false)
		w.WriteHeader(errInt)
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.Write(msg)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    tokenString,
		Path:     "/",
		Expires:  expirationTime,
		HttpOnly: true,
	})
	// записываем в формате json JWT-токен и отправляем пользователю.
	response := map[string]string{"token": tokenString}
	answear, err := json.Marshal(response)
	if err != nil {
		msg, errInt := http_server.JsonErrorMarshal(http_server.TaskResponseError{"sirelization error"}, false)
		w.WriteHeader(errInt)
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.Write(msg)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Write(answear)
}

// CheckToken function for user authentication.
// CheckToken функция для аутентификации пользователей.
func CheckToken(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		envPassword := os.Getenv("TODO_PASSWORD")
		// парсим JWT-токен из куки и если он действителен пропускаем на следующую страничку
		if len(envPassword) > 0 {
			var cookieToken string
			cookie, err := req.Cookie("token")
			if err != nil {
				msg, errInt := http_server.JsonErrorMarshal(http_server.TaskResponseError{"unauthorized user"}, true)
				w.WriteHeader(errInt)
				w.Header().Set("Content-Type", "application/json; charset=UTF-8")
				w.Write(msg)
				return
			}
			cookieToken = cookie.Value
			claims := &Claims{}
			token, err := jwt.ParseWithClaims(cookieToken, claims, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return []byte(envPassword), nil
			})
			if err != nil {
				http.Error(w, "Authentication required", http.StatusUnauthorized)
				return
			}

			if !token.Valid {
				http.Error(w, "Authentification required", http.StatusUnauthorized)
				return
			}
		}
		next(w, req)
	})
}
