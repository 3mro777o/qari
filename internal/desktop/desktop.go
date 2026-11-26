// Package desktop connects qari to the desktop: media keys and bar widgets
// through MPRIS, and a notification when a new surah starts. It is a no-op
// outside Linux.
package desktop

// Cmd is a request from the desktop, such as a media key press. Name is one
// of playpause, play, pause, next, previous, stop, seek (Arg = relative
// seconds), setpos (Arg = absolute seconds) or quit.
type Cmd struct {
	Name string
	Arg  float64
}

// State is what the desktop is told about playback.
type State struct {
	Status           string // Playing, Paused or Stopped
	Title, Artist    string
	Album            string
	Length, Pos      float64 // seconds
	Rate, Volume     float64
	CanNext, CanPrev bool
}
