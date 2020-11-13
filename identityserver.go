package id4client

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"

	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

type cachedToken struct {
	timeCached int64
	token      string
	expiresIn  int64
}

// IdentityConfig defines the settings required to use IdentityServer.
type IdentityConfig struct {
	BaseURL        string
	TokenPath      string
	IntrospectPath string
	ID             string
	Secret         string
	ServiceName    string
	ServiceVersion string
	CommitHash     string
}

func introspectURL() string {
	return fmt.Sprintf("%s/%s", idcfg.BaseURL, idcfg.IntrospectPath)
}

func tokenURL() string {
	return fmt.Sprintf("%s/%s", idcfg.BaseURL, idcfg.TokenPath)
}

var (
	idcfg      IdentityConfig
	once       sync.Once
	tokenCache map[string]cachedToken
	tmu        sync.RWMutex
)

// Initialize bootstraps the necessary settings use IdentityServer.
// This func only runs once during the lifetime of the app. Subsequent calls
// are a NOOP.
func Initialize(c IdentityConfig) (err error) {
	once.Do(func() {
		if err = initConfig(c); err != nil {
			return
		}

		tokenCache = make(map[string]cachedToken)
		go purgeOldTokens()
		logrus.Debug("purgeOldTokens routine started")
	})

	return
}

func initConfig(c IdentityConfig) (err error) {
	switch {
	case len(c.ID) == 0:
		err = errors.New("missing id")
	case len(c.Secret) == 0:
		err = errors.New("missing secret")
	case len(c.ServiceName) == 0:
		err = errors.New("missing service name")
	case len(c.ServiceVersion) == 0:
		err = errors.New("missing service version")
	case len(c.BaseURL) == 0:
		err = errors.New("missing base URL")
	case len(c.IntrospectPath) == 0:
		err = errors.New("missing instrospection path")
	case len(c.TokenPath) == 0:
		err = errors.New("missing token path")
	default:
		idcfg = c
	}
	return
}

func getCachedToken(scope string) (t *string) {
	tmu.RLock()
	defer tmu.RUnlock()
	for s, ct := range tokenCache {
		if s == scope {
			at := ct.token
			t = &at
			return
		}
	}
	return
}

func setCachedToken(s, t string, ttl int64) {
	tmu.Lock()
	defer tmu.Unlock()
	tokenCache[s] = cachedToken{
		timeCached: time.Now().Unix(),
		token:      t,
		expiresIn:  ttl,
	}
}

func scanCachedTokens() {
	tmu.RLock()
	defer tmu.RUnlock()
	for s, ct := range tokenCache {
		evalCachedToken(s, ct)
	}
}

func evalCachedToken(s string, ct cachedToken) {
	diff := time.Now().Unix() - ct.timeCached
	per := float64(diff) / float64(ct.expiresIn)
	// if the expiresIn time is within 99% of the total
	// time cached then go ahead and delete it.
	if per >= 0.99 {
		func() {
			tmu.Lock()
			defer tmu.Unlock()
			delete(tokenCache, s)
		}()
	}
}

func purgeOldTokens() {
	for range time.Tick(time.Minute) {
		scanCachedTokens()
	}
}

// AuthResponse contains the response from the STS.
type AuthResponse struct {
	AccessToken string  `json:"access_token"`
	ExpiresIn   int     `json:"expires_in"`
	TokenType   string  `json:"token_type"`
	Scope       string  `json:"scope"`
	Error       *string `json:"error,omitempty"`
}

// AuthRequestData contains the data needed to request an access token.
type AuthRequestData struct {
	ClientID     string
	ClientSecret string
	GrantType    string
	Scope        string
}

// reader encodes the data into "URL Encoded" form sorted by key.
func (arb *AuthRequestData) reader() *strings.Reader {
	dat := url.Values{
		"client_id":     {arb.ClientID},
		"client_secret": {arb.ClientSecret},
		"grant_type":    {arb.GrantType},
		"scope":         {arb.Scope},
	}.Encode()
	return strings.NewReader(dat)
}

// requestToken returns a new authorized http.Client.
func requestToken(arb AuthRequestData) (ar *AuthResponse, err error) {
	req, err := http.NewRequest("POST", tokenURL(), arb.reader())
	if err != nil {
		return
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	client := newHTTPClient()
	res, err := client.Do(req)
	if err != nil {
		return
	}
	defer res.Body.Close()

	raw, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return
	}

	err = json.Unmarshal(raw, &ar)
	return
}

// NewAuthenticatedRequest returns a new request that has been authenticated.
func NewAuthenticatedRequest(arb AuthRequestData, httpMethod string, url string, headers map[string]string, body io.Reader) (req *http.Request, err error) {
	ar, err := requestToken(arb)
	if err != nil {
		return
	}

	req, err = http.NewRequest(httpMethod, url, body)
	if err != nil {
		return
	}

	bearer := "Bearer " + ar.AccessToken
	req.Header.Set("Authorization", bearer)

	for k, v := range headers {
		req.Header.Add(k, v)
	}

	return req, nil
}

func newHTTPClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).Dial,
			TLSHandshakeTimeout:   10 * time.Second,
			ResponseHeaderTimeout: 10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}
}
