package health_check

import (
	"encoding/json"
	"github.com/caring/go-packages/v2/pkg/logging"
	"net/http"
)

type Endpoint struct {
	Service     string `json:"service"`
	Branch 		string `json:"branch"`
	SHA1   		string `json:"sha1"`
	Tag    		string `json:"tag"`
	GoVersion 	string `json:"go_version"`
	ContainerID string `json:"container_id"`
	StartedAt string `json:"started_at"`
	Log         *logging.Logger
}

// NewEndpoint returns a initialized Endpoint struct
func NewEndpoint(service, branch, sha1, tag, goVersion *string, l *logging.Logger) *Endpoint {
	var Tag string

	if  len(*tag) < 1 {
		Tag = *tag
	} else {
		Tag = "N/A"
	}

	return &Endpoint{
		Service: *service,
		Branch: *branch,
		SHA1:   *sha1,
		Tag: Tag,
		GoVersion: *goVersion,
		ContainerID: "",
		StartedAt: "",
		Log: l,
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

func(e *Endpoint) Status(w http.ResponseWriter, r *http.ResponseWriter) {
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
