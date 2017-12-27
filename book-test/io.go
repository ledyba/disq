package book_test

import (
	"github.com/ledyba/disq/book"
	"github.com/ledyba/disq/conf"
	"github.com/ledyba/disq/util-test"
)

func ReadBook(t util_test.Tester, relpath string) *book.Book {
	c, err := conf.Load(util_test.ReadAll(t, relpath))
	if err != nil {
		t.Fatal(err)
	}
	b, err := book.FromConfig(c)
	if err != nil {
		t.Fatal(err)
	}
	return b
}
