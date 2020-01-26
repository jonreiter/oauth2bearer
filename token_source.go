package oauth2bearer

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

type controlMessage struct {
	action  int
	channel chan controlMessage
	token   *oauth2.Token
}

const (
	getToken        int = 0
	refresh             = 1
	registerChannel     = 2
	sendToken           = 3
)

// TokenSource describes a place to get tokens
type TokenSource struct {
	config          clientcredentials.Config
	client          *http.Client
	ctx             context.Context
	controlChannels []chan controlMessage
	refreshChannel  chan *oauth2.Token
	params          TokenSourceParams
	currentToken    *oauth2.Token
}

// RawBearerToken is the blob of data returned, in JSON, when
// you request a new bearer token
type RawBearerToken struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
	expiryTime  time.Time
}

// TokenSourceWithChannel attaches a channel to a token source
// Access through one of these to get correctly-serialized responses
type TokenSourceWithChannel struct {
	source  *TokenSource
	channel chan controlMessage
}

// Token retrieves a new fresh bearer token from the
// requested source
func (source *TokenSource) retrieveRawToken() (*oauth2.Token, error) {
	for i := 0; i < source.params.NumRetries; i++ {
		tok, err := source.config.Token(source.ctx)
		if err == nil {
			return tok, nil
		}
		if i == (refreshRetries - 1) {
			return nil, err
		}
		time.Sleep(time.Duration(source.params.RetrySleep) * time.Millisecond)
	}
	log.Panic("somehow after for loop in retrieveRawToken")
	return nil, nil
}

// NewTokenSource sets up a new token source
func NewTokenSource(ctx context.Context, config clientcredentials.Config, params TokenSourceParams) *TokenSource {
	ts := TokenSource{ctx: ctx, config: config}
	ts.client = config.Client(ctx)
	//	ts := TokenSource{Credentials: creds, ScopesList: scopes, URL: url}
	ts.refreshChannel = make(chan *oauth2.Token)
	ts.controlChannels = make([]chan controlMessage, 0)
	ts.params = params

	ts.controlChannels = append(ts.controlChannels, make(chan controlMessage))

	go tokenControllerLoop(ts)
	go mainRefreshLoop(ts)

	return &ts
}

// NewTokenSourceWithChannel sets up a new channel and wrapper around an existing
// TokenSource
func (source *TokenSource) NewTokenSourceWithChannel() *TokenSourceWithChannel {
	newc := make(chan controlMessage)
	psrc := TokenSourceWithChannel{
		source:  source,
		channel: newc,
	}
	adminMessage := controlMessage{
		action:  registerChannel,
		channel: newc,
	}
	source.controlChannels[0] <- adminMessage
	return &psrc
}

// Refresh causes the underlying token source to refresh
func (source *TokenSourceWithChannel) Refresh() {
	source.channel <- controlMessage{
		action: refresh,
	}
}

// Token returns an access token
func (source *TokenSourceWithChannel) Token() *oauth2.Token {
	res := getAccessToken(source.channel)
	if res == nil {
		fmt.Println("got back nil!")
	}
	return res
}

// eof
