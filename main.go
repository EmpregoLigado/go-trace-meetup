package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"contrib.go.opencensus.io/exporter/jaeger"
	"github.com/gorilla/mux"
	exp "go.opencensus.io/examples/exporter"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/trace"
)

const collectorURL = "http://0.0.0.0:14268/api/traces"

//initializes and replaces global constants.
//keep in mind that some of the overwritten variables
//here could result in unexpected behaviors.
//that's  ok for an example, but may cause side effects
//on bigger applications.
func init() {
	exporter, err := jaeger.NewExporter(jaeger.Options{
		CollectorEndpoint: collectorURL,
		Process: jaeger.Process{
			ServiceName: "Fibonacci",
		},
	})
	if err != nil {
		log.Printf("error initializing jaeger tracing backend: %s", err)
		return
	}
	fmt.Printf("Tracing initialized: %s\n", collectorURL)
	trace.ApplyConfig(trace.Config{
		DefaultSampler: trace.AlwaysSample(),
	})
	trace.RegisterExporter(exporter)
	trace.RegisterExporter(&exp.PrintExporter{})

	http.DefaultClient = &http.Client{
		Transport: &ochttp.Transport{
			NewClientTrace: ochttp.NewSpanAnnotatingClientTrace,
		},
	}
}

//main entry point of our example.
//needless to say that it's not safe for production environments.
func main() {
	mux := mux.NewRouter()
	mux.Path("/fib").HandlerFunc(fibHandler)
	handler := &ochttp.Handler{
		Handler: mux,
	}
	if err := http.ListenAndServe(":8080", handler); err != nil {
		log.Println(err)
	}
}

//the format of the http answers
type result struct {
	Result int `json:"result"`
}

//fibHandler its the entrypoint for the http requests.
//it converts params to the right types and formats the response.
func fibHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	ranks, ok := r.Form["rank"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Missing rank param"))
		return
	}
	rank, err := strconv.Atoi(ranks[0])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	res := result{
		Result: fibCalc(r.Context(), rank),
	}
	result, err := json.Marshal(res)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(result)
}

//fibCalc is the Fibonacci implementation
//doing the recursion using http calls.
//It's inefficient but shows in one app http traces.
//The scope of the span will include all the sub calls.
func fibCalc(ctx context.Context, rank int) int {
	cCtx, cSpan := trace.StartSpan(ctx, strconv.Itoa(rank))
	defer cSpan.End()
	if rank == 0 || rank == 1 {
		return rank
	}
	return fibReq(cCtx, rank-2) + fibReq(cCtx, rank-1)
}

//wrapper for the http request for the app itself.
//note that the request receives the context to pass it down.
func fibReq(ctx context.Context, rank int) int {
	url := fmt.Sprintf("http://0.0.0.0:8080/fib?rank=%d", rank)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}
	req = req.WithContext(ctx)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	res := result{}
	err = json.Unmarshal(body, &res)
	if err != nil {
		log.Fatal(err)
	}
	return res.Result
}
