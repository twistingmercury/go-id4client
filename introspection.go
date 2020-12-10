package id4client

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"strconv"
	"strings"
)

// IntrospectResponse contains the response data from IdentityServer4 introspect call.
type IntrospectResponse struct {
	ISS      string `json:"iss,omitempty"`
	NBF      int64  `json:"nbf,omitempty"`
	EXP      int64  `json:"exp,omitempty"`
	AUD      string `json:"aud,omitempty"`
	ClientID string `json:"client_id"`
	Active   bool   `json:"active"`
	Scope    string `json:"scope,omitempty"`
}

// Introspect adds a simple client bearer token authentication against a security token service (sts).
func Introspect(req *http.Request) (active bool, statusCode int, status string) {
	active = false
	statusCode = 500
	status = "internal server error"
	t, statusCode, status := extractToken(req.Header)
	if statusCode != 200 {
		return
	}

	ar, err := NewIntrospectRequest(t)
	if err != nil {
		return
	}

	authClient := newHTTPClient()
	res, err := authClient.Do(ar)
	if err != nil {
		status = err.Error()
		return
	}
	statusCode = res.StatusCode
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return
	}

	var ir IntrospectResponse
	err = json.Unmarshal(body, &ir)
	if err != nil {
		return
	}

	active = ir.Active
	return
}

func extractToken(h http.Header) (t string, statusCode int, status string) {
	statusCode = 401
	bearer := h.Get("Authorization")
	if len(bearer) == 0 {
		status = "no authentication header"
		return
	}

	jwt := strings.Split(bearer, " ")
	if len(jwt) != 2 {
		status = "invalid authentication header"
		return
	}

	t = jwt[1]
	if len(t) == 0 {
		status = "invalid bearer token"
		return
	}

	statusCode = 200
	status = "ok"
	return
}

// NewIntrospectRequest creates a new http request to be used to validate the access token
// that is passed in.
func NewIntrospectRequest(token string) (ar *http.Request, err error) {
	data := url.Values{"token": {token}}.Encode()
	ar, err = http.NewRequest("POST", introspectURL(), strings.NewReader(data))
	if err != nil {
		return
	}

	ar.SetBasicAuth(idcfg.ID, idcfg.Secret)
	ar.Header.Add("Accept", "*/*")
	ar.Header.Add("Accept-Encoding", "gzip, deflate, br")
	ar.Header.Add("Cache-Control", "no-cache")
	ar.Header.Add("Connection", "keep-alive")
	ar.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	ar.Header.Add("Content-Length", strconv.Itoa(len(data)))
	ar.Header.Add("User-Agent", fmt.Sprintf("app: %s, ver: %s, commit: %s", idcfg.ServiceName, idcfg.ServiceVersion, idcfg.CommitHash))

	return
}
