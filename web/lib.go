package main

import (
    "bytes"
    "fmt"
    "time"
    "context"
    "net/http"
    "html/template"
    "encoding/json"
)


var ctx = context.Background()

func serveStaticTemplate(w http.ResponseWriter, templatePath string) {
    tmpl, err := template.ParseFiles("templates/layout.tmpl.html", templatePath)
    if err != nil {
        http.Error(w, "Failed to parse template: " + err.Error(), http.StatusInternalServerError)
        return
    }

    err = tmpl.Execute(w, nil)
}

func fetchData[T any](url string) (T, error) {
    var zeroVal T

    res, err := http.Get(url)
    if err != nil {
        return zeroVal, err
    }
    defer res.Body.Close()
    if res.StatusCode != http.StatusOK {
        return zeroVal, fmt.Errorf("Unexpected status: %d", res.StatusCode)
    }

    var data T
    if err = json.NewDecoder(res.Body).Decode(&data); err != nil {
        return zeroVal, err
    }

    return data, nil
}

func sendTemplateWithAPIData[T any](w http.ResponseWriter, url string, templatePath string) {
    w.Header().Set("Content-Type", "text/html")

    rdb := getRedisClient()
    defer rdb.Close()

    cachedPage, err := rdb.Get(ctx, "PAGE:" + url).Result()
    if err == nil { // cache hit
        w.Write([]byte(cachedPage))
        return
    }

    data, err := fetchData[T](url)
    if err != nil {
        http.Error(w, "Failed to fetch data: " + err.Error(), http.StatusInternalServerError)
        return
    }

    tmpl, err := template.ParseFiles("templates/layout.tmpl.html", templatePath)
    if err != nil {
        http.Error(w, "Failed to parse template: " + err.Error(), http.StatusInternalServerError)
        return
    }

    var pageBuf bytes.Buffer
    err = tmpl.Execute(&pageBuf, data)
    if err != nil {
        http.Error(w, "Failed to render template: " + err.Error(), http.StatusInternalServerError)
        return
    }

    rdb.Set(ctx, "PAGE:" + url, pageBuf.String(), time.Hour*24)
    w.Write(pageBuf.Bytes())
}

