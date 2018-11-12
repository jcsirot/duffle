package repo

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"os"

	"github.com/Masterminds/semver"
)

var (
	// ErrNoAPIVersion indicates that an API version was not specified.
	ErrNoAPIVersion = errors.New("no API version specified")
	// ErrNoBundleVersion indicates that a bundle with the given version is not found.
	ErrNoBundleVersion = errors.New("no bundle with the given version found")
	// ErrNoBundleName indicates that a bundle with the given name is not found.
	ErrNoBundleName = errors.New("no bundle name found")
)

// Index defines a list of bundle repositories, each repository's respective tags and the digest reference.
type Index map[string]map[string]string

// LoadIndex takes a file at the given path and returns an Index object
func LoadIndex(path string) (Index, error) {
	f, err := os.OpenFile(path, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return loadIndex(f)
}

// LoadIndexReader takes a reader and returns an Index object
func LoadIndexReader(r io.Reader) (Index, error) {
	return loadIndex(r)
}

// LoadIndexBuffer reads repository metadata from a JSON byte stream
func LoadIndexBuffer(data []byte) (Index, error) {
	return loadIndex(bytes.NewBuffer(data))
}

// Add adds a new entry to the index
func (i Index) Add(name, version string, digest string) {
	if tags, ok := i[name]; ok {
		tags[version] = digest
	} else {
		i[name] = map[string]string{
			version: digest,
		}
	}
}

// Has returns true if the index has an entry for a bundle with the given name and exact version.
func (i Index) Has(name, version string) bool {
	_, err := i.Get(name, version)
	return err == nil
}

// Get returns the digest for the given name.
//
// If version is empty, this will return the digest for the bundle with the highest version.
func (i Index) Get(name, version string) (string, error) {
	vs, ok := i[name]
	if !ok {
		return "", ErrNoBundleName
	}
	if len(vs) == 0 {
		return "", ErrNoBundleVersion
	}

	var constraint *semver.Constraints
	if len(version) == 0 {
		constraint, _ = semver.NewConstraint("*")
	} else {
		var err error
		constraint, err = semver.NewConstraint(version)
		if err != nil {
			return "", err
		}
	}

	for ver, digest := range vs {
		test, err := semver.NewVersion(ver)
		if err != nil {
			continue
		}

		if constraint.Check(test) {
			return digest, nil
		}
	}
	return "", ErrNoBundleVersion
}

// WriteFile writes an index file to the given destination path.
//
// The mode on the file is set to 'mode'.
func (i Index) WriteFile(dest string, mode os.FileMode) error {
	b, err := json.MarshalIndent(i, "", "    ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(dest, b, mode)
}

// Merge merges the src index into i (dest).
//
// This merges by name and version.
//
// If one of the entries in the destination index does _not_ already exist, it is added.
// In all other cases, the existing record is preserved.
func (i *Index) Merge(src Index) {
	for name, versionMap := range src {
		for version, digest := range versionMap {
			if !i.Has(name, version) {
				i.Add(name, version, digest)
			}
		}
	}
}

// loadIndex loads an index file and does minimal validity checking.
func loadIndex(r io.Reader) (Index, error) {
	i := Index{}
	if err := json.NewDecoder(r).Decode(&i); err != nil && err != io.EOF {
		return i, err
	}
	return i, nil
}