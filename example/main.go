package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"time"

	"golang.org/x/oauth2/clientcredentials"

	"github.com/jonreiter/oauth2bearer"
)

const defaultTokenURL = "https://apiauth.dfnapp.com/oauth2/token"
const defaultScope1 = "clientapi/basicsearch"
const defaultScope2 = "clientapi/advancedsearch"

func sourceLoop(src *oauth2bearer.TokenSourceWithChannel) {
	for {
		accessToken := src.Token()
		fmt.Println("got:", len(accessToken.AccessToken))
		fmt.Println("working")
		time.Sleep(1 * time.Second)
		if rand.Float64() < 0.1 {
			src.Refresh()
		}
	}
}

func main() {
	scopes := []string{defaultScope1, defaultScope2}
	config := clientcredentials.Config{
		ClientID:     os.Getenv("DF_CLIENT_ID"),
		ClientSecret: os.Getenv("DF_CLIENT_SECRET"),
		TokenURL:     defaultTokenURL,
		Scopes:       scopes,
	}
	params := oauth2bearer.NewDefaultTokenSourceParams()
	params.NumRetries = 10
	params.RefreshMargin = 10
	params.RetrySleep = 100
	src := oauth2bearer.NewTokenSource(context.Background(), config, params)

	nRoutines := 10

	for i := 0; i < nRoutines; i++ {
		perRoutine := src.NewTokenSourceWithChannel()
		go sourceLoop(perRoutine)
	}
	perRoutine := src.NewTokenSourceWithChannel()
	sourceLoop(perRoutine)
}

// eof
