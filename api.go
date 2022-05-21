package osuapiv2

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/semaphore"
)

const (
	BASE_URL  = "https://osu.ppy.sh/api/v2"
	TOKEN_URL = "https://osu.ppy.sh/oauth/token"
)

type Api struct {
	httpClient *http.Client
	lock       *semaphore.Weighted
	token      string
	expires    time.Time
	config     *Config

	tokenLock       sync.RWMutex
	isFetchingToken bool
}

type Config struct {
	ClientId     string
	ClientSecret string
}

type OsuToken struct {
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	AccessToken string `json:"access_token"`
}

func New(config *Config) *Api {
	client := &http.Client{
		Timeout: 9 * time.Second,
	}

	// want to cap at around 1000 requests a minute, OSU cap is 1200
	lock := semaphore.NewWeighted(1000)

	return &Api{
		httpClient: client,
		lock:       lock,
		expires:    time.Now(),
		config:     config,
	}
}

func (api *Api) Token() (token string, err error) {
	if time.Now().Before(api.expires) {
		token = api.token
		return
	}

	if api.isFetchingToken {
		api.tokenLock.RLock()
		token = api.token
		api.tokenLock.RUnlock()
		return
	}

	api.tokenLock.Lock()
	api.isFetchingToken = true

	data := fmt.Sprintf(
		"client_id=%s&client_secret=%s&grant_type=client_credentials&scope=public",
		api.config.ClientId,
		api.config.ClientSecret,
	)

	resp, err := http.Post(
		TOKEN_URL,
		"application/x-www-form-urlencoded",
		strings.NewReader(data),
	)
	if err != nil {
		return
	}

	var osuToken OsuToken
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	err = json.Unmarshal(respBody, &osuToken)
	if err != nil {
		return
	}

	log.Println("got new access token", osuToken.AccessToken[:12]+"...")
	api.token = osuToken.AccessToken
	api.expires = time.Now().Add(time.Duration(osuToken.ExpiresIn) * time.Second)

	token = api.token
	api.tokenLock.Unlock()
	return
}

func (api *Api) Request0(action string, url string) (resp *http.Response, err error) {
	err = api.lock.Acquire(context.TODO(), 1)
	if err != nil {
		return
	}
	apiUrl := BASE_URL + url
	req, err := http.NewRequest(action, apiUrl, nil)

	token, err := api.Token()
	if err != nil {
		return
	}

	req.Header.Add("Authorization", "Bearer "+token)
	if err != nil {
		return
	}

	resp, err = api.httpClient.Do(req)
	if err != nil {
		return
	}

	if resp.StatusCode != 200 {
		var respBody []byte
		respBody, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return
		}

		err = fmt.Errorf("not 200: %s", string(respBody))
		return
	}

	// release the lock after 1 minute
	go func() {
		time.Sleep(time.Minute)
		api.lock.Release(1)
	}()
	return
}

func (api *Api) Request(action string, url string, result interface{}) (err error) {
	resp, err := api.Request0(action, url)
	if err != nil {
		return
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	err = json.Unmarshal(data, result)
	if err != nil {
		return
	}

	return
}
