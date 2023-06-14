package tests

import (
	"github.com/brianvoe/gofakeit/v6"
	"testing"
)

func TestSelf(t *testing.T) {
	t.Log(gofakeit.Name())

}
