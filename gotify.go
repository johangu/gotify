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

func call(call string, verb string, params url.Values, token Token) (string, error) {
	verb = strings.ToUpper(verb)
	var body string
	var Url string

	if verb == "GET" {
		Url = fmt.Sprintf("%s%s?%s", APIUrl, call, params.Encode())
	} else {
		Url = fmt.Sprintf("%s%s", APIUrl, call)
	}
	client := &http.Client{}
	resp := &http.Response{}

	req, err := http.NewRequest(verb, Url, strings.NewReader(params.Encode()))
	if err == nil {
		if verb == "POST" {
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		}
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))
		resp, err = client.Do(req)
		defer resp.Body.Close()
		fmt.Println(resp.Request)
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
	params := url.Values{}
	params.Add("ids", ids)
	return call("albums", "GET", params, token)
}

// Returns tracks information from an album given an album id
func AlbumTracks(id string, token Token) (string, error) {
	id = parseId("album", id)
	return call(fmt.Sprintf("albums/%s/tracks", id), "GET", url.Values{}, token)
}

// Returns an artist
func Artist(id string, token Token) (string, error) {
	id = parseId("artist", id)
	return call(fmt.Sprintf("artists/%s", id), "GET", url.Values{}, token)
}

// Returns albums from an artist, given an artist id
func ArtistAlbums(id string, token Token) (string, error) {
	id = parseId("artist", id)
	return call(fmt.Sprintf("artists/%s/albums", id), "GET", url.Values{}, token)
}

// Returns an artists top tracks given artist id and country
func ArtistTopTracks(id string, country string, token Token) (string, error) {
	id = parseId("artist", id)
	params := url.Values{}
	params.Add("country", country)
	return call(fmt.Sprintf("artists/%s/top-tracks", id), "GET", params, token)
}

// Returns a list of artists
func Artists(artists []string, token Token) (string, error) {
	ids := buildIdString(artists, "artist")
	params := url.Values{}
	params.Add("ids", ids)
	return call("artists", "GET", params, token)
}

// Returns information about me
func Me(token Token) (string, error) {
	return call("me", "GET", url.Values{}, token)
}

// Returns a playlist
func Playlist(userId string, playlistId string, fields string, token Token) (string, error) {
	params := url.Values{}
	params.Add("fields", url.QueryEscape(fields))
	return call(fmt.Sprintf("users/%s/playlists/%s", userId, playlistId), "GET", params, token)
}

// Returns a users playlists
func Playlists(userId string, token Token) (string, error) {
	return call(fmt.Sprintf("users/%s/playlists", userId), "GET", url.Values{}, token)
}

// Returns a tracks from playlists
func Playlists(userId string, playlistId string, fields string, token Token) (string, error) {
	params := url.Values{}
	params.Add("fields", url.QueryEscape(fields))
	return call(fmt.Sprintf("users/%s/playlists/%s/tracks", userId, playlistId), "GET", params, token)
}

func Search(q string, qType string, limit int, offset int, token Token) (string, error) {
	params := url.Values{}
	params.Add("q", q)
	params.Add("type", url.QueryEscape(qType))
	params.Add("limit", fmt.Sprintf("%d", limit))
	params.Add("offset", fmt.Sprintf("%d", offset))

	return call("search", "GET", params, token)
}

// Returns a single track
func Track(id string, token Token) (string, error) {
	return call(fmt.Sprintf("tracks/%s", id), "GET", url.Values{}, token)
}

// Returns a list of tracks
func Tracks(tracks []string, token Token) (string, error) {
	ids := buildIdString(tracks, "track")
	params := url.Values{}
	params.Add("ids", ids)
	return call("tracks", "GET", params, token)
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
