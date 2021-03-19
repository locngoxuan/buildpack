package utils

import "fmt"

const (
	colorReset = "\033[0m"

	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
	colorWhite  = "\033[37m"
)

func TextColor(color, s string) string {
	return fmt.Sprintf("%s%s%s", color, s, string(colorReset))
}

func TextRed(s string) string {
	return TextColor(string(colorRed), s)
}

func TextGreen(s string) string {
	return TextColor(string(colorGreen), s)
}

func TextYello(s string) string {
	return TextColor(string(colorYellow), s)
}

func TextBlue(s string) string {
	return TextColor(string(colorBlue), s)
}

func TextPurple(s string) string {
	return TextColor(string(colorPurple), s)
}

func TextCyan(s string) string {
	return TextColor(string(colorCyan), s)
}

func TextWhite(s string) string {
	return TextColor(string(colorWhite), s)
}
