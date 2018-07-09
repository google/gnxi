package gnoi

import (
	"strings"

	"github.com/google/go-cmp/cmp"
)

func FilterInternalPB(p cmp.Path) bool {
	return strings.Contains(p.String(), "XXX")
}

func Equal(x, y interface{}) bool {
	return cmp.Equal(x, y, cmp.FilterPath(FilterInternalPB, cmp.Ignore()))
}
func Diff(x, y interface{}) string {
	return cmp.Diff(x, y, cmp.FilterPath(FilterInternalPB, cmp.Ignore()))
}
