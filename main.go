package main

import (
	"fmt"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/superbet-group/proto.registry/pkg/registry"
)

var topWindow fyne.Window

func main() {
	a := app.NewWithID("kafka.tool")
	a.SetIcon(theme.FyneLogo())
	logLifecycle(a)
	makeTray(a)
	w := a.NewWindow("Kafka Tool")
	topWindow = w

	w.SetMaster()

	tabs := container.NewAppTabs(
		container.NewTabItem("Connections", setContent("Connections")),
		container.NewTabItem("Produce", setContent("Produce")),
		container.NewTabItem("Consume", setContent("Consume")),
	)
	tabs.SetTabLocation(container.TabLocationLeading)

	themes := container.NewGridWithColumns(2,
		widget.NewButton("Dark", func() {
			a.Settings().SetTheme(theme.DarkTheme())
		}),
		widget.NewButton("Light", func() {
			a.Settings().SetTheme(theme.LightTheme())
		}),
	)

	c := container.NewBorder(nil, themes, nil, nil, tabs)
	w.SetContent(c)
	w.Resize(fyne.NewSize(1920, 1080))
	w.ShowAndRun()
}

func logLifecycle(a fyne.App) {
	a.Lifecycle().SetOnStarted(func() {
		log.Println("Lifecycle: Started")
	})
	a.Lifecycle().SetOnStopped(func() {
		log.Println("Lifecycle: Stopped")
	})
	a.Lifecycle().SetOnEnteredForeground(func() {
		log.Println("Lifecycle: Entered Foreground")
	})
	a.Lifecycle().SetOnExitedForeground(func() {
		log.Println("Lifecycle: Exited Foreground")
	})
}

func makeTray(a fyne.App) {
	if desk, ok := a.(desktop.App); ok {
		h := fyne.NewMenuItem("Hello", func() {})
		menu := fyne.NewMenu("Hello World", h)
		h.Action = func() {
			log.Println("System tray menu tapped")
			h.Label = "Welcome"
			menu.Refresh()
		}
		desk.SetSystemTrayMenu(menu)
	}
}

func setContent(tab string) *fyne.Container {
	content := container.NewMax()
	title := widget.NewLabel("")
	intro := widget.NewLabel("")

	switch tab {
	case "Connections":
		title.SetText("Edit Connections")
		intro.SetText("Here you can edit your Kafka connections")
		content.Objects = []fyne.CanvasObject{createConnections()}

	case "Produce":
		title.SetText("Produce a message")
		intro.SetText("Here you can produce Kafka messages")
		content.Objects = []fyne.CanvasObject{createProduce()}

	case "Consume":
		title.SetText("Consume messages")
		intro.SetText("Here you can consume Kafka messages")
		content.Objects = []fyne.CanvasObject{createConsume()}
	}

	l := container.NewBorder(
		container.NewVBox(title, widget.NewSeparator(), intro), nil, nil, nil, content)

	return l
}

func createConnections() *widget.Entry {
	return widget.NewEntry()
}

func createProduce() *fyne.Container {
	r := registry.NewWithDefaults()
	r.RegisterFiles("../betting.contracts/schema")
	protos := r.GetFullNameList()
	protobufSelection := widget.NewSelect(protos, func(selected string) { fmt.Println(selected) })
	newMessageScreen := widget.NewMultiLineEntry()
	return container.NewVBox(protobufSelection, newMessageScreen)
}

func createConsume() *fyne.Container {
	r := registry.NewWithDefaults()
	r.RegisterFiles("../betting.contracts/schema")
	protos := r.GetFullNameList()
	protobufSelection := widget.NewSelect(protos, func(selected string) { fmt.Println(selected) })
	newMessageScreen := widget.NewTextGridFromString("this is a text\ntestesteste\ntestesetesetsentounthuasontehusoan\nsnotaeuhnaoehtu")
	return container.NewVBox(protobufSelection, newMessageScreen)
}
