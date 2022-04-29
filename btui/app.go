package btui

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/treethought/boba"
	"github.com/treethought/mammut/mastodon"
)

// App is a wrapper around boba.App used to manage the applications models
type App struct {
	boba   *boba.App
	client mastodon.Client
}

func NewApp() *App {
	app := &App{}

	app.boba = boba.NewApp()

	app.client = mastodon.NewClient()

	box := boba.NewBox("vertical", 100, 100)
	timeline := NewTimeline(app, TimelineHome)

	status := NewStatus(app, nil)

	box.AddNode(timeline, 100, 50)
	box.AddNode(status, 100, 50)

	app.boba.Add("box", box)
	app.boba.Register("status", status)
	app.boba.Register("timeline", timeline)

	// begin with input prompting for url
	app.boba.SetFocus("box")
	return app

}

// delegate is used handle messages as needed before they are passed to the
// currently focused model.
func (a *App) delegate(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	// show a simple view of any errors that occurred
	// case errMsg:
	// 	m, _ := a.boba.Get("error").(*ErrorView)
	// 	m.message = msg.msg
	// 	return a.boba, boba.ChangeState("error")

	// return to input view on ESC
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEscape:
			return a.boba, boba.ChangeState("input")
		}

	}
	return a.boba, nil
}

func (a *App) Start() {

	// title := fmt.Sprintf(" Timeline - %s ", strings.Title(ttype))
	// if selected != nil {
	// 	t.SetCurrentToot(selected)
	// }
	// app.boba.SetDelegate(app.delegate)
	// start the app
	p := tea.NewProgram(a.boba, tea.WithAltScreen())
	if err := p.Start(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
