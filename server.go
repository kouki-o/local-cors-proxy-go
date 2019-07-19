package main

import (
	"fmt"
	"log"
	"regexp"

	. "github.com/logrusorgru/aurora"
	"github.com/valyala/fasthttp"
)

func getServer(options *Options) *fasthttp.Server {
	proxyHttpClient := &fasthttp.Client{}
	proxyPathPattern := regexp.MustCompile(`^\/` + options.cleanUrlSection)

	requestHandler := func(ctx *fasthttp.RequestCtx) {
		proxyPath := string(ctx.Path())

		if proxyPathPattern.MatchString(proxyPath) {
			proxyRequestHandler(
				ctx,
				proxyHttpClient,
				options,
				proxyPathPattern.ReplaceAllString(proxyPath, ``),
			)
		} else {
			ctx.Error("Not found", fasthttp.StatusNotFound)
		}
	}

	server := &fasthttp.Server{
		Handler: requestHandler,
	}

	go func() {
		err := server.ListenAndServe(options.addr)
		if err != nil {
			log.Fatal(err)
		} else {
			fmt.Println(Red("Shutted down!\n").Bold())
		}
	}()

	return server
}

func proxyRequestHandler(ctx *fasthttp.RequestCtx, proxyHttpClient *fasthttp.Client, options *Options, proxyPath string) {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)
	res := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(res)

	proxiedUri := options.cleanUrl + proxyPath

	ctx.Request.CopyTo(req)
	req.SetRequestURI(proxiedUri)
	for headerName, headerValue := range options.parsedHeaders {
		req.Header.Set(headerName, headerValue)
	}

	if err := proxyHttpClient.Do(req, res); err != nil {
		ctx.Error(err.Error(), 500)
	}

	res.Header.Set("Access-Control-Allow-Origin", "*")
	res.WriteTo(ctx.Conn())

	defer fmt.Printf(
		Sprintf(
			"%s %s %s %s %s %s %d\n",
			Magenta(ctx.Method()),
			Blue("request proxied:"),
			Green(ctx.RequestURI()),
			Blue("->"),
			Green(proxiedUri),
			Blue("with status code"),
			White(res.StatusCode()),
		),
	)
}
