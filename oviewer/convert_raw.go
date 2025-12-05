package oviewer

type rawConverter struct{}

func newRawConverter() *rawConverter {
	return &rawConverter{}
}

// convert converts only escape sequence codes to display characters and returns them as is.
// Returns true if it is an escape sequence and a non-printing character.
func (rawConverter) convert(st *parseState) bool {
	if st.str != "\x1b" {
		return false
	}
	// ESC -> '^', '['
	c := DefaultContent
	c.str = "^"
	st.lc = append(st.lc, c)
	c.str = "["
	st.lc = append(st.lc, c)
	return true
}
