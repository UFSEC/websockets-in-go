# Websockets In Go
A tutorial on how to use websockets in... Go! Presented to the University of Florida Software Development Club.

# Basics
## Go (http://golang.org)
This is a relatively new language designed primarily for systems programming. It has a syntax very similar to C/C++. Even if you aren't familiar with Go specifically, you should be able to understand the example.

The most interesting features we'll use from Go in this talk are
* Go routines
* channels

I'll explain what these are as we need to use them.

## Websockets
Websockets are also a relatively new technology. They fill a need that is missing from the standard HTTP request/response communication paradigm, which is bi-directional stateful communication. Basically, they're like a TCP connection for a web browser.

### Why Websockets?
Web pages don't have the ability to open TCP connections. Their primary way to get data is to initiate an HTTP request via Javascript. This HTTP request of course does happen over a TCP connection, but the TCP part is handled by the browser. This was generally the only kind of communication a web browser needed in the early days of the Internet. But now, fully-featured applications in the browser are the norm, and there are some kinds of situations that don't work well using the requst/response model of HTTP. Namely, this model makes it impossible for the server to *push* data to the client.

### Simple Example
Think for a minute about how Twitter might deliver tweets to your browser.
