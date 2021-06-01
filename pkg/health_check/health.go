package health_check

type Endpoint struct {
	Branch 		string `json:"branch"`
	SHA1   		string `json:"sha1"`
	Tag    		string `json:"tag"`
	GoVersion 	string `json:"go_version"`
}

// NewEndpoint returns a initialized Endpoint struct
func NewEndpoint(branch, sha1, tag, goVersion *string) *Endpoint {
	var Tag string

	if  len(*tag) < 1 {
		Tag = *tag
	} else {
		Tag = "N/A"
	}
	
	return &Endpoint{
		Branch: *branch,
		SHA1:   *sha1,
		Tag: Tag,
		GoVersion: *goVersion,
	}
}

