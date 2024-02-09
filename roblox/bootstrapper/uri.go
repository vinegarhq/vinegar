package bootstrapper

import (
	"strings"
)

// ProtocolURI is a representation of a Roblox web provided
// URI which is sent to Player using MIME as a command line argument.
//
// The protocol URI is formatted in a hacky way so it is easy to parse
// as a key:value seperated with the colon character.
//
// The order may matter, but it seems that the client parses it normal.
//
// Example:
// 	roblox-player:1+channel:ganesh
// 	roblox-player:1+channel:
type ProtocolURI map[string]string

// ParseProtocolURI returns the given MIME protocol URI in [ProtocolURI] form.
func ParseProtocolURI(mime string) ProtocolURI {
	puri := make(ProtocolURI)
	uris := strings.Split(mime, "+")

	for _, uri := range uris {
		kv := strings.Split(uri, ":")

		if len(kv) == 2 {
			puri[kv[0]] = kv[1]
		}
	}

	return puri
}

// String returns the single command line argument of the [ProtocolURI].
func (puri ProtocolURI) String() string {
	var uris []string
	for k, v := range puri {
		uris = append(uris, k+":"+v)
	}
	return strings.Join(uris, "+")
}
