package httpclient

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestJoinGroup(t *testing.T) {
	base, _ := url.Parse("https://dev-gateway.fox.one")

	r := joinGroup(base, "/merchant")
	assert.Equal(t, "https://dev-gateway.fox.one/merchant", r.String())
	assert.Equal(t, "/merchant", r.Path)

	r = joinGroup(base, "/merchant/member?a=b")
	assert.Equal(t, "https://dev-gateway.fox.one/merchant/member?a=b", r.String())
	assert.Equal(t, "/merchant/member", r.Path)

	baseWithPath, _ := url.Parse("https://dev-gateway.fox.one/merchant")
	r = joinGroup(baseWithPath, "/member?a=b")
	assert.Equal(t, "https://dev-gateway.fox.one/merchant/member?a=b", r.String())
	assert.Equal(t, "/merchant/member", r.Path)

	baseWithQuery, _ := url.Parse("https://dev-gateway.fox.one/merchant?k=v")
	r = joinGroup(baseWithQuery, "https://cloud.fox.one/member?a=b")
	assert.Equal(t, "https://dev-gateway.fox.one/merchant/member?k=v&a=b", r.String())
	assert.Equal(t, "/merchant/member", r.Path)
}

func TestRequest(t *testing.T) {
	type Resp struct {
		Method string `json:"method"`
		Uri    string `json:"uri"`
		Body   string `json:"body"`
		Type   string `json:"type"`
	}

	e := gin.Default()
	e.Any("/test", func(c *gin.Context) {
		method := c.Request.Method
		uri := c.Request.URL.String()
		body, _ := ioutil.ReadAll(c.Request.Body)
		contentType := c.GetHeader("Content-Type")
		c.JSON(http.StatusOK, Resp{
			Method: method,
			Uri:    uri,
			Body:   string(body),
			Type:   contentType,
		})
	})

	host := "http://localhost"
	client := NewClient(host)
	w := httptest.NewRecorder()

	readResp := func() Resp {
		resp := Resp{}
		json.NewDecoder(w.Body).Decode(&resp)
		return resp
	}

	{
		req, err := client.GET("/test").P("foo", "bar").HTTPRequest()
		if assert.Nil(t, err) {
			e.ServeHTTP(w, req)
			resp := readResp()
			assert.Equal(t, "GET", resp.Method)
			assert.Equal(t, host+"/test?foo=bar", resp.Uri)
			assert.Equal(t, "", resp.Body)
			assert.Equal(t, "", resp.Type)
		}
	}

	{
		req, err := client.POST("/test").P("foo", "bar").HTTPRequest()
		if assert.Nil(t, err) {
			e.ServeHTTP(w, req)
			resp := readResp()
			assert.Equal(t, "POST", resp.Method)
			assert.Equal(t, host+"/test", resp.Uri)
			assert.Equal(t, `{"foo":"bar"}`, resp.Body)
			assert.Equal(t, jsonContentType, resp.Type)
		}
	}

	{
		body := map[string]interface{}{"foo": "bar"}
		req, err := client.POST("/test").Q("foo", "bar").Body(body).HTTPRequest()
		if assert.Nil(t, err) {
			e.ServeHTTP(w, req)
			resp := readResp()
			assert.Equal(t, "POST", resp.Method)
			assert.Equal(t, host+"/test?foo=bar", resp.Uri)
			assert.Equal(t, `{"foo":"bar"}`, resp.Body)
			assert.Equal(t, jsonContentType, resp.Type)
		}
	}

	{
		body := strings.NewReader("foo=bar")
		contentType := "application/x-www-form-urlencoded"
		req, err := client.POST("/test").P("foo", "bar").Body(body, contentType).HTTPRequest()
		if assert.Nil(t, err) {
			e.ServeHTTP(w, req)
			resp := readResp()
			assert.Equal(t, "POST", resp.Method)
			assert.Equal(t, host+"/test", resp.Uri)
			assert.Equal(t, `foo=bar`, resp.Body)
			assert.Equal(t, contentType, resp.Type)
		}
	}
}
