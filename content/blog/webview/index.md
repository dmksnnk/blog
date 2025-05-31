---
date: '2025-05-28T22:02:23+02:00'
draft: true
title: 'Webview'
showToc: true

tags:
    - webview
    - Go
---

Today we will lear how to build a cross-platform GUI with [Webview](https://github.com/webview/webview) using its [Go bindings](https://github.com/webview/webview_go) and how to package all of it into a singe executable for easier distribution.

Webview launches a new window which uses rendering engines of a browser already present on the system to show HTMLs and run JavaSripts. This could be a nice alternative to the bulky Electon apps, but still be able to build a beautiful app using Web technologies.

If you take a look at the [library itself](https://pkg.go.dev/github.com/webview/webview_go),
it has a really small footprint, just creating a window and bunch of operations with it: destorying window, setting size, setting HTML, JavaScript and navigating to the page. We will use the latter.

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
{{< details summary="webserver.go" >}}


```go
package main

import (
	webview "github.com/webview/webview_go"
)

func main() {
	w := webview.New(false)
	defer w.Destroy()

	w.SetTitle("Hello, WebView!")
	w.SetSize(480, 480, webview.HintNone)
	w.Run()
}

```

{{< /details >}}

`go run ./webserver.go` and here is how it looks like:

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

{{< details summary="webserver.go" >}}

```go
package main

import (
    "errors"
    "log/slog"
    "net"
    "net/http"
    "os"

    webview "github.com/webview/webview_go"
)

func main() {
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
}

```

{{< /details >}}

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

## Glamour

Let's learn how to serve static files. We will serve CSS to add some beauty 💅 to the pages.

Create a file `style.css` in the `static` folder and put there some CSS magic:

{{< details summary="style.css" >}}

```css
body {
    font-family: Consolas, monospace;
    margin: 40px auto;
    padding: 20px;
    text-align: center;
}

input {
    padding: 8px;
    margin: 10px;
    border-radius: 4px;
    border: 1px solid #ddd;
}

button {
    background-color: #4CAF50;
    color: white;
    padding: 10px 20px;
    border: none;
    border-radius: 4px;
    cursor: pointer;
}
```

{{< /details >}}

Now, embed all files in the static folder:

```go
//go:embed static/*
var static embed.FS
```

And serve them:

```go
mux.Handle("/static/", http.FileServerFS(static))
```

Last step is to reference this CSS in our HTML. Add the next lines to the files in `templates` directory, just after `html` tag:

```html
<head>
    <link rel="stylesheet" href="/static/style.css">
</head>
```

Run the webserver again and see how it shines! ✨

You can use the same technique to serve images, JavaScript or anything you'd like.

## Cross-compiling for Windows

To be able to your application from Linux or Mac for Windows, 
you will need the [mingw-w64](https://www.mingw-w64.org/).

For various flavors of Linux, check the link above for the instructions. For example, here is for [Debian/Ubuntu](https://www.mingw-w64.org/getting-started/debian/).

For Mac, install it via brew:

```sh
brew install mingw-w64
```

### Building

To build it, you must use CGO and provide correct flags to the compiler, 
here what's works for me:

```sh
GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc CXX=x86_64-w64-mingw32-g++ \
    go build -ldflags "-H windowsgui" -v -o webview.exe ./webview.go
```

## Recap

Hopefully, we have learned how to build a simple Webview app using local web server for renderign pages and serving static files.

Now go and explose the endless sea of possibilities you can do with it.
