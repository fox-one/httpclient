package httpclient

import (
	"net/http"
	"net/url"
	"strings"
)

type Client struct {
	base   *url.URL
	client *http.Client

	OnRequest func(req *Request, method, uri string)
}

func parseAPIBase(apiBase string) (*url.URL, error) {
	u, err := url.Parse(apiBase)
	if err != nil {
		return nil, err
	}

	if u.Scheme == "" {
		panic("api base has no scheme")
	}

	return u, nil
}

func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}

func joinGroup(base *url.URL, group string) *url.URL {
	u, err := url.Parse(group)
	if err != nil {
		panic(err)
	}

	baseQuery := base.RawQuery
	u.Scheme = base.Scheme
	u.Host = base.Host
	u.Path = singleJoiningSlash(base.Path, u.Path)
	if baseQuery == "" || u.RawQuery == "" {
		u.RawQuery = baseQuery + u.RawQuery
	} else {
		u.RawQuery = baseQuery + "&" + u.RawQuery
	}

	return u
}

func NewClient(apiBase string) *Client {
	u, err := parseAPIBase(apiBase)
	if err != nil {
		panic(err)
	}

	return &Client{
		base:   u,
		client: http.DefaultClient,
	}
}

func (c *Client) Group(group string) *Client {
	return &Client{
		base:      joinGroup(c.base, group),
		client:    c.client,
		OnRequest: c.OnRequest,
	}
}

func (c *Client) UseClient(client *http.Client) {
	c.client = client
}

func (c *Client) Client() *http.Client {
	if c.client != nil {
		return c.client
	}

	return http.DefaultClient
}

func (c *Client) req(method, uri string) *Request {
	req := &Request{
		client:  c,
		method:  method,
		uri:     uri,
		params:  map[string]interface{}{},
		headers: http.Header{},
	}

	if c.OnRequest != nil {
		c.OnRequest(req, method, uri)
	}

	return req
}

func (c *Client) GET(uri string) *Request {
	return c.req(http.MethodGet, uri)
}

func (c *Client) DELETE(uri string) *Request {
	return c.req(http.MethodDelete, uri)
}

func (c *Client) PUT(uri string) *Request {
	return c.req(http.MethodPut, uri)
}

func (c *Client) POST(uri string) *Request {
	return c.req(http.MethodPost, uri)
}

func (c *Client) M(method string, uri string) *Request {
	return c.req(method, uri)
}
