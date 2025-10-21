package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"gorm.io/gorm"

	"github.com/lockb0x-llc/relayforge/internal/models"
)

type AuthService struct {
	githubConfig *oauth2.Config
	jwtSecret    []byte
}

type GitHubUser struct {
	ID        int64  `json:"id"`
	Login     string `json:"login"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
}

type Claims struct {
	UserID uint `json:"user_id"`
	jwt.RegisteredClaims
}

func NewAuthService(clientID, clientSecret, jwtSecret string) *AuthService {
	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scopes:       []string{"user:email"},
		Endpoint:     github.Endpoint,
		RedirectURL:  "http://localhost:8080/api/auth/callback",
	}

	return &AuthService{
		githubConfig: config,
		jwtSecret:    []byte(jwtSecret),
	}
}

func (a *AuthService) GetGitHubAuthURL() string {
	return a.githubConfig.AuthCodeURL("state", oauth2.AccessTypeOffline)
}

func (a *AuthService) HandleGitHubCallback(code string, db *gorm.DB) (*models.User, string, error) {
	token, err := a.githubConfig.Exchange(context.Background(), code)
	if err != nil {
		return nil, "", fmt.Errorf("failed to exchange code: %v", err)
	}

	client := a.githubConfig.Client(context.Background(), token)
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		return nil, "", fmt.Errorf("failed to get user info: %v", err)
	}
	defer resp.Body.Close()

	var githubUser GitHubUser
	if err := json.NewDecoder(resp.Body).Decode(&githubUser); err != nil {
		return nil, "", fmt.Errorf("failed to decode user info: %v", err)
	}

	// Get user email if not public
	if githubUser.Email == "" {
		emailResp, err := client.Get("https://api.github.com/user/emails")
		if err == nil {
			defer emailResp.Body.Close()
			var emails []struct {
				Email   string `json:"email"`
				Primary bool   `json:"primary"`
			}
			if json.NewDecoder(emailResp.Body).Decode(&emails) == nil {
				for _, email := range emails {
					if email.Primary {
						githubUser.Email = email.Email
						break
					}
				}
			}
		}
	}

	// Find or create user
	var user models.User
	result := db.Where("github_id = ?", githubUser.ID).First(&user)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			// Create new user
			user = models.User{
				GitHubID:    githubUser.ID,
				Username:    githubUser.Login,
				Email:       githubUser.Email,
				AvatarURL:   githubUser.AvatarURL,
				AccessToken: token.AccessToken,
			}
			if err := db.Create(&user).Error; err != nil {
				return nil, "", fmt.Errorf("failed to create user: %v", err)
			}
		} else {
			return nil, "", fmt.Errorf("database error: %v", result.Error)
		}
	} else {
		// Update existing user
		user.Email = githubUser.Email
		user.AvatarURL = githubUser.AvatarURL
		user.AccessToken = token.AccessToken
		if err := db.Save(&user).Error; err != nil {
			return nil, "", fmt.Errorf("failed to update user: %v", err)
		}
	}

	// Generate JWT token
	jwtToken, err := a.GenerateToken(user.ID)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate token: %v", err)
	}

	return &user, jwtToken, nil
}

func (a *AuthService) GenerateToken(userID uint) (string, error) {
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(a.jwtSecret)
}

func (a *AuthService) ValidateToken(tokenString string, db *gorm.DB) (*models.User, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return a.jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		var user models.User
		if err := db.First(&user, claims.UserID).Error; err != nil {
			return nil, err
		}
		return &user, nil
	}

	return nil, fmt.Errorf("invalid token")
}