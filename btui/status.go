package btui

import (
	"fmt"
	"image/color"
	"log"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/eliukblau/pixterm/pkg/ansimage"
	"github.com/kyokomi/emoji"
	"github.com/mattn/go-mastodon"
	"gitlab.com/tslocum/cview"
)

var statusStyle = lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder())

type Status struct {
	status *mastodon.Status
	vp     viewport.Model
	app    *App
}

// func formatContent(html string) string {
// 	converter := md.NewConverter("", true, nil)

// 	mdContent, err := converter.ConvertString(html)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	g, err := glamour.NewTermRenderer(glamour.WithAutoStyle())
// 	if err != nil {
// 		return mdContent
// 	}

// 	out, err := g.Render(mdContent)
// 	if err != nil {
// 		return mdContent
// 	}
// 	return out

// }

type StatusMsg struct {
	status *mastodon.Status
}

func (m *Status) IsFavorite() bool {
	favorited, ok := m.status.Favourited.(bool)
	if !ok {
		return false
	}
	return favorited

}

func (m *Status) header() string {
	header := m.status.Account.DisplayName

	if m.IsFavorite() {
		header += emoji.Sprint(" :heart:")
	} else {
		header += emoji.Sprint(" :white_heart:")
	}

	if m.status.Reblog != nil {
		header = emoji.Sprintf("%s  || :repeat_button:@%s", header, m.status.Reblog.Account.DisplayName)
	}
	return header
}

func (t *Status) Content() string {
	return formatContent(t.status.Content)
}

func NewStatus(app *App, status *mastodon.Status) *Status {

	t := &Status{
		status: status,
		vp:     viewport.New(0, 0),
		app:    app,
	}

	return t
}

func (m *Status) View() string {
	if m.status == nil {
		return "no status"
	}

	mdContent := formatContent(m.status.Content)
	// lines := strings.Split(mdContent, "\n")
	return statusStyle.Render(mdContent)
	// return m.vp.View()
	return m.vp.View()
	head := m.status.Account.DisplayName
	content := formatContent(m.status.Content)

	if m.status.Reblog != nil {
		head = emoji.Sprintf("%s  :repeat_button:@%s", head, m.status.Reblog.Account.DisplayName)
	}

	return fmt.Sprintf("%s\n%s", head, content)

}

func (m *Status) updateStatus(status *mastodon.Status) {
	m.vp.SetContent("yoo")
	viewport.Sync(m.vp)

	m.status = status

	mdContent := formatContent(status.Content)

	// tr, err := glamour.NewTermRenderer(glamour.WithEmoji())
	// if err != nil {
	// 	log.Fatal(err)
	// }

	g, err := glamour.NewTermRenderer(glamour.WithAutoStyle())
	if err != nil {
		log.Fatal(err)
	}
	mdAnsi, err := g.Render(mdContent)
	if err != nil {
		log.Fatal(err)
	}

	content := cview.TranslateANSI(mdAnsi)
	m.vp.SetContent(content)

	w, h := m.vp.Width, m.vp.Width

	for _, m := range m.status.MediaAttachments {
		if m.Type == "image" {

			w = w - 5
			h = h - len(strings.Split(content, "\n")) - 5

			img := translateImage(m.URL, w, h)
			content = fmt.Sprintf("%s\n%s", content, img)
		}
	}

	ct := status.CreatedAt

	created := fmt.Sprintf("%02d:%02d %d-%02d-%02d",
		ct.Hour(), ct.Minute(),
		ct.Year(), ct.Month(), ct.Day())

	replies := emoji.Sprintf(":speech_balloon: %d", status.RepliesCount)
	boosts := emoji.Sprintf(":repeat_button: %d", status.ReblogsCount)

	likes := ""
	if m.IsFavorite() {
		likes += emoji.Sprintf(":heart: %d", status.FavouritesCount)
	} else {
		likes += emoji.Sprintf(":white_heart: %d", status.FavouritesCount)
	}

	info := strings.Join([]string{replies, boosts, likes}, " | ")

	avatar := translateImage(status.Account.AvatarStatic, 4, 8)

	content = fmt.Sprintf("%s\n%s",
		content,
		lipgloss.NewStyle().Align(lipgloss.Left).Render(status.Account.DisplayName),
	)
	content = fmt.Sprintf("%s\n%s",
		content,
		lipgloss.NewStyle().Align(lipgloss.Left).Render(avatar),
	)
	content = fmt.Sprintf("%s\n%s",
		content,
		lipgloss.NewStyle().Align(lipgloss.Left).Render(created),
	)
	content = fmt.Sprintf("%s\n%s",
		content,
		lipgloss.NewStyle().Align(lipgloss.Left).Render(info),
	)

	m.vp.SetContent(content)

	// 	f.AddText(status.Account.DisplayName, true, cview.AlignLeft, tcell.ColorWhite)
	// 	f.AddText(status.Account.Acct, true, cview.AlignCenter, tcell.ColorWhite)
	// 	f.AddText(status.Account.Username, true, cview.AlignRight, tcell.ColorWhite)
	// 	f.AddText(avatar, true, cview.AlignLeft, tcell.ColorWhite)
	// 	f.AddText(created, true, cview.AlignCenter, tcell.ColorWhite)

	// 	f.AddText(info, false, cview.AlignCenter, tcell.ColorWhite)
	// 	if status.Reblog != nil {
	// 		boosted := fmt.Sprintf("Boosted from %s", status.Reblog.Account.DisplayName)
	// 		f.AddText(boosted, false, cview.AlignRight, tcell.ColorLightCyan)
	// 	}

}

func (m *Status) Init() tea.Cmd {
	return nil
}

func (m *Status) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case StatusMsg:
		m.updateStatus(msg.status)
		return m, nil

	case tea.WindowSizeMsg:
		x, y := timelineStyle.GetFrameSize()
		m.vp = viewport.New(msg.Width-x, msg.Height-y)
		m.vp.HighPerformanceRendering = true
		return m, viewport.Sync(m.vp)
	}
	return m, nil
}

func translateImage(url string, x, y int) string {
	img, err := buildImage(url, x, y)
	if err != nil {
		return ""
	}
	ansi := img.Render()
	return cview.TranslateANSI(ansi)

}

func buildImage(url string, x, y int) (*ansimage.ANSImage, error) {
	pix, err := ansimage.NewScaledFromURL(url, y, x, color.Transparent, ansimage.ScaleModeResize, ansimage.NoDithering)
	if err != nil {
		return nil, err
	}
	return pix, nil

}
