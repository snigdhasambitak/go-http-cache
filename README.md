# go-http-cache
Create a simple calculator and try to render the page from a cache the 2nd time a user tries to load

## server side cache

For the static webpages which is serving millions of customers sometimes we need to do a server side caching which enables us to render the webpage faster.
After the HTML of the index page is built for an user it’s possible to cache it on the server and use this cached version to respond to all subsequent requests to the same page. By doing this on the server, we have full control on when to invalidate a given set of cached content when a new content is published.

It does now save the user from sending a HTTP request like the browser cache does, but it’ll certainly speed up the way the server responds to it

## In-Memory cache in go

A simple cache implementation in go can be done in a few lines of code. The difficult part is to decide where you want to cache the page. Common strategies are usually to store it in process memory, disk or a database. Either of these approaches are fine, but understanding the drawbacks of each of them is important to make a decision.

In our example we will be using In-Memory cache where every page is cached on your web application’s process memory, which makes it an excellent candidate for the fastest cache you’ll ever have and the easiest to implement. The drawback is that if you have multiple servers (which you should probably have), you’ll end with N copies of these cached content. If the process restarts for any reason, it’ll lose all the cached content and thus slowing down the first request again

## Implementation


First thing we have created is an interface named Storage that my application can use to get/set cached pages. It’s an interface because the application doesn’t care where it’s going to be stored.

```go
type Storage interface {
Get(key string) []byte
Set(key string, content []byte, duration time.Duration)
}
```

Then we have one structs that implement this interface, memory.Storage that uses a map object to store all the content.

The implementaton of these structs are pretty straightforward, so I’ll skip and go to the important part. If you want to give the disk strategy a try, just create a new struct that implements the interface just like the others.

`cached` is an http middleware that runs before the http handler and returns the content straight away if the page is already cached. If it’s not, the handler is executed and its body is cached for a given period of time. Because it’s a middleware, it’s really easy to enable and disable it for certain routes. Keep reading for a concrete example.

I’m using RequestURI as the key for my storage because I want to cache the pages based on different paths and querystring. This means that a page with url `/withCache` and `/withoutCache` are cached independently.

The code of the middleware is as follow.


```go
func cached(duration string, handler func(w http.ResponseWriter, r *http.Request)) http.Handler {
return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		content := storage.Get(r.RequestURI)
		if content != nil {
			fmt.Print("Cache Hit!\n")
			w.Write(content)
		} else {
			c := httptest.NewRecorder()
			handler(c, r)

			for k, v := range c.HeaderMap {
				w.Header()[k] = v
			}

			w.WriteHeader(c.Code)
			content := c.Body.Bytes()

			if d, err := time.ParseDuration(duration); err == nil {
				fmt.Printf("New page cached: %s for %s\n", r.RequestURI, duration)
				storage.Set(r.RequestURI, content, d)
			} else {
				fmt.Printf("Page not cached. err: %s\n", err)
			}

			w.Write(content)
		}

	})
}
```

To use it we just need to wrap our HTTP handler function inside a cached call, like the following. Thanks to Go’s time package, we can use human friendly string to represent a duration. For instance, 10s is much easier to understand than 10 * 1000. On the following example, only the index route is being cached.

```go
    fmt.Println("Listening on 80...")
	http.HandleFunc("/withoutCache", serveStaticFileWithoutCache)
	http.Handle("/withCache", cached("10s", serveStaticFileWithCache))
	log.Fatal(http.ListenAndServe(":80", nil))
```