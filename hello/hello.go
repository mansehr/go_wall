package hello

import (
    "appengine"
    "appengine/datastore"
    "appengine/user"
    "fmt"
    "html/template"
    "net/http"
    "time"
)

type Greeting struct {
    Author  string
    Content string
    Date    time.Time
}

func init() {
    http.HandleFunc("/", root)
    http.HandleFunc("/sign", sign)
}

func root(w http.ResponseWriter, r *http.Request) {
    c := appengine.NewContext(r)
    u := user.Current(c)
    q := datastore.NewQuery("Greeting").Order("-Date").Limit(100)
    greetings := make([]Greeting, 0, 100)
    if _, err := q.GetAll(c, &greetings); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    fmt.Fprint(w, headerHTML)
    if err := guestbookTemplate.Execute(w, greetings); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
    if u == nil {
	url, err := user.LoginURL(c, "/")
	if err != nil {
		fmt.Fprintf(w, `Could not connect to google autentication`)
	} else {
	        fmt.Fprintf(w, `<a href="%s">Sign in or register to post a message</a>`, url)
	}
    } else {
	url, err := user.LogoutURL(c, "/")
	if err != nil {
		fmt.Fprint(w, `Could not connect to google autentication`)
	} else {
	        fmt.Fprintf(w, `Welcome, %s! (<a href="%s">sign out</a>)`, u, url)
	}
	fmt.Fprint(w, formHTML)
    }
    fmt.Fprint(w, footerHTML)
}

var guestbookTemplate = template.Must(template.New("book").Parse(guestbookTemplateHTML))

const headerHTML = `
<html>
  <head>
  <title>Guestbook</title>
  <script src="https://ajax.googleapis.com/ajax/libs/jquery/2.1.4/jquery.min.js"></script>
  <script src="https://cdnjs.cloudflare.com/ajax/libs/jQuery-linkify/1.1.7/jquery.linkify.min.js"></script>
  </head>
  <body>
	<h3>Feel free to login and post a message!</h3>
	<hr/>
`

const guestbookTemplateHTML = `
    {{range .}}
      {{with .Author}}
        <p><b>{{.}}</b> wrote:</p>
      {{else}}
        <p>An anonymous person wrote:</p>
      {{end}}
      <p class="comment">{{.Content}}</p>
    {{end}}
`

const formHTML = `
    <form action="/sign" method="post">
      <div><textarea name="content" rows="4" cols="60"></textarea></div>
      <div><input type="submit" value="Post message"></div>
    </form>
`

const footerHTML = `
  </body>
  <script>
  $('p.comment').linkify({
    target: "_blank"
    });
  </script>
</html>
`

func sign(w http.ResponseWriter, r *http.Request) {
    c := appengine.NewContext(r)
    g := Greeting{
        Content: r.FormValue("content"),
        Date:    time.Now(),
    }
    if u := user.Current(c); u != nil {
        g.Author = u.String()
    } else {
	redirectToLogin(w, r, c)
	return
    }
    if g.Content == "" {
	http.Error(w, "No content in post", http.StatusNoContent)
	return
    }
    _, err := datastore.Put(c, datastore.NewIncompleteKey(c, "Greeting", nil), &g)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    http.Redirect(w, r, "/", http.StatusFound)
}

func redirectToLogin(w http.ResponseWriter, r *http.Request, c appengine.Context) {
	url, err := user.LoginURL(c, r.URL.String())
        if err != nil {
        	http.Error(w, err.Error(), http.StatusInternalServerError)
                return
        }
        w.Header().Set("Location", url)
        w.WriteHeader(http.StatusFound)
        return
}
