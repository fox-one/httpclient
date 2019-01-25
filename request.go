package httpclient

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	jsoniter "github.com/json-iterator/go"
)

type Authenticator interface {
	Auth(req *Request, method, uri string, body []byte)
}

type Request struct {
	client *Client

	method  string
	uri     string
	params  map[string]interface{}
	headers http.Header

	auth Authenticator
}

func (r *Request) H(key, value string) *Request {
	r.headers.Add(key, value)
	return r
}

func (r *Request) P(key string, value interface{}) *Request {
	if value != nil {
		r.params[key] = value
	} else {
		delete(r.params, key)
	}

	return r
}

func (r *Request) Auth(auth Authenticator) *Request {
	r.auth = auth
	return r
}

func (r *Request) Do(ctx context.Context) (int, io.Reader, error) {
	u := joinGroup(r.client.base, r.uri)

	var body []byte

	switch r.method {
	case http.MethodPut, http.MethodPost:
		body, _ = jsoniter.Marshal(r.params)
	default:
		query := u.Query()
		for k, v := range r.params {
			value := fmt.Sprint(v)
			query.Add(k, value)
		}
		u.RawQuery = query.Encode()
	}

	if r.auth != nil {
		uri := u.Path
		if u.RawQuery != "" {
			uri += "?" + u.RawQuery
		}

		r.auth.Auth(r, r.method, uri, body)
	}

	request, err := http.NewRequest(r.method, u.String(), bytes.NewReader(body))
	if err != nil {
		return 0, nil, err
	}

	request.Header = r.headers
	request = request.WithContext(ctx)
	resp, err := r.client.Client().Do(request)
	if err != nil {
		return resp.StatusCode, nil, err
	}

	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)

	return resp.StatusCode, bytes.NewReader(data), err
}
