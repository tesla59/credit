package main

func Check(e error) {
	if e != nil {
		panic(e)
	}
}

func Markup(text string) string {
	return "<code>" + text + "</code>"
}
