package forms

import (
	"fmt"
)

func NewContactForm() *Form {
	return NewFormWithFields(
		map[string]FormField{
			"email":   NewBasicTextField("Email", "email", true),
			"message": NewBasicTextAreaField("Message", "message", true),
		},
	)
}

func NewSignupForm() *Form {
	passwordField := NewBasicTextField("Password", "password", true)
	passwordField.InputType = "password"
	repeatField := NewBasicTextField("Repeat Password", "repeatPassword", true)
	repeatField.InputType = "password"
	return NewFormWithFields(
		map[string]FormField{
			"email":          NewBasicTextField("Email", "email", false),
			"password":       passwordField,
			"repeatPassword": repeatField,
		},
	)
}

func NewSigninForm() *Form {
	passwordField := NewBasicTextField("Password", "password", true)
	passwordField.InputType = "password"
	return NewFormWithFields(
		map[string]FormField{
			"email":    NewBasicTextField("Email", "email", false),
			"password": passwordField,
		},
	)
}

func NewRequestResetPasswordForm() *Form {
	return NewFormWithFields(
		map[string]FormField{
			"email": NewBasicTextField("Email address", "email", true),
		},
	)
}

func NewResetPasswordForm() *Form {
	passwordField := NewBasicTextField("New Password", "newPassword", true)
	passwordField.InputType = "password"
	repeatField := NewBasicTextField("Repeat Password", "repeatPassword", true)
	repeatField.InputType = "password"
	return NewFormWithFields(
		map[string]FormField{
			"newPassword":    passwordField,
			"repeatPassword": repeatField,
		},
	)
}

func NewWorkForm(characterOptions []map[string]string, settingOptions []map[string]string) *Form {
	currentCharIds := GetCurrentIds(characterOptions)
	characters := NewSelectField("Characters", "characters", false, characterOptions...)
	currentSettingIds := GetCurrentIds(settingOptions)
	settings := NewSelectField("Settings", "settings", false, settingOptions...)
	return NewFormWithFields(
		map[string]FormField{
			"title":             NewBasicTextField("Title", "title", true),
			"blurb":             NewBasicTextAreaField("Blurb", "blurb", false),
			"characters":        characters,
			"currentCharIds":    &HiddenField{Name: "currentCharIds", Value: currentCharIds},
			"settings":          settings,
			"currentSettingIds": &HiddenField{Name: "currentSettingIds", Value: currentSettingIds},
		},
	)
}

func NewSectionForm(characterOptions []map[string]string, settingOptions []map[string]string) *Form {
	currentCharIds := GetCurrentIds(characterOptions)
	currentSettingIds := GetCurrentIds(settingOptions)
	characters := NewSelectField("Characters", "characters", false, characterOptions...)
	settings := NewSelectField("Settings", "settings", false, settingOptions...)
	snippet := &CheckField{Name: "snippet", Label: "This is a snippet"}
	return NewFormWithFields(
		map[string]FormField{
			"title":             NewBasicTextField("Section Title", "title", true),
			"blurb":             NewBasicTextAreaField("Blurb", "blurb", false),
			"body":              NewBasicTextAreaField("Body", "body", false),
			"characters":        characters,
			"settings":          settings,
			"snippet":           snippet,
			"currentCharIds":    &HiddenField{Name: "currentCharIds", Value: currentCharIds},
			"currentSettingIds": &HiddenField{Name: "currentSettingIds", Value: currentSettingIds},
		},
	)
}

func NewCharacterForm() *Form {
	return NewFormWithFields(
		map[string]FormField{
			"name":  NewBasicTextField("Name", "name", true),
			"blurb": NewBasicTextAreaField("Blurb", "blurb", false),
			"body":  NewBasicTextAreaField("Body", "body", false),
		},
	)
}

func NewSettingForm() *Form {
	return NewFormWithFields(
		map[string]FormField{
			"name":    NewBasicTextField("Name", "name", true),
			"blurb":   NewBasicTextAreaField("Blurb", "blurb", false),
			"body":    NewBasicTextAreaField("Body", "body", false),
			"work_id": &HiddenField{Name: "work_id"},
		},
	)
}

func NewDeleteForm(objId int) *Form {
	return NewFormWithFields(
		map[string]FormField{
			"id": &HiddenField{Name: "object_id", Value: fmt.Sprintf("%v", objId)},
		})
}
