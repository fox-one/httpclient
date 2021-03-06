package httpclient

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"

	jsoniter "github.com/json-iterator/go"
)

const (
	jsonContentType = "application/json"
)

type AuthPreparer interface {
	PrepareAuth(req *Request)
}

type Authenticator interface {
	Auth(req *Request, method, uri string, body []byte)
}

type Request struct {
	client *Client

	method  string
	uri     string
	params  map[string]interface{}
	headers http.Header
	query   url.Values

	body            []byte
	bodyContentType string

	auth Authenticator
}

// H add key value into headers
func (r *Request) H(key, value string) *Request {
	r.headers.Add(key, value)
	return r
}

// p add key value into parameters
func (r *Request) P(key string, value interface{}) *Request {
	if value != nil {
		r.params[key] = value
	} else {
		delete(r.params, key)
	}

	return r
}

func (r *Request) Q(key string, value interface{}) *Request {
	if value != nil {
		r.query.Set(key, fmt.Sprint(value))
	} else {
		r.query.Del(key)
	}

	return r
}

// Header overwrite exist headers
func (r *Request) Header(h http.Header) *Request {
	r.headers = h
	return r
}

func (r *Request) Body(b interface{}, contentType ...string) *Request {
	t := jsonContentType
	if len(contentType) > 0 {
		t = contentType[0]
	}

	if reader, ok := b.(io.Reader); ok {
		r.body, _ = ioutil.ReadAll(reader)
	} else {
		r.body, _ = jsoniter.Marshal(b)
	}

	r.bodyContentType = t
	return r
}

func (r *Request) Auth(auth Authenticator) *Request {
	r.auth = auth
	return r
}

type Result struct {
	req *Request

	status     string
	statusCode int
	data       []byte
	err        error
}

func (r *Result) IsSuccess() bool {
	return r.statusCode > 199 && r.statusCode < 300
}

func (r *Result) StatusErr() error {
	return errors.New(r.status)
}

func (r *Result) String() string {
	return fmt.Sprintf("%s %s %s", r.req.method, r.req.uri, r.status)
}

func (r *Result) Err() error {
	return r.err
}

func (r *Result) Status() (int, string) {
	return r.statusCode, r.status
}

func (r *Result) Bytes() ([]byte, error) {
	return r.data, r.err
}

func (r *Result) Reader() (io.Reader, error) {
	return bytes.NewReader(r.data), r.err
}

func (r *Request) HTTPRequest() (*http.Request, error) {
	u := joinGroup(r.client.base, r.uri)
	if r.auth != nil {
		if preparer, ok := r.auth.(AuthPreparer); ok {
			preparer.PrepareAuth(r)
		}
	}

	if query := r.query.Encode(); query != "" {
		if u.RawQuery == "" {
			u.RawQuery = query
		} else {
			u.RawQuery += "&" + query
		}
	}

	var body []byte

	switch r.method {
	case http.MethodPut, http.MethodPost, http.MethodPatch:
		if r.body == nil && len(r.params) > 0 {
			r.Body(r.params)
		}

		if r.body != nil {
			body = r.body
			r.H("Content-Type", r.bodyContentType)
		}
	default:
		query := url.Values{}
		for k, v := range r.params {
			value := fmt.Sprint(v)
			query.Add(k, value)
		}

		if raw := query.Encode(); raw != "" {
			if u.RawQuery == "" {
				u.RawQuery = raw
			} else {
				u.RawQuery += "&" + raw
			}
		}
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
		return nil, err
	}

	request.Header = r.headers
	return request, nil
}

func (r *Request) Do(ctx context.Context) *Result {
	result := &Result{req: r}
	req, err := r.HTTPRequest()
	if err != nil {
		result.err = err
		return result
	}

	req = req.WithContext(ctx)
	resp, err := r.client.Client().Do(req)

	if resp != nil {
		result.statusCode, result.status = resp.StatusCode, resp.Status

		if resp.Body != nil {
			defer resp.Body.Close()
		}
	}

	if err != nil {
		result.err = err
		return result
	}

	result.data, result.err = ioutil.ReadAll(resp.Body)
	return result
}
