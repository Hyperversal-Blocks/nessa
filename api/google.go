package api

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"os"

	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"keyless-auth/repository/user"
)

type GoogleHandler struct {
	oauthConfig *oauth2.Config
	repository  user.Repo
}

func NewGoogleHandler(r user.Repo) *GoogleHandler {
	config := &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.profile",
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/photoslibrary.readonly",
		},
		Endpoint: google.Endpoint,
	}

	return &GoogleHandler{
		oauthConfig: config,
		repository:  r,
	}
}

func generateState() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return base64.StdEncoding.EncodeToString(b)
}

func (h *GoogleHandler) HandleGoogleLogin(w http.ResponseWriter, r *http.Request) {
	oauthState := generateState()
	url := h.oauthConfig.AuthCodeURL(oauthState, oauth2.AccessTypeOffline, oauth2.ApprovalForce)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (h *GoogleHandler) HandleGoogleCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Redirect(w, r, "/auth/failure", http.StatusTemporaryRedirect)
		return
	}

	token, err := h.oauthConfig.Exchange(r.Context(), code)
	if err != nil {
		http.Redirect(w, r, "/auth/failure", http.StatusTemporaryRedirect)
		return
	}

	client := h.oauthConfig.Client(r.Context(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		http.Redirect(w, r, "/auth/failure", http.StatusTemporaryRedirect)
		return
	}
	defer resp.Body.Close()

	var userInfo struct {
		ID    string `json:"id"`
		Email string `json:"email"`
		Name  string `json:"name"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		http.Redirect(w, r, "/auth/failure", http.StatusTemporaryRedirect)
		return
	}

	userObj := &user.OAuthUser{
		ID:             uuid.New().String(),
		GoogleID:       userInfo.ID,
		Email:          userInfo.Email,
		Name:           userInfo.Name,
		ProfilePicture: "",
		AccessToken:    token.AccessToken,
		RefreshToken:   token.RefreshToken,
	}

	if err := h.repository.SaveoAuthUser(userObj); err != nil {
		http.Redirect(w, r, "/auth/failure", http.StatusTemporaryRedirect)
		return
	}

	http.Redirect(w, r, "/auth/success", http.StatusTemporaryRedirect)
}

// WithGoogle TODO: google
type WithGoogle struct {
	ID             string `json:"id"`
	GoogleID       string `json:"google_id"`
	Email          string `json:"email"`
	Name           string `json:"name"`
	ProfilePicture string `json:"profile_picture"`
	AccessToken    string `json:"-"`
	RefreshToken   string `json:"-"`
}

type WithMetamask struct {
}

// TODO metamask
type WithNFC struct{}

// TODO: NCF

// WithKeyExchange is a simple key exchange to generate a shared key.
type WithKeyExchange struct {
	PublicKey string `json:"public_key"`
	Challenge string `json:"challenge"`
}

type WithKeyExchangeResponse struct {
	HashedKey       string `json:"hashed_key"`
	HashedSignature string `json:"hashed_signature"`
}

type KeyExchangeLoginRequest struct {
	PublicKey string `json:"public_key"`
}

type KeyExchangeLoginByChallengeResponse struct {
	HashedKey string `json:"hashed_key"`
}

type KeyExchangeLoginBySignatureResponse struct {
	Signature string `json:"signature"`
}
