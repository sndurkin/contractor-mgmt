package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"

	"github.com/decred/contractor-mgmt/cmswww/api/v1"
	"github.com/decred/politeia/util"
)

// isLoggedIn ensures that a user is logged in before calling the next
// function.
func (c *cmswww) isLoggedIn(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Debugf("isLoggedIn: %v %v %v %v", remoteAddr(r), r.Method,
			r.URL, r.Proto)

		email, err := c.GetSessionEmail(r)
		if err != nil {
			util.RespondWithJSON(w, http.StatusUnauthorized, v1.ErrorReply{
				ErrorCode: int64(v1.ErrorStatusNotLoggedIn),
			})
			return
		}

		// Check if user is authenticated
		if email == "" {
			util.RespondWithJSON(w, http.StatusUnauthorized, v1.ErrorReply{
				ErrorCode: int64(v1.ErrorStatusNotLoggedIn),
			})
			return
		}

		f(w, r)
	}
}

// isLoggedInAsAdmin ensures that a user is logged in as an admin user
// before calling the next function.
func (c *cmswww) isLoggedInAsAdmin(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check if user is admin
		isAdmin, err := c.isAdmin(r)
		log.Debugf("isLoggedInAsAdmin: %v %v %v %v %v", isAdmin, remoteAddr(r),
			r.Method, r.URL, r.Proto)
		if err != nil {
			log.Errorf("isLoggedInAsAdmin: isAdmin %v", err)
			util.RespondWithJSON(w, http.StatusUnauthorized, v1.ErrorReply{
				ErrorCode: int64(v1.ErrorStatusNotLoggedIn),
			})
			return
		}
		if !isAdmin {
			util.RespondWithJSON(w, http.StatusForbidden, v1.ErrorReply{})
			return
		}

		f(w, r)
	}
}

// logging logs all incoming commands before calling the next funxtion.
//
// NOTE: LOGGING WILL LOG PASSWORDS IF TRACING IS ENABLED.
func logging(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Trace incoming request
		log.Tracef("%v", newLogClosure(func() string {
			trace, err := httputil.DumpRequest(r, true)
			if err != nil {
				trace = []byte(fmt.Sprintf("logging: "+
					"DumpRequest %v", err))
			}
			return string(trace)
		}))

		// Log incoming connection
		log.Infof("%v %v %v %v", remoteAddr(r), r.Method, r.URL, r.Proto)
		f(w, r)
	}
}

// closeBody closes the request body.
func closeBody(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		f(w, r)
		r.Body.Close()
	}
}

func remoteAddr(r *http.Request) string {
	via := r.RemoteAddr
	xff := r.Header.Get(v1.Forward)
	if xff != "" {
		return fmt.Sprintf("%v via %v", xff, r.RemoteAddr)
	}
	return via
}

func (c *cmswww) loadInventory(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := c.LoadInventory(); err != nil {
			RespondWithError(w, r, 0,
				"failed to get Load Inventory", err)
			return
		}
		f(w, r)
	}
}