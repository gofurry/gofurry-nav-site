package models

import (
	"reflect"
	"testing"
)

func TestSearchPageItemAppIDColumn(t *testing.T) {
	t.Parallel()

	field, ok := reflect.TypeOf(GameV2SearchPageItem{}).FieldByName("AppID")
	if !ok {
		t.Fatal("AppID field is missing")
	}
	if column := field.Tag.Get("db"); column != "appid" {
		t.Fatalf("AppID column = %q, want appid", column)
	}
}
