package variable_replacer

import (
	"fmt"
	"github.com/google/uuid"
	"strings"
	"time"
)

var replacerMap = make(map[string]string, 0)

func VariableReplacer(t time.Time, s string) string {
	initTimeVariables(t)
	replacerMap["%uuid%"] = uuid.New().String()

	newString := s
	for k, v := range replacerMap {
		newString = strings.ReplaceAll(newString, k, v)
	}

	return newString
}

func initTimeVariables(t time.Time) {
	replacerMap["%year%"] = fmt.Sprintf("%s", t.Format("2006"))
	replacerMap["%year_short%"] = fmt.Sprintf("%s", t.Format("06"))
	replacerMap["%month%"] = fmt.Sprintf("%s", t.Format("01"))
	replacerMap["%month_name%"] = fmt.Sprintf("%s", t.Format("January"))
	replacerMap["%month_name_short%"] = fmt.Sprintf("%s", t.Format("Jan"))
	replacerMap["%day%"] = fmt.Sprintf("%s", t.Format("02"))
	replacerMap["%hour%"] = fmt.Sprintf("%s", t.Format("15"))
	replacerMap["%minute%"] = fmt.Sprintf("%s", t.Format("04"))
	replacerMap["%second%"] = fmt.Sprintf("%s", t.Format("05"))
	replacerMap["%timezone%"] = fmt.Sprintf("%s", t.Format("Z07:00:00"))
	replacerMap["%unix%"] = fmt.Sprintf("%d", t.Unix())
}
