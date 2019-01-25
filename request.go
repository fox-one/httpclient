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

type Result struct {
	code int
	data []byte
	err  error
}

func (r *Result) Err() error {
	return r.err
}

func (r *Result) StatusCode() int {
	return r.code
}

func (r *Result) Bytes() ([]byte, error) {
	return r.data, r.err
}

func (r *Result) Reader() (io.Reader, error) {
	return bytes.NewBuffer(r.data), r.err
}

func (r *Request) Do(ctx context.Context) *Result {
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

	result := &Result{}

	request, err := http.NewRequest(r.method, u.String(), bytes.NewReader(body))
	if err != nil {
		result.err = err
		return result
	}

	request.Header = r.headers
	request = request.WithContext(ctx)
	resp, err := r.client.Client().Do(request)

	result.code = resp.StatusCode
	if err != nil {
		result.err = err
		return result
	}

	defer resp.Body.Close()
	result.data, result.err = ioutil.ReadAll(resp.Body)
	return result
}
