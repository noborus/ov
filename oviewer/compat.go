package oviewer

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/viper"
)

// compatEntry holds the mapping between an existing internal config key name
// and its new canonical name, plus display names for deprecation messages.
type compatEntry struct {
	oldKey  string // existing viper key (lowercase, no change)
	newKey  string // new canonical viper key (lowercase)
	oldName string // human-readable old key for deprecation message
	newName string // human-readable new key for deprecation message
}

// compatViperEntries enumerates every config key rename from the naming audit.
// oldKey is the existing internal name that stays unchanged; newKey is the new
// name being introduced. Deprecation warnings fire on oldKey; aliases are
// registered new→old so either name works in config files and env vars.
var compatViperEntries = []compatEntry{
	// A2: QuitSmall → QuitIfOneScreen (matches CLI --quit-if-one-screen)
	{"quitsmall", "quitifonescreen", "QuitSmall", "QuitIfOneScreen"},
	// A3: LineNumMode → LineNumberMode (abbreviation → full word)
	{"general.linenumode", "general.linenumbermode", "general.LineNumMode", "general.LineNumberMode"},
	// A5/B6: IsWriteOnExit → WriteOnExit (drop Go Is-prefix from YAML key)
	{"iswriteonexit", "writeonexit", "IsWriteOnExit", "WriteOnExit"},
	// B6: IsWriteOriginal → WriteOriginal (drop Go Is-prefix from YAML key)
	{"iswriteoriginal", "writeoriginal", "IsWriteOriginal", "WriteOriginal"},
	// A6: BeforeWriteOriginal → WriteContextBefore (drop impl-detail "Original")
	{"beforewriteoriginal", "writecontextbefore", "BeforeWriteOriginal", "WriteContextBefore"},
	// A6: AfterWriteOriginal → WriteContextAfter (drop impl-detail "Original")
	{"afterwriteoriginal", "writecontextafter", "AfterWriteOriginal", "WriteContextAfter"},
	// B1: VerticalHeader → RowHeader (misnomer: freezes row-identifying content, not column labels)
	{"general.verticalheader", "general.rowheader", "general.VerticalHeader", "general.RowHeader"},
	// B1: HeaderColumn → RowHeaderColumn (paired rename with RowHeader)
	{"general.headercolumn", "general.rowheadercolumn", "general.HeaderColumn", "general.RowHeaderColumn"},
	// B1: Style.VerticalHeader → Style.RowHeader
	{"general.style.verticalheader", "general.style.rowheader", "general.style.VerticalHeader", "general.style.RowHeader"},
	// B1: Style.VerticalHeaderBorder → Style.RowHeaderBorder
	{"general.style.verticalheaderborder", "general.style.rowheaderborder", "general.style.VerticalHeaderBorder", "general.style.RowHeaderBorder"},
	// B4: ColumnWidth → ColumnFixedWidth (clarify it switches detection method, not column width value)
	{"general.columnwidth", "general.columnfixedwidth", "general.ColumnWidth", "general.ColumnFixedWidth"},
	// B5: Incsearch → IncrementalSearch (sole abbreviation among fully-spelled keys)
	{"incsearch", "incrementalsearch", "Incsearch", "IncrementalSearch"},
	// B8: HScrollWidth → HScrollStep (it is a step size, not a display width)
	{"general.hscrollwidth", "general.hscrollstep", "general.HScrollWidth", "general.HScrollStep"},
}

// compatActionEntries maps new canonical action names to existing internal ones.
// Users may bind either name in the keybind section of their config.
var compatActionEntries = map[string]string{
	// A1: alter_rows_mode → alternate_rows_mode (truncated word → full word)
	"alternate_rows_mode": actionAlternate,
	// A4: rainbow_mode → column_rainbow_mode (disambiguate from multi-color search highlight)
	"column_rainbow_mode": actionRainbow,
	// A5: write_exit → write_on_exit (match CLI --exit-write and YAML WriteOnExit)
	"write_on_exit": actionWriteExit,
	// A8: hide_other → hide_other_section (drop ambiguous truncation)
	"hide_other_section": actionHideOther,
	// B1: vertical_header → row_header (paired with config key rename)
	"row_header": actionVerticalHeader,
	// B1: header_column → row_header_column (paired with config key rename)
	"row_header_column": actionHeaderColumn,
	// D3: set_write_exit → write_exit_range (describes what it does: configure line range)
	"write_exit_range": actionWriteBA,
	// D4: logdoc → debug_log (opaque internal name → intent-revealing name)
	"debug_log": actionLogDoc,
}

// RegisterCompatAliases registers new canonical config key names as viper aliases
// for existing internal names and emits deprecation notices for any old names
// found in the loaded configuration. Must be called after viper.ReadInConfig
// and before viper.Unmarshal.
func RegisterCompatAliases(v *viper.Viper) {
	warnDeprecatedKeys(v)
	for _, e := range compatViperEntries {
		v.RegisterAlias(e.newKey, e.oldKey)
	}
}

// warnDeprecatedKeys writes a deprecation notice to stderr for each old config
// key name detected in the loaded configuration (file, env var, or CLI flag).
// Called before alias registration so IsSet reflects direct usage only.
func warnDeprecatedKeys(v *viper.Viper) {
	for _, e := range compatViperEntries {
		if v.IsSet(e.oldKey) {
			fmt.Fprintf(os.Stderr, "ov: config key %q is deprecated, use %q instead\n", e.oldName, e.newName)
		}
	}
}

// addCompatActionAliases injects new canonical action names into the handler
// map, pointing each to the handler registered under the existing internal name.
func addCompatActionAliases(handlers map[string]func(context.Context)) {
	for newName, oldName := range compatActionEntries {
		if h, ok := handlers[oldName]; ok {
			handlers[newName] = h
		}
	}
}
