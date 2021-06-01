package health_check

type Endpoint struct {
	Branch 		string
	SHA1   		string
	Tag    		string
	GoVersion 	string
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

