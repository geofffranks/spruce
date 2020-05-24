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

package bunt_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/gonvenience/bunt"
)

var _ = Describe("color specific tests", func() {
	Context("fallback to 4 bit colors for non true color terminals", func() {
		BeforeEach(func() {
			ColorSetting = ON
			TrueColorSetting = OFF
		})

		AfterEach(func() {
			ColorSetting = AUTO
			TrueColorSetting = AUTO
		})

		var (
			f1 = func(color string) string {
				input := fmt.Sprintf("%s{%s}", color, "text")
				result, err := ParseString(input, ProcessTextAnnotations())
				Expect(err).ToNot(HaveOccurred())
				Expect(result).ToNot(BeNil())
				return result.String()
			}

			f2 = func(color uint8) string {
				return fmt.Sprintf("\x1b[%dm%s\x1b[0m", color, "text")
			}
		)

		It("should find a suitable 4 bit equivalent color for both foreground and background", func() {
			input := "Example: \x1b[38;2;133;247;7mforeground\x1b[0m, and \x1b[48;2;133;247;7mbackground\x1b[0m."
			result, err := ParseString(input)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).ToNot(BeNil())
			Expect(result.String()).To(BeEquivalentTo("Example: \x1b[92mforeground\x1b[0m, and \x1b[102mbackground\x1b[0m."))
		})

		It("should match known combinations of a true color to a 4 bit color", func() {
			Expect(f1("Black")).To(BeEquivalentTo(f2(30)))                // Black matches Black (#30)
			Expect(f1("Brown")).To(BeEquivalentTo(f2(31)))                // Brown matches Red (#31)
			Expect(f1("DarkRed")).To(BeEquivalentTo(f2(31)))              // DarkRed matches Red (#31)
			Expect(f1("FireBrick")).To(BeEquivalentTo(f2(31)))            // FireBrick matches Red (#31)
			Expect(f1("Maroon")).To(BeEquivalentTo(f2(31)))               // Maroon matches Red (#31)
			Expect(f1("SaddleBrown")).To(BeEquivalentTo(f2(31)))          // SaddleBrown matches Red (#31)
			Expect(f1("Sienna")).To(BeEquivalentTo(f2(31)))               // Sienna matches Red (#31)
			Expect(f1("DarkGreen")).To(BeEquivalentTo(f2(32)))            // DarkGreen matches Green (#32)
			Expect(f1("DarkSeaGreen")).To(BeEquivalentTo(f2(32)))         // DarkSeaGreen matches Green (#32)
			Expect(f1("ForestGreen")).To(BeEquivalentTo(f2(32)))          // ForestGreen matches Green (#32)
			Expect(f1("Green")).To(BeEquivalentTo(f2(32)))                // Green matches Green (#32)
			Expect(f1("LimeGreen")).To(BeEquivalentTo(f2(32)))            // LimeGreen matches Green (#32)
			Expect(f1("MediumSeaGreen")).To(BeEquivalentTo(f2(32)))       // MediumSeaGreen matches Green (#32)
			Expect(f1("Olive")).To(BeEquivalentTo(f2(32)))                // Olive matches Green (#32)
			Expect(f1("OliveDrab")).To(BeEquivalentTo(f2(32)))            // OliveDrab matches Green (#32)
			Expect(f1("SeaGreen")).To(BeEquivalentTo(f2(32)))             // SeaGreen matches Green (#32)
			Expect(f1("Gold")).To(BeEquivalentTo(f2(33)))                 // Gold matches Yellow (#33)
			Expect(f1("Yellow")).To(BeEquivalentTo(f2(33)))               // Yellow matches Yellow (#33)
			Expect(f1("Blue")).To(BeEquivalentTo(f2(34)))                 // Blue matches Blue (#34)
			Expect(f1("DarkBlue")).To(BeEquivalentTo(f2(34)))             // DarkBlue matches Blue (#34)
			Expect(f1("DarkSlateBlue")).To(BeEquivalentTo(f2(34)))        // DarkSlateBlue matches Blue (#34)
			Expect(f1("Indigo")).To(BeEquivalentTo(f2(34)))               // Indigo matches Blue (#34)
			Expect(f1("MediumBlue")).To(BeEquivalentTo(f2(34)))           // MediumBlue matches Blue (#34)
			Expect(f1("MidnightBlue")).To(BeEquivalentTo(f2(34)))         // MidnightBlue matches Blue (#34)
			Expect(f1("Navy")).To(BeEquivalentTo(f2(34)))                 // Navy matches Blue (#34)
			Expect(f1("BlueViolet")).To(BeEquivalentTo(f2(35)))           // BlueViolet matches Magenta (#35)
			Expect(f1("DarkMagenta")).To(BeEquivalentTo(f2(35)))          // DarkMagenta matches Magenta (#35)
			Expect(f1("DarkOrchid")).To(BeEquivalentTo(f2(35)))           // DarkOrchid matches Magenta (#35)
			Expect(f1("DarkViolet")).To(BeEquivalentTo(f2(35)))           // DarkViolet matches Magenta (#35)
			Expect(f1("MediumVioletRed")).To(BeEquivalentTo(f2(35)))      // MediumVioletRed matches Magenta (#35)
			Expect(f1("Purple")).To(BeEquivalentTo(f2(35)))               // Purple matches Magenta (#35)
			Expect(f1("CadetBlue")).To(BeEquivalentTo(f2(36)))            // CadetBlue matches Cyan (#36)
			Expect(f1("DarkCyan")).To(BeEquivalentTo(f2(36)))             // DarkCyan matches Cyan (#36)
			Expect(f1("DarkTurquoise")).To(BeEquivalentTo(f2(36)))        // DarkTurquoise matches Cyan (#36)
			Expect(f1("DeepSkyBlue")).To(BeEquivalentTo(f2(36)))          // DeepSkyBlue matches Cyan (#36)
			Expect(f1("LightSeaGreen")).To(BeEquivalentTo(f2(36)))        // LightSeaGreen matches Cyan (#36)
			Expect(f1("MediumAquamarine")).To(BeEquivalentTo(f2(36)))     // MediumAquamarine matches Cyan (#36)
			Expect(f1("Teal")).To(BeEquivalentTo(f2(36)))                 // Teal matches Cyan (#36)
			Expect(f1("BurlyWood")).To(BeEquivalentTo(f2(37)))            // BurlyWood matches White (#37)
			Expect(f1("DarkGoldenrod")).To(BeEquivalentTo(f2(37)))        // DarkGoldenrod matches White (#37)
			Expect(f1("DarkGray")).To(BeEquivalentTo(f2(37)))             // DarkGray matches White (#37)
			Expect(f1("Gray")).To(BeEquivalentTo(f2(37)))                 // Gray matches White (#37)
			Expect(f1("LightPink")).To(BeEquivalentTo(f2(37)))            // LightPink matches White (#37)
			Expect(f1("LightSkyBlue")).To(BeEquivalentTo(f2(37)))         // LightSkyBlue matches White (#37)
			Expect(f1("LightSlateGray")).To(BeEquivalentTo(f2(37)))       // LightSlateGray matches White (#37)
			Expect(f1("LightSteelBlue")).To(BeEquivalentTo(f2(37)))       // LightSteelBlue matches White (#37)
			Expect(f1("Pink")).To(BeEquivalentTo(f2(37)))                 // Pink matches White (#37)
			Expect(f1("RosyBrown")).To(BeEquivalentTo(f2(37)))            // RosyBrown matches White (#37)
			Expect(f1("SandyBrown")).To(BeEquivalentTo(f2(37)))           // SandyBrown matches White (#37)
			Expect(f1("Silver")).To(BeEquivalentTo(f2(37)))               // Silver matches White (#37)
			Expect(f1("Tan")).To(BeEquivalentTo(f2(37)))                  // Tan matches White (#37)
			Expect(f1("Thistle")).To(BeEquivalentTo(f2(37)))              // Thistle matches White (#37)
			Expect(f1("DarkOliveGreen")).To(BeEquivalentTo(f2(90)))       // DarkOliveGreen matches BrightBlack (#90)
			Expect(f1("DarkSlateGray")).To(BeEquivalentTo(f2(90)))        // DarkSlateGray matches BrightBlack (#90)
			Expect(f1("DimGray")).To(BeEquivalentTo(f2(90)))              // DimGray matches BrightBlack (#90)
			Expect(f1("Chocolate")).To(BeEquivalentTo(f2(91)))            // Chocolate matches BrightRed (#91)
			Expect(f1("Coral")).To(BeEquivalentTo(f2(91)))                // Coral matches BrightRed (#91)
			Expect(f1("Crimson")).To(BeEquivalentTo(f2(91)))              // Crimson matches BrightRed (#91)
			Expect(f1("DarkOrange")).To(BeEquivalentTo(f2(91)))           // DarkOrange matches BrightRed (#91)
			Expect(f1("DarkSalmon")).To(BeEquivalentTo(f2(91)))           // DarkSalmon matches BrightRed (#91)
			Expect(f1("IndianRed")).To(BeEquivalentTo(f2(91)))            // IndianRed matches BrightRed (#91)
			Expect(f1("LightCoral")).To(BeEquivalentTo(f2(91)))           // LightCoral matches BrightRed (#91)
			Expect(f1("LightSalmon")).To(BeEquivalentTo(f2(91)))          // LightSalmon matches BrightRed (#91)
			Expect(f1("OrangeRed")).To(BeEquivalentTo(f2(91)))            // OrangeRed matches BrightRed (#91)
			Expect(f1("PaleVioletRed")).To(BeEquivalentTo(f2(91)))        // PaleVioletRed matches BrightRed (#91)
			Expect(f1("Peru")).To(BeEquivalentTo(f2(91)))                 // Peru matches BrightRed (#91)
			Expect(f1("Red")).To(BeEquivalentTo(f2(91)))                  // Red matches BrightRed (#91)
			Expect(f1("Salmon")).To(BeEquivalentTo(f2(91)))               // Salmon matches BrightRed (#91)
			Expect(f1("Tomato")).To(BeEquivalentTo(f2(91)))               // Tomato matches BrightRed (#91)
			Expect(f1("Chartreuse")).To(BeEquivalentTo(f2(92)))           // Chartreuse matches BrightGreen (#92)
			Expect(f1("GreenYellow")).To(BeEquivalentTo(f2(92)))          // GreenYellow matches BrightGreen (#92)
			Expect(f1("LawnGreen")).To(BeEquivalentTo(f2(92)))            // LawnGreen matches BrightGreen (#92)
			Expect(f1("LightGreen")).To(BeEquivalentTo(f2(92)))           // LightGreen matches BrightGreen (#92)
			Expect(f1("Lime")).To(BeEquivalentTo(f2(92)))                 // Lime matches BrightGreen (#92)
			Expect(f1("MediumSpringGreen")).To(BeEquivalentTo(f2(92)))    // MediumSpringGreen matches BrightGreen (#92)
			Expect(f1("PaleGreen")).To(BeEquivalentTo(f2(92)))            // PaleGreen matches BrightGreen (#92)
			Expect(f1("SpringGreen")).To(BeEquivalentTo(f2(92)))          // SpringGreen matches BrightGreen (#92)
			Expect(f1("YellowGreen")).To(BeEquivalentTo(f2(92)))          // YellowGreen matches BrightGreen (#92)
			Expect(f1("DarkKhaki")).To(BeEquivalentTo(f2(93)))            // DarkKhaki matches BrightYellow (#93)
			Expect(f1("Goldenrod")).To(BeEquivalentTo(f2(93)))            // Goldenrod matches BrightYellow (#93)
			Expect(f1("Khaki")).To(BeEquivalentTo(f2(93)))                // Khaki matches BrightYellow (#93)
			Expect(f1("Orange")).To(BeEquivalentTo(f2(93)))               // Orange matches BrightYellow (#93)
			Expect(f1("PaleGoldenrod")).To(BeEquivalentTo(f2(93)))        // PaleGoldenrod matches BrightYellow (#93)
			Expect(f1("CornflowerBlue")).To(BeEquivalentTo(f2(94)))       // CornflowerBlue matches BrightBlue (#94)
			Expect(f1("DodgerBlue")).To(BeEquivalentTo(f2(94)))           // DodgerBlue matches BrightBlue (#94)
			Expect(f1("MediumPurple")).To(BeEquivalentTo(f2(94)))         // MediumPurple matches BrightBlue (#94)
			Expect(f1("MediumSlateBlue")).To(BeEquivalentTo(f2(94)))      // MediumSlateBlue matches BrightBlue (#94)
			Expect(f1("RoyalBlue")).To(BeEquivalentTo(f2(94)))            // RoyalBlue matches BrightBlue (#94)
			Expect(f1("SlateBlue")).To(BeEquivalentTo(f2(94)))            // SlateBlue matches BrightBlue (#94)
			Expect(f1("SlateGray")).To(BeEquivalentTo(f2(94)))            // SlateGray matches BrightBlue (#94)
			Expect(f1("SteelBlue")).To(BeEquivalentTo(f2(94)))            // SteelBlue matches BrightBlue (#94)
			Expect(f1("DeepPink")).To(BeEquivalentTo(f2(95)))             // DeepPink matches BrightMagenta (#95)
			Expect(f1("Fuchsia")).To(BeEquivalentTo(f2(95)))              // Fuchsia matches BrightMagenta (#95)
			Expect(f1("HotPink")).To(BeEquivalentTo(f2(95)))              // HotPink matches BrightMagenta (#95)
			Expect(f1("Magenta")).To(BeEquivalentTo(f2(95)))              // Magenta matches BrightMagenta (#95)
			Expect(f1("MediumOrchid")).To(BeEquivalentTo(f2(95)))         // MediumOrchid matches BrightMagenta (#95)
			Expect(f1("Orchid")).To(BeEquivalentTo(f2(95)))               // Orchid matches BrightMagenta (#95)
			Expect(f1("Plum")).To(BeEquivalentTo(f2(95)))                 // Plum matches BrightMagenta (#95)
			Expect(f1("Violet")).To(BeEquivalentTo(f2(95)))               // Violet matches BrightMagenta (#95)
			Expect(f1("Aqua")).To(BeEquivalentTo(f2(96)))                 // Aqua matches BrightCyan (#96)
			Expect(f1("Aquamarine")).To(BeEquivalentTo(f2(96)))           // Aquamarine matches BrightCyan (#96)
			Expect(f1("Cyan")).To(BeEquivalentTo(f2(96)))                 // Cyan matches BrightCyan (#96)
			Expect(f1("LightBlue")).To(BeEquivalentTo(f2(96)))            // LightBlue matches BrightCyan (#96)
			Expect(f1("MediumTurquoise")).To(BeEquivalentTo(f2(96)))      // MediumTurquoise matches BrightCyan (#96)
			Expect(f1("PaleTurquoise")).To(BeEquivalentTo(f2(96)))        // PaleTurquoise matches BrightCyan (#96)
			Expect(f1("PowderBlue")).To(BeEquivalentTo(f2(96)))           // PowderBlue matches BrightCyan (#96)
			Expect(f1("SkyBlue")).To(BeEquivalentTo(f2(96)))              // SkyBlue matches BrightCyan (#96)
			Expect(f1("Turquoise")).To(BeEquivalentTo(f2(96)))            // Turquoise matches BrightCyan (#96)
			Expect(f1("AliceBlue")).To(BeEquivalentTo(f2(97)))            // AliceBlue matches BrightWhite (#97)
			Expect(f1("AntiqueWhite")).To(BeEquivalentTo(f2(97)))         // AntiqueWhite matches BrightWhite (#97)
			Expect(f1("Azure")).To(BeEquivalentTo(f2(97)))                // Azure matches BrightWhite (#97)
			Expect(f1("Beige")).To(BeEquivalentTo(f2(97)))                // Beige matches BrightWhite (#97)
			Expect(f1("Bisque")).To(BeEquivalentTo(f2(97)))               // Bisque matches BrightWhite (#97)
			Expect(f1("BlanchedAlmond")).To(BeEquivalentTo(f2(97)))       // BlanchedAlmond matches BrightWhite (#97)
			Expect(f1("Cornsilk")).To(BeEquivalentTo(f2(97)))             // Cornsilk matches BrightWhite (#97)
			Expect(f1("FloralWhite")).To(BeEquivalentTo(f2(97)))          // FloralWhite matches BrightWhite (#97)
			Expect(f1("Gainsboro")).To(BeEquivalentTo(f2(97)))            // Gainsboro matches BrightWhite (#97)
			Expect(f1("GhostWhite")).To(BeEquivalentTo(f2(97)))           // GhostWhite matches BrightWhite (#97)
			Expect(f1("Honeydew")).To(BeEquivalentTo(f2(97)))             // Honeydew matches BrightWhite (#97)
			Expect(f1("Ivory")).To(BeEquivalentTo(f2(97)))                // Ivory matches BrightWhite (#97)
			Expect(f1("Lavender")).To(BeEquivalentTo(f2(97)))             // Lavender matches BrightWhite (#97)
			Expect(f1("LavenderBlush")).To(BeEquivalentTo(f2(97)))        // LavenderBlush matches BrightWhite (#97)
			Expect(f1("LemonChiffon")).To(BeEquivalentTo(f2(97)))         // LemonChiffon matches BrightWhite (#97)
			Expect(f1("LightCyan")).To(BeEquivalentTo(f2(97)))            // LightCyan matches BrightWhite (#97)
			Expect(f1("LightGoldenrodYellow")).To(BeEquivalentTo(f2(97))) // LightGoldenrodYellow matches BrightWhite (#97)
			Expect(f1("LightGray")).To(BeEquivalentTo(f2(97)))            // LightGray matches BrightWhite (#97)
			Expect(f1("LightYellow")).To(BeEquivalentTo(f2(97)))          // LightYellow matches BrightWhite (#97)
			Expect(f1("Linen")).To(BeEquivalentTo(f2(97)))                // Linen matches BrightWhite (#97)
			Expect(f1("MintCream")).To(BeEquivalentTo(f2(97)))            // MintCream matches BrightWhite (#97)
			Expect(f1("MistyRose")).To(BeEquivalentTo(f2(97)))            // MistyRose matches BrightWhite (#97)
			Expect(f1("Moccasin")).To(BeEquivalentTo(f2(97)))             // Moccasin matches BrightWhite (#97)
			Expect(f1("NavajoWhite")).To(BeEquivalentTo(f2(97)))          // NavajoWhite matches BrightWhite (#97)
			Expect(f1("OldLace")).To(BeEquivalentTo(f2(97)))              // OldLace matches BrightWhite (#97)
			Expect(f1("PapayaWhip")).To(BeEquivalentTo(f2(97)))           // PapayaWhip matches BrightWhite (#97)
			Expect(f1("PeachPuff")).To(BeEquivalentTo(f2(97)))            // PeachPuff matches BrightWhite (#97)
			Expect(f1("Seashell")).To(BeEquivalentTo(f2(97)))             // Seashell matches BrightWhite (#97)
			Expect(f1("Snow")).To(BeEquivalentTo(f2(97)))                 // Snow matches BrightWhite (#97)
			Expect(f1("Wheat")).To(BeEquivalentTo(f2(97)))                // Wheat matches BrightWhite (#97)
			Expect(f1("White")).To(BeEquivalentTo(f2(97)))                // White matches BrightWhite (#97)
			Expect(f1("WhiteSmoke")).To(BeEquivalentTo(f2(97)))           // WhiteSmoke matches BrightWhite (#97)
		})
	})

	Context("custom colors in text annotation", func() {
		BeforeEach(func() {
			ColorSetting = ON
			TrueColorSetting = ON
		})

		AfterEach(func() {
			ColorSetting = AUTO
			TrueColorSetting = AUTO
		})

		It("should parse hexcolors in text annotations", func() {
			Expect(Sprint("#6495ED{CornflowerBlue}")).To(
				BeEquivalentTo(Sprint("CornflowerBlue{CornflowerBlue}")))
		})
	})

	Context("random colors", func() {
		BeforeEach(func() {
			ColorSetting = ON
			TrueColorSetting = OFF
		})

		AfterEach(func() {
			ColorSetting = AUTO
			TrueColorSetting = AUTO
		})

		It("should create a list of random terminal friendly colors", func() {
			colors := RandomTerminalFriendlyColors(16)
			Expect(len(colors)).To(BeEquivalentTo(16))
		})

		It("should panic if negative number is given", func() {
			Expect(func() {
				RandomTerminalFriendlyColors(-1)
			}).To(Panic())
		})
	})
})
