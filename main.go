package main

import (
	"context"
	"fmt"
	"log"
	"sort"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	kafkaclients "github.com/superbet-group/kafka.clients/v3"
	registry "github.com/superbet-group/proto.registry"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/dynamicpb"
)

var topWindow fyne.Window
var state State
var producer *kafkaclients.Producer
var consumer *kafkaclients.Consumer
var r *registry.Registry
var stringList binding.StringList

type State struct {
	brokers             string
	marshalerType       string
	descriptors         []string
	protoConversionMap  map[string]string
	chosenProtoFullName string
	chosenProtoMessage  dynamicpb.Message
}

func main() {
	a := app.NewWithID("kafka.tool")
	a.SetIcon(theme.FyneLogo())
	logLifecycle(a)
	makeTray(a)
	w := a.NewWindow("Kafka Tool")
	topWindow = w

	w.SetMaster()

	state = State{}

	r = registry.NewWithDefaults()
	r.RegisterFiles("../betting.contracts/schema")
	state.protoConversionMap = r.GetNameConversionMap()
	protos := r.GetFullNameList()
	sort.Strings(protos)
	state.descriptors = protos
	stringList = binding.NewStringList()

	tabs := container.NewAppTabs(
		container.NewTabItem("Connections", setContent("Connections")),
		container.NewTabItem("Protobuf", setContent("Protobuf")),
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

	guiContainer := container.NewBorder(nil, themes, nil, nil, tabs)
	w.SetContent(guiContainer)
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

	case "Protobuf":
		title.SetText("Choose a Protobuf descriptor")
		intro.SetText("Here you can choose the Protobuf descriptor with which to unmarshal a message")
		content.Objects = []fyne.CanvasObject{createChooseProtoDescriptor()}

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

func createConnections() *fyne.Container {
	entry := widget.NewEntry()
	entry.OnSubmitted = func(brokers string) {
		fmt.Println("test")
		state.brokers = brokers
		producer = newKafkaProducer(brokers, "")
	}

	return container.NewBorder(entry, nil, nil, nil)
}

func createChooseProtoDescriptor() *widget.List {
	protobufSelection := widget.NewList(
		func() int {
			return len(state.descriptors)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("protobuf")
		},
		func(i int, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(state.descriptors[i])
		},
	)

	protobufSelection.OnSelected = func(id int) {
		chosenProtoFullName := state.descriptors[id]

		chosenProtoMessage, err := r.FindCorrectProtoDefinition(chosenProtoFullName, state.protoConversionMap[chosenProtoFullName])
		if err != nil {
			fmt.Println(err)
			return
		}

		state.chosenProtoMessage = chosenProtoMessage
		state.chosenProtoFullName = chosenProtoFullName
	}

	return protobufSelection
}

func createProduce() *fyne.Container {
	topicName := widget.NewLabel("Topic name:")
	topicField := widget.NewEntry()
	marshalerType := widget.NewSelect(
		[]string{"text", "json"},
		func(option string) {
			state.marshalerType = option
		},
	)

	marshalerType.SetSelected("json")
	marshalerTypeName := widget.NewLabel("Marshaler type name:")

	messageValueField := widget.NewMultiLineEntry()
	messageFieldName := widget.NewLabel("Message:")
	submitButton := widget.NewButton("Submit", func() {
		val := []byte(messageValueField.Text)

		if state.chosenProtoFullName != "" {
			switch state.marshalerType {
			case "text":
				err := prototext.Unmarshal(val, &state.chosenProtoMessage)
				if err != nil {
					fmt.Println(err)
					return
				}

			case "json":
				err := protojson.Unmarshal(val, &state.chosenProtoMessage)
				if err != nil {
					fmt.Println(err)
					return
				}

			default:
				fmt.Println("unknown producer type, not going to use protobuf")
			}

			var err error
			val, err = proto.Marshal(&state.chosenProtoMessage)
			if err != nil {
				fmt.Println(err)
				return
			}
		}

		_, err := producer.ProduceSync(kafkaclients.Message{
			Metadata: kafkaclients.Metadata{Topic: topicField.Text},
			Value:    val,
		})
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println("submitted")
	})

	topicBorder := container.NewBorder(nil, nil, topicName, nil, topicField)
	marshalerTypeBorder := container.NewBorder(nil, nil, marshalerTypeName, nil, marshalerType)
	messageBorder := container.NewBorder(nil, nil, messageFieldName, nil, messageValueField)
	return container.NewVBox(topicBorder, marshalerTypeBorder, messageBorder, submitButton)
}

func createConsume() *fyne.Container {
	newMessageScreen := widget.NewListWithData(
		stringList,
		func() fyne.CanvasObject {
			return widget.NewLabel("messages")
		},
		func(i binding.DataItem, o fyne.CanvasObject) {
			o.(*widget.Label).Bind(i.(binding.String))
		},
	)
	topicName := widget.NewLabel("Topic name:")
	topicField := widget.NewEntry()
	consumeButton := widget.NewButton("Consume", func() {
		go func() {
			consumer = newKafkaConsumer(state.brokers, topicField.Text, 0, 1, "")
			for {
				kafkaMessages, _, err := consumer.Consume(context.TODO(), 0, 0, make(map[string]interface{}))
				if err != nil {
					fmt.Println(err)
				}

				for message := range kafkaMessages {
					val := string(message.Value)
					if state.chosenProtoFullName != "" {
						byteVal, err := r.Unmarshal(message.Value, state.chosenProtoMessage)
						if err != nil {
							fmt.Println(err)
							return
						}

						val = string(byteVal)
					}

					err := stringList.Append(val)
					if err != nil {
						fmt.Println(err)
						return
					}
				}
			}
		}()
	})

	topicBorder := container.NewBorder(nil, nil, topicName, nil, topicField)
	return container.NewBorder(container.NewVBox(topicBorder, consumeButton), nil, nil, nil, newMessageScreen)
}
