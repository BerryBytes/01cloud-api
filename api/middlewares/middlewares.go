package middlewares

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"01cloud-api/api/auth"
	"01cloud-api/api/models"
	"01cloud-api/api/responses"
	"01cloud-api/api/utils/cache"

	log "github.com/sirupsen/logrus"

	"github.com/jinzhu/gorm"
)

var (
	DB     *gorm.DB
	UserI  = models.NewUser()
	TokenI = models.NewToken()
)

func SetMiddlewareJSON(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next(w, r)
	}
}

func HideFolders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fi, err := os.Stat("/data/uploads" + r.URL.Path)
		if err != nil || fi.Mode().IsDir() || strings.HasSuffix(r.URL.Path, "json") {
			http.NotFound(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func CORS(next http.Handler) http.Handler {
	allowedOrigins := strings.Split(os.Getenv("ALLOWED_ORIGINS"), ",")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		allowed := false

		for _, o := range allowedOrigins {
			o = strings.TrimSpace(o)
			if o == "*" || origin == o {
				allowed = true
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Vary", "Origin")
				break
			}
		}

		if allowed {
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers",
			"Origin, X-Requested-With, Content-Type, Accept, Authorization, X-API-VERSION, X-PERSONAL-TOKEN, X-ORG-ID")
		w.Header().Set("Access-Control-Max-Age", "86400")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func SetAdminMiddlewareAuthentication(che cache.ICache, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := auth.TokenValid(r)
		if err != nil {
			responses.ERROR(w, http.StatusUnauthorized, err)
			return
		}
		userId, _, err := auth.ExtractTokenID(r)
		if err != nil {
			responses.ERROR(w, http.StatusBadRequest, err)
			return
		}
		key := fmt.Sprintf("admin-auth-middleware-%d", userId)
		var value interface{}
		if ok := che.Get(key, &value); ok {
			next(w, r)
			return
		}
		userGotten, err := UserI.FindUserByID(DB, userId)
		if err != nil {
			responses.ERROR(w, http.StatusUnauthorized, err)
			return
		}
		if !userGotten.IsAdmin {
			responses.ERROR(w, http.StatusUnauthorized, errors.New("you are not authorized to use this api"))
			return
		}
		che.Set(key, userId)
		next(w, r)
	}
}

func CheckRoles(modelType string, modelId int64, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := auth.TokenValid(r)
		if err != nil {
			responses.ERROR(w, http.StatusUnauthorized, err)
			return
		}
		userId, _, err := auth.ExtractTokenID(r)
		if err != nil {
			responses.ERROR(w, http.StatusBadRequest, err)
			return
		}
		usr := models.NewUser()
		userGotten, err := usr.FindUserByID(DB, userId)
		if err != nil {
			responses.ERROR(w, http.StatusUnauthorized, err)
			return
		}
		if !userGotten.IsAdmin {
			responses.ERROR(w, http.StatusUnauthorized, errors.New("you are not authorized to use this api"))
			return
		}
		next(w, r)
	}
}

// SetMiddlewareAuthentication routes to appropriate authentication based on MODE env variable
func SetMiddlewareAuthentication(che cache.ICache, next http.HandlerFunc) http.HandlerFunc {
	mode := strings.ToLower(os.Getenv("MODE"))

	if mode == "legacy" {
		return SetMiddlewareAuthenticationLegacy(che, next)
	}
	return setMiddlewareAuthenticationCurrent(che, next)
}
func isSwaggerPath(path string) bool {
	return strings.HasPrefix(path, "/swagger") || strings.HasPrefix(path, "/docs")
}

// setMiddlewareAuthenticationCurrent handles the current authentication logic
func setMiddlewareAuthenticationCurrent(che cache.ICache, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// ✅ Allow Swagger UI: skip auth for OPTIONS requests or Swagger docs paths
		if r.Method == http.MethodOptions || isSwaggerPath(r.URL.Path) {
			next(w, r)
			return
		}
		referer := r.Header.Get("Referer")
		if strings.Contains(referer, "/swagger") {
			// Likely coming from Swagger UI page
			next(w, r)
			return
		}
		if r.Header.Get("X-PERSONAL-TOKEN") != "" {
			token, err := TokenI.GetToken(DB, r.Header.Get("X-PERSONAL-TOKEN"))
			if err != nil {
				responses.ERROR(w, http.StatusUnauthorized, errors.New("invalid token"))
				return
			}
			if token.ExpiryDate != nil {
				expiry := token.ExpiryDate.Add(time.Duration(time.Hour * 24))
				if expiry.Before(time.Now()) {
					responses.ERROR(w, http.StatusUnauthorized, errors.New("invalid token"))
					return
				}
			}
			key := fmt.Sprintf("user-auth-middleware-personal-%d", token.UserId)
			var value interface{}
			if ok := che.Get(key, &value); ok {
				r.Header.Set("X-USER-ID", fmt.Sprint(token.UserId))
				next(w, r)
				return
			}
			_, err = UserI.FindUserByID(DB, uint(token.UserId))
			if err != nil {
				responses.ERROR(w, http.StatusUnauthorized, errors.New("invalid user token"))
				return
			}
			// if !userGotten.Active {
			// 	responses.ERROR(w, http.StatusUnauthorized, errors.New("user is not active, please contact support team"))
			// 	return
			// }
			r.Header.Set("X-USER-ID", fmt.Sprint(token.UserId))
			che.Set(key, token.UserId)
			next(w, r)
			return
		}

		headerValue := r.Header.Get("X-Email")
		if headerValue == "" {
			http.Error(w, "User email header missing", http.StatusUnauthorized)
			return
		}

		userGotten, err := UserI.FindUserByEmail(DB, headerValue)
		if err != nil {
			responses.ERROR(w, http.StatusNotFound, errors.New("invalid user token"))
			return
		}
		// if !userGotten.Active {
		// 	responses.ERROR(w, http.StatusNotFound, errors.New("user is not active, please contact support team"))
		// 	return
		// }

		key := fmt.Sprintf("user-auth-middleware-%d", userGotten.ID)
		var value interface{}
		if ok := che.Get(key, &value); ok {
			next(w, r)
			return
		}

		sessionId, err := auth.ExtractSessionID(r)
		if err != nil {
			responses.ERROR(w, http.StatusNotFound, errors.New("invalid session"))
			return
		}
		_, err = UserI.GetSessionById(DB, sessionId)
		if err != nil {
			responses.ERROR(w, http.StatusNotFound, errors.New("invalid session"))
			return
		}
		che.Set(key, userGotten.ID)
		next(w, r)
	}
}

// setMiddlewareAuthenticationLegacy handles the legacy authentication logic
func SetMiddlewareAuthenticationLegacy(che cache.ICache, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-PERSONAL-TOKEN") != "" {
			token, err := TokenI.GetToken(DB, r.Header.Get("X-PERSONAL-TOKEN"))
			if err != nil {
				responses.ERROR(w, http.StatusUnauthorized, errors.New("invalid token"))
				return
			}
			if token.ExpiryDate != nil {
				expiry := token.ExpiryDate.Add(time.Duration(time.Hour * 24))
				if expiry.Before(time.Now()) {
					responses.ERROR(w, http.StatusUnauthorized, errors.New("invalid token"))
					return
				}
			}
			key := fmt.Sprintf("user-auth-middleware-personal-%d", token.UserId)
			var value interface{}
			if ok := che.Get(key, &value); ok {
				r.Header.Set("X-USER-ID", fmt.Sprint(token.UserId))
				next(w, r)
				return
			}
			userGotten, err := UserI.FindUserByID(DB, uint(token.UserId))
			if err != nil {
				responses.ERROR(w, http.StatusUnauthorized, errors.New("invalid user token"))
				return
			}
			if !userGotten.Active {
				responses.ERROR(w, http.StatusUnauthorized, errors.New("user is not active, please contact support team"))
				return
			}
			r.Header.Set("X-USER-ID", fmt.Sprint(token.UserId))
			che.Set(key, token.UserId)
			next(w, r)
			return
		}
		err := auth.TokenValid(r)
		if err != nil {
			responses.ERROR(w, http.StatusUnauthorized, errors.New("invalid token"))
			return
		}
		userId, _, err := auth.ExtractTokenID(r)
		if err != nil {
			responses.ERROR(w, http.StatusUnauthorized, errors.New("invalid token"))
			return
		}
		key := fmt.Sprintf("user-auth-middleware-%d", userId)
		var value interface{}
		if ok := che.Get(key, &value); ok {
			next(w, r)
			return
		}
		userGotten, err := UserI.FindUserByID(DB, userId)
		if err != nil {
			responses.ERROR(w, http.StatusUnauthorized, errors.New("invalid user token"))
			return
		}
		if !userGotten.Active {
			responses.ERROR(w, http.StatusUnauthorized, errors.New("user is not active, please contact support team"))
			return
		}
		sessionId, err := auth.ExtractSessionID(r)
		if err != nil {
			responses.ERROR(w, http.StatusUnauthorized, errors.New("invalid session"))
			return
		}
		_, err = UserI.GetSessionById(DB, sessionId)
		if err != nil {
			responses.ERROR(w, http.StatusUnauthorized, errors.New("invalid session"))
			return
		}
		che.Set(key, userId)
		next(w, r)
	}
}

func SetServerMiddlewareAuthentication(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		keys := r.URL.Query()
		token := keys.Get("token")
		if token == "" || token != os.Getenv("API_SECRET") {
			responses.ERROR(w, http.StatusUnauthorized, errors.New("invalid token unable to connnect to server"))
			return
		}
		next(w, r)
	}
}

func WithLogging(h http.Handler) http.Handler {
	logFn := func(rw http.ResponseWriter, r *http.Request) {
		start := time.Now()
		uri := r.RequestURI
		method := r.Method
		h.ServeHTTP(rw, r) // serve the original request
		duration := time.Since(start)
		uid, _, _ := auth.ExtractTokenID(r)
		// log request details
		log.Infof("%d :: [%s] %s :: %v", uid, method, uri, duration)
	}
	return http.HandlerFunc(logFn)
}
