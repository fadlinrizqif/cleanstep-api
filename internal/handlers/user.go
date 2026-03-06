package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/fadlinrizqif/cleanstep-api/internal/app"
	"github.com/fadlinrizqif/cleanstep-api/internal/auth"
	"github.com/fadlinrizqif/cleanstep-api/internal/database"
	"github.com/fadlinrizqif/cleanstep-api/internal/dto"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type GoogleUserInfo struct {
	ID            string `json:"sub"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"email_verified"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}

type UserHandler struct {
	App *app.App
}

func NewUserHandler(app *app.App) *UserHandler {
	return &UserHandler{App: app}
}

func (h *UserHandler) CreateUser(c *gin.Context) {
	type params struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var userDetail params

	if err := c.ShouldBindJSON(&userDetail); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hashedPassword, err := auth.HashPassword(userDetail.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	users, err := h.App.DBqueries.CreateUser(c.Request.Context(), database.CreateUserParams{
		Name:           userDetail.Name,
		Email:          userDetail.Email,
		HashedPassword: hashedPassword,
		Provider:       "manual",
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, dto.CreateUser{
		ID:        users.ID,
		CreatedAt: users.CreatedAt,
		UpdatedAt: users.UpdatedAt,
		Name:      users.Name,
		Email:     users.Email,
	})

}

func (h *UserHandler) LoginUser(c *gin.Context) {
	type params struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var LoginDetail params

	if err := c.ShouldBindJSON(&LoginDetail); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	getUser, err := h.App.DBqueries.GetUser(c.Request.Context(), LoginDetail.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	isCorrect, _ := auth.CheckPassword(LoginDetail.Password, getUser.HashedPassword)
	if !isCorrect {
		c.JSON(http.StatusForbidden, gin.H{"error": "wrong password"})
		return
	}

	jwtDuration := time.Duration(60) * time.Minute
	newJwt, err := auth.MakeJWT(getUser.ID, h.App.SeverSecret, jwtDuration)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	getRefreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	_, err = h.App.DBqueries.CreateRefreshToken(c.Request.Context(), database.CreateRefreshTokenParams{
		Token:     getRefreshToken,
		UserID:    getUser.ID,
		ExpiresAt: time.Now().AddDate(0, 0, 60),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	header := c.Request.Header
	getForwardedProto := header.Get("X-Forwarded-Proto")
	secure := c.Request.TLS != nil || getForwardedProto == "https"
	c.SetCookieData(&http.Cookie{
		Name:     "access_token",
		Value:    newJwt,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   60 * 60,
	})

	c.SetCookieData(&http.Cookie{
		Name:     "refresh_token",
		Value:    getRefreshToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   60 * 60 * 24 * 60,
	})

	c.JSON(http.StatusOK, gin.H{
		"body": "login successfully",
	})

	//c.JSON(http.StatusOK, dto.LoginUser{
	//	Token:        newJwt,
	//	RefreshToken: getRefreshToken,
	//	User: dto.UserDetail{
	//		ID:    getUser.ID,
	//		Name:  getUser.Name,
	//		Email: getUser.Email,
	//	},
	//})
}

func (h *UserHandler) LogoutUser(c *gin.Context) {
	refreshTOken, err := c.Cookie("refresh_token")
	if err == nil {
		_ = h.App.DBqueries.RevokeRefreshToken(c.Request.Context(), refreshTOken)
	}

	secure := c.Request.TLS != nil

	c.SetCookieData(&http.Cookie{
		Name:     "access_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})

	c.SetCookieData(&http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})

	c.JSON(http.StatusOK, gin.H{
		"body": "logout successfully",
	})
}

func (h *UserHandler) OauthLogin(c *gin.Context) {

	conf := &oauth2.Config{
		ClientID:     h.App.GoogleID,
		ClientSecret: h.App.GoogleSecret,
		RedirectURL:  h.App.RedirectURL,
		Scopes: []string{
			"openid",
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}

	verifier := oauth2.GenerateVerifier()

	//c.SetCookie("oauth_verifier", verifier, 300, "/", "", false, true)
	c.SetCookieData(&http.Cookie{
		Name:     "oauth_verifier",
		Value:    verifier,
		MaxAge:   300,
		Path:     "/",
		Secure:   false,
		HttpOnly: true,
	})

	url := conf.AuthCodeURL("state", oauth2.AccessTypeOffline, oauth2.S256ChallengeOption(verifier))

	c.Redirect(http.StatusTemporaryRedirect, url)

}

func (h *UserHandler) OauthCallback(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No code found"})
		return
	}

	verifier, err := c.Cookie("oauth_verifier")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	conf := &oauth2.Config{
		ClientID:     h.App.GoogleID,
		ClientSecret: h.App.GoogleSecret,
		RedirectURL:  h.App.RedirectURL,
		Scopes: []string{
			"openid",
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}

	token, err := conf.Exchange(c.Request.Context(), code, oauth2.VerifierOption(verifier))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	client := conf.Client(c.Request.Context(), token)
	res, err := client.Get("https://openidconnect.googleapis.com/v1/userinfo")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	defer res.Body.Close()

	var userInfo GoogleUserInfo

	if err := json.NewDecoder(res.Body).Decode(&userInfo); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	getUser, err := h.App.DBqueries.GetUser(c.Request.Context(), userInfo.Email)
	if err == sql.ErrNoRows {
		getUser, err = h.App.DBqueries.CreateUser(c.Request.Context(), database.CreateUserParams{
			Name:     userInfo.Name,
			Email:    userInfo.Email,
			Provider: "google",
		})

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	jwtDuration := time.Duration(60) * time.Minute
	newJwt, err := auth.MakeJWT(getUser.ID, h.App.SeverSecret, jwtDuration)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	newRefreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	_, err = h.App.DBqueries.CreateRefreshToken(c.Request.Context(), database.CreateRefreshTokenParams{
		Token:     newRefreshToken,
		UserID:    getUser.ID,
		ExpiresAt: time.Now().AddDate(0, 0, 60),
	})

	secure := c.Request.TLS != nil

	c.SetCookie("oauth_verifier", "", -1, "/", "localhost", false, true)

	c.SetCookieData(&http.Cookie{
		Name:     "access_token",
		Value:    newJwt,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   60 * 60,
	})

	c.SetCookieData(&http.Cookie{
		Name:     "refresh_token",
		Value:    newRefreshToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   60 * 60 * 24 * 60,
	})

	c.JSON(http.StatusOK, gin.H{"body": "login successfully"})

}
