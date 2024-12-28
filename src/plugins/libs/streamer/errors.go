package streamer

type StreamerStoppedError struct{}

var ErrStreamerStopped = &StreamerStoppedError{}

func (e *StreamerStoppedError) Error() string {
	return "streamer has been stopped"
}

func (e *StreamerStoppedError) Is(target error) bool {
	_, ok := target.(*StreamerStoppedError)
	return ok
}
