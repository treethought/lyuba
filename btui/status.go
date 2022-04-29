package btui

import (
	"fmt"
	"image/color"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/eliukblau/pixterm/pkg/ansimage"
	"github.com/kyokomi/emoji"
	"github.com/mattn/go-mastodon"
)

var statusStyle = lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()).Padding(0).MarginBackground(lipgloss.Color("300"))

const useHighPerformanceRenderer = true

const img = "https://i.redd.it/65fmdbh1ja951.jpg"

type Status struct {
	status *mastodon.Status
	vp     viewport.Model
	app    *App
	// img    *imgcat.Model
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

func (t *Status) Content() string {
	return formatContent(t.status.Content)
}

func NewStatus(app *App, status *mastodon.Status) *Status {

	t := &Status{
		status: status,
		vp:     viewport.New(0, 0),
		app:    app,
		// img:    nil,
	}

	return t
}

func (m *Status) View() string {
	if m.status == nil {
		return "no status"
	}

	// mdContent := formatContent(m.status.Content)
	// lines := strings.Split(mdContent, "\n")
	return statusStyle.Render(m.render())
	return m.vp.View()
	head := m.status.Account.DisplayName
	content := formatContent(m.status.Content)

	if m.status.Reblog != nil {
		head = emoji.Sprintf("%s  :repeat_button:@%s", head, m.status.Reblog.Account.DisplayName)
	}

	return fmt.Sprintf("%s\n%s", head, content)

}

func (m Status) buildEngagements() string {
	status := m.status
	replies := emoji.Sprintf(":speech_balloon: %d", status.RepliesCount)
	boosts := emoji.Sprintf(":repeat_button: %d", status.ReblogsCount)

	likes := ""
	if m.IsFavorite() {
		likes += emoji.Sprintf(":heart: %d", status.FavouritesCount)
	} else {
		likes += emoji.Sprintf(":white_heart: %d", status.FavouritesCount)
	}

	return strings.Join([]string{replies, boosts, likes}, " | ")
}

func (m Status) buildheader() string {
	content := ""
	status := m.status
	ct := status.CreatedAt

	created := fmt.Sprintf("%02d:%02d %d-%02d-%02d",
		ct.Hour(), ct.Minute(),
		ct.Year(), ct.Month(), ct.Day())

	accountStyle := lipgloss.NewStyle().Align(lipgloss.Left)
	boostStyle := lipgloss.NewStyle().Align(lipgloss.Right).Padding(0, 0, 0, 2).Foreground(lipgloss.Color("128"))
	createdStyle := lipgloss.NewStyle().Align(lipgloss.Right).Padding(0, 0, 0, 2)

	avatar := translateImage(status.Account.AvatarStatic, 8, 8)

	content = lipgloss.JoinHorizontal(lipgloss.Center,
		accountStyle.Render(avatar),
		accountStyle.Render(status.Account.DisplayName),
	)

	if m.status.Reblog != nil {
		boostContent := emoji.Sprintf(" || :repeat_button:@%s", m.status.Reblog.Account.DisplayName)

		content = lipgloss.JoinHorizontal(lipgloss.Center,
			content,
			boostStyle.Render(boostContent),
		)
	}

	content = lipgloss.JoinHorizontal(lipgloss.Center,
		content,
		createdStyle.Render(created),
	)

	return content
}

func (m Status) buildMedia() string {
	mediaStyle := lipgloss.NewStyle().Align(lipgloss.Center)
	// w, h := m.vp.Width, m.vp.Width

	content := ""
	for _, m := range m.status.MediaAttachments {
		if m.Type == "image" {

			// w = w - 5
			// h = h - len(strings.Split(content, "\n")) - 5

			img := translateImage(m.URL, 10, 10)
			content = fmt.Sprintf("%s\n%s", content, img)
		}
	}
	return mediaStyle.Render(content)
}

func (m *Status) render() string {
	status := m.status

	header := m.buildheader()
	media := m.buildMedia()
	content := formatContent(status.Content)

	content = lipgloss.JoinVertical(lipgloss.Top, header, media, content)

	info := m.buildEngagements()

	content = lipgloss.JoinVertical(lipgloss.Center,
		content,
		info,
	)
	return content

}

func (m *Status) Init() tea.Cmd {
	return nil
}

func (m *Status) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	cmds := []tea.Cmd{}
	if useHighPerformanceRenderer {
		// Render (or re-render) the whole viewport. Necessary both to
		// initialize the viewport and when the window is resized.
		//
		// This is needed for high-performance rendering only.
		cmds = append(cmds, viewport.Sync(m.vp))
	}
	switch msg := msg.(type) {

	case StatusMsg:
		m.status = msg.status
		return m, nil

	case tea.WindowSizeMsg:
		x, y := timelineStyle.GetFrameSize()
		m.vp = viewport.New(msg.Width-x, msg.Height-y)
		m.vp.HighPerformanceRendering = useHighPerformanceRenderer
	}

	return m, tea.Batch(cmds...)
}

func translateImage(url string, x, y int) string {
	img, err := buildImage(url, x, y)
	if err != nil {
		return ""
	}
	ansi := img.Render()
	return ansi

}

func buildImage(url string, x, y int) (*ansimage.ANSImage, error) {
	pix, err := ansimage.NewScaledFromURL(url, y, x, color.Transparent, ansimage.ScaleModeFit, ansimage.NoDithering)
	if err != nil {
		return nil, err
	}
	return pix, nil

}
