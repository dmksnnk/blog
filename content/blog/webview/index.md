---
date: '2025-05-28T22:02:23+02:00'
draft: true
title: 'Webview'
---


Today we will lear how to build a cross-platform GUI with [Webview](https://github.com/webview/webview) using its [Go bindings](https://github.com/webview/webview_go) and how to package all of it into a singe executable for easier distribution.

Webview launches a new window which uses rendering engines of a browser already present on the system to show HTMLs and run JavaSripts. This could be a nice alternative to the bulky Electon apps.

If you take a look at the [library itself](https://pkg.go.dev/github.com/webview/webview_go),
it has a really small footprint, just creating a window and bunch of operations with it: destorying window, setting size, setting HTML, JS and navigating to the page. We will use the latter.

## Architecture

The application architecture will be simple:

![Application architecture](webview_architecture.png#center)

There will be a webserver that will do all the heavy-lifting.
We will do server-side rendring using HTML templates and try to avoid JavaScript as much as possible. We will use plain HTML forms and the webserser will be responsible for all the logic. Also, it will serve us assets, like CSS and images.

The webview itself will be just a view. No Electron required.

## Showing the window

Starting up with webview is pretty simple. Here is a code to run a new window 
with a title and specific size:

```go
w := webview.New(false)
defer w.Destroy()

w.SetTitle("Hello, WebView!")
w.SetSize(480, 480, webview.HintNone)
w.Run()

```
And here is how it looks like:

![Webview window](hello_webview.png#center)

Pretty neat, huh?

## Serving the HTML

Now, let's serve our first page. For this we will need to have 
an HTTP server (let's call it a webserver) which will give us content and instruct webview to navigate to our webserver.

Usually, when you run an HTTP server, you specify on which port it will run.
We could choose a random port number, hoping the port is not already taken on the host's system. Instead, we will let the system to choose it for us.

First, we will create a TCP listener which will bind any port on localhost:

```go
// port 0 means port will be automatically chosen
listener, err := net.Listen("tcp4", "127.0.0.1:0")
```

And give the listener to HTTP server:

```go
srv := http.Server{...}
srv.Serve(listener)
```

Then, instruct webview window to navigate to the server's addres:

```go
w.Navigate("http://" + u.String())
```

Putting it all together, we are ready to serve our fist page:

```go
listener, err := net.Listen("tcp4", "127.0.0.1:0")
if err != nil {
    slog.Error("listen on TCP", "error", err)
    os.Exit(1)
}

mux := http.NewServeMux()
mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("Hello, WebView!\n"))
})
srv := http.Server{
    Handler: mux,
}

go func() {
    if err := srv.Serve(listener); err != nil {
        if !errors.Is(err, http.ErrServerClosed) {
            slog.Error("server error", "error", err)
            os.Exit(1)
        }
    }
}()

w := webview.New(false)
defer w.Destroy()

w.SetTitle("Hello, WebView!")
w.SetSize(480, 480, webview.HintNone)
w.Navigate("http://" + listener.Addr().String())
w.Run()
```

![First page](first_page.png#center)

Easy as this you now have your fist HTML served on webview!

## Interactivity

Let's add some interactivity to our page: a form and a button! We ask user a name to greed it.

We will define two HTML pages, one with a form asking for a name and another with a greeting.

Here is a simple HTML we will use. First one is a form:

{{< details summary="index.html" >}}

```html
<!DOCTYPE html>
<html lang="en">

<body>
    <h1>Hello human, what is your name?</h1>
    <form id="nameForm" action="/greet" method="post">
        <label for="name">Name:</label>
        <input type="text" id="name" name="name" placeholder="Enter your name" required>

        <button type="submit">Submit</button>
    </form>
</body>

</html>
```

{{< /details >}}

And second one is a Go [template/html](https://pkg.go.dev/html/template) for adding a name of a person to greet:

{{< details summary="greeting.html" >}}

```html
<!DOCTYPE html>
<html lang="en">

<body>
    <h1>Hello {{.name}}!</h1>
</body>

</html>
```

{{< /details >}}


Because we want our package to be self-contained, we will embed our templates into [embeddable file system](https://pkg.go.dev/embed#FS). If we put all our templates under `templates` folder, we can do next:

```go
//go:embed templates/*
var templates embed.FS
```

Now, as out templates as safely embedded, we can to parse the templates and serve them:

```go
tmpl, err := template.ParseFS(templates, "templates/*.html")
...

mux := http.NewServeMux()
mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    tmpl.ExecuteTemplate(w, "index.html", nil)
})
mux.HandleFunc("/greet", func(w http.ResponseWriter, r *http.Request) {
    if err := r.ParseForm(); err != nil {
        http.Error(w, fmt.Sprintf("whoopsies: %s", err), http.StatusInternalServerError)
        return
    }

    name := r.FormValue("name")
    data := map[string]string{"name": name}
    tmpl.ExecuteTemplate(w, "greeting.html", data)
})
```

Give it a quick try:

{{< sides >}}

![Index page](index_page.png)

![Greeting page](greeting_page.png)

{{< /sides >}}


TODO: building for windows from linux and mac