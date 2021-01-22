package functions

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"log"
	"net/http"
	"os"
)

const (
	clientIdEnv     = "MS_CLIENT_ID"
	clientSecretEnv = "MS_CLIENT_SECRET"
	authRedirectEnv = "MS_REDIRECT"
)

var (
	clientId                  string
	clientSecret              string
	authRedirectUrl           string
	microsoftConsumerEndpoint oauth2.Endpoint
	statesList                []string
	msConfig                  *oauth2.Config
)

func init() {
	clientId = os.Getenv(clientIdEnv)
	if clientId == "" {
		log.Fatalf("%q environment variable must be set!", clientIdEnv)
	}

	clientSecret = os.Getenv(clientSecretEnv)
	if clientSecret == "" {
		log.Fatalf("%q environment variable must be set!", clientSecretEnv)
	}

	authRedirectUrl = os.Getenv(authRedirectEnv)
	if authRedirectUrl == "" {
		log.Fatalf("%q environment variable must be set!", authRedirectEnv)
	}

	microsoftConsumerEndpoint.AuthStyle = oauth2.AuthStyleInHeader
	microsoftConsumerEndpoint.AuthURL = "https://login.microsoftonline.com/consumers/oauth2/v2.0/authorize"
	microsoftConsumerEndpoint.TokenURL = "https://login.microsoftonline.com/consumers/oauth2/v2.0/token"

	msConfig = &oauth2.Config{
		RedirectURL:  authRedirectUrl,
		ClientID:     clientId,
		ClientSecret: clientSecret,
		Scopes:       []string{"offline_access tasks.readwrite"},
		Endpoint:     microsoftConsumerEndpoint,
	}
}

func Login(w http.ResponseWriter, r *http.Request) {
	newState := uuid.New().String()

	statesList = append(statesList, newState)
	url := msConfig.AuthCodeURL(newState)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func LoginRedirect(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	token, err := getAccessToken(query.Get("state"), query.Get("code"))
	if err != nil {
		fmt.Println(err.Error())
		http.Redirect(w, r, "/", http.StatusInternalServerError)
	}

	_, _ = fmt.Fprintf(w, token.AccessToken)
}

func getAccessToken(state string, code string) (*oauth2.Token, error) {
	found := false

	for _, savedState := range statesList {
		if state == savedState {
			found = true
			break
		}
	}

	if !found {
		return nil, fmt.Errorf("invalid oauth states, try logging in again")
	}

	token, err := msConfig.Exchange(context.Background(), code)
	if err != nil {
		return nil, fmt.Errorf("code exchange failed: %s", err.Error())
	}

	return token, nil
}
