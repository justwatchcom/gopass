package gpg

// Identity is a GPG identity, one key can have many IDs
type Identity struct {
	Name    string
	Comment string
	Email   string
}

// ID returns the GPG ID format
func (i Identity) ID() string {
	out := i.Name
	if i.Comment != "" {
		out += " (" + i.Comment + ")"
	}
	out += " <" + i.Email + ">"
	return out
}

// String implement fmt.Stringer. This method resembels the output gpg uses
// for user-ids
func (i Identity) String() string {
	return "uid                            " + i.ID()
}
