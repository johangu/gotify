# Gotify - a Go wrapper for the Spotify API

This is my first attempt at using Go, so I'm sure there are lots of things that could have been written differently, please feel free to let me know, and to improve upon it.

## Documentation
Documentation will be available soon

## Installation
Add the following to your project:

    import (
      "github.com/johangu/gotify"
    )

## Quick start
To get an access token

    var oauth = gotify.SpotifyOauth{
      ClientId,
      ClientSecret,
      CallbackURL,
      "",
      Scope,
      "<path for saving token>"
    }
    var token gotify.Token
    
    func main() {
      url, err := gotify.GetAuthorizeURL(oauth)
      if err != nil {
        // Handle error
      }
      
      http.Handlefunc("/callback", callbackHandler)
      ... // add different handlers you may need.
      
      http.ListenAndServe(":8888", nil)
      http.Get(url)
    }
    
    func callbackHandler(w http.ResponseWriter, r *http.Request) {
      code, _ := gotify.ParseResponseCode(r.RequestURI)
      token, _ = gotify.GetAccessToken(code, oauth)
    }

## Report Issues
Feel free to report any issues [here](https://github.com/johangu/gotify/issues) or send a pull request.

## Version

 - 0.0.1 - 2014-06-21 - Initial pre-release

