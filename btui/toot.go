package btui

import (
	"fmt"
	"log"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/charmbracelet/glamour"
	"github.com/kyokomi/emoji"
	"github.com/mattn/go-mastodon"
)

type Toot struct {
	status *mastodon.Status
	app    *App
}

func (t *Toot) FilterValue() string { return "" }

func (t *Toot) Title() string {
	return t.header()
}

func (t *Toot) Description() string {
	return t.Content()
}

func formatContent(html string) string {
	converter := md.NewConverter("", true, nil)

	mdContent, err := converter.ConvertString(html)
	if err != nil {
		log.Fatal(err)
	}

	out, err := glamour.Render(mdContent, "dark")
	if err != nil {
		return mdContent
	}
	return out

}

func (t *Toot) IsFavorite() bool {
	favorited, ok := t.status.Favourited.(bool)
	if !ok {
		return false
	}
	return favorited

}

func (t *Toot) header() string {
	header := t.status.Account.DisplayName

	if t.IsFavorite() {
		header += emoji.Sprint(" :heart:")
	} else {
		header += emoji.Sprint(" :white_heart:")
	}

	if t.status.Reblog != nil {
		header = emoji.Sprintf("%s  || :repeat_button:@%s", header, t.status.Reblog.Account.DisplayName)
	}
	return header
}

func (t *Toot) Content() string {
	return formatContent(t.status.Content)
}

func NewToot(app *App, status *mastodon.Status) *Toot {

	t := &Toot{
		status: status,
		app:    app,
	}

	return t
}

func (m *Toot) View() string {
	head := m.status.Account.DisplayName
	content := formatContent(m.status.Content)

	if m.status.Reblog != nil {
		head = emoji.Sprintf("%s  :repeat_button:@%s", head, m.status.Reblog.Account.DisplayName)
	}

	return fmt.Sprintf("%s\n%s", head, content)

}
