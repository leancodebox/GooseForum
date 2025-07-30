package searchservice

import (
	"testing"
)

// TestSearchRequest 测试搜索请求结构
func TestSearchRequest(t *testing.T) {
	req := SearchRequest{
		Query:      "测试",
		Categories: []uint64{1, 2},
		Limit:      10,
		Offset:     0,
	}

	if req.Query != "测试" {
		t.Errorf("Expected query to be '测试', got %s", req.Query)
	}

	if len(req.Categories) != 2 {
		t.Errorf("Expected 2 categories, got %d", len(req.Categories))
	}

	if req.Limit != 10 {
		t.Errorf("Expected limit to be 10, got %d", req.Limit)
	}
}

// TestJoinFilters 测试过滤条件连接函数
func TestJoinFilters(t *testing.T) {
	// 测试空数组
	result := joinFilters([]string{}, " OR ")
	if result != "" {
		t.Errorf("Expected empty string, got %s", result)
	}

	// 测试单个过滤条件
	result = joinFilters([]string{"category = 1"}, " OR ")
	if result != "category = 1" {
		t.Errorf("Expected 'category = 1', got %s", result)
	}

	// 测试多个过滤条件
	result = joinFilters([]string{"category = 1", "category = 2"}, " OR ")
	expected := "category = 1 OR category = 2"
	if result != expected {
		t.Errorf("Expected '%s', got %s", expected, result)
	}

	// 测试三个过滤条件
	result = joinFilters([]string{"category = 1", "category = 2", "category = 3"}, " OR ")
	expected = "category = 1 OR category = 2 OR category = 3"
	if result != expected {
		t.Errorf("Expected '%s', got %s", expected, result)
	}
}
