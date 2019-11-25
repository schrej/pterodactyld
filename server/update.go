package server

import (
	"encoding/json"
	"github.com/buger/jsonparser"
	"github.com/imdario/mergo"
	"github.com/pkg/errors"
)

// Merges data passed through in JSON form into the existing server object.
// Any changes to the build settings will apply immediately in the environment
// if the environment supports it.
//
// The server will be marked as requiring a rebuild on the next boot sequence,
// it is up to the specific environment to determine what needs to happen when
// that is the case.
func (s *Server) UpdateDataStructure(data []byte) error {
	src := Server{}
	if err := json.Unmarshal(data, &src); err != nil {
		return errors.WithStack(err)
	}

	// Don't allow obviously corrupted data to pass through into this function. If the UUID
	// doesn't match something has gone wrong and the API is attempting to meld this server
	// instance into a totally different one, which would be bad.
	if src.Uuid != s.Uuid {
		return errors.New("attempting to merge a data stack with an invalid UUID")
	}

	// Merge the new data object that we have received with the existing server data object
	// and then save it to the disk so it is persistent.
	if err := mergo.Merge(s, src, mergo.WithOverride); err != nil {
		return errors.WithStack(err)
	}

	// Mergo can't quite handle this boolean value correctly, so for now we'll just
	// handle this edge case manually since none of the other data passed through in this
	// request is going to be boolean. Allegedly.
	if v, err := jsonparser.GetBoolean(data, "container", "oom_disabled"); err != nil {
		if err != jsonparser.KeyPathNotFoundError {
			return errors.WithStack(err)
		}
	} else {
		s.Container.OomDisabled = v
	}

	s.Container.RebuildRequired = true
	if _, err := s.WriteConfigurationToDisk(); err != nil {
		return errors.WithStack(err)
	}

	return s.Environment.InSituUpdate()
}