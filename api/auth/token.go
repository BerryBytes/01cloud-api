package auth

import (
	"01cloud-api/api/models"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

func CreateToken(userId uint, organizationId, sessionId uint) (string, error) {
	claims := jwt.MapClaims{}
	claims["authorized"] = true
	claims["user_id"] = userId
	claims["session_id"] = sessionId
	claims["organization_id"] = organizationId
	claims["exp"] = time.Now().Add(time.Hour * 24 * 7).Unix() //Token expires after 1 hour
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(os.Getenv("API_SECRET")))
}

func GetSession(userAgent string, uid uint, action string) models.Session {
	os := parseUserAgentOS(userAgent)
	browser := parseUserAgentBrowser(userAgent)
	return models.Session{
		UserId:  uid,
		OS:      os,
		Browser: browser,
		Action:  action,
	}
}

func parseUserAgentOS(userAgent string) string {
	if contains(userAgent, "Windows") {
		return "Windows"
	} else if contains(userAgent, "Macintosh") || contains(userAgent, "Mac OS X") {
		return "macOS"
	} else if contains(userAgent, "Linux") {
		return "Linux"
	} else if contains(userAgent, "Android") {
		return "Android"
	} else if contains(userAgent, "iPhone") {
		return "iPhone"
	}

	return "Unknown"
}

func parseUserAgentBrowser(userAgent string) string {
	if contains(userAgent, "Firefox") {
		return "Firefox"
	} else if contains(userAgent, "Chrome") {
		return "Chrome"
	} else if contains(userAgent, "Safari") {
		return "Safari"
	} else if contains(userAgent, "OPR") {
		return "Opera"
	} else if contains(userAgent, "Edg") {
		return "Microsoft Edge"
	} else if contains(userAgent, "cli") {
		return "01cloud-cli"
	}

	return "Unknown"
}

func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
func TokenValid(r *http.Request) error {
	tokenString := ExtractToken(r)
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("API_SECRET")), nil
	})
	if err != nil {
		return err
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		Pretty(claims)
	}
	return nil
}

func ExtractToken(r *http.Request) string {
	keys := r.URL.Query()
	token := keys.Get("token")
	if token != "" {
		return token
	}
	bearerToken := r.Header.Get("Authorization")
	if len(strings.Split(bearerToken, " ")) == 2 {
		return strings.Split(bearerToken, " ")[1]
	}
	return ""
}

func ExtractIdFromToken(tokenString string) (uint, uint, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("API_SECRET")), nil
	})
	if err != nil {
		return 0, 0, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return 0, 0, fmt.Errorf("invalid token claims")
	}

	// Extract user_id
	uidFloat, ok := claims["user_id"].(float64)
	if !ok {
		return 0, 0, fmt.Errorf("user_id not found or invalid")
	}
	uid := uint(uidFloat)

	// Extract organization_id
	oidFloat, ok := claims["organization_id"].(float64)
	if !ok {
		return uid, 0, fmt.Errorf("organization_id not found or invalid")
	}
	oid := uint(oidFloat)

	return uid, oid, nil
}

func ExtractSessionIdFromToken(tokenString string) (uint, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("API_SECRET")), nil
	})
	if err != nil {
		return 0, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		sid, err := strconv.ParseUint(fmt.Sprintf("%.0f", claims["session_id"]), 10, 64)
		if err != nil {
			return 0, err
		}
		return uint(sid), nil
	}
	return 0, nil
}

func ExtractTokenID(r *http.Request) (uint, uint, error) {
	if r.Header.Get("X-PERSONAL-TOKEN") != "" {
		return ExtractPersonalTokenID(r)
	}
	tokenString := ExtractToken(r)
	return ExtractIdFromToken(tokenString)
}

func ExtractSessionID(r *http.Request) (uint, error) {
	tokenString := ExtractToken(r)
	return ExtractSessionIdFromToken(tokenString)
}

func Pretty(data interface{}) {
	_, err := json.MarshalIndent(data, "", " ")
	if err != nil {
		log.Println(err)
		return
	}
}

func ExtractPersonalTokenID(r *http.Request) (uint, uint, error) {
	userID, _ := strconv.ParseUint(r.Header.Get("X-USER-ID"), 10, 64)
	if r.Header.Get("X-ORG-ID") != "" {
		orgID, _ := strconv.ParseUint(r.Header.Get("X-ORG-ID"), 10, 64)
		return uint(userID), uint(orgID), nil
	}
	return uint(userID), 0, nil
}

func GeneratePassword() string {
	b := make([]byte, 10)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return base64.StdEncoding.EncodeToString(b)
}
