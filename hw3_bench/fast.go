package main

import (
	"bufio"
	json "encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"

	easyjson "github.com/mailru/easyjson"
	jlexer "github.com/mailru/easyjson/jlexer"
	jwriter "github.com/mailru/easyjson/jwriter"
)

func main() {
	//FastSearch(os.Stdout)
}

var (
	_ *json.RawMessage
	_ *jlexer.Lexer
	_ *jwriter.Writer
	_ easyjson.Marshaler
)

func easyjson3486653aDecodeCourseraHw3Bench(in *jlexer.Lexer, out *User) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeString()
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "name":
			out.Name = string(in.String())
		case "email":
			out.Email = string(in.String())
		case "browsers":
			if in.IsNull() {
				in.Skip()
				out.Browsers = nil
			} else {
				in.Delim('[')
				if out.Browsers == nil {
					if !in.IsDelim(']') {
						out.Browsers = make([]string, 0, 4)
					} else {
						out.Browsers = []string{}
					}
				} else {
					out.Browsers = (out.Browsers)[:0]
				}
				for !in.IsDelim(']') {
					var v1 string
					v1 = string(in.String())
					out.Browsers = append(out.Browsers, v1)
					in.WantComma()
				}
				in.Delim(']')
			}
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjson3486653aEncodeCourseraHw3Bench(out *jwriter.Writer, in User) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"name\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.String(string(in.Name))
	}
	{
		const prefix string = ",\"email\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.String(string(in.Email))
	}
	{
		const prefix string = ",\"browsers\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		if in.Browsers == nil && (out.Flags&jwriter.NilSliceAsEmpty) == 0 {
			out.RawString("null")
		} else {
			out.RawByte('[')
			for v2, v3 := range in.Browsers {
				if v2 > 0 {
					out.RawByte(',')
				}
				out.String(string(v3))
			}
			out.RawByte(']')
		}
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v User) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjson3486653aEncodeCourseraHw3Bench(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v User) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson3486653aEncodeCourseraHw3Bench(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *User) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson3486653aDecodeCourseraHw3Bench(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *User) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson3486653aDecodeCourseraHw3Bench(l, v)
}

var dataPool = sync.Pool{
	New: func() interface{} {
		return new(User)
	},
}

//easyjson:json
type User struct {
	Name     string   `json:"name"`
	Email    string   `json:"email"`
	Browsers []string `json:"browsers"`
}

func FastSearch(out io.Writer) {
	counter := -1
	uniq := 0
	seen := make(map[string]bool, 256)

	file, err := os.Open(filePath)
	defer file.Close()
	if err != nil {
		panic(err)
	}

	fmt.Fprintln(out, "found users:")

	reader := bufio.NewReader(file)
	for {
		line, err := reader.ReadSlice('\n')
		if err != nil {
			break
		}

		user := dataPool.Get().(*User)

		err = user.UnmarshalJSON(line)
		if err != nil {
			panic(err)
		}
		dataPool.Put(user)

		userAndroid, userMSIE := false, false
		for _, browser := range user.Browsers {
			isSeen, ok := seen[browser]

			m := strings.Contains(browser, "MSIE")
			a := strings.Contains(browser, "Android")

			if !ok {
				seen[browser] = m || a
				if a || m {
					uniq++
				}
			} else {
				if !isSeen {
					a, m = false, false
				}
			}
			userAndroid = userAndroid || m
			userMSIE = userMSIE || a
		}

		counter++
		if !(userMSIE && userAndroid) {
			continue
		}

		info := strings.Builder{}
		info.Grow(7 + len(user.Name) + len(user.Email) + 5)
		info.WriteByte('[')
		info.WriteString(strconv.Itoa(counter))
		info.WriteByte(']')
		info.WriteByte(' ')
		info.WriteString(user.Name)
		info.WriteByte(' ')
		info.WriteByte('<')
		info.WriteString(strings.Replace(user.Email, "@", " [at] ", 1))
		info.WriteByte('>')
		fmt.Fprintln(out, info.String())
		info.Reset()
	}

	fmt.Fprintln(out, "\nTotal unique browsers", uniq)
}
