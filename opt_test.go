package opt

import (
	"testing"

	rawomit "github.com/aarondl/opt/omit"
)

func TestOpt_1(t *testing.T) {
	opt := rawomit.From(1)
	t.Log(opt)
}
