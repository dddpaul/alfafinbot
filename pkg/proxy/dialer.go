package proxy

import (
	"net"
	"net/http"
	"net/url"

	log "github.com/sirupsen/logrus"
	"golang.org/x/net/proxy"
)

var directDialer = &net.Dialer{}

func NewTransport(socks string) http.RoundTripper {
	return &http.Transport{
		Dial: newDialer(socks).Dial,
	}
}

func newDialer(socks string) proxy.Dialer {
	if len(socks) == 0 {
		log.Debugf("SOCKS5 proxy URL is empty, use DIRECT connection")
		return directDialer
	}

	u, err := url.Parse(socks)
	if err != nil {
		log.Infof("SOCKS5 proxy URL parse error, use DIRECT connection")
		return directDialer
	}

	var auth *proxy.Auth
	if u.User != nil {
		auth = &proxy.Auth{
			User: u.User.Username(),
		}
		if p, ok := u.User.Password(); ok {
			auth.Password = p
		}
	}

	d, err := proxy.SOCKS5("tcp", u.Host, auth, proxy.Direct)
	if err != nil {
		log.Infof("SOCKS5 proxy init error, use DIRECT connection")
		return directDialer
	}
	return d
}
