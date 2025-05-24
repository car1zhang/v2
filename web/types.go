package main

import (
	"time"
)

type Photo struct {
	ID         string     `json:"id"`
	Title      string     `json:"title"`
	Timestamp  *time.Time `json:"timestamp,omitempty"`
	Precedence *int       `json:"precedence,omitempty"`
}
type Collection struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	Precedence *int   `json:"precedence,omitempty"`
}

type Response struct {
	Message string `json:"message"`
}
type PhotoResponse struct {
	Photo       Photo        `json:"photo"`
	Collections []Collection `json:"collections"`
}
type CollectionResponse struct {
	Collection Collection `json:"collection"`
	Photos     []Photo    `json:"photos"`
}
