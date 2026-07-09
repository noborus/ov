package oviewer

import (
	"context"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/gdamore/tcell/v3"
)

func TestRoot_MoveLine(t *testing.T) {
	type args struct {
		num int
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "move line",
			args: args{
				num: 1,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := rootHelper(t)
			root.MoveLine(tt.args.num)
		})
	}
}

func TestRoot_SetDocument(t *testing.T) {
	type args struct {
		docNum int
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "valid document number",
			args: args{
				docNum: 0,
			},
			wantErr: false,
		},
		{
			name: "invalid document number - negative",
			args: args{
				docNum: -1,
			},
			wantErr: true,
		},
		{
			name: "invalid document number - out of range",
			args: args{
				docNum: 100,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := rootHelper(t)
			if err := root.SetDocument(tt.args.docNum); (err != nil) != tt.wantErr {
				t.Errorf("Root.SetDocument() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRoot_event(t *testing.T) {
	type args struct {
		ev tcell.Event
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "event app quit",
			args: args{
				ev: &eventAppQuit{},
			},
			want: true,
		},
		{
			name: "event update end num",
			args: args{
				ev: &eventUpdateEndNum{},
			},
			want: false,
		},
		{
			name: "event document",
			args: args{
				ev: &eventDocument{},
			},
			want: false,
		},
		{
			name: "event resize",
			args: args{
				ev: tcell.NewEventResize(20, 20),
			},
			want: false,
		},
		{
			name: "event mouse",
			args: args{
				ev: tcell.NewEventMouse(0, 0, tcell.Button1, tcell.ModNone),
			},
			want: false,
		},
		{
			name: "event key",
			args: args{
				ev: tcell.NewEventKey(tcell.KeyRune, "a", tcell.ModNone),
			},
			want: false,
		},
		{
			name: "event nil",
			args: args{
				ev: nil,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := rootHelper(t)
			ctx := context.Background()
			if got := root.event(ctx, tt.args.ev); got != tt.want {
				t.Errorf("Root.event() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRoot_loadEventQueue_doesNotDropQueuedEvents(t *testing.T) {
	root := rootHelper(t)
	for {
		select {
		case <-root.Screen.EventQ():
		default:
			goto drained
		}
	}

drained:
	root.screenState.Store(int32(ScreenStateTransition))
	root.eventQueue = []tcell.Event{
		tcell.NewEventKey(tcell.KeyF1, "", tcell.ModNone),
	}

	root.eventMu.Lock()
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		root.postEvent(tcell.NewEventKey(tcell.KeyF2, "", tcell.ModNone))
	}()
	go func() {
		defer wg.Done()
		root.postEvent(tcell.NewEventKey(tcell.KeyF3, "", tcell.ModNone))
	}()
	root.eventMu.Unlock()

	wg.Wait()

	got := make(map[string]bool)
	for i := 0; i < 3; i++ {
		select {
		case ev := <-root.Screen.EventQ():
			key, ok := ev.(*tcell.EventKey)
			if !ok {
				t.Fatalf("unexpected event type %T", ev)
			}
			got[key.Name()] = true
		case <-time.After(time.Second):
			t.Fatal("timed out waiting for posted events")
		}
	}

	want := map[string]bool{"F1": true, "F2": true, "F3": true}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("posted events = %#v, want %#v", got, want)
	}

	if len(root.eventQueue) != 0 {
		t.Fatalf("eventQueue length = %d, want 0", len(root.eventQueue))
	}
	if gotState := root.screenState.Load(); gotState != int32(ScreenStateReady) {
		t.Fatalf("screenState = %d, want %d", gotState, ScreenStateReady)
	}
}
