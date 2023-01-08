package transport

import (
	"net/http"
	"net/http/httptrace"
	"net/url"
	"time"

	log "github.com/sirupsen/logrus"

	"golang.org/x/net/proxy"
)

func NewSocksClient(socks string) *http.Client {
	u, err := url.Parse(socks)
	if err != nil {
		log.Panic(err)
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

	dialer, err := proxy.SOCKS5("tcp", u.Host, auth, proxy.Direct)
	if err != nil {
		log.Panic(err)
	}
	transport := &http.Transport{
		Dial: dialer.Dial,
	}

	return &http.Client{
		Transport: transport,
	}
}

func NewRedirectClient() *http.Client {
	return &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			log.Debugf("REDIRECT: %v", req.URL.String())
			return nil
		},
	}

}

func NewTrace() *httptrace.ClientTrace {
	var dns time.Time
	return &httptrace.ClientTrace{
		DNSStart: func(dsi httptrace.DNSStartInfo) { dns = time.Now() },
		GotFirstResponseByte: func() {
			log.Debugf("RESPONSE: time to first byte received %v", time.Since(dns))
		},
	}
}
