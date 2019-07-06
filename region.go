package lol

import "errors"

var (
	// ErrNoSuchRegion is returned if region not found.
	ErrNoSuchRegion = errors.New("No such region")
)

// Region represents a league of legends service area.
//
// Number is defined based on launched date.
// If two servers launched at same date, alphabet must be used for sort.
// Example: Global=1, PBE = 2, NA = 10 ....
type Region int32

// RegionByName gets a region by name.
func RegionByName(name string) (Region, error) {
	r, ok := regionByName[name]
	if !ok {
		return 0, ErrNoSuchRegion
	}
	return r, nil
}

// RegionByPlatformID gets a region by platform id.
func RegionByPlatformID(id string) (Region, error) {
	r, ok := regionByPlatformID[id]
	if !ok {
		return 0, ErrNoSuchRegion
	}
	return r, nil
}

// Name returns name of the region.
func (r Region) Name() string {
	n, _ := nameByRegion[r]
	return n
}

// IsGlobal returns true if this region is global.
func (r Region) IsGlobal() bool {
	return r == Global
}

// String implements fmt.Stringer
func (r Region) String() string {
	return r.Name()
}

// PlatformID returns the ID for observer api.
func (r Region) PlatformID() string {
	p, _ := platformIDByRegion[r]
	return p
}

// Host returns hostname for api call.
func (r Region) Host() string {
	p, _ := hostByRegion[r]
	return p
}

func (r Region) baseURL() string {
	return "https://" + r.Host()
}

// Regions returns all region except 'Global'
func Regions() []Region {
	return regions
}
