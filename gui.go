package db

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"
)

var DataSets []*DataSet

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
	DataSets = LoadDatasets()
	endpoints := []*Endpoint{
		&Endpoint{"/gitdb", overview},
	}

	for _, ds := range DataSets {
		endpoints = append(endpoints, &Endpoint{"/gitdb/view/" + ds.Name, view})
	}

	return endpoints
}

type OverviewViewModel struct {
	Title    string
	DataSets []*DataSet
}

type ViewDataSetViewModel struct {
	DataSets []*DataSet
	DataSet *DataSet
	Content string
}

var css =`div {
    box-sizing: border-box;
  }

  h1 {
    padding: 0;
    margin:  0;
  }

  h1 a {
    text-decoration:none;
    color: darkseagreen
  }

  .sidebar {
    float:            left;
    width:            20%;
    height:           800px;
    background-color: #ccc;
    padding:          5px;
  }

  .content {
    padding:     30px;
    padding-top: 10px;
    float:       left;
    width:       80%;
    height:      800px;
  }

  .nav {
    list-style: none;
    margin:     0;
    padding:    0;
  }

  .nav li {
    color: #000;
  }

  .nav a {
    color:            #000;
    text-decoration:  none;
    display:          block;
    padding:          12px;
    border-bottom:    1px solid #ddd;
  }

  .nav a:hover {
    background-color: #ddd;
  }

  table tr:hover td {
    background-color: #ccc;
  }

  table th {
    background-color: darkseagreen;
    color:#fff;
  }

  table {
    width:          100%;
    border:         1px solid #000;
    border-spacing: 0px;
  }

  table td, table th {
    padding: 10px;
    border:  1px solid #000;
  }

  pre {
    background-color: #333;
    color:            #fff;
    padding:          10px;
    font-size:        14px;
  }`

func overview(w http.ResponseWriter, r *http.Request) {

	var html = `<html>
<head></head>
<style type="text/css">
  `+css+`
</style>
<body>

<div class="sidebar">
  <h1><a href="/gitdb">GitDb</a></h1>
  <hr>
  <strong>Data Sets</strong>
  <ul class="nav">
  {{range $key, $value := .DataSets}}
    <li><a href="/gitdb/view/{{ $value.Name }}">{{ $value.Name }}</a></li>
  {{end}}
  </ul>
</div>
<div class="content">
  <h1>{{.Title}}</h1>

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
      <td>{{ $value.Name }}</td>
      <td>{{ $value.BlockCount }}</td>
      <td>{{ $value.RecordCount }}</td>
      <td>{{ $value.HumanSize }}</td>
      <td>-</td>
      <td>-</td>
    </tr>
  {{end}}
  </table>
</div>

</body>
</html>
`

	viewModel := &OverviewViewModel{Title: "DB GUI", DataSets: DataSets}
	t, _ := template.New("tt").Parse(html)
	t.Execute(w, viewModel)
}

func view(w http.ResponseWriter, r *http.Request) {

	var html = `<html>
<head></head>
<style type="text/css">
  `+css+`
</style>
<body>

<div class="sidebar">
  <h1><a href="/gitdb">GitDb</a></h1>
  <hr>
  <strong>Data Sets</strong>
  <ul class="nav">
  {{range $key, $value := .DataSets}}
    <li><a href="/gitdb/view/{{ $value.Name }}">{{ $value.Name }}</a></li>
  {{end}}
  </ul>
</div>
<div class="content">
  <h1>{{.DataSet.Name}}</h1>
  <div><span>{{.DataSet.BlockCount}} blocks</span> <span>{{.DataSet.HumanSize}}</span></div>

  <pre>
    {{.Content}}
  </pre>
</div>


</body>
</html>
`

	fmt.Println(r.URL.Path)

	viewDs := strings.Replace(r.URL.Path, "/gitdb/view/","", -1)
	var dataSet *DataSet
	for _, ds := range DataSets {
		if ds.Name == viewDs {
			dataSet = ds
			break
		}
	}

	block := dataSet.Blocks[0]
	block.LoadRecords()
	content := block.Records[0].Content

	viewModel := &ViewDataSetViewModel{DataSets: DataSets, DataSet: dataSet, Content: content}
	t, _ := template.New("tt").Parse(html)
	t.Execute(w, viewModel)
}
