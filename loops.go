package oauth2bearer

import (
	"fmt"
	"log"
	"reflect"
	"time"

	"golang.org/x/oauth2"
)

// this loops on one goroutine, waiting until the token is about to expire
// and then grabbing a new one
func mainRefreshLoop(source TokenSource) {
	timeToWait := 0.0
	for {
		time.Sleep(time.Duration(timeToWait) * time.Second)
		initRaw, err := source.retrieveRawToken()
		if err != nil {
			log.Panic("cannot refresh raw token")
		}
		log.Print("got new raw token")
		source.refreshChannel <- initRaw
		//		log.Print("send complete")
		waitDurationS := initRaw.Expiry.Sub(time.Now()).Seconds()
		timeToWait = waitDurationS - source.params.RefreshMargin
	}
}

// this loop receives updates from the mainRefreshLoop and sends back
// up-to-date tokens when queries by the user code
func tokenControllerLoop(ts TokenSource) {
	userChans := ts.controlChannels
	refreshChan := ts.refreshChannel
	rawToken := <-refreshChan
	//	log.Print("got first raw token")
	for {

		cases := make([]reflect.SelectCase, 0)
		cases = append(cases, reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(refreshChan),
		})
		for _, v := range userChans {
			cases = append(cases, reflect.SelectCase{
				Dir:  reflect.SelectRecv,
				Chan: reflect.ValueOf(v),
			})
		}

		chosen, val, ok := reflect.Select(cases)

		if !ok {
			panic("got error in select")
		}

		if chosen == 0 {
			// got refresh message
			fmt.Println("got refresh message")
			asToken, ok := val.Interface().(*oauth2.Token)
			if !ok {
				panic("token type assert failed")
			}
			rawToken = asToken
			//			fmt.Println("done refresh message")
		} else {
			chosenChan := userChans[chosen-1]
			// got control message
			asControl, ok := val.Interface().(controlMessage)
			if !ok {
				panic("control type assert failed")
			}
			action := asControl.action
			if action == getToken {
				//				fmt.Println("get token")
				chosenChan <- controlMessage{
					action: sendToken,
					token:  rawToken,
				}
				//				fmt.Println("token sent")
			} else if action == refresh {
				fmt.Println("control refresh")
				// do refresh
				newRawToken, err := ts.retrieveRawToken()
				if err != nil {
					panic("force refresh got error")
				}
				rawToken = newRawToken
			} else if action == registerChannel {
				userChans = append(userChans, asControl.channel)
			}
		}
	}
}

// this is how we get a token from the controller
// it sends a message on the controller channel and then
// returns what comes back
func getAccessToken(controllerChan chan controlMessage) *oauth2.Token {
	//	fmt.Println("sending request for token")
	controllerChan <- controlMessage{
		action: getToken,
	}
	//	fmt.Println("reading reply")
	reply := <-controllerChan
	if reply.action != sendToken {
		panic("didnt get token back")
	} else {
		//		fmt.Println("got token back")
	}
	return reply.token
}

// eof
