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
	if !(int(time.Since(t.ExpiresAt)) >= 0) {
		return false
	}
	return true
}

// Reads a cached token
func GetCachedToken(oauth SpotifyOauth) (Token, error) {
	var token Token
	var err error
	if oauth.CachePath != "" {
		cachedData, err := ioutil.ReadFile(oauth.CachePath)
		if err == nil {
			err = json.Unmarshal(cachedData, &token)
			if token.isExpired() {
				token, err = RefreshAccessToken(token.RefreshToken, oauth)
			}
			if err == nil {
				return token, nil
			}
		}
	}
	return Token{}, err
}

// Takes a SpotifyOauth struct and returns the appropriate AuthorizeUrl for
// your request
func GetAuthorizeURL(oauth SpotifyOauth) (string, error) {
	var Url *url.URL
	Url, err := url.Parse(OauthAuthorizeURL)
	if err == nil {
		parameters := url.Values{}
		parameters.Add("client_id", oauth.ClientId)
		parameters.Add("response_type", "code")
		parameters.Add("redirect_uri", oauth.RedirectUri)
		if oauth.Scope != "" {
			parameters.Add("scope", oauth.Scope)
		}
		if oauth.State != "" {
			parameters.Add("state", oauth.State)
		}
		Url.RawQuery = parameters.Encode()
		return Url.String(), nil
	}
	return "", err
}

// Takes the authorization code and a SpotifyOauth
// Returns an access token
func GetAccessToken(code string, oauth SpotifyOauth) (Token, error) {
	var err error
	parameters := url.Values{}
	parameters.Add("redirect_uri", oauth.RedirectUri)
	parameters.Add("code", code)
	parameters.Add("grant_type", "authorization_code")

	token, err := sendAccessTokenRequest(parameters, oauth)
	if err == nil {
		err = saveTokenInfo(token, oauth)
		if err == nil {
			return token, nil
		}
	}
	return Token{}, err
}

// Refreshes an expired AccessToken
func RefreshAccessToken(refreshToken string, oauth SpotifyOauth) (Token, error) {
	parameters := url.Values{}
	parameters.Add("refresh_token", refreshToken)
	parameters.Add("grant_type", "refresh_token")

	token, err := sendAccessTokenRequest(parameters, oauth)
	if err == nil {
		if token.RefreshToken == "" {
			token.RefreshToken = refreshToken
		}
		err = saveTokenInfo(token, oauth)
		if err == nil {
			return token, nil
		}
	}
	return Token{}, err
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

func sendAccessTokenRequest(parameters url.Values, oauth SpotifyOauth) (Token, error) {
	client := &http.Client{}
	req, err := http.NewRequest("POST", OauthTokenUrl, strings.NewReader(parameters.Encode()))
	if err == nil {
		req.SetBasicAuth(oauth.ClientId, oauth.ClientSecret)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		resp, err := client.Do(req)
		defer resp.Body.Close()
		if err == nil {
			if resp.StatusCode == http.StatusOK {
				var token Token
				err = json.NewDecoder(resp.Body).Decode(&token)
				if err == nil {
					token.ExpiresAt = time.Now().Add(time.Duration(token.TTL) * time.Second)
					return token, nil
				}
			} else {
				err = errors.New(resp.Status)
			}
		}
	}
	return Token{}, err
}

func saveTokenInfo(token Token, oauth SpotifyOauth) error {
	var err error
	if oauth.CachePath != "" {
		marshaledToken, err := json.Marshal(token)
		if err == nil {
			err = ioutil.WriteFile(oauth.CachePath, marshaledToken, 0x777)
			return nil
		}
	}
	return err
}
