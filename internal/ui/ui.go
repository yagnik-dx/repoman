package ui

import "github.com/AlecAivazis/survey/v2"

// MultiSelect shows a checkbox list. defaults are pre-selected option values.
func MultiSelect(message string, options []string, defaults []string) ([]string, error) {
	var selected []string
	err := survey.AskOne(&survey.MultiSelect{
		Message: message,
		Options: options,
		Default: defaults,
	}, &selected)
	return selected, err
}

// Confirm shows a yes/no prompt.
func Confirm(message string) (bool, error) {
	var result bool
	err := survey.AskOne(&survey.Confirm{Message: message, Default: false}, &result)
	return result, err
}

// AskString shows a text input prompt with an optional default value.
func AskString(message, defaultVal string) (string, error) {
	var result string
	err := survey.AskOne(&survey.Input{Message: message, Default: defaultVal}, &result)
	return result, err
}

// Select shows a single-choice list.
func Select(message string, options []string) (string, error) {
	var result string
	err := survey.AskOne(&survey.Select{Message: message, Options: options}, &result)
	return result, err
}
