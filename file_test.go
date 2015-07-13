package main

import "testing"
import . "github.com/smartystreets/goconvey/convey"

func TestReadFile(t *testing.T) {
	Convey("When reading file <512 bytes", t, func() {
		path := "assets/readFile/small"
		content, err := readFile(path)
		Convey("There should be no errors", func() {
			So(err, ShouldBeNil)
		})
		Convey("The correct data was read", func() {
			So(string(content), ShouldEqual, "under 512 bytes of data to read\n")
		})
	})

	Convey("When reading a file >512 bytes", t, func() {
		path := "assets/readFile/large"
		content, err := readFile(path)
		Convey("There should be no error", func() {
			So(err, ShouldBeNil)
		})
		Convey("The correct data was read", func() {
			So(string(content), ShouldEqual,
				"over 512 bytes of data to readover 512 bytes of data to readover 512 bytes of data to readover 512 bytes of data to readover 512 bytes of data to readover 512 bytes of data to readover 512 bytes of data to readover 512 bytes of data to readover 512 bytes of data to readover 512 bytes of data to readover 512 bytes of data to readover 512 bytes of data to readover 512 bytes of data to readover 512 bytes of data to readover 512 bytes of data to readover 512 bytes of data to readover 512 bytes of data to readover 512 bytes of data to readover 512 bytes of data to readover 512 bytes of data to readover 512 bytes of data to readover 512 bytes of data to readover 512 bytes of data to readover 512 bytes of data to readover 512 bytes of data to readover 512 bytes of data to readover 512 bytes of data to readover 512 bytes of data to readover 512 bytes of data to readover 512 bytes of data to readover 512 bytes of data to read\n",
			)
		})
	})

	Convey("When reading a nonexistent file", t, func() {
		_, err := readFile("nonexistent/path")
		Convey("readFile() should fail", func() {
			So(err, ShouldNotBeNil)
		})
	})
}

func BenchmarkReadFile(b *testing.B) {
	for i := 0; i < b.N; i++ {
		path := "assets/readFile/small"
		_, err := readFile(path)
		if err != nil {
			b.Fatalf("Error reading %s: %s - benchmarking will be invalid\n", path, err.Error())
		}

		path = "assets/readFile/large"
		_, err = readFile(path)
		if err != nil {
			b.Fatalf("Error reading %s: %s - benchmarking will be invalid\n", path, err.Error())
		}
	}
}
