package forms

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"

	"bitbucket.org/jtyburke/pathfork/app/auth"
	"bitbucket.org/jtyburke/pathfork/app/config"
	"bitbucket.org/jtyburke/pathfork/app/sessionManager"
	"github.com/golang/glog"
)

type Form struct {
	Fields map[string]FormField
	Errors []string
}

func (f *Form) Populate(r *http.Request) error {
	if err := r.ParseForm(); err != nil {
		return err
	}
	//r.Form is url.Values => map[string][]string
	for _, field := range f.Fields {
		postVal, ok := r.Form[field.GetName()]
		if ok {
			field.SetData(postVal...)
		}
	}
	return nil
}

func (f *Form) Validate() (valid bool) {
	valid = true
	for i := range f.Fields {
		if v, err := f.Fields[i].Validate(); !v {
			glog.Info("Field error: %v", err)
			valid = false
		}
	}
	return
}

func (f *Form) AddError(err string) {
	f.Errors = append(f.Errors, err)
}

func NewFormWithFields(fields map[string]FormField) *Form {
	return &Form{
		Fields: fields,
		Errors: make([]string, 0),
	}
}

/*
..
..
*/

type FormField interface {
	Render(...string) template.HTML
	Validate() (bool, error)
	GetData() []interface{}
	GetName() string
	SetData(...string)
}

/*
..
..
*/

type StringField struct {
	Label        string
	Name         string
	Required     bool
	Error        error
	data         []string
	InputType    string
	renderFunc   func(*StringField, ...string) template.HTML
	validateFunc func(*StringField) (bool, error)
}

func NewBasicTextField(label, name string, required bool) *StringField {
	return &StringField{
		renderFunc:   basicTextFieldRender,
		validateFunc: basicTextFieldValidate,
		Label:        label,
		Name:         name,
		Required:     required,
		InputType:    "text",
	}
}

func basicTextFieldRender(t *StringField, args ...string) template.HTML {
	var str string
	required := ""
	if t.Required {
		required = "required"
	}
	if len(t.data) > 0 && t.InputType != "password" {
		str += fmt.Sprintf("<label>%v</label> <input class=\"form-control\" type=\"%v\" name=\"%v\" value=\"%v\" %v>", t.Label, t.InputType, t.Name, t.data[0], required)
	} else {
		str += fmt.Sprintf("<label>%v</label> <input class=\"form-control\" type=\"%v\" name=\"%v\" %v>", t.Label, t.InputType, t.Name, required)
	}
	if t.Error != nil {
		str += fmt.Sprintf("<br /><span class=\"form-error\">%v</span>", t.Error.Error())
	}
	return template.HTML(str)
}

func basicTextFieldValidate(sf *StringField) (bool, error) {
	if sf.Required && (len(sf.data) == 0 || len(sf.data[0]) == 0) {
		err := errors.New("This field is Required.")
		sf.Error = err
		return false, err
	}
	return true, nil
}

func (sf *StringField) Render(args ...string) template.HTML {
	return sf.renderFunc(sf, args...)
}

func (sf *StringField) Validate() (bool, error) {
	return sf.validateFunc(sf)
}

func (sf *StringField) GetData() []interface{} {
	output := make([]interface{}, len(sf.data))
	for i := range sf.data {
		output[i] = sf.data[i]
	}
	return output
}

func (sf *StringField) SetData(d ...string) {
	sf.data = d
}

func (sf *StringField) GetName() string {
	return sf.Name
}

/*
..
..
*/

func NewBasicTextAreaField(label, name string, required bool) *StringField {
	return &StringField{
		renderFunc:   basicTextAreaFieldRender,
		validateFunc: basicTextFieldValidate,
		Label:        label,
		Name:         name,
		Required:     required,
	}
}

func basicTextAreaFieldRender(sf *StringField, args ...string) template.HTML {
	var str string
	if len(args) < 2 {
		glog.Error("TextAreaField.Render() takes rows, cols args")
		return template.HTML("There's an error with this form on the server side.")
	}
	required := ""
	if sf.Required {
		required = "required"
	}
	if len(sf.data) > 0 {
		str = fmt.Sprintf(
			`<label>%v</label> <textarea rows="%v" cols="%v" class="form-control" name="%v" %v>%v</textarea>`,
			sf.Label, args[0], args[1], sf.Name, required, sf.data[0],
		)
	} else {
		str = fmt.Sprintf(
			`<label>%v</label> <textarea rows="%v" cols="%v" class="form-control" name="%v" %v></textarea>`,
			sf.Label, args[0], args[1], sf.Name, required,
		)
	}
	if sf.Error != nil {
		str += fmt.Sprintf("<br /><span class=\"form-error\">%v</span>", sf.Error.Error())
	}
	return template.HTML(str)
}

/*
..
..
*/

type CheckField struct {
	Label string
	Name  string
	Error error
	data  bool
}

func (c *CheckField) SetData(d ...string) {
	if len(d) > 0 && d[0] == "on" {
		c.data = true
	} else {
		c.data = false
	}
}

func (c *CheckField) GetData() []interface{} {
	return []interface{}{c.data}
}

func (c *CheckField) GetName() string {
	return c.Name
}

func (c *CheckField) Render(args ...string) template.HTML {
	str := ""
	if c.data == true {
		str = fmt.Sprintf("<label><input type=\"checkbox\" name=\"%v\" checked> %v</label>", c.Name, c.Label)
	} else {
		str = fmt.Sprintf("<label><input type=\"checkbox\" name=\"%v\"> %v</label>", c.Name, c.Label)

	}
	return template.HTML(str)
}

func (c *CheckField) Validate() (bool, error) {
	return true, nil
}

/*
..
..
*/

type HiddenField struct {
	Name  string
	Value string
}

func (h *HiddenField) SetData(d ...string) {
	if len(d) > 0 {
		h.Value = d[0]
	} else {
		h.Value = ""
	}
}

func (h *HiddenField) GetData() []interface{} {
	data := make([]interface{}, 1)
	data[0] = h.Value
	return data
}

func (h *HiddenField) GetName() string {
	return h.Name
}

func (h *HiddenField) Render(args ...string) template.HTML {
	str := fmt.Sprintf("<input type=\"hidden\" name=\"%v\" value=\"%v\">", h.Name, h.Value)
	return template.HTML(str)
}

func (h *HiddenField) Validate() (bool, error) {
	return true, nil
}

/*
..
..
*/

type SelectField struct {
	Label    string
	Name     string
	Error    error
	data     []string
	Required bool
	Multiple bool
	Options  []map[string]string
}

func NewSelectField(label, name string, required bool, options ...map[string]string) *SelectField {
	field := &SelectField{
		Label:    label,
		Name:     name,
		Required: required,
		Options:  options,
		Multiple: true,
	}
	newData := []string{}
	for _, option := range options {
		if option["selected"] == "true" {
			newData = append(newData, option["value"])
		}
	}
	field.data = newData
	return field
}

func (s *SelectField) SetData(d ...string) {
	s.data = d
	for _, id := range s.data {
		for i := range s.Options { // index loop keeps original object
			if s.Options[i]["value"] == id {
				s.Options[i]["selected"] = "true"
				break
			}
		}
	}
}

func (s *SelectField) GetData() []interface{} {
	data := make([]interface{}, len(s.data))
	for i := range s.data {
		data[i] = s.data[i]
	}
	return data
}

func (s *SelectField) GetName() string {
	return s.Name
}

func (s *SelectField) Validate() (bool, error) {
	if s.Required && len(s.data) == 0 {
		err := errors.New("This field is required.")
		s.Error = err
		return false, err
	}
	if !s.Multiple && len(s.data) > 1 {
		err := errors.New("Please only select one.")
		s.Error = err
		return false, err
	}
	return true, nil
}

func (s *SelectField) Render(args ...string) template.HTML {
	required := ""
	if s.Required {
		required = "required"
	}
	output := fmt.Sprintf(`<label>%v</label><select class="chosen-select form-control" name="%v" multiple="%v" %v>`, s.Label, s.Name, s.Multiple, required)
	for i := range s.Options {
		if s.Options[i]["selected"] == "true" {
			output += fmt.Sprintf(`<option value="%v" selected="true">%v</option>`, s.Options[i]["value"], s.Options[i]["text"])
		} else {
			output += fmt.Sprintf(`<option value="%v">%v</option>`, s.Options[i]["value"], s.Options[i]["text"])
		}
	}
	if s.Error != nil {
		output += fmt.Sprintf(`<span class="form-error">%v</span>`, s.Error.Error())
	}
	output += "</select>"
	return template.HTML(output)
}

/*
.
.
*/

type CSRFField struct {
	Name    string
	Value   string
	Email   string
	Manager sessionManager.SessionManager
	Error   error
}

func NewCSRFField(manager sessionManager.SessionManager) *CSRFField {
	email := manager.GetUserEmail()
	token := auth.NewTSToken(email, "csrf")
	return &CSRFField{
		Name:    "csrf",
		Value:   token,
		Email:   email,
		Manager: manager,
	}
}

func (f *CSRFField) SetData(d ...string) {
	if len(d) > 0 {
		f.Value = d[0]
	} else {
		f.Value = ""
	}
}

func (f *CSRFField) GetData() []interface{} {
	data := make([]interface{}, 1)
	data[0] = f.Value
	return data
}

func (f *CSRFField) GetName() string {
	return f.Name
}

func (f *CSRFField) Render(args ...string) template.HTML {
	str := fmt.Sprintf("<input type=\"hidden\" name=\"%v\" value=\"%v\">", f.Name, f.Value)
	if f.Error != nil {
		str += `<br /><span class="form-error">Sorry, this form expired. Please submit it again.</span>`
	}
	return template.HTML(str)
}

func (f *CSRFField) Validate() (bool, error) {
	value, valid := auth.VerifyTSToken("csrf", f.Value, config.CSRFValidTime)
	email := f.Manager.GetUserEmail()
	if !valid || value != email {
		f.Error = errors.New("Expired CSRF token")
		return false, nil
	}
	return true, nil
}
