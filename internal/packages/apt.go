package packages

// NewDeb constructs a new Deb instance.
func NewDeb(name string) *Deb {
	return &Deb{Name: name}
}

// Deb is a simple representation of a package installed from the Ubuntu archive.
type Deb struct {
	Name string
}
