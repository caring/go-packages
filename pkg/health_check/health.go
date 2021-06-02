package health_check

import (
	"encoding/json"
	"github.com/caring/go-packages/v2/pkg/logging"
	"net/http"
	"runtime"
)

var (
	service, branch, sha1, tag string
)

type Endpoint struct {
	Service   string `json:"service"`
	Branch    string `json:"branch"`
	SHA1      string `json:"sha1"`
	Tag       string `json:"tag"`
	GoVersion string `json:"go_version"`
	Log       *logging.Logger `json:"-"`
}

// NewEndpoint returns a initialized Endpoint struct
func NewEndpoint(l *logging.Logger) *Endpoint {
	var Tag string

	if len(tag) < 1 {
		Tag = "N/A"
	} else {
		Tag = tag
	}

	return &Endpoint{
		Service:   service,
		Branch:    branch,
		SHA1:      sha1,
		Tag:       Tag,
		GoVersion: runtime.Version(),
		Log:       l,
	}
}

func (e *Endpoint) LogStatus() error {
	msg, err := json.Marshal(e)
	if err != nil {
		return err
	}
	e.Log.Info(string(msg))
	return nil
}

func (e *Endpoint) Status(w http.ResponseWriter, r *http.Request) {
	msg, err := json.Marshal(e)
	if err != nil {
		e.Log.Error("Error encountered during status check: " + err.Error())
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(append(msg, []byte("\r\n")...))
	if err != nil {
		e.Log.Error("Error encountered during status check: " + err.Error())
		return
	}
	return
}
