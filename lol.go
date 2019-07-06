package lol

//go:generate go run go-lol-generator/main.go

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"golang.org/x/net/context"
	"golang.org/x/net/context/ctxhttp"
)

var (
	// ErrInvalidArguement is returned if invalid argument was passed.
	ErrInvalidArguement = errors.New("Invalid argument")
	// ErrNotSupportedRegion is returned if operation is not supported in a region.
	ErrNotSupportedRegion = errors.New("This operation does not work for such region")

	// ErrAPIKeyRequired is returned if riot api server returns HTTP 401.
	ErrAPIKeyRequired error = RiotError{Status: 401}
	// ErrAPILimitExceeded is returned if riot api server returns HTTP 429 Too Many Requests.
	ErrAPILimitExceeded error = RiotError{Status: 429}
	// ErrServiceUnavailable is returned if riot api server returns HTTP 503 Service unavailable.
	ErrServiceUnavailable error = RiotError{Status: 503}
)

// ClientProviderFunc is used to get a http client.
// This must NOT return nil.
type ClientProviderFunc func(context.Context) *http.Client

// Client is a league of legend api fetcher.
type Client struct {
	getClient ClientProviderFunc
	apiKey    string
}

// New creates a new league of legends client.
func New(clientProvider ClientProviderFunc, key string) (*Client, error) {
	if len(key) == 0 {
		return nil, ErrInvalidArguement
	}
	if clientProvider == nil {
		clientProvider = DefaultClientProvider
	}

	return &Client{getClient: clientProvider, apiKey: key}, nil
}

func (c *Client) doRequest(ctx context.Context, method, urlStr string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, urlStr, body)
	if err != nil {
		return nil, err
	}

	httpClient := c.getClient(ctx)

	if ctx != nil {
		return ctxhttp.Do(ctx, httpClient, req)
	}
	return httpClient.Do(req)
}

// RiotError represents an error returned from riot api server.
//
// Predeclared errors:
//	ErrAPIKeyRequired - HTTP 401 Unauthorized
type RiotError struct {
	Status int
	// This is provided for debugging.
	Body string
}

func (e RiotError) Error() string {
	return fmt.Sprintf("Riot api returned HTTP %d\nBody: %s", e.Status, e.Body)
}

// verifyAPIResponse returns nil if no error found.
func verifyAPIResponse(resp *http.Response) error {
	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		return nil
	}

	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return ErrAPIKeyRequired
	case 429:
		return ErrAPILimitExceeded
	case http.StatusServiceUnavailable:
		return ErrServiceUnavailable
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return RiotError{Status: resp.StatusCode, Body: string(data)}
}
