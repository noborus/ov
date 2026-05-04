package oviewer

import (
	"bytes"
	"context"
	"io"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/gdamore/tcell/v3"
	"github.com/gdamore/tcell/v3/color"
	"github.com/gdamore/tcell/v3/vt"
	"github.com/spf13/viper"
)

const cwd = ".."

var testdata = filepath.Join(cwd, "testdata")

// fakeScreen returns a fake screen.
func fakeScreen() (tcell.Screen, error) {
	// width, height := 80, 25
	mt := vt.NewMockTerm(vt.MockOptSize{X: 80, Y: 25})
	return tcell.NewTerminfoScreenFromTty(mt)
}

func rootHelper(t *testing.T) *Root {
	t.Helper()
	tcellNewScreen = fakeScreen
	defer func() {
		tcellNewScreen = tcell.NewScreen
	}()
	root, err := NewRoot(bytes.NewBufferString("test"))
	if err != nil {
		t.Fatal(err)
	}
	setFixedTestScreen(t, root)
	return root
}

func rootFileReadHelper(t *testing.T, fileNames ...string) *Root {
	t.Helper()
	root, err := Open(fileNames...)
	if err != nil {
		t.Fatal(err)
	}
	setFixedTestScreen(t, root)
	root.mu.RLock()
	for _, doc := range root.DocList {
		doc.WaitEOF()
	}
	root.mu.RUnlock()
	return root
}

func setFixedTestScreen(t *testing.T, root *Root) {
	t.Helper()
	screen, err := virtualScreen(80, 25)
	if err != nil {
		t.Fatal(err)
	}
	root.Screen = screen
}

func TestNewOviewer(t *testing.T) {
	tcellNewScreen = fakeScreen
	defer func() {
		tcellNewScreen = tcell.NewScreen
	}()
	type args struct {
		docs []*Document
	}
	tests := []struct {
		name    string
		args    args
		want    *Root
		wantErr bool
	}{
		{
			name:    "testEmpty",
			args:    args{},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewOviewer(tt.args.docs...)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewOviewer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewOviewer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOpen(t *testing.T) {
	tcellNewScreen = fakeScreen
	defer func() {
		tcellNewScreen = tcell.NewScreen
	}()
	type args struct {
		fileNames []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test1",
			args: args{
				fileNames: []string{filepath.Join(testdata, "test.txt")},
			},
			wantErr: false,
		},
		{
			name: "test2",
			args: args{
				fileNames: []string{
					filepath.Join(testdata, "test.txt"),
					filepath.Join(testdata, "test2.txt"),
				},
			},
			wantErr: false,
		},
		{
			name: "testErr",
			args: args{
				fileNames: []string{filepath.Join(testdata, "err.txt")},
			},
			wantErr: true,
		},
		{
			name: "testDir",
			args: args{
				fileNames: []string{testdata},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root, err := Open(tt.args.fileNames...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Open() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			root.Quit(context.Background())
		})
	}
}

func TestNewRoot(t *testing.T) {
	tcellNewScreen = fakeScreen
	defer func() {
		tcellNewScreen = tcell.NewScreen
	}()
	type args struct {
		read io.Reader
	}
	tests := []struct {
		name    string
		args    args
		want    *Root
		wantErr bool
	}{
		{
			name:    "test1",
			args:    args{read: bytes.NewBufferString("test")},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewRoot(tt.args.read)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewRoot() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestRoot_Run(t *testing.T) {
	tcellNewScreen = fakeScreen
	defer func() {
		tcellNewScreen = tcell.NewScreen
	}()
	tests := []struct {
		name    string
		general RunTimeSettings
		ovArgs  []string
		wantErr bool
	}{
		{
			name: "testEmpty",
			general: RunTimeSettings{
				TabWidth:       8,
				MarkStyleWidth: 1,
			},
			ovArgs:  []string{},
			wantErr: false,
		},
		{
			name: "test1",
			general: RunTimeSettings{
				TabWidth:       8,
				MarkStyleWidth: 1,
			},
			ovArgs:  []string{filepath.Join(testdata, "test.txt")},
			wantErr: false,
		},
		{
			name: "testHeader",
			general: RunTimeSettings{
				TabWidth:       8,
				MarkStyleWidth: 1,
				Header:         1,
			},
			ovArgs:  []string{filepath.Join(testdata, "test.txt")},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root, err := Open(tt.ovArgs...)
			if err != nil {
				t.Fatalf("NewOviewer error = %v", err)
			}

			done := make(chan error, 1)
			go func() {
				done <- root.Run()
			}()

			root.Quit(context.Background())

			runErr := <-done
			if (runErr != nil) != tt.wantErr {
				t.Errorf("Root.Run() error = %v, wantErr %v", runErr, tt.wantErr)
			}
		})
	}
}

func Test_applyStyle(t *testing.T) {
	t.Parallel()
	type args struct {
		style tcell.Style
		s     OVStyle
	}
	tests := []struct {
		name string
		args args
		want tcell.Style
	}{
		{
			name: "test1",
			args: args{
				style: tcell.StyleDefault,
				s: OVStyle{
					Background: "red",
					Foreground: "white",
				},
			},
			want: tcell.StyleDefault.Foreground(color.White).Background(color.Red),
		},
		{
			name: "test2",
			args: args{
				style: tcell.StyleDefault,
				s: OVStyle{
					UnDim: true,
				},
			},
			want: tcell.StyleDefault.Dim(false),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := applyStyle(tt.args.style, tt.args.s); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("applyStyle() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRoot_setKeyConfig(t *testing.T) {
	tcellNewScreen = fakeScreen
	defer func() {
		tcellNewScreen = tcell.NewScreen
	}()
	tests := []struct {
		name    string
		cfgFile string
		want    []string
		wantErr bool
	}{
		{
			name:    "test-ov.yaml",
			cfgFile: filepath.Join(cwd, "ov.yaml"),
			want:    []string{"Enter", "Down", "ctrl+n"},
			wantErr: false,
		},
		{
			name:    "test-ov-less.yaml",
			cfgFile: filepath.Join(cwd, "ov-less.yaml"),
			want:    []string{"e", "ctrl+e", "j", "J", "ctrl+j", "Enter", "Down"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root, err := Open(filepath.Join(testdata, "test.txt"))
			if err != nil {
				t.Fatalf("NewOviewer error = %v", err)
			}
			mt := vt.NewMockTerm(vt.MockOptSize{X: 80, Y: 25})
			scr, err := tcell.NewTerminfoScreenFromTty(mt)
			if err != nil {
				t.Fatalf("NewTerminfoScreenFromTty error = %v", err)
			}
			root.Screen = scr
			viper.SetConfigFile(tt.cfgFile)
			var config Config
			viper.AutomaticEnv() // read in environment variables that match
			if err := viper.ReadInConfig(); err != nil {
				t.Fatal("failed to read config file:", err)
			}
			if err := viper.Unmarshal(&config); err != nil {
				t.Fatal("failed to unmarshal config:", err)
			}
			root.SetConfig(config)
			got, err := root.setKeyConfig(context.Background())
			if (err != nil) != tt.wantErr {
				t.Errorf("Root.setKeyConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			action := "down"
			if !reflect.DeepEqual(got[action], tt.want) {
				t.Errorf("Root.setKeyConfig() = %v, want %v", got[action], tt.want)
			}
		})
	}
}

func TestSetConfig_WithViewMode(t *testing.T) {
	// prepare config with a view mode that sets FollowAll and TabWidth
	cfg := NewConfig()
	if cfg.Mode == nil {
		cfg.Mode = make(map[string]General)
	}
	var g General
	g.SetFollowAll(true)
	g.SetTabWidth(4)
	cfg.Mode["myview"] = g
	cfg.ViewMode = "myview"
	cfg.MinStartX = 12

	root := &Root{settings: NewRunTimeSettings()}
	root.SetConfig(cfg)

	if !root.settings.FollowAll {
		t.Fatalf("expected settings.FollowAll=true, got false")
	}
	if !root.FollowAll {
		t.Fatalf("expected root.FollowAll=true, got false")
	}
	if root.settings.TabWidth != 4 {
		t.Fatalf("expected TabWidth=4, got %d", root.settings.TabWidth)
	}
	if root.minStartX != 12 {
		t.Fatalf("expected minStartX=12, got %d", root.minStartX)
	}
}

func TestRoot_writeOriginal(t *testing.T) {
	tcellNewScreen = fakeScreen
	defer func() {
		tcellNewScreen = tcell.NewScreen
	}()
	type fields struct {
		header             int
		topLN              int
		sectionDelimiter   string
		sectionNum         int
		AfterWriteOriginal int
	}
	type args struct {
		fileNames []string
	}
	type want struct {
		output string
	}
	tests := []struct {
		name   string
		args   args
		fields fields
		want   want
	}{
		{
			name: "test1",
			fields: fields{
				topLN:              0,
				header:             0,
				sectionDelimiter:   "",
				sectionNum:         0,
				AfterWriteOriginal: 3,
			},
			args: args{
				fileNames: []string{filepath.Join(testdata, "test.txt")},
			},
			want: want{
				output: "test\n",
			},
		},
		{
			name: "test3-1",
			fields: fields{
				topLN:              4,
				header:             0,
				sectionDelimiter:   "1",
				sectionNum:         1,
				AfterWriteOriginal: 4,
			},
			args: args{
				fileNames: []string{filepath.Join(testdata, "test3.txt")},
			},
			want: want{
				output: "1\n6\n7\n8\n",
			},
		},
		{
			name: "test3-2",
			fields: fields{
				topLN:              0,
				header:             3,
				sectionDelimiter:   "2",
				sectionNum:         1,
				AfterWriteOriginal: 4,
			},
			args: args{
				fileNames: []string{filepath.Join(testdata, "test3.txt")},
			},
			want: want{
				output: "1\n2\n3\n4\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := rootFileReadHelper(t, tt.args.fileNames...)
			root.Doc.Header = tt.fields.header
			root.Doc.topLN = tt.fields.topLN
			root.setSectionDelimiter(tt.fields.sectionDelimiter)
			if tt.fields.sectionNum > 0 {
				root.Doc.SectionHeaderNum = tt.fields.sectionNum
				root.Doc.SectionHeader = true
			}
			root.prepareScreen()
			root.Config.AfterWriteOriginal = tt.fields.AfterWriteOriginal
			output := &bytes.Buffer{}
			root.writeOriginal(output)
			if gotOutput := output.String(); gotOutput != tt.want.output {
				t.Errorf("Root.writeOriginal() = %v, want %v", gotOutput, tt.want.output)
			}
		})
	}
}

func TestRoot_docSmall(t *testing.T) {
	tcellNewScreen = fakeScreen
	defer func() {
		tcellNewScreen = tcell.NewScreen
	}()
	type fields struct{}
	type args struct {
		fileNames []string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "test1",
			args: args{
				fileNames: []string{filepath.Join(testdata, "test.txt")},
			},
			want: true,
		},
		{
			name: "test3",
			args: args{
				fileNames: []string{filepath.Join(testdata, "test3.txt")},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := rootFileReadHelper(t, tt.args.fileNames...)
			if got := root.docSmall(); got != tt.want {
				t.Errorf("Root.docSmall() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRoot_setCaption(t *testing.T) {
	type fields struct {
		caption string
		manpn   string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "testnil",
			fields: fields{
				caption: "",
				manpn:   "",
			},
			want: "",
		},
		{
			name: "test1",
			fields: fields{
				caption: "test",
				manpn:   "",
			},
			want: "test",
		},
		{
			name: "testMan",
			fields: fields{
				caption: "",
				manpn:   "man",
			},
			want: "man",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("MAN_PN", tt.fields.manpn)
			root := rootHelper(t)
			root.settings.Caption = tt.fields.caption
			root.setCaption()
			if got := root.Doc.Caption; got != tt.want {
				t.Errorf("Root.setCaption() = %v, want %v", got, "test")
			}
		})
	}
}

func TestRoot_setViewModeConfig(t *testing.T) {
	type fields struct {
		viewMode map[string]General
	}
	tests := []struct {
		name     string
		fields   fields
		wantList []string
	}{
		{
			name: "test1",
			fields: fields{
				viewMode: map[string]General{
					"view1": {},
				},
			},
			wantList: []string{nameGeneral, "view1"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := rootHelper(t)
			root.Config.Mode = tt.fields.viewMode
			root.setViewModeConfig()
			if !reflect.DeepEqual(root.input.Candidate[ViewMode].list, tt.wantList) {
				t.Errorf("Root.setViewModeConfig() = %v, want %v", root.input.Candidate[ViewMode].list, tt.wantList)
			}
		})
	}
}

func TestRoot_prepareRun(t *testing.T) {
	type fields struct {
		QuitSmall       bool
		QuitSmallFilter bool
	}
	tests := []struct {
		name     string
		fields   fields
		wantErr  bool
		wantQuit bool
	}{
		{
			name: "test1",
			fields: fields{
				QuitSmall:       false,
				QuitSmallFilter: false,
			},
			wantErr:  false,
			wantQuit: false,
		},
		{
			name: "test2",
			fields: fields{
				QuitSmall:       true,
				QuitSmallFilter: true,
			},
			wantErr:  false,
			wantQuit: false,
		},
		{
			name: "test2",
			fields: fields{
				QuitSmall:       true,
				QuitSmallFilter: false,
			},
			wantErr:  false,
			wantQuit: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := rootHelper(t)
			root.Config.QuitSmall = tt.fields.QuitSmall
			root.Config.QuitSmallFilter = tt.fields.QuitSmallFilter
			if err := root.prepareRun(context.Background()); (err != nil) != tt.wantErr {
				t.Errorf("Root.prepareRun() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got := root.Config.QuitSmall; got != tt.wantQuit {
				t.Errorf("Root.prepareRun() = %v, want %v", got, tt.wantQuit)
			}
		})
	}
}

func TestRoot_outputOnExit(t *testing.T) {
	tcellNewScreen = fakeScreen
	defer func() {
		tcellNewScreen = tcell.NewScreen
	}()
	type fields struct {
		fileName        string
		IsWriteOnExit   bool
		IsWriteOriginal bool
	}
	tests := []struct {
		name       string
		fields     fields
		wantOutput string
	}{
		{
			name: "no output",
			fields: fields{
				fileName:        filepath.Join(testdata, "test.txt"),
				IsWriteOnExit:   false,
				IsWriteOriginal: false,
			},
			wantOutput: "",
		},
		{
			name: "writeCurrentScreen",
			fields: fields{
				fileName:        filepath.Join(testdata, "test.txt"),
				IsWriteOnExit:   true,
				IsWriteOriginal: false,
			},
			wantOutput: "test\n",
		},
		{
			name: "writeCurrentScreen_test4",
			fields: fields{
				fileName:        filepath.Join(testdata, "test4.txt"),
				IsWriteOnExit:   true,
				IsWriteOriginal: false,
			},
			wantOutput: "\x1b[38;2;255;175;135m\x1b[1mHello\033[0m\n",
		},
		{
			name: "writeOriginal",
			fields: fields{
				fileName:        filepath.Join(testdata, "test.txt"),
				IsWriteOnExit:   true,
				IsWriteOriginal: true,
			},
			wantOutput: "test\n",
		},
		{
			name: "writeLog when Debug is true",
			fields: fields{
				fileName:        filepath.Join(testdata, "test.txt"),
				IsWriteOnExit:   false,
				IsWriteOriginal: false,
			},
			wantOutput: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := rootFileReadHelper(t, tt.fields.fileName)
			root.prepareScreen()
			root.Config.IsWriteOnExit = tt.fields.IsWriteOnExit
			root.Config.IsWriteOriginal = tt.fields.IsWriteOriginal
			buf := &bytes.Buffer{}
			root.outputOnExit(buf)
			gotOutput := buf.String()
			if gotOutput != tt.wantOutput {
				t.Errorf("Root.outputOnExit() = \n%v, want \n%v", []byte(gotOutput), []byte(tt.wantOutput))
			}
		})
	}
}

func Test_openFiles_AllFail(t *testing.T) {
	tcellNewScreen = fakeScreen
	defer func() { tcellNewScreen = tcell.NewScreen }()

	files := []string{filepath.Join(testdata, "err.txt")}
	root, err := openFiles(files)
	if root != nil {
		t.Fatalf("expected nil root, got %v", root)
	}
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func Test_openFiles_PartialSuccess(t *testing.T) {
	tcellNewScreen = fakeScreen
	defer func() { tcellNewScreen = tcell.NewScreen }()

	files := []string{
		filepath.Join(testdata, "test.txt"),
		filepath.Join(testdata, "err.txt"),
	}
	root, err := openFiles(files)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if root == nil {
		t.Fatalf("expected root, got nil")
	}
	if len(root.DocList) != 1 {
		t.Fatalf("expected 1 document, got %d", len(root.DocList))
	}
}

func Test_openFiles_AllSuccess(t *testing.T) {
	tcellNewScreen = fakeScreen
	defer func() { tcellNewScreen = tcell.NewScreen }()

	files := []string{
		filepath.Join(testdata, "test.txt"),
		filepath.Join(testdata, "test2.txt"),
	}
	root, err := openFiles(files)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if root == nil {
		t.Fatalf("expected root, got nil")
	}
	if len(root.DocList) != 2 {
		t.Fatalf("expected 2 documents, got %d", len(root.DocList))
	}
}
