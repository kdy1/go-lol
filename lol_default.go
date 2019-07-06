// +build !appengine

package lol

import (
	"net/http"

	"golang.org/x/net/context"
)

var (
	// DefaultClientProvider is a simple ClientProviderFunc which returns http.DefaultClient
	DefaultClientProvider ClientProviderFunc = func(context.Context) *http.Client {
		return http.DefaultClient
	}
)
