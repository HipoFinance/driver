package util

import (
	"fmt"

	"github.com/dustin/go-humanize"
)

func GramToTonString(gram int64) string {
	return fmt.Sprintf("%v Ton", humanize.Commaf(float64(gram)/1000000000))
}

func GramString(gram int64) string {
	return fmt.Sprintf("%v Gram", humanize.Comma(gram))
}
