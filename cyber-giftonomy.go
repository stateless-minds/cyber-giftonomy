package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
	"github.com/mitchellh/mapstructure"
	shell "github.com/stateless-minds/go-ipfs-api"
)

const dbNameGift = "gift"

const (
	topicCreateGift  = "create-gift"
	topicCollectGift = "collect-gift"
)

type cybergiftonomy struct {
	app.Compo
	sh    *shell.Shell
	sub   *shell.PubSubSubscription
	gifts map[string]Gift
	alert string
}

type Gift struct {
	ID          string   `mapstructure:"_id" json:"_id" validate:"uuid_rfc4122"`
	Title       string   `mapstructure:"title" json:"title" validate:"uuid_rfc4122"`
	Description string   `mapstructure:"description" json:"description" validate:"uuid_rfc4122"`
	Coordinates string   `mapstructure:"coordinates" json:"coordinates" validate:"uuid_rfc4122"`
	Photo       string   `mapstructure:"photo" json:"photo" validate:"uuid_rfc4122"`
	Tags        []string `mapstructure:"tags" json:"tags" validate:"uuid_rfc4122"`
}

func (c *cybergiftonomy) OnMount(ctx app.Context) {
	sh := shell.NewShell("localhost:5001")
	c.sh = sh

	c.subscribeToCreateGiftTopic(ctx)
	c.subscribeToCollectGiftTopic(ctx)

	c.gifts = make(map[string]Gift)

	ctx.Async(func() {
		// c.DeleteGifts(ctx)
		v := c.FetchGifts(ctx)
		var vv []interface{}
		err := json.Unmarshal(v, &vv)
		if err != nil {
			log.Fatal(err)
		}

		for _, ii := range vv {
			g := Gift{}
			err = mapstructure.Decode(ii, &g)
			if err != nil {
				log.Fatal(err)
			}
			ctx.Dispatch(func(ctx app.Context) {
				c.gifts[g.ID] = g

				// sort.SliceStable(c.gifts, func(i, j int) bool {
				// 	return c.gifts[i].ID < c.gifts[j].ID
				// })
			})
		}
	})
}

func (c *cybergiftonomy) DeleteGifts(ctx app.Context) {
	err := c.sh.OrbitDocsDelete(dbNameGift, "all")
	if err != nil {
		log.Fatal(err)
	}
}

func (c *cybergiftonomy) FetchGifts(ctx app.Context) []byte {
	v, err := c.sh.OrbitDocsQuery(dbNameGift, "all", "")
	if err != nil {
		log.Fatal(err)
	}

	return v
}

func (c *cybergiftonomy) subscribeToCreateGiftTopic(ctx app.Context) {
	ctx.Async(func() {
		topic := topicCreateGift
		subscription, err := c.sh.PubSubSubscribe(topic)
		if err != nil {
			log.Fatal(err)
		}
		c.sub = subscription
		c.subscriptionCreateGift(ctx)
	})
}

func (c *cybergiftonomy) subscriptionCreateGift(ctx app.Context) {
	ctx.Async(func() {
		defer c.sub.Cancel()
		// wait on pubsub
		res, err := c.sub.Next()
		if err != nil {
			log.Fatal(err)
		}
		// Decode the string data.
		str := string(res.Data)
		// log.Println("Subscriber of topic create-gift received message: " + str)
		ctx.Async(func() {
			c.subscribeToCreateGiftTopic(ctx)
		})

		g := Gift{}
		err = json.Unmarshal([]byte(str), &g)
		if err != nil {
			log.Fatal(err)
		}

		ctx.Dispatch(func(ctx app.Context) {
			c.gifts[g.ID] = g
		})
	})
}

func (c *cybergiftonomy) subscribeToCollectGiftTopic(ctx app.Context) {
	ctx.Async(func() {
		topic := topicCollectGift
		subscription, err := c.sh.PubSubSubscribe(topic)
		if err != nil {
			log.Fatal(err)
		}
		c.sub = subscription
		c.subscriptionCollectGift(ctx)
	})
}

func (c *cybergiftonomy) subscriptionCollectGift(ctx app.Context) {
	ctx.Async(func() {
		defer c.sub.Cancel()
		// wait on pubsub
		res, err := c.sub.Next()
		if err != nil {
			log.Fatal(err)
		}
		// Decode the string data.
		str := string(res.Data)
		// log.Println("Subscriber of topic collect-gift received message: " + str)
		ctx.Async(func() {
			c.subscribeToCollectGiftTopic(ctx)
		})

		g := Gift{}
		err = json.Unmarshal([]byte(str), &g)
		if err != nil {
			log.Fatal(err)
		}

		ctx.Dispatch(func(ctx app.Context) {
			delete(c.gifts, g.ID)
		})
	})
}

func (c *cybergiftonomy) Render() app.UI {
	return app.Div().Class("app page-wrap xyz-in").Body(
		app.If(len(c.alert) > 0, func() app.UI {
			return app.Div().Class("container alert").Body(
				app.Text(c.alert),
			)
		}),
		app.Div().Class("col center-x space-y-0 pt-50 page-hero").Attr("xyz", "fade small stagger ease-out-back").Body(
			app.H1().Class("title hero-logo xyz-nested").Text("CyberGiftonomy"),
			app.H2().Class("subtitle hero-text xyz-nested pt-5").Text("Give what you can, get what you need!"),
		),
		app.Div().ID("what-is").Class("col center-x").Body(
			app.Div().Class("container").Body(
				app.Div().Class("page-section").Attr("xyz", "fade small stagger delay-4 ease-in-out").Body(
					app.Div().Class("section row space-10 xyz-nested").Attr("xyz", "fade left stagger").Body(
						app.Div().Class("card section-item xyz-nested").Body(
							app.Header().Text("Sociable"),
							app.Text("Local-first global community organized around gift economy."),
							app.Footer().Body(app.Strong().Text("Everything shared is free.")),
						),
						app.Div().Class("card with-leaves section-item xyz-nested").Body(
							app.Header().Text("Eco-Friendly"),
							app.Text("We post the coordinates of what we leave behind so that it can be picked up."),
							app.Footer().Body(app.Strong().Text("If it works give it to someone.")),
						),
					),
					app.Div().Class("section row space-10 xyz-nested").Attr("xyz", "fade left stagger").Body(
						app.Div().Class("card with-scales section-item xyz-nested").Body(
							app.Header().Text("Transparent"),
							app.Text("No registration, no ads, no tracking, no data collection."),
							app.Footer().Body(app.Strong().Text("No strings attached.")),
						),
						app.Div().Class("card with-wood section-item xyz-nested").Body(
							app.Header().Text("Sustainable"),
							app.Text("By reusing functional products we extend their life."),
							app.Footer().Body(app.Strong().Text("Promote functional economy.")),
						),
					),
				),
			),
		),
		app.Div().Class("col center-x space-y-0 pt-25 page-hero").Attr("xyz", "fade small stagger ease-out-back").Body(
			app.H1().Class("title hero-logo xyz-nested").Text("Give"),
			app.H2().Class("subtitle hero-text xyz-nested").Text("Someone a treat!"),
		),
		app.Div().ID("give").Class("col center-x space-y-0 page-hero").Attr("xyz", "fade small stagger ease-out-back").Body(
			app.Div().Class("col container").Body(
				app.Form().Class("col center-x space-10").Body(
					app.Input().ID("title").Class("input is-success xyz-nested").Size(50).Placeholder("Title: Armchair").Required(true),
					app.Textarea().ID("description").Class("textarea is-success xyz-nested").Rows(4).Cols(50).Placeholder("Description: Natural leather, mint condition, thoroughly cleaned.").Required(true),
					app.Input().ID("coordinates").Class("input is-success xyz-nested").Size(50).Placeholder("Pickup coordinates: 42.13963553947106, 24.76160045085874").Required(true),
					app.Input().ID("tags").Class("input is-success xyz-nested").Size(50).Placeholder("Tags: #armchair,#chair,#furniture").Required(true),
					app.Label().ID("label-photo").For("photo").Class("pr-100 xyz-nested").Body(
						app.Text("Image: "),
						app.Input().Type("file").ID("photo").Name("photo").Required(true).Accept("image/*").OnChange(c.ReadFile),
					),
					app.Img().ID("preview").Src("").Height(100).Alt("Image preview").Hidden(true),
					app.Button().Class("button mt-30 is-info xyz-nested").Type("submit").Text("Submit"),
				).OnSubmit(c.onSubmitGift),
			),
		),
		app.Div().Class("col center-x space-y-0 pt-25 page-hero").Attr("xyz", "fade small stagger ease-out-back").Body(
			app.H1().Class("title hero-logo xyz-nested").Text("Find"),
			app.H2().Class("subtitle hero-text xyz-nested").Text("Your own treasure!"),
		),
		app.Div().ID("gallery").Class("col center-x space-y-0 page-hero").Attr("xyz", "fade small stagger ease-out-back").Body(
			app.Div().Class("col container gallery-container").Body(
				app.Div().Class("gallery-section row space-10 xyz-nested").Attr("xyz", "fade left stagger").Body(
					app.Range(c.gifts).Map(func(i string) app.UI {
						return app.Div().ID(i).Class("card card-gallery section-item xyz-nested").Body(
							app.Div().Class("row p-10").Body(
								app.Header().Text(c.gifts[i].Title),
							),
							app.Div().Class("row p-10").Body(
								app.Span().Body(app.Img().Src(c.gifts[i].Photo).Alt(c.gifts[i].Title).Height(250).Width(250)),
							),
							app.Div().Class("row p-10").Body(
								app.Span().Body(app.Text(c.gifts[i].Description)),
							),
							app.Div().Class("tags").Body(
								app.Range(c.gifts[i].Tags).Slice(func(n int) app.UI {
									return app.Span().Class("badge").Text(c.gifts[i].Tags[n])
								}),
							),
							app.Div().Class("row p-10").Body(
								app.Span().Body(app.Strong().Text("Pickup coordinates: "+c.gifts[i].Coordinates)),
							),
							app.Button().Class("button mt-30 is-success xyz-nested").Text("Collect").OnClick(c.onCollectGift),
						)
					}),
				).Style("--carousel-start", "-"+strconv.Itoa(len(c.gifts)*250)+"px").Style("--carousel-end", strconv.Itoa(len(c.gifts)*250)+"px"),
			),
		),
		app.Div().Class("row center-x pb-20").Body(
			app.Span().Body(
				app.Text("Made with "),
				app.I().Class("fa fa-heart pulse").Style("color", "red"),
				app.Text(" by "),
				app.A().Href("https://github.com/stateless-minds").Target("_blank").Text("Stateless Minds"),
			),
		),
	)
}

func (c *cybergiftonomy) Alert(ctx app.Context, msg string) {
	c.alert = msg
	ctx.Async(func() {
		time.Sleep(3 * time.Second)
		c.alert = ""
		ctx.Dispatch(func(ctx app.Context) {})
	})
}

func (c *cybergiftonomy) ReadFile(ctx app.Context, e app.Event) {
	files := ctx.JSSrc().Get("files")
	if !files.Truthy() || files.Get("length").Int() == 0 {
		c.Alert(ctx, "Error: No file attached.")
		return
	} else if files.Get("length").Int() > 1 {
		c.Alert(ctx, "Error: You can upload only one image.")
		return
	}

	file := files.Index(0)

	if !strings.Contains(file.Get("type").String(), "image/") {
		c.Alert(ctx, "Error: Only images are supported.")
		return
	} else if file.Get("size").Int() > 5000000 {
		c.Alert(ctx, "Error: Max image size 5MB.")
		return
	}

	preview := app.Window().GetElementByID("preview")
	var close func()
	reader := app.Window().Get("FileReader").New()

	onFileLoad := app.FuncOf(func(this app.Value, args []app.Value) interface{} {
		event := args[0]
		content := event.Get("target").Get("result").String()

		preview.Set("src", content)
		preview.Set("hidden", false)

		close()
		return nil
	})

	onFileLoadError := app.FuncOf(func(this app.Value, args []app.Value) interface{} {
		// Your error handling...

		close()
		return nil
	})

	// To release resources when callback are called.
	close = func() {
		onFileLoad.Release()
		onFileLoadError.Release()
	}

	reader.Set("onload", onFileLoad)
	reader.Set("onerror", onFileLoadError)
	reader.Call("readAsDataURL", file, "UTF-8")
}

func (c *cybergiftonomy) onSubmitGift(ctx app.Context, e app.Event) {
	e.PreventDefault()
	title := app.Window().GetElementByID("title").Get("value").String()
	description := app.Window().GetElementByID("description").Get("value").String()
	coordinates := app.Window().GetElementByID("coordinates").Get("value").String()
	tags := app.Window().GetElementByID("tags").Get("value").String()
	var tagsSlice []string
	if strings.Contains(tags, ",") {
		tagsSlice = strings.Split(tags, ",")
	} else {
		tagsSlice = append(tagsSlice, tags)
	}

	var id int
	if len(c.gifts) > 0 {
		id = len(c.gifts)
	} else {
		id = 0
	}

	g := Gift{
		ID:          strconv.Itoa(id),
		Title:       title,
		Description: description,
		Coordinates: coordinates,
		Photo:       app.Window().GetElementByID("preview").Get("src").String(),
		Tags:        tagsSlice,
	}

	gft, err := json.Marshal(g)
	if err != nil {
		log.Fatal(err)
	}

	ctx.Async(func() {
		err = c.sh.OrbitDocsPut(dbNameGift, gft)
		if err != nil {
			ctx.Dispatch(func(ctx app.Context) {
				c.Alert(ctx, "Error: could not save gift.")
				log.Fatal(err)
			})
		}
		err = c.sh.PubSubPublish(topicCreateGift, string(gft))
		if err != nil {
			log.Fatal(err)
		}

		ctx.Dispatch(func(ctx app.Context) {
			c.Alert(ctx, "Success: Gift submited.")
		})
	})
}

func (c *cybergiftonomy) onCollectGift(ctx app.Context, e app.Event) {
	e.PreventDefault()
	pid := ctx.JSSrc().Get("parentElement").Get("id").String()

	g := Gift{
		ID:          pid,
		Title:       c.gifts[pid].Title,
		Description: c.gifts[pid].Description,
		Coordinates: c.gifts[pid].Coordinates,
		Photo:       c.gifts[pid].Photo,
		Tags:        c.gifts[pid].Tags,
	}

	gft, err := json.Marshal(g)
	if err != nil {
		log.Fatal(err)
	}

	ctx.Async(func() {
		err = c.sh.OrbitDocsDelete(dbNameGift, g.ID)
		if err != nil {
			ctx.Dispatch(func(ctx app.Context) {
				c.Alert(ctx, "Error: could not delete collected gift.")
				log.Fatal(err)
			})
		}
		err = c.sh.PubSubPublish(topicCollectGift, string(gft))
		if err != nil {
			log.Fatal(err)
		}

		ctx.Dispatch(func(ctx app.Context) {
			c.Alert(ctx, "Success: Gift collected.")
		})
	})
}

func main() {
	app.Route("/", func() app.Composer{
		return &cybergiftonomy{}
	})
	app.RunWhenOnBrowser()
	http.Handle("/", &app.Handler{
		Name:        "cybergiftonomy",
		Description: "P2P gift economy community",
		Styles: []string{
			"https://cdnjs.cloudflare.com/ajax/libs/normalize/8.0.1/normalize.min.css",
			"web/app.css",
			"https://cdn.jsdelivr.net/npm/retro.css@1.0.0/css/index.min.css",
			"https://unpkg.com/pattern.css",
			"https://cdn.jsdelivr.net/npm/@animxyz/core",
			"https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.4.0/css/all.min.css",
		},
	})

	if err := http.ListenAndServe(":8000", nil); err != nil {
		log.Fatal(err)
	}
}
