package oviewer

type rawConverter struct{}

func newRawConverter() *rawConverter {
	return &rawConverter{}
}

// convert converts only escape sequence codes to display characters and returns them as is.
// Returns true if it is an escape sequence and a non-printing character.
func (rawConverter) convert(st *parseState) bool {
	if st.mainc != 0x1b {
		return false
	}
	// ESC -> '^', '['
	c := DefaultContent
	c.mainc = '^'
	st.lc = append(st.lc, c)
	c.mainc = '['
	st.lc = append(st.lc, c)
	return true
}
