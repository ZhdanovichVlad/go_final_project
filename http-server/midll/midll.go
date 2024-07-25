package midll

import (
	authorization "github.com/ZhdanovichVlad/go_final_project/http-server/auth"
	"net/http"
	"os"
)

type AuthMiddleware struct {
	authService *authorization.Service
}

func NewAuthMiddleware(authService *authorization.Service) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
	}
}

func (a AuthMiddleware) CheckToken(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// смотрим наличие пароля
		pass := os.Getenv("TODO_PASSWORD")
		if len(pass) > 0 {
			var cookieToken string // JWT-токен из куки
			cookie, err := r.Cookie("accessToken")
			if err == nil {
				cookieToken = cookie.Value
			}

			var valid bool
			valid = a.authService.VerifyUser(cookieToken) // здесь код для валидации и проверки JWT-токена

			if !valid {
				// возвращаем ошибку авторизации 401
				http.Error(w, "Authentification required", http.StatusUnauthorized)
				return
			}
		}
		next(w, r)
	})
}
