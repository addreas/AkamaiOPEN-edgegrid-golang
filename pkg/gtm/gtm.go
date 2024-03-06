// Package gtm provides access to the Akamai GTM V1_4 APIs
//
// See: https://techdocs.akamai.com/gtm/reference/api
package gtm

import (
	"net/http"

	"github.com/akamai/AkamaiOPEN-edgegrid-golang/v7/pkg/session"
)

type (
	// GTM is the gtm api interface
	GTM interface {
		Domains
		Properties
		Datacenters
		Resources
		ASMaps
		GeoMaps
		CIDRMaps
	}

	gtm struct {
		session.Session
	}

	// Option defines a GTM option
	Option func(*gtm)

	// ClientFunc is a gtm client new method, this can used for mocking
	ClientFunc func(sess session.Session, opts ...Option) GTM
)

// Client returns a new dns Client instance with the specified controller
func Client(sess session.Session, opts ...Option) GTM {
	p := &gtm{
		Session: sess,
	}

	for _, opt := range opts {
		opt(p)
	}
	return p
}

// Exec overrides the session.Exec to add dns options
func (g *gtm) Exec(r *http.Request, out interface{}, in ...interface{}) (*http.Response, error) {
	return g.Session.Exec(r, out, in...)
}
