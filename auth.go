package httpclient

func (r *Request) AddToken(token string) *Request {
	return r.H("Authorization", "Bearer "+token)
}

type tokenStringAuth string

func (token tokenStringAuth) Auth(req *Request, method, uri string, body []byte) {
	req.AddToken(string(token))
}
