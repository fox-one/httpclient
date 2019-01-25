package httpclient

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJoinGroup(t *testing.T) {
	base, _ := url.Parse("https://dev-gateway.fox.one")

	r := joinGroup(base, "merchant")
	assert.Equal(t, "https://dev-gateway.fox.one/merchant", r.String())

	r = joinGroup(base, "merchant/member?a=b")
	assert.Equal(t, "https://dev-gateway.fox.one/merchant/member?a=b", r.String())

	baseWithPath, _ := url.Parse("https://dev-gateway.fox.one/merchant")
	r = joinGroup(baseWithPath, "/member?a=b")
	assert.Equal(t, "https://dev-gateway.fox.one/merchant/member?a=b", r.String())

	baseWithQuery, _ := url.Parse("https://dev-gateway.fox.one/merchant?k=v")
	r = joinGroup(baseWithQuery, "https://cloud.fox.one/member?a=b")
	assert.Equal(t, "https://dev-gateway.fox.one/merchant/member?k=v&a=b", r.String())
}