# WebSockets In Go
A tutorial on how to use WebSockets in... Go! Presented to the University of Florida Software Development Club.

# Basics
## Go (http://golang.org)
This is a relatively new language designed primarily for systems programming. It has a syntax very similar to C/C++. Even if you aren't familiar with Go specifically, you should be able to understand the example.

The most interesting features we'll use from Go in this talk are
* Go routines
* channels

I'll explain what these are as we need to use them.

## WebSockets
WebSockets are also a relatively new technology. They fill a need that is missing from the standard HTTP request/response communication paradigm, which is bi-directional stateful communication. Basically, they're like a TCP connection for a web browser.

### Why WebSockets?
Web pages don't have the ability to open TCP connections. Their primary way to get data is to initiate a HTTP request via Javascript. This HTTP request of course does happen over a TCP connection, but the TCP part is handled by the browser. TCP connections cannot be opened arbitrarily from the code of a web page. The HTTP request/response paradigm was generally the only kind of communication a web browser needed in the early days of the Internet. But now, fully-featured applications in the browser are the norm, and there are some kinds of situations that don't work well using the requst/response model of HTTP. Namely, this model makes it impossible for the server to *push* data to the client.

### Simple Example
Think for a minute about how Twitter might deliver tweets to your browser. Twitter tries to deliver tweets as quick as possible when they are tweeted out. To keep your timeline updating in real-time, data needs to arrive to the browser as quick as possible.

If you're using the request/response paradigm of HTTP, then you have to use *polling*, which is where you periodically send a request to the server, asking for new data. There are two main disadvantages to this approach:
1. If there is no new data, then the request is essentially a waste of resources and causes unnecessary load on the server and client.
2. To have truly real-time data with polling, you need to poll very frequently, which magnifies the wasted resources even more.

If you don't need real time data, then polling might be a perfectly acceptable solution. But if you app is time-sensitive with respect to its data, then polling can be very wasteful.

What if instead of having to ask the server for new data, without knowing if your request will actually yield new data, the server could just send the new data to the client as soon as it became available? This idea sounds great, but unfortunately it doesn't fit into the request/response paradigm of HTTP, where the client requests data, receives data, and then doesn't get any more data unless it specifically asks for it. Clearly we need some other way of communicating with the server, and that other way is **WebSockets**.
