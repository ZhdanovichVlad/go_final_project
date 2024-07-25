package authorization

import (
	"bytes"
	"encoding/json"
	"fmt"
	http_server "github.com/ZhdanovichVlad/go_final_project/http-server"
	"github.com/ZhdanovichVlad/go_final_project/http-server/token"
	"github.com/ZhdanovichVlad/go_final_project/storage/users"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"time"
)

type Service struct {
	claims      token.AuthClaims
	accessToken string
}

func NewService() *Service {
	return &Service{}
}

func (s Service) GetToken() (string, error) {
	if s.accessToken == "" {
		return "", fmt.Errorf("accessToken is emty")
	} else {
		return s.accessToken, nil
	}
}

func (s Service) Authorization(w http.ResponseWriter, req *http.Request) {
	fmt.Println("Вход в ручку Authorization")
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
	now := time.Now()
	s.claims = token.AuthClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(8 * time.Hour)),
		},
		TokenSum: 500,
	}

	accessToken, err := token.GenerateAccessToken(s.claims)
	if err != nil {
		msg, errInt := http_server.JsonErrorMarshal(http_server.TaskResponseError{"access token generation error"}, false)
		w.WriteHeader(errInt)
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.Write(msg)
		return
	}
	s.accessToken = accessToken
	accessTokenCookie := http.Cookie{
		Name:     "accessToken",
		Value:    s.accessToken,
		HttpOnly: true,
	}
	http.SetCookie(w, &accessTokenCookie)
}

func (s Service) VerifyUser(cookieToken string) bool {
	//userToken, err := s.GetToken()
	//if err != nil {
	//	return false
	//}
	userToken := s.accessToken
	fmt.Println("Куки токен", cookieToken)
	fmt.Println("user токен", userToken)
	fmt.Println("Сравнение", userToken == cookieToken)
	return userToken == cookieToken
}
