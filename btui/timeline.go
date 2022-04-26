package btui

import (
	"fmt"
	"io"
	"log"
	"os/exec"
	"runtime"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-mastodon"
)

var docStyle = lipgloss.NewStyle().Margin(1, 2)

const listHeight = 14

var (
	tootItemStyle     = lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()).Padding(0, 1, 0)
	selectedItemStyle = tootItemStyle.Copy().BorderForeground(lipgloss.Color("170"))
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	quitTextStyle     = lipgloss.NewStyle().Margin(1, 0, 2, 4)
)

type TimelineType int

type TimelineMsg struct {
	action string
	toots  []*mastodon.Status
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

type timelineKeyMap struct {
	refresh  key.Binding
	favorite key.Binding
	// Down key.Binding
}

var defaultTimelineKeyMap = timelineKeyMap{
	refresh: key.NewBinding(
		key.WithKeys("r"),                     // actual keybindings
		key.WithHelp("r", "refresh timeline"), // corresponding help text
	),
	favorite: key.NewBinding(
		key.WithKeys("f"),
		key.WithHelp("f", "favorite"),
	),
}

func (t TimelineType) String() string {
	return [...]string{"home", "local", "federated", "profile", "likes", "tags", "media"}[t]
}

type Timeline struct {
	list   list.Model
	cursor int
	Toots  []*mastodon.Status
	app    *App
	ttype  TimelineType
	// inputHandler *cbind.Configuration
}

type itemDelegate struct{}

func (d itemDelegate) Height() int                               { return 2 }
func (d itemDelegate) Spacing() int                              { return 0 }
func (d itemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }

func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(*Toot)
	if !ok {
		return
	}

	s := i.View()

	fn := tootItemStyle.Render
	if index == m.Index() {
		fn = func(s string) string {
			return selectedItemStyle.Render(s)
		}
	}

	fmt.Fprintf(w, fn(s))
}

func NewTimeline(app *App, ttype TimelineType) *Timeline {
	items := []list.Item{}
	t := &Timeline{
		list:  list.New(items, itemDelegate{}, 0, 0),
		app:   app,
		ttype: ttype,
	}
	t.list.Title = ttype.String()
	t.list.Select(0)
	return t
}

func (m *Timeline) Init() tea.Cmd {
	return m.RefreshCmd
}

func (m *Timeline) Update(msg tea.Msg) (model tea.Model, cmd tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		switch {
		case key.Matches(msg, defaultTimelineKeyMap.refresh):
			var cmds []tea.Cmd

			cmds = append(cmds, m.list.SetItems([]list.Item{}), m.list.NewStatusMessage("refreshing"), m.RefreshCmd)
			return m, tea.Batch(cmds...)

		case key.Matches(msg, defaultTimelineKeyMap.favorite):
			var cmds []tea.Cmd

			cmds = append(cmds, m.list.SetItems([]list.Item{}), m.list.NewStatusMessage("favoriting toot"), m.RefreshCmd)
			return m, tea.Batch(cmds...)
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)

	case TimelineMsg:

		items := []list.Item{}
		for _, toot := range m.Toots {
			tc := NewToot(m.app, toot)
			items = append(items, tc)
		}
		return m, tea.Batch(
			m.list.SetItems(items),
			m.list.NewStatusMessage("updating timeline"),
		)
	}

	var cmds []tea.Cmd
	m.list, cmd = m.list.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m *Timeline) View() string {
	// out := ""
	// for _, t := range m.Toots {
	// }
	return m.list.View()
}

func (t *Timeline) handleSelect(ev *tcell.EventKey) *tcell.EventKey {
	// toot := t.GetCurrentToot()
	// t.app.ViewThread(toot)

	return nil
}

func (t *Timeline) handleDelete(ev *tcell.EventKey) *tcell.EventKey {

	// toot := t.GetCurrentToot()
	// status := toot.status

	// if t.app.client.IsOwnStatus(status) {
	// 	t.app.Notify("Deleting toot!")
	// 	t.app.client.Delete(status)
	// 	t.app.FocusTimeline()
	// 	return nil
	// }
	return ev
}

func (t *Timeline) handleRefresh(ev *tcell.EventKey) *tcell.EventKey {
	// t.app.FocusTimeline()
	return nil

}

func (t *Timeline) handleFollow(ev *tcell.EventKey) *tcell.EventKey {
	toot := t.GetCurrentToot()
	status := toot.status

	// t.app.Notify("Following %s", status.Account.Acct)
	t.app.client.Follow(status.Account.ID)
	// t.app.FocusTimeline()
	return nil

}

func (t *Timeline) handleUnfollow(ev *tcell.EventKey) *tcell.EventKey {
	// toot := t.GetCurrentToot()
	// status := toot.status

	// t.app.Notify("Unfollowing %s", status.Account.Acct)
	// t.app.client.Unfollow(status.Account.ID)
	// t.app.FocusTimeline()
	return nil

}

func (t *Timeline) handleLike(ev *tcell.EventKey) *tcell.EventKey {
	toot := t.GetCurrentToot()
	// status := toot.status

	if toot.IsFavorite() {
		// t.app.Notify("Unliking toot!")
		// t.app.client.Unlike(status)
		// t.app.FocusTimeline()
		return nil
	}

	// t.app.Notify("Liking toot!")
	// t.app.client.Like(status)
	// t.app.FocusTimeline()
	return nil
}

func (t *Timeline) handleBoost(ev *tcell.EventKey) *tcell.EventKey {
	toot := t.GetCurrentToot()
	status := toot.status

	t.app.client.Boost(status)
	// t.app.FocusTimeline()
	return nil
}

func (t *Timeline) handleOpen(ev *tcell.EventKey) *tcell.EventKey {
	// t.app.Notify("Opening in browser")
	toot := t.GetCurrentToot()
	status := toot.status
	openbrowser(status.URL)
	return nil

}

func (t *Timeline) SetTimeline(ttype TimelineType) {
	t.ttype = ttype
}

func (t *Timeline) GetCurrentToot() *Toot {
	ref := t.list.SelectedItem()
	toot, ok := ref.(*Toot)
	if !ok {
		return nil
	}
	return toot
}
func (t *Timeline) SetCurrentToot(toot *Toot) {
	// for i, item := range t.AddItem() {
	// 	ref := item.GetReference()
	// 	tootc, ok := ref.(*Toot)
	// 	if !ok {
	// 		continue
	// 	}
	// 	if tootc.status.ID == toot.status.ID {
	// 		t.SetCurrentItem(i)
	// 	}
	// }
}

func (t *Timeline) FavoriteCmd() tea.Msg {
	ref := t.list.SelectedItem()
	toot, ok := ref.(*Toot)
	if !ok {
		return nil
	}
	status := toot.status

	if toot.IsFavorite() {
		t.app.client.Unlike(status)
		return tea.Batch(
			t.list.NewStatusMessage("unfavoriting"),
			t.RefreshCmd,
		)
	}

	t.app.client.Like(status)
	return tea.Batch(
		t.list.NewStatusMessage("favoriting toot!"),
		t.RefreshCmd,
	)
}

func (t *Timeline) RefreshCmd() tea.Msg {
	toots := t.app.client.GetTimeline(t.ttype.String())
	t.Toots = toots
	return TimelineMsg{
		toots: toots,
	}

}

func (t *Timeline) Refresh() {
	// selected := t.list.Cursor

	// toots := t.app.client.GetTimeline(t.ttype.String())
	// t.fillToots(toots)
	// title := fmt.Sprintf(" Timeline - %s ", strings.Title(t.ttype.String()))
	// t.SetTitle(title)
	// t.SetTitleColor(tcell.ColorLightCyan)
	// if selected != nil {
	// 	t.SetCurrentToot(selected)
	// }
}

func (t *Timeline) fillToots(toots []*mastodon.Status) {
	t.list.SetItems([]list.Item{})
	t.Toots = toots
	for i, toot := range t.Toots {
		tc := NewToot(t.app, toot)
		t.list.InsertItem(i, tc)
	}

}

func openbrowser(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		log.Fatal(err)
	}

}
