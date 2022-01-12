package http

import (
	"encoding/json"
	"fmt"
	"hashing/src/services"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"syscall"
)

var passwordFieldName = "password"

func recordStats(createHandler func(res http.ResponseWriter, req *http.Request), stats services.StatsService) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		stats.Update(func() {
			createHandler(res, req)
		})
	}
}

func handleShutdown(done chan os.Signal) func(res http.ResponseWriter, req *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		done <- syscall.SIGTERM
	}
}

func tryWriteResponse(res http.ResponseWriter, data []byte) {
	if _, err := res.Write(data); err != nil {
		http.Error(res, fmt.Sprint(err), http.StatusBadRequest)
	}
}

func handleStats(stats services.StatsService) func(res http.ResponseWriter, req *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		data := stats.AsMap()
		if jsonResult, err := json.Marshal(data); err != nil {
			http.Error(res, fmt.Sprint(err), http.StatusInternalServerError)
		} else {
			tryWriteResponse(res, jsonResult)
		}
	}
}

func handleGet(hashingService services.HashingService) func(res http.ResponseWriter, req *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			http.Error(res, fmt.Sprintf("Invalid request method=%s. Only GET is supported", req.Method), http.StatusBadRequest)
			return
		}
		var id, rel string
		// TODO: Clean this up + better validation
		_, rel = ShiftPath(req.URL.Path)
		id, _ = ShiftPath(rel)
		if id == "" {
			// TODO: Better validation of input id param
			http.Error(res, fmt.Sprintf("Path param for resource id cannot be blank"), http.StatusBadRequest)
			return
		}
		if itemId, err := strconv.ParseUint(id, 10, 64); err != nil {
			// We'll assume that if we can't parse the input as a unit, it's a client error:
			http.Error(res, fmt.Sprint(err), http.StatusBadRequest)
		} else {
			if hashedItem, err := hashingService.Get(itemId); err != nil {
				http.Error(res, fmt.Sprint(err), http.StatusInternalServerError)
			} else {
				tryWriteResponse(res, []byte(hashedItem.GetHash()))
			}
		}
	}
}

func handleCreate(hashingService services.HashingService) func(res http.ResponseWriter, req *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			http.Error(res, fmt.Sprintf("Invalid request method, only POST is supported."), http.StatusBadRequest)
			return
		}
		if err := req.ParseForm(); err != nil {
			// Assume client error if the form data can't be parsed
			http.Error(res, fmt.Sprint(err), http.StatusBadRequest)
			return
		}
		password := strings.TrimSuffix(req.PostForm.Get(passwordFieldName), "\n")
		if hashedEntity, err := hashingService.Create(password); err != nil {
			switch err.(type) {
			case *services.InputLengthError:
				http.Error(res, fmt.Sprint(err), http.StatusBadRequest)
			case *services.PublishingError:
				http.Error(res, fmt.Sprint(err), http.StatusTooManyRequests)
			default:
				http.Error(res, fmt.Sprintf("An unknown error occurred while processing the request: %s", err), http.StatusInternalServerError)
			}
		} else {
			tryWriteResponse(res, []byte(strconv.FormatUint(hashedEntity.GetId(), 10)))
		}
	}
}

// ShiftPath splits off the first component of p, which will be cleaned of
// relative components before processing. head will never contain a slash and
// tail will always be a rooted path without trailing slash.
//
// Note: This approach was "borrowed" from: https://blog.merovius.de/2017/06/18/how-not-to-use-an-http-router.html
func ShiftPath(p string) (head, tail string) {
	p = path.Clean("/" + p)
	i := strings.Index(p[1:], "/") + 1
	if i <= 0 {
		return p[1:], "/"
	}
	return p[1:i], p[i:]
}
