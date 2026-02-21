package models

import (
	"net/http"
	"strconv"
)

const (
	defaultPage    = 1
	defaultPerPage = 20
	maxPerPage     = 100
)

type PaginationParams struct {
	Page    int
	PerPage int
}

func (p PaginationParams) Offset() int {
	return (p.Page - 1) * p.PerPage
}

type PaginationMeta struct {
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

func NewPaginationMeta(params PaginationParams, total int) PaginationMeta {
	totalPages := 0
	if total > 0 && params.PerPage > 0 {
		totalPages = (total + params.PerPage - 1) / params.PerPage
	}
	return PaginationMeta{
		Page:       params.Page,
		PerPage:    params.PerPage,
		Total:      total,
		TotalPages: totalPages,
	}
}

func ParsePaginationParams(r *http.Request) PaginationParams {
	page := defaultPage
	perPage := defaultPerPage

	if p := r.URL.Query().Get("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}

	if pp := r.URL.Query().Get("per_page"); pp != "" {
		if v, err := strconv.Atoi(pp); err == nil && v > 0 {
			perPage = v
		}
	}

	if perPage > maxPerPage {
		perPage = maxPerPage
	}

	return PaginationParams{
		Page:    page,
		PerPage: perPage,
	}
}
