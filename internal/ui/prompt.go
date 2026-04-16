package ui

import survey "github.com/AlecAivazis/survey/v2"

// Confirm prompts the user with a yes/no question and returns the answer.
func Confirm(message string) (bool, error) {
	var result bool
	prompt := &survey.Confirm{Message: message}
	err := survey.AskOne(prompt, &result)
	return result, err
}
