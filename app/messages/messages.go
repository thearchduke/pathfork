package messages

import "bitbucket.org/jtyburke/pathfork/app/forms"

type Message struct {
	From    string
	Message string
}

func (m Message) Send() error {
	return nil
}

func NewFromContactForm(f *forms.Form) Message {
	return Message{
		From:    f.Fields["name"].GetData()[0].(string),
		Message: f.Fields["message"].GetData()[0].(string),
	}
}
