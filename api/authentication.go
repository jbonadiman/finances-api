package handler

import (
	"context"
	"fmt"
	"golang.org/x/oauth2"
	"log"
	"net/http"
	"os"
)

func LoginRedirect(w http.ResponseWriter, r *http.Request) {
	microsoftConsumerEndpoint := oauth2.Endpoint{}

	clientId := os.Getenv("MS_CLIENT_ID")
	if clientId == "" {
		log.Fatalf("%q environment variable must be set!", "MS_CLIENT_ID")
	}

	clientSecret := os.Getenv("MS_CLIENT_SECRET")
	if clientSecret == "" {
		log.Fatalf("%q environment variable must be set!", "MS_CLIENT_SECRET")
	}

	authRedirectUrl := os.Getenv("MS_REDIRECT")
	if authRedirectUrl == "" {
		log.Fatalf("%q environment variable must be set!", "MS_REDIRECT")
	}

	microsoftConsumerEndpoint.AuthStyle = oauth2.AuthStyleInHeader
	microsoftConsumerEndpoint.AuthURL = "https://login.microsoftonline.com/consumers/oauth2/v2.0/authorize"
	microsoftConsumerEndpoint.TokenURL = "https://login.microsoftonline.com/consumers/oauth2/v2.0/token"

	msConfig := &oauth2.Config{
		RedirectURL:  authRedirectUrl,
		ClientID:     clientId,
		ClientSecret: clientSecret,
		Scopes:       []string{"offline_access tasks.readwrite"},
		Endpoint:     microsoftConsumerEndpoint,
	}

	query := r.URL.Query()

	token, err := getAccessToken(query.Get("code"), msConfig)
	if err != nil {
		fmt.Println(err.Error())
		http.Redirect(w, r, "/", http.StatusInternalServerError)
	}

	_, _ = fmt.Fprintf(w, token.AccessToken)
}

func getAccessToken(code string, config *oauth2.Config) (*oauth2.Token, error) {
	token, err := config.Exchange(context.Background(), code)
	if err != nil {
		return nil, fmt.Errorf("code exchange failed: %s", err.Error())
	}

	return token, nil
}
