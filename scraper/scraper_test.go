package scraper

import (
	"golang.org/x/net/html"
	"strings"
	"testing"
)

func MustParse(doc string) *html.Node {
	n, err := html.Parse(strings.NewReader(doc))
	if err != nil {
		panic("cannot parse document")
	}
	return n
}

func TestParseContent(t *testing.T) {
	type args struct {
		root *html.Node
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "basic",
			args: args{
				root: MustParse(`<p>test</p>`),
			},
			want: "test ",
		},
		{
			name: "nested",
			args: args{
				root: MustParse(`<p>Hello <b>World</b></p>`),
			},
			want: "Hello World ",
		},
		{
			name: "ignore",
			args: args{
				root: MustParse(`<body>
	<header>
		<p>test</p>
	</header>
	<main>
		<p>test</p>
	</main>
</body>`),
			},
			want: "test ",
		},
		{
			name: "complex",
			args: args{
				root: MustParse(`<body>
	<header>
		<p>test</p>
	</header>
	<main>
		<p>test</p>
		<p>test</p>
	</main>
</body>`),
			},
			want: "test test ",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseContent(tt.args.root); got != tt.want {
				t.Errorf("parseContent() = %v, want %v", got, tt.want)
			}
		})
	}
}
