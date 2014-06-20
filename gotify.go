// gotify is a client library for the Spotify API
package gotify

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
)

const Version = 1

var APIUrl = fmt.Sprintf("https://api.spotify.com/v%d/", Version)

type SpotifyError struct {
	HttpStatus int
	Code       int
	Message    string
}

func (e SpotifyError) Error() string {
	return fmt.Sprintf("%d, code: %d - %s", e.HttpStatus, e.Code, e.Message)
}

func call(call string, verb string, parameters url.Values, token Token) (string, error) {
	Url := fmt.Sprintf("%s%s", APIUrl, call)
	client := &http.Client{}
	resp := &http.Response{}
	var body string
	req, err := http.NewRequest(strings.ToUpper(verb), Url, strings.NewReader(parameters.Encode()))
	if err == nil {
		if strings.ToUpper(verb) == "POST" {
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		}
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))
		resp, err = client.Do(req)
		defer resp.Body.Close()
		if err == nil {
			if resp.StatusCode == http.StatusOK {
				tmp, err := ioutil.ReadAll(resp.Body)
				if err == nil {
					body = string(tmp)
					return body, nil
				}
			} else {
				err = SpotifyError{
					resp.StatusCode,
					-1,
					"Requested resource could not be found"}
			}
		}
	}
	return "", err
}

// Returns an album
func Album(id string, token Token) (string, error) {
	id = parseId("album", id)
	return call(fmt.Sprintf("albums/%s", id), "GET", url.Values{}, token)
}

// Returns a list of albums
func Albums(albums []string, token Token) (string, error) {
	ids := buildIdString(albums, "album")
	return call(fmt.Sprintf("albums?ids=%s", ids), "GET", url.Values{}, token)
}

// Returns an artist
func Artist(id string, token Token) (string, error) {
	id = parseId("artist", id)
	return call(fmt.Sprintf("artists/%s", id), "GET", url.Values{}, token)
}

// Returns a list of artists
func Artists(artists []string, token Token) (string, error) {
	ids := buildIdString(artists, "artist")
	return call(fmt.Sprintf("artists?ids=%s", ids), "GET", url.Values{}, token)
}

// Returns information about me
func Me(token Token) (string, error) {
	return call("me", "GET", url.Values{}, token)
}

// Returns a playlist
func Playlist(userId string, playlistId string, fields string, token Token) (string, error) {
	params := url.Values{}
	params.Add("fields", fields)
	return call(fmt.Sprintf("users/%s/playlists/%s", userId, playlistId), "GET", params, token)
}

// Returns a users playlists
func Playlists(userId string, token Token) (string, error) {
	return call(fmt.Sprintf("users/%s/playlists", userId), "GET", url.Values{}, token)
}

// Returns a single track
func Track(trackId string, token Token) (string, error) {
	return call(fmt.Sprintf("tracks/%s", trackId), "GET", url.Values{}, token)
}

// Returns a list of tracks
func Tracks(tracks []string, token Token) (string, error) {
	ids := buildIdString(tracks, "track")
	return call(fmt.Sprintf("tracks?ids=%s", ids), "GET", url.Values{}, token)
}

// Gets basic information about a user
func User(userId string, token Token) (string, error) {
	return call(fmt.Sprintf("users/%s", userId), "GET", url.Values{}, token)
}

func buildIdString(slice []string, idType string) string {
	ids := make([]string, len(slice))
	for i, item := range slice {
		ids[i] = parseId(idType, item)
	}
	return strings.Join(ids, ",")
}

// Parses the id from spotify uri's or http links before making api calls
func parseId(idType string, uri string) string {
	fields := strings.Split(uri, ":")
	if len(fields) == 3 {
		if fields[1] != idType {
			log.Print(fmt.Sprintf("Warning: expected if of type %s but found id of type %s", idType, fields[1]))
		}
		return fields[2]
	}
	fields = strings.Split(uri, "/")
	fmt.Println(len(fields))
	if len(fields) >= 3 {
		if idType != fields[len(fields)-2] {
			log.Print(fmt.Sprintf("Warning: expected if of type %s but found id of type %s", idType, fields[1]))
		}
		return fields[len(fields)-1]
	}
	return uri
}
