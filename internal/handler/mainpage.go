package handler

import "text/template"

// AI-generated, as I couldn't figure out how to dereference pointers.
const mainPageTpl = `
<!DOCTYPE html>
<html>
<head>
    <title>Metrics Service</title>
</head>
<body>
    <h2>Welcome to the Metrics service</h2>
    {{if .}}
        <p>ID&nbsp;&nbsp;&nbsp;&nbsp;Value</p>
        {{range .}}
            <p>
                {{.ID}}&nbsp;&nbsp;&nbsp;&nbsp;
                {{if eq .MType "counter"}}
                    {{derefInt .Delta}}
                {{else if eq .MType "gauge"}}
                    {{printf "%.2f" (derefFloat .Value)}}
                {{end}}
            </p>
        {{end}}
    {{else}}
        <p>Currently there are no metrics to display</p>
    {{end}}
</body>
</html>
`

var templateFuncs = template.FuncMap{
	"derefFloat": func(f *float64) float64 {
		if f == nil {
			return 0.0
		}
		return *f
	},
	"derefInt": func(i *int64) int64 {
		if i == nil {
			return 0
		}
		return *i
	},
}
