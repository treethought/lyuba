package ui

import (
	"log"
	"time"

	"github.com/gdamore/tcell/v2"
	ma "github.com/treethought/mammut/mastodon"
	"gitlab.com/tslocum/cbind"
	"gitlab.com/tslocum/cview"
)

type App struct {
	client     ma.Client
	ui         *cview.Application
	root       *cview.Flex
	timeline   *Timeline
	info       *cview.TextView
	statusView *StatusFrame
	menu       *Menu
}

func New() *App {
	client := ma.NewClient()
	return &App{
		client: client,
	}
}

func (app *App) FocusTimeline() {
	// Set the grid as the application root and focus the timeline
	app.ui.SetRoot(app.root, true)
	app.ui.SetFocus(app.timeline)

	app.timeline.SetTitle("...")

	go app.ui.QueueUpdateDraw(func() {
		app.timeline.Refresh()
	})
}

func (app *App) SetStatus(toot *Toot) {
	if app.statusView != nil {
		app.statusView.SetStatus(toot)
	}

	// app.Notify(fmt.Sprintf("Viewing status by: %s", toot.status.Account.DisplayName))

}

func (app *App) Notify(msg string) {
	if app.info == nil {
		return
	}
	app.info.Clear()
	app.info.SetText(msg)
	go app.ui.QueueUpdateDraw(func() {
		time.Sleep(2 * time.Second)
		app.info.Clear()

	})
}

func (app *App) Start() {
	// Initialize application
	app.ui = cview.NewApplication()

	toots := app.client.GetTimeline("local")
	if len(toots) == 0 {
		log.Fatal("Failed to get toots")
	}
	app.timeline = NewTimeline(app, toots, TimelineLocal)

	app.menu = NewMenu(app)

	app.statusView = NewStatusFrame(app)

	app.info = cview.NewTextView()
	app.info.SetBackgroundColor(tcell.ColorDefault)

	mid := cview.NewFlex()
	mid.SetBackgroundColor(tcell.ColorDefault)
	mid.SetDirection(cview.FlexRow)
	mid.AddItem(app.timeline, 0, 4, true)
	mid.AddItem(app.statusView, 0, 4, false)
	mid.AddItem(app.info, 0, 1, false)

	flex := cview.NewFlex()
	flex.SetBackgroundTransparent(false)
	flex.SetBackgroundColor(tcell.ColorDefault)

	flex.AddItem(app.menu, 0, 1, false)
	flex.AddItem(mid, 0, 4, false)

	focusManager := cview.NewFocusManager(app.ui.SetFocus)
	focusManager.SetWrapAround(true)
	focusManager.Add(app.menu, app.timeline)

	inputHandler := cbind.NewConfiguration()
	for _, key := range cview.Keys.MovePreviousField {
		err := inputHandler.Set(key, wrap(focusManager.FocusPrevious))
		if err != nil {
			log.Fatal(err)
		}
	}
	for _, key := range cview.Keys.MoveNextField {
		err := inputHandler.Set(key, wrap(focusManager.FocusNext))
		if err != nil {
			log.Fatal(err)
		}
	}

	app.root = flex
	app.ui.SetInputCapture(inputHandler.Capture)

	app.FocusTimeline()

	err := app.ui.Run()
	if err != nil {
		log.Fatal(err)
	}
}
func wrap(f func()) func(ev *tcell.EventKey) *tcell.EventKey {
	return func(ev *tcell.EventKey) *tcell.EventKey {
		f()
		return nil
	}
}
