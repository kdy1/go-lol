// +build appengine

package lol

import (
	"net/http"

	"golang.org/x/net/context"
	"google.golang.org/appengine/urlfetch"
)

var (
	// DefaultClientProvider is a ClientProviderFunc which use urlfetch.
	DefaultClientProvider ClientProviderFunc = func(ctx context.Context) *http.Client {
		return urlfetch.Client(ctx)
	}
)
