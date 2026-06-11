package http

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/pikoci/registry/pkreg"
)

func fetchVersion(s pkreg.Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		ns := vars["namespace"]
		name := vars["name"]
		ver := vars["version"]

		ip := extractIP(r)

		v, err := s.FetchVersionContent(r.Context(), ns, name, ver, "", ip)
		if err != nil {
			encodeError(w, http.StatusBadRequest, err.Error())
			return
		}

		encodeJSON(w, http.StatusOK, v)
	})
}

func publishVersion(s pkreg.Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		ns := vars["namespace"]
		name := vars["name"]
		ver := vars["version"]
		userID := r.Context().Value(UserIDContextKey).(string)

		// Parse multipart form; max 10MB
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			encodeError(w, http.StatusBadRequest, "invalid multipart form: "+err.Error())
			return
		}

		// Read manifest file
		manifestFile, _, err := r.FormFile("manifest")
		if err != nil {
			encodeError(w, http.StatusBadRequest, "manifest file is required")
			return
		}
		defer manifestFile.Close()

		manifestBytes, err := io.ReadAll(manifestFile)
		if err != nil {
			encodeError(w, http.StatusInternalServerError, "failed to read manifest")
			return
		}

		// Read optional readme file
		var readmeBytes []byte
		readmeFile, _, err := r.FormFile("readme")
		if err == nil {
			defer readmeFile.Close()
			readmeBytes, err = io.ReadAll(readmeFile)
			if err != nil {
				encodeError(w, http.StatusInternalServerError, "failed to read readme")
				return
			}
		}

		// Allow version override from form field
		if formVer := r.FormValue("version"); formVer != "" {
			ver = formVer
		}

		v, err := s.PublishVersion(r.Context(), userID, ns, name, ver, manifestBytes, readmeBytes)
		if err != nil {
			status := http.StatusBadRequest
			msg := err.Error()
			if strings.Contains(msg, "access denied") {
				status = http.StatusForbidden
			} else if strings.Contains(msg, "already exists") {
				status = http.StatusConflict
			}
			encodeError(w, status, msg)
			return
		}

		encodeJSON(w, http.StatusCreated, v)
	})
}

func yankVersion(s pkreg.Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		ns := vars["namespace"]
		name := vars["name"]
		ver := vars["version"]
		userID := r.Context().Value(UserIDContextKey).(string)

		if err := s.YankVersion(r.Context(), userID, ns, name, ver); err != nil {
			status := http.StatusBadRequest
			msg := err.Error()
			if strings.Contains(msg, "access denied") {
				status = http.StatusForbidden
			}
			encodeError(w, status, msg)
			return
		}

		encodeJSON(w, http.StatusOK, map[string]string{"status": "yanked"})
	})
}

func deprecateVersion(s pkreg.Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		ns := vars["namespace"]
		name := vars["name"]
		ver := vars["version"]
		userID := r.Context().Value(UserIDContextKey).(string)

		var body struct {
			Message   string `json:"message"`
			Successor string `json:"successor"`
		}
		json.NewDecoder(r.Body).Decode(&body)

		if err := s.DeprecateVersion(r.Context(), userID, ns, name, ver, body.Message, body.Successor); err != nil {
			status := http.StatusBadRequest
			if strings.Contains(err.Error(), "access denied") {
				status = http.StatusForbidden
			}
			encodeError(w, status, err.Error())
			return
		}

		encodeJSON(w, http.StatusOK, map[string]string{"status": "deprecated"})
	})
}

func deleteVersion(s pkreg.Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		ns := vars["namespace"]
		name := vars["name"]
		ver := vars["version"]
		userID := r.Context().Value(UserIDContextKey).(string)

		if err := s.DeleteVersion(r.Context(), userID, ns, name, ver); err != nil {
			status := http.StatusBadRequest
			if strings.Contains(err.Error(), "access denied") {
				status = http.StatusForbidden
			}
			encodeError(w, status, err.Error())
			return
		}

		w.WriteHeader(http.StatusNoContent)
	})
}

func deleteType(s pkreg.Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		ns := vars["namespace"]
		name := vars["name"]
		userID := r.Context().Value(UserIDContextKey).(string)

		if err := s.DeleteType(r.Context(), userID, ns, name); err != nil {
			status := http.StatusBadRequest
			if strings.Contains(err.Error(), "access denied") {
				status = http.StatusForbidden
			}
			encodeError(w, status, err.Error())
			return
		}

		w.WriteHeader(http.StatusNoContent)
	})
}

// extractIP returns the client IP from X-Forwarded-For or RemoteAddr.
func extractIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if idx := strings.IndexByte(xff, ','); idx != -1 {
			return strings.TrimSpace(xff[:idx])
		}
		return strings.TrimSpace(xff)
	}
	addr := r.RemoteAddr
	if idx := strings.LastIndex(addr, ":"); idx != -1 {
		return addr[:idx]
	}
	return addr
}
