package ui

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

const useHighPerformanceRenderer = true

type Status struct {
	status *mastodon.Status
	vp     viewport.Model
	app    *App
}

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
	statusStyle := lipgloss.NewStyle().
		Width(m.vp.Width).Height(m.vp.Height).
		MaxHeight(m.vp.Height).MaxWidth(m.vp.Width).
		BorderStyle(lipgloss.NormalBorder()).
		Padding(0).MarginBackground(lipgloss.Color("300")).

	return statusStyle.Render(m.render())

}

func (m Status) buildEngagements() string {
	status := m.status
	replies := emoji.Sprintf(":speech_balloon:%d", status.RepliesCount)
	boosts := emoji.Sprintf(":repeat_button:%d", status.ReblogsCount)

	likes := ""
	if m.IsFavorite() {
		likes += emoji.Sprintf(":heart: %d", status.FavouritesCount)
	} else {
		likes += emoji.Sprintf(":blue_heart: %d", status.FavouritesCount)
	}

	return strings.Join([]string{replies, boosts, likes}, "  ")
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
	createdStyle := lipgloss.NewStyle().Align(lipgloss.Right).Padding(0, 0, 0, 5)

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

func buildMedia(status *mastodon.Status, x, y int) string {

	content := ""
	for _, m := range status.MediaAttachments {
		if m.Type == "image" {

			// w = w - 5
			// h = h - len(strings.Split(content, "\n")) - 5

			img := translateImage(m.URL, x, y)
			content = fmt.Sprintf("%s\n%s", content, img)
		}
	}
	return content
}

func (m *Status) render() string {
	status := m.status

	header := m.buildheader()
	info := m.buildEngagements()

	bodyHeight := m.vp.Height - lipgloss.Height(header) - lipgloss.Height(info)

	text := formatContent(status.Content)

	textStyle := lipgloss.NewStyle().
		Align(lipgloss.Left).
		// Width(m.vp.Width / 2). //.Height(bodyHeight).
		MaxWidth(m.vp.Width).MaxHeight(bodyHeight).
		Padding(0).Margin(0).
		BorderStyle(lipgloss.NormalBorder()).
		Align(lipgloss.Center)

	text = textStyle.Render(text)

	media := buildMedia(m.status, m.vp.Width/3, bodyHeight)

	mediaStyle := lipgloss.NewStyle().
		Align(lipgloss.Center).
		// Width(m.vp.Width / 2).
		MaxWidth(m.vp.Width / 2).MaxHeight(bodyHeight).
		Padding(0).Margin(0).
		Align(lipgloss.Center)
	// BorderStyle(lipgloss.NormalBorder())

	media = mediaStyle.Render(media)

	body := lipgloss.JoinHorizontal(lipgloss.Top, text, media)
	body = lipgloss.NewStyle().MaxHeight(bodyHeight).Render(body)

	content := lipgloss.JoinVertical(lipgloss.Top, header, info, body)
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
