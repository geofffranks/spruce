// Copyright © 2019 The Homeport Team
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package text

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/homeport/gonvenience/pkg/v1/bunt"
)

const chars = "abcdefghijklmnopqrstuvwxyz"

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

// RandomString creates a string with random content and a given length
func RandomString(length int) string {
	if length < 0 {
		panic(fmt.Errorf("negative length value"))
	}

	tmp := make([]byte, length)
	for i := range tmp {
		tmp[i] = chars[rand.Intn(len(chars))]
	}

	return string(tmp)
}

// RandomStringWithPrefix creates a string with the provided prefix and
// additional random content so that the whole string has the given length.
func RandomStringWithPrefix(prefix string, length int) string {
	if length < 0 {
		panic(fmt.Errorf("negative length value"))
	}

	if len(prefix) > length {
		panic(fmt.Errorf("given prefix length exceeds given length"))
	}

	return prefix + RandomString(length-len(prefix))
}

// FixedLength expands or trims the given text to the provided length
func FixedLength(text string, length int) string {
	textLength := bunt.PlainTextLength(text)

	switch {
	case textLength < length: // padding required
		return text + strings.Repeat(" ", length-textLength)

	case textLength > length:
		const ellipsis = " [...]"
		return bunt.Substring(text, 0, length-len(ellipsis)) + ellipsis

	default:
		return text
	}
}

// Plural returns a string with the number and noun in either singular or plural form.
// If one text argument is given, the plural will be done with the plural s. If two
// arguments are provided, the second text is the irregular plural. If more than two
// are provided, then the additional ones are simply ignored.
func Plural(amount int, text ...string) string {
	words := [...]string{"no", "one", "two", "three", "four", "five", "six", "seven", "eight", "nine", "ten", "eleven", "twelve"}

	var number string
	if amount < len(words) {
		number = words[amount]
	} else {
		number = strconv.Itoa(amount)
	}

	switch len(text) {
	case 1:
		if amount == 1 {
			return fmt.Sprintf("%s %s", number, text[0])
		}

		return fmt.Sprintf("%s %ss", number, text[0])

	default:
		if amount == 1 {
			return fmt.Sprintf("%s %s", number, text[0])
		}

		return fmt.Sprintf("%s %s", number, text[1])
	}
}
