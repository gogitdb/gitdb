package db

import (
	"fmt"
	"html/template"
	"net/http"
)

func StartGUI() {
	appHost := "localhost"
	port := 4120

	eps := GetGUIEndpoints()
	for _, ep := range eps {
		http.HandleFunc(ep.Path, ep.Handler)
	}

	addr := fmt.Sprintf("%s:%d", appHost, port)
	fmt.Println("GitDB GUI will run at http://" + addr)

	err := http.ListenAndServe(":4120", nil)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Server started!")
	}
}

type Endpoint struct {
	Path    string
	Handler func(w http.ResponseWriter, r *http.Request)
}

func GetGUIEndpoints() []*Endpoint {
	return []*Endpoint{
		&Endpoint{"/gitdb", handler},
	}
}

type ViewModel struct {
	Title    string
	DataSets []string
}

var html = `<html>
<head></head>
<body>

	<h1>{{.Title}}</h1>

	<h2>Data Sets</h2>
	<table border="1">
		<tr>
			<th>Name</th>
			<th>No. of blocks</th>
			<th>No. of records</th>
			<th>Size</th>
			<th>Created</th>
			<th>Last Modified</th>
		</tr>
		{{range $key, $value := .DataSets}} 
		<tr>
			<td>{{ $value }}</td>
			<td>-</td>
			<td>-</td>
			<td>-</td>
			<td>-</td>
			<td>-</td>
		</tr>
		{{end}}
	</table>


</body>
</html>
`

func handler(w http.ResponseWriter, r *http.Request) {

	viewModel := &ViewModel{Title: "DB GUI", DataSets: listDatasets()}
	t, _ := template.New("tt").Parse(html)
	t.Execute(w, viewModel)
}
