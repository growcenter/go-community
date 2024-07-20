package google

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go-community/internal/config"
	"io/ioutil"

	"golang.org/x/oauth2"
)

type Google interface {
	Redirect() string
	Fetch(state, code string) (*GoogleUser, error)
}

type GoogleAuth struct {
	authState    string
	clientId     string
	clientSecret string
	redirectUrl  string
}

var oauth = &oauth2.Config{}

func NewGoogle(config *config.Configuration) (*GoogleAuth, error) {
	if config.Google.State == "" {
		return nil, errors.New("empty signing key")
	}

	return &GoogleAuth{authState: config.Google.State, clientId: config.Google.ClientID, clientSecret: config.Google.ClientSecret, redirectUrl: config.Google.Redirect}, nil
}

func (g *GoogleAuth) Redirect() string {
	oauth = &oauth2.Config{
		ClientID:     g.clientId,
		ClientSecret: g.clientSecret,
		RedirectURL:  g.redirectUrl,
		Scopes:       []string{"profile", "email"},
	}

	url := oauth.AuthCodeURL(g.authState)
	return url
}

func (g *GoogleAuth) Fetch(state, code string) (*GoogleUser, error) {
	if state != g.authState {
		return nil, errors.New("state are not the same")
	}

	token, err := oauth.Exchange(context.Background(), code)
	if err != nil {
		return nil, err
	}

	client := oauth.Client(context.Background(), token)
	response, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	byteData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	var data map[string]interface{}
	err = json.Unmarshal(byteData, &data)
	if err != nil {
		return nil, err
	}

	user := &GoogleUser{
		Email: data["email"].(string),
		Name:  data["name"].(string),
	}

	fmt.Printf("email: %s\nname: %s", user.Email, user.Name)

	return user, nil
}
