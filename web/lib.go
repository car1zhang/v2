package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
)

var ctx = context.Background()

func serveStaticTemplate(w http.ResponseWriter, templatePath string) {
	tmpl, err := template.ParseFiles("templates/layout.tmpl.html", templatePath)
	if err != nil {
		http.Error(w, "Failed to parse template: "+err.Error(), http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, "Failed to serve template: "+err.Error(), http.StatusInternalServerError)
	}
}

func fetchData[T any](url string) (T, error) {
	var zeroVal T

	fmt.Println(url)
	res, err := http.Get(url)
	if err != nil {
		return zeroVal, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return zeroVal, fmt.Errorf("unexpected status: %s", res.Status)
	}

	var data T
	if err = json.NewDecoder(res.Body).Decode(&data); err != nil {
		return zeroVal, err
	}

	return data, nil
}

func fillTemplateWithData[T any](templatePath string, data T) (bytes.Buffer, error) {
	var zeroVal bytes.Buffer

	tmpl, err := template.ParseFiles("templates/layout.tmpl.html", templatePath)
	if err != nil {
		return zeroVal, err
	}

	var pageBuf bytes.Buffer
	err = tmpl.Execute(&pageBuf, data)
	if err != nil {
		return zeroVal, err
	}

	return pageBuf, nil
}

func sendTemplateWithAPIData[T any](w http.ResponseWriter, urlPath string, apiEndpoint string, templatePath string) {
	w.Header().Set("Content-Type", "text/html")

	if sendPageIfCached(w, urlPath) {
		return
	}

	data, err := fetchData[T](apiEndpoint)
	if err != nil {
		http.Error(w, "Failed to fetch data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	pageBuf, err := fillTemplateWithData(templatePath, data)
	if err != nil {
		http.Error(w, "Failed to fill template: "+err.Error(), http.StatusInternalServerError)
		return
	}

	writePageToCache(pageBuf, urlPath)
	w.Write(pageBuf.Bytes())
}
