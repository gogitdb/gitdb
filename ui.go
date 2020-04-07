//go:generate gitdb embed-ui -o ./ui_static.go
package gitdb

import (
	"context"
	"encoding/base64"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

var fs *fileSystem

func getFs() *fileSystem {
	if fs == nil {
		fs = &fileSystem{}
	}
	return fs
}

func (g *gitdb) startUI() {

	server := &http.Server{
		Addr:    fmt.Sprintf("localhost:%d", g.config.UIPort),
		Handler: new(router).configure(g.config),
	}

	log("GitDB GUI will run at http://" + server.Addr)
	go func() {
		if err := server.ListenAndServe(); err != nil {
			logError(err.Error())
		}
	}()

	//listen for shutdown event
	go func() {
		<-g.shutdown
		if server != nil {
			server.Shutdown(context.TODO())
		}
		logTest("shutting down UI server")
		return
	}()
}

type fileSystem struct {
	files map[string]string
}

func (e *fileSystem) embed(name, src string) {
	if e.files == nil {
		e.files = make(map[string]string)
	}
	e.files[name] = src
}

func (e *fileSystem) has(name string) bool {
	_, ok := e.files[name]
	return ok
}

func (e *fileSystem) get(name string) []byte {
	content := e.files[name]
	decoded, err := base64.StdEncoding.DecodeString(content)
	if err != nil {
		logError(err.Error())
		return []byte("")
	}

	return decoded
}

//router provides all the http handlers for the UI
type router struct {
	datasets  []*dataset
	refreshAt time.Time
}

func (u *router) configure(cfg Config) *mux.Router {
	router := mux.NewRouter()
	for path, handler := range u.getEndpoints() {
		router.HandleFunc(path, handler)
	}

	//refresh dataset after 1 minute
	router.Use(func(h http.Handler) http.Handler {
		if u.refreshAt.IsZero() || u.refreshAt.Before(time.Now()) {
			u.datasets = loadDatasets(cfg)
			u.refreshAt = time.Now().Add(time.Second * 10)
		}

		return h
	})

	return router
}

//getEndpoints maps a path to a http handler
func (u *router) getEndpoints() map[string]http.HandlerFunc {
	return map[string]http.HandlerFunc{
		"/css/app.css":              u.appCSS,
		"/js/app.js":                u.appJS,
		"/":                         u.overview,
		"/errors/{dataset}":         u.viewErrors,
		"/list/{dataset}":           u.list,
		"/view/{dataset}":           u.view,
		"/view/{dataset}/b{b}/r{r}": u.view,
	}
}

func (u *router) appCSS(w http.ResponseWriter, r *http.Request) {
	src := readView("static/css/app.css")
	w.Header().Set("Content-Type", "text/css")
	w.Write(src)
}

func (u *router) appJS(w http.ResponseWriter, r *http.Request) {
	src := readView("static/js/app.js")
	w.Header().Set("Content-Type", "text/javascript")
	w.Write(src)
}

func (u *router) overview(w http.ResponseWriter, r *http.Request) {
	viewModel := &overviewViewModel{}
	viewModel.Title = "Overview"
	viewModel.DataSets = u.datasets

	render(w, viewModel, "static/index.html", "static/sidebar.html")
}

func (u *router) list(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	viewDs := vars["dataset"]

	dataset := u.findDataset(viewDs)
	if dataset == nil {
		w.Write([]byte("Dataset (" + viewDs + ") does not exist"))
		return
	}

	block := dataset.Block(0)
	table := block.table()
	viewModel := &listDataSetViewModel{DataSet: dataset, Table: table}
	viewModel.DataSets = u.datasets

	render(w, viewModel, "static/list.html", "static/sidebar.html")
}

func (u *router) view(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	viewDs := vars["dataset"]

	dataset := u.findDataset(viewDs)
	if dataset == nil {
		w.Write([]byte("Dataset (" + viewDs + ") does not exist"))
		return
	}

	viewModel := &viewDataSetViewModel{
		DataSet: dataset,
		Content: "No record found",
		Pager:   &pager{totalBlocks: dataset.BlockCount()},
	}
	viewModel.DataSets = u.datasets
	if vars["b"] != "" && vars["r"] != "" {
		viewModel.Pager.set(vars["b"], vars["r"])
	}
	block := dataset.Block(viewModel.Pager.blockPage)
	viewModel.Block = block
	viewModel.Pager.totalRecords = block.RecordCount()
	if viewModel.Pager.totalRecords > viewModel.Pager.recordPage {
		viewModel.Content = block.Records[viewModel.Pager.recordPage].data
	}

	render(w, viewModel, "static/view.html", "static/sidebar.html")
}

func (u *router) viewErrors(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	viewDs := vars["dataset"]

	dataset := u.findDataset(viewDs)
	if dataset == nil {
		w.Write([]byte("Dataset (" + viewDs + ") does not exist"))
		return
	}
	viewModel := &errorsViewModel{DataSet: dataset}
	viewModel.Title = "Errors"
	viewModel.DataSets = u.datasets

	render(w, viewModel, "static/errors.html", "static/sidebar.html")
}

func (u *router) findDataset(name string) *dataset {
	for _, ds := range u.datasets {
		if ds.Name == name {
			return ds
		}
	}
	return nil
}

func render(w http.ResponseWriter, data interface{}, templates ...string) {

	parseFiles := false
	for _, template := range templates {
		if !getFs().has(template) {
			parseFiles = true
		}
	}

	var t *template.Template
	var err error
	if parseFiles {
		t, err = template.ParseFiles(templates...)
		if err != nil {
			logError(err.Error())
		}
	} else {
		t = template.New("overview")
		for _, template := range templates {
			logTest("Reading EMBEDDED file - " + template)
			t, err = t.Parse(string(getFs().get(template)))
			if err != nil {
				logError(err.Error())
			}
		}
	}

	t.Execute(w, data)
}

func readView(fileName string) []byte {
	if getFs().has(fileName) {
		return getFs().get(fileName)
	}

	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		logError(err.Error())
		return []byte("")
	}

	return data
}
