package main

import (
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	// "fmt"
)

// pre-parse all templates (to help with caching)
var templates = template.Must(template.ParseFiles("tmpl/edit.html", "tmpl/view.html", "tmpl/front.html"))

// regular expression to validate page paths
var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)(/?)$")

func main() {
	// p1 := &Page{Title: "TestPage", Body: []byte("This is a sample page.")}
	// p1.save()
	// p2, _ := loadPage("TestPage")
	// fmt.Println(string(p2.Body))

	// handler bindings for
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))

	// handler to show /view/frontpage for /
	http.HandleFunc("/", frontPageHandler)

	// serve directory of static resources
	// http.Handle("/", http.FileServer(http.Dir("css/")))
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// run server
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// struct to store a web page
type Page struct {
	Title string
	Body  []byte	// page body if []byte instead of string because that's what the IO libraries we use would require
	Pages []string	
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	err := templates.ExecuteTemplate(w, tmpl+".html", p)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// save body of a Page to a text file <pagetitle>.txt
func (p *Page) save() error {
	filename := "./data/" + p.Title + ".txt"
	return ioutil.WriteFile(filename, p.Body, 0600)
}

// construct page filename for a gives page, load from file, and return pointer to Page literal
func loadPage(title string) (*Page, error) {
	filename := "./data/" + title + ".txt"
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

// wrapper taking a function literal of xHandler(resp, req, title) type and returns a function of type http.HandlerFunc
func makeHandler(fn func (http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		
		// extract page title from request and call the provided handler 'fn'
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}

		fn(w, r, m[2])
	}
}

// ----------------- handler functions for views, edits and saves -----------------
// show a frontpage with a list of already created wiki pages 
func frontPageHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/view/FrontPage/", http.StatusFound)
}

// view a webpage specified as http://server/view/webpage
func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	title = r.URL.Path[len("/view/"):]
	if title == "FrontPage/" || title == "FrontPage" {		
		// fmt.Fprintf(w, "Welcome to the %s!", title)
		// return
		
		pageTitle := "FrontPage"

		// get list of pages created
		pages, err := ioutil.ReadDir("./data")
		if err != nil {
			log.Fatal(err)
		}

		pagesList := []string{}

		if len(pages) <= 0 {
			// provide link to create first page

		} else {
			for _, page := range pages {
				pagesList = append(pagesList, page.Name())
			}
		}

		p := &Page{Title: pageTitle, Pages: pagesList}
		renderTemplate(w, "front", p)
	}

	p, err := loadPage(title)

	// requested nonexistent page - redirect to edit page to create
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}

	renderTemplate(w, "view", p)
}

// edithandler loads a given page (or creates struct if it doesn't exist) in editable form
func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)

	if err != nil {
		p = &Page{Title: title}
	}

	// fmt.Fprintf(w, "<h1>Editing %s</h1>" + "<form action=\"/save/%s\" method=\"POST\">" + "<textarea name=\"body\">%s</textarea><br>" + "<input type=\"submit\" value=\"Save\">" + "</form>", p.Title, p.Title, p.Body)
	renderTemplate(w, "edit", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	body := r.FormValue("body")

	p := &Page{Title: title, Body: []byte(body)}

	err := p.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}
