package btui

import "github.com/mattn/go-mastodon"

type TimelineType int

type TimelineMsg struct {
	action string
	toots  []*mastodon.Status
}

type ErrorMsg struct {
	action string
	msg    string
}

const (
	TimelineHome TimelineType = iota
	TimelineLocal
	TimelineFederated
	TimelineProfile
	TimelineLikes
	TimelineTag
	TimelineMedia

	TimelineTootContext
)

var TimelineTypes = []TimelineType{
	TimelineHome,
	TimelineLocal,
	TimelineFederated,
	TimelineProfile,
	TimelineLikes,
	TimelineTag,
	TimelineMedia,
}

func (t TimelineType) String() string {
	return [...]string{"home", "local", "federated", "profile", "likes", "tags", "media"}[t]
}
