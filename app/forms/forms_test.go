package forms

import (
	"testing"

	"bitbucket.org/jtyburke/pathfork/app/models"
)

func TestFormValidate(t *testing.T) {
	field := NewBasicTextField("Username", "username", false)
	form := NewFormWithFields(map[string]FormField{"username": field})
	if v := form.Validate(); !v {
		t.Errorf("Valid form vailed to validate")
	}
	fieldNotBlank := NewBasicTextField("Email", "email", true)
	form = NewFormWithFields(map[string]FormField{
		"username": field,
		"email":    fieldNotBlank,
	})
	if v := form.Validate(); v {
		t.Errorf("Invalid form validated")
	}
}

func TestWorksToFormOptions(t *testing.T) {
	works := []*models.Work{
		&models.Work{Id: 1, Title: "Title 1"},
		&models.Work{Id: 2, Title: "Title 2"},
	}
	selected := []*models.Work{
		&models.Work{Id: 2, Title: "Title 2"},
	}
	options := WorksToFormOptions(works, selected...)
	if options[0]["value"] != "1" && options[0]["name"] != "Title 1" && options[0]["selected"] != "" {
		t.Errorf("Value 1 does not match")
	}
	if options[1]["value"] != "2" && options[1]["name"] != "Title 2" && options[1]["selected"] != "true" {
		t.Errorf("Value 1 does not match")
	}
}

func TestGetCurrentIds(t *testing.T) {
	options := make([]map[string]string, 2)
	options[0] = map[string]string{"value": "1", "name": "Title 1"}
	options[1] = map[string]string{"value": "2", "name": "Title 2", "selected": "true"}
	ids := GetCurrentIds(options)
	if ids != "2" {
		t.Errorf("Expected 2, got %v", ids)
	}
	options[1] = map[string]string{"value": "2", "name": "Title 2"}
	ids = GetCurrentIds(options)
	if ids != "" {
		t.Errorf("Expected nil string, got %v", ids)
	}
}
