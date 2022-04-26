package btui

import (
	"fmt"
	"io"
	"log"
	"os/exec"
	"runtime"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-mastodon"
)

var timelineStyle = lipgloss.NewStyle() //.BorderStyle(lipgloss.NormalBorder())

const listHeight = 14

var (
	tootItemStyle     = lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()).Padding(0, 0, 0).Margin(0)
	selectedItemStyle = tootItemStyle.Copy().BorderForeground(lipgloss.Color("170"))
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	quitTextStyle     = lipgloss.NewStyle().Margin(1, 0, 2, 4)

	headerStyle  = lipgloss.NewStyle().Bold(true).BorderBottom(true).Align(lipgloss.Center)
	contentStyle = lipgloss.NewStyle().Align(lipgloss.Center)
)

type timelineKeyMap struct {
	refresh  key.Binding
	favorite key.Binding
	open     key.Binding
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
	open: key.NewBinding(
		key.WithKeys("o"),
		key.WithHelp("o", "open in browser"),
	),
}

type Timeline struct {
	list   list.Model
	cursor int
	Toots  []*mastodon.Status
	app    *App
	ttype  TimelineType
	// inputHandler *cbind.Configuration
}

type itemDelegate struct {
	del list.DefaultDelegate
}

func (d itemDelegate) Height() int                               { return 3 } // d.del.Height() }
func (d itemDelegate) Spacing() int                              { return d.del.Spacing() }
func (d itemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return d.del.Update(msg, m) }

func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	// perform default render
	var b strings.Builder
	d.del.Render(&b, m, index, listItem)

	// then apply our styles to the rendered item
	s := b.String()

	width := timelineStyle.GetWidth()

	fn := tootItemStyle.Copy().Width(width).Render
	if index == m.Index() {
		fn = func(s string) string {
			return selectedItemStyle.Copy().Width(width).Render(s)
		}
	}

	fmt.Fprintf(w, fn(s))
}

func NewTimeline(app *App, ttype TimelineType) *Timeline {
	items := []list.Item{}

	delegate := itemDelegate{
		del: list.NewDefaultDelegate(),
	}

	delegate.del.Styles.NormalTitle = headerStyle
	delegate.del.Styles.SelectedTitle = headerStyle

	delegate.del.Styles.NormalDesc = contentStyle.Inline((true))
	delegate.del.Styles.SelectedDesc = contentStyle

	t := &Timeline{
		list:  list.New(items, delegate, 0, 10),
		app:   app,
		ttype: ttype,
	}
	t.list.Title = ttype.String()
	t.list.SetHeight(timelineStyle.GetHeight())
	t.list.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			defaultTimelineKeyMap.favorite,
			defaultTimelineKeyMap.refresh,
			defaultTimelineKeyMap.open,
		}
	}
	t.list.Select(0)
	return t
}

func (m *Timeline) Init() tea.Cmd {
	return m.RefreshCmd
}

func (m *Timeline) handleKeyBinding(msg tea.KeyMsg) (model tea.Model, cmd tea.Cmd) {
	switch {
	case key.Matches(msg, defaultTimelineKeyMap.refresh):
		var cmds []tea.Cmd

		cmds = append(cmds, m.list.SetItems([]list.Item{}), m.list.NewStatusMessage("refreshing"), m.RefreshCmd)
		return m, tea.Batch(cmds...)

	case key.Matches(msg, defaultTimelineKeyMap.favorite):
		var cmds []tea.Cmd
		cmds = append(cmds, m.list.NewStatusMessage("favoriting toot"), m.FavoriteCmd, m.list.SetItems([]list.Item{}))
		return m, tea.Batch(cmds...)
	case key.Matches(msg, defaultTimelineKeyMap.open):
		var cmds []tea.Cmd
		cmds = append(cmds, m.list.NewStatusMessage("opening in browser"), m.OpenCmd)
		return m, tea.Batch(cmds...)
	}
	return m, nil
}

func (m *Timeline) Update(msg tea.Msg) (model tea.Model, cmd tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

		if !m.list.SettingFilter() {

			model, cmd = m.handleKeyBinding(msg)
			if cmd != nil {
				return model, cmd
			}
		}

	case tea.WindowSizeMsg:
		// h, v := timelineStyle.GetFrameSize()
		// m.list.SetSize(msg.Width-h, msg.Height-v)
		m.list.SetSize(msg.Width, msg.Height)

	case ErrorMsg:
		return m, m.list.NewStatusMessage(fmt.Sprintf("failed to %s: %s", msg.action, msg.msg))

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
	return timelineStyle.Render(m.list.View())
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

func (t *Timeline) OpenCmd() tea.Msg {
	toot := t.GetCurrentToot()
	if toot == nil {
		return ErrorMsg{action: "get toot", msg: "selected item not a toot"}
		return nil
	}
	status := toot.status
	openbrowser(status.URL)
	return nil
}

func (t *Timeline) FavoriteCmd() tea.Msg {
	toot := t.GetCurrentToot()
	if toot == nil {
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
