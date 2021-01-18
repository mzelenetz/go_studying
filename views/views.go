package views

import (
	"html/template"
	"path/filepath"
	"net/http"
)

var (
	LayoutDir string = "views/layouts/"
	TemplateBot string = ".gohtml"
	TemplateDir string = "views/"
	TemplateExt string = ".gohtml"
)

func NewView(layout string, files ...string) *View {
	addTemplatePath(files)
	addTemplateExt(files)
	
	files = append(files, layoutFiles()...)

	t, err := template.ParseFiles(files...)
	if err != nil {
		panic(err)
	}

	return &View{
		Template: t,
		Layout: layout,
	}
}

type View struct {
	Template *template.Template
	Layout string //you can call this whatever you want
}

func (v *View) ServeHTTP(w http.ResponseWriter, r *http.Request){
	if err := v.Render(w, nil); err != nil{
		panic(err)
	}
}

// Render is used to to render the view with predefined layout
func (v *View) Render(w http.ResponseWriter, data interface{}) error {
	w.Header().Set("Content-Type", "text/html")
	return v.Template.ExecuteTemplate(w, v.Layout, data)
}

// layoutFiles returns a slice of strings with
// the layoutfiles that are used in our application
func layoutFiles() []string {
	files, err := filepath.Glob( LayoutDir + "*" + TemplateBot) 
	if err != nil {
		panic(err)
	}
	return files
}

// addTemplatePath takes in a slice of strings of file paths
// for templates and it prepends the 
// TemplateDir directory to each string in the slice
//
// Ex the input {"home"} would output {"views/home"}
// If TemplateDir == "views/"
func addTemplatePath(files []string){
	for i, f := range files {
		files[i] = TemplateDir + f
	}
}

// addTemplateExt takes in a slice of stirngs
// of files paths for templates
// it appends the TemplateExt extension for each string in
// the slice
// 
// Ex. The input {"home"} outputs {"home.gohtml"}
// assuming TemplateExt == ".gohtml"
func addTemplateExt(files []string){
	for i, f := range files {
		files[i] = f + TemplateExt
	}
}