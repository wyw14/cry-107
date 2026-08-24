package console

import (
	"html/template"
	"io"
)

type Page struct {
	Path   string
	Title  string
	Active string
	API    string
	Action string
	Fields []Field
	Stages []Stage
}

type Field struct {
	Name  string
	Label string
	Value string
}

type Stage struct {
	Name  string
	Value string
	Unit  string
}

var pageTemplate = template.Must(template.New("operator").Parse(`<!doctype html>
<html lang="zh-CN"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<title>{{.Title}} | KilnGuard</title><style>` + styles + `</style></head>
<body data-api="{{.API}}"><header><div class="brand">KilnGuard</div><div class="plant">Line K-07 / Control room</div>
<nav><a href="/kiln" class="{{if eq .Active "kiln"}}active{{end}}">Kiln</a><a href="/burner" class="{{if eq .Active "burner"}}active{{end}}">Burner</a><a href="/cooler" class="{{if eq .Active "cooler"}}active{{end}}">Cooler</a><a href="/incidents" class="{{if eq .Active "incidents"}}active{{end}}">Incidents</a></nav></header>
<main><div class="titlebar"><h1>{{.Title}}</h1><div class="status"><i></i> Control service online</div></div>
<section class="process">{{range .Stages}}<div class="stage"><small>{{.Name}}</small><strong>{{.Value}}</strong><span>{{.Unit}}</span><div class="flowline"></div></div>{{end}}</section>
<section class="grid"><div class="panel"><h2>Live process state</h2><pre id="state">Loading process state...</pre></div>
<aside class="panel"><h2>Control target</h2>{{if .Action}}<form id="control" action="{{.Action}}">{{range .Fields}}<label>{{.Label}}<input name="{{.Name}}" type="number" step="0.1" value="{{.Value}}"></label>{{end}}<button type="submit">Apply target</button></form>{{end}}<p id="message" class="message"></p></aside></section>
</main><script>` + script + `</script></body></html>`))

func Render(writer io.Writer, page Page) error {
	return pageTemplate.Execute(writer, page)
}

func KilnPage() Page {
	return Page{Path: "/kiln", Title: "Kiln feed and rotation", Active: "kiln", API: "/api/kiln/state", Action: "/api/kiln/load", Fields: []Field{{Name: "load_percent", Label: "Kiln load (%)", Value: "82"}}, Stages: defaultStages()}
}

func BurnerPage() Page {
	return Page{Path: "/burner", Title: "Combustion and draft", Active: "burner", API: "/api/burner/state", Action: "/api/kiln/air-proof", Fields: []Field{{Name: "actual_air", Label: "Confirmed secondary air", Value: "103"}}, Stages: defaultStages()}
}

func CoolerPage() Page {
	return Page{Path: "/cooler", Title: "Clinker cooling and heat recovery", Active: "cooler", API: "/api/cooler/state", Action: "/api/cooler/air", Fields: []Field{{Name: "outlet_temp", Label: "Outlet temperature (C)", Value: "145"}, {Name: "requested_air", Label: "Cooling air target", Value: "76"}}, Stages: defaultStages()}
}

func IncidentPage() Page {
	return Page{Path: "/incidents", Title: "Trips and recovery", Active: "incidents", API: "/api/incidents", Stages: defaultStages()}
}

func defaultStages() []Stage {
	return []Stage{{Name: "Raw feed", Value: "210", Unit: "t/h"}, {Name: "Kiln", Value: "2.4", Unit: "rpm"}, {Name: "Burning", Value: "1450", Unit: "C"}, {Name: "Cooling", Value: "118", Unit: "C"}, {Name: "Stack CO", Value: "286", Unit: "ppm"}}
}
