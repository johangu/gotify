// gotify is a client library for the Spotify API
package gotify

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const OauthAuthorizeURL = "https://accounts.spotify.com/authorize"
const OauthTokenUrl = "https://accounts.spotify.com/api/token"

type SpotifyOauth struct {
	ClientId     string
	ClientSecret string
	RedirectUri  string
	State        string
	Scope        string
	CachePath    string
}

type Token struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresAt    time.Time
	RefreshToken string `json:"refresh_token"`
	TTL          int    `json:"expires_in"`
}

func (t Token) isExpired() bool {
	if int(time.Since(t.ExpiresAt)) >= 0 {
		return true
	}
	return false
}

// TODO: TEST
func GetCachedToken(oauth SpotifyOauth) (Token, error) {
	var token Token
	if oauth.CachePath != "" {
		cachedData, err := ioutil.ReadFile(oauth.CachePath)
		if err != nil {
			return Token{}, err
		}
		err = json.Unmarshal(cachedData, &token)
		if err != nil {
			return Token{}, err
		}

		if token.isExpired() {
			token, err = RefreshAccessToken(token.RefreshToken, oauth)
			if err != nil {
				return Token{}, err
			}
		}
	}
	return token, nil
}

// TOOO: TEST
func SaveTokenInfo(token Token, oauth SpotifyOauth) error {
	if oauth.CachePath != "" {
		marshaledToken, err := json.Marshal(token)
		if err != nil {
			return err
		}
		ioutil.WriteFile(oauth.CachePath, marshaledToken, 0x777)
	}
	return nil
}

// Takes a SpotifyOauth struct and returns the appropriate AuthorizeUrl for
// your request
func GetAuthorizeURL(oauth SpotifyOauth) (string, error) {
	var Url *url.URL
	Url, err := url.Parse(OauthAuthorizeURL)
	if err != nil {
		return "", err
	}

	parameters := url.Values{}
	parameters.Add("client_id", oauth.ClientId)
	parameters.Add("response_type", "code")
	parameters.Add("redirect_url", oauth.RedirectUri)
	if oauth.Scope != "" {
		parameters.Add("scope", oauth.Scope)
	}
	if oauth.State != "" {
		parameters.Add("state", oauth.State)
	}
	Url.RawQuery = parameters.Encode()

	return Url.String(), nil
}

// TODO: TEST
func GetAccessToken(code string, oauth SpotifyOauth) (Token, error) {
	parameters := url.Values{}
	parameters.Add("redirect_uri", oauth.RedirectUri)
	parameters.Add("code", code)
	parameters.Add("grant_type", "authorization_code")
	if oauth.Scope != "" {
		parameters.Add("scope", oauth.Scope)
	}
	if oauth.State != "" {
		parameters.Add("state", oauth.State)
	}

	token, err := sendAccessTokenRequest(parameters, oauth)
	if err != nil {
		return Token{}, err
	}
	err = SaveTokenInfo(token, oauth)
	if err != nil {
		return Token{}, err
	}
	return token, nil
}

// Refreshes an expired AccessToken
// TODO: TEST
func RefreshAccessToken(refreshToken string, oauth SpotifyOauth) (Token, error) {
	parameters := url.Values{}
	parameters.Add("refresh_token", refreshToken)
	parameters.Add("grant_type", "refresh_token")

	token, err := sendAccessTokenRequest(parameters, oauth)
	if err != nil {
		return Token{}, err
	}

	if token.RefreshToken == "" {
		token.RefreshToken = refreshToken
	}
	err = SaveTokenInfo(token, oauth)
	if err != nil {
		return Token{}, err
	}

	return token, nil
}

func sendAccessTokenRequest(parameters url.Values, oauth SpotifyOauth) (Token, error) {
	client := &http.Client{}
	req, err := http.NewRequest("POST", OauthTokenUrl, strings.NewReader(parameters.Encode()))
	if err != nil {
		return Token{}, err
	}
	req.SetBasicAuth(oauth.ClientId, oauth.ClientSecret)
	resp, err := client.Do(req)
	if err != nil {
		return Token{}, err
	}
	if resp.StatusCode != http.StatusOK {
		return Token{}, errors.New(resp.Status)
	}
	var token Token
	err = json.NewDecoder(resp.Body).Decode(token)
	if err != nil {
		return Token{}, err
	}
	token.ExpiresAt = time.Now().Add(time.Duration(token.TTL) * time.Second)

	return token, nil
}

// Parses the response code from from the query string when user is redirected
// back to the application
func ParseResponseCode(response string) (string, error) {
	u, err := url.Parse(response)
	if err != nil {
		return "", err
	}
	q, _ := url.ParseQuery(u.RawQuery)
	code := q["code"][0]
	return code, nil
}
