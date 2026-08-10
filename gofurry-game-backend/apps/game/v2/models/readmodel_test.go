package models

import (
	"sync"
	"testing"

	"gorm.io/gorm/schema"
)

func TestSearchPageItemAppIDColumn(t *testing.T) {
	t.Parallel()

	parsed, err := schema.Parse(&GameV2SearchPageItem{}, &sync.Map{}, schema.NamingStrategy{})
	if err != nil {
		t.Fatalf("parse search page item schema: %v", err)
	}
	field := parsed.LookUpField("AppID")
	if field == nil {
		t.Fatal("AppID field is missing")
	}
	if field.DBName != "appid" {
		t.Fatalf("AppID column = %q, want appid", field.DBName)
	}
}
