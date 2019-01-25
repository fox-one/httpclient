package httpclient

import (
	"context"
	"io"
	"net/http"
)

func (c *Client) Redirect(ctx context.Context, req *http.Request, w http.ResponseWriter) error {
	req = req.WithContext(ctx)
	req.URL = joinGroup(c.base, req.URL.String())

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	for name, values := range resp.Header {
		w.Header()[name] = values
	}

	w.WriteHeader(resp.StatusCode)

	_, err = io.Copy(w, resp.Body)
	return err
}
