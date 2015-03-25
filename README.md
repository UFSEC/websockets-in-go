# WebSockets In Go
A tutorial on how to use WebSockets in... Go! Presented to the University of Florida Software Development Club.

# Basics
## Go
[Go](http://golang.org) is a relatively new language designed primarily for systems programming. It has a syntax very similar to C/C++. Even if you aren't familiar with Go specifically, you should be able to understand the example.

The most interesting features we'll use from Go in this talk are
* goroutines
* channels

I'll explain what these are as we need to use them.

## HTML/Javascript
I've intentionally kept all of the demos as simple as possible, so you should be able to follow along.

## WebSockets
WebSockets are also a relatively new technology. A predecessor called Comet was an unstandardized way protocol to achieve a similar communication mechanism. Its IETF [RFC](http://tools.ietf.org/html/rfc6455) was only standardized in 2011. They fill a need that is missing from the standard HTTP request/response communication paradigm, which is bi-directional stateful communication. Basically, they're like a TCP connection for a web browser.

### Why WebSockets?
Web pages don't have the ability to open TCP connections. Their primary way to get data is to initiate a HTTP request via Javascript. This HTTP request of course does happen over a TCP connection, but the TCP part is handled by the browser. TCP connections cannot be opened arbitrarily from the code of a web page. The HTTP request/response paradigm was generally the only kind of communication a web browser needed in the early days of the Internet. But now, fully-featured applications in the browser are the norm, and there are some kinds of situations that don't work well using the requst/response model of HTTP. Namely, this model makes it impossible for the server to *push* data to the client.

### Simple Example
Think for a minute about how Twitter might deliver tweets to your browser. Twitter tries to deliver tweets as quick as possible when they are tweeted out. To keep your timeline updating in real-time, data needs to arrive to the browser as quick as possible.

If you're using the request/response paradigm of HTTP, then you have to use *polling*, which is where you periodically send a request to the server, asking for new data. There are two main disadvantages to this approach:
1. If there is no new data, then the request is essentially a waste of resources and causes unnecessary load on the server and client.
2. To have truly real-time data with polling, you need to poll very frequently, which magnifies the wasted resources even more.

If you don't need real time data, then polling might be a perfectly acceptable solution. But if you app is time-sensitive with respect to its data, then polling can be very wasteful.

What if instead of having to ask the server for new data, without knowing if your request will actually yield new data, the server could just send the new data to the client as soon as it became available? This idea sounds great, but unfortunately it doesn't fit into the request/response paradigm of HTTP, where the client requests data, receives data, and then doesn't get any more data unless it specifically asks for it. Clearly we need some other way of communicating with the server, and that other way is **WebSockets**.

### Like TCP For Your Webpages
WebSockets can be conceptually thought of as a TCP connection. For those who don't know, the basic properties of a TCP connection are
* it's bi-directional (full-duplex)
* arbitrary bytes can be sent on it
* data is guaranteed to be delivered in the order it was sent

*Note*: if your client is not a web page, then just go ahead and use TCP, you probably don't need to use WebSockets.

The biggest difference between a WebSocket and a TCP connection is how they are initiated. A WebSocket begins its life as an HTTP request which includes special headers in the request telling the server that the client wishes to upgrade this HTTP request into a full WebSocket connection. We'll see this happen in our demo later.

### How do we use WebSockets?
In the browser, we can initiate a connection server using Javascript. We'll see the syntax for exactly how to do this in the demo.

On the server side, it entirely depends on the language/web framework you're using. Most popular web frameworks have support for WebSockets. In our case  we're using Go, and we're going to use the WebSocket library from the [Gorilla web toolkit](http://www.gorillatoolkit.org/pkg/websocket).

But enough blabbing, let's see the code!

## Demos
### Example 1
In this example we have a very simple echo server to show you the basics of how things work.

### Example 2
This example requires a bit of explanation on a couple of the features of Go.
#### Channels
Go has a construct called *channels*. Channels are are "conduits" to pass information through. Channels have the following properties:
* have a type
* uni-directional (insert at one end, remove from the other)
* are FIFO (first-in, first-out, just like queue)
* have a size
* removes will block if channel is empty
* inserts will block if channel is full
* safe to concurrent access

Here's an example using a channel
```
var channel = make(chan int, 10) // a channel of ints, size 10
channel <- 5            // insert 5 into the channel
var num int = <-channel // remove first value from channel
fmt.Println(num)        // prints "5"
```

Note the use of the `<-` operator. This operator is used for both inserting into and removing from a channel. Which operation it performs depends on which side of the operator the channel/value are on.

If you use it like this,
```
channel <- value
```
this is inserting the value into the channel.

If you flip the statement,
```
value = <-channel
```
you are
1. Removing the first value from the channel
2. assigning the value to the variable `value`

Note that `<-channel` returns the value of the channel, so you can use it anywhere you want to put that value. So you could pass it to a function call
```
functionCall(<-channel)
```

Just remember, the data flows in the direction of the arrow.

Try out the example [here](http://play.golang.org/p/yQYRIHPjw1).

#### Goroutines
The other important feature of Go we need to talk about is goroutines. A *goroutine* is a "lighweight thread" that is managed by the Go runtime. So it isn't technically an actual OS thread, but can conceptually be thought of as a thread when using one.

Go has a really simple syntax for starting goroutines, which is to prefix any function call with the keyword `go`. How nice!

```
func countTo(x int) {
    for i := 0; i < x; i++ {
        fmt.Println(i)
    }
}

func main() {
    countTo(4)      // will execute on this current goroutine
    go countTo(4)   // will execute in its own goroutine
}
```
When you run this, you'll see the prints from the "main" thread and "goroutine" thread interleaved. This is because they are executing concurrently.

Try out the example [here](http://play.golang.org/p/ro2rGtNjlE).


#### Gotcha! JSON Serialization
