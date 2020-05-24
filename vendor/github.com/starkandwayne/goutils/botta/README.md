# botta - A simple to use JSON API parser in Go

`botta` is a simple-to-use library for sending/retrieving
arbitrary data to/from arbitrary HTTP APIs using JSON.

It has helper functions for retreiving data paths, taking
most of the hassle out of type-casting.

# Examples:

```
    // let's assume that http://example.com returns the following json:
	// {
	//    "stringVal": "this is a string value",
	//    "numVal":   1234,
	//    "mapVal": {
	//        "key": "value"
	//    },
	//    "arrayVal": [
	//        "first",
	//        "second",
	//        {
	//            "subkey": "subval"
	//        }
	//    ]
	// }
    req, err := botta.Get("http://example.com")
    if err != nil {
        panic("Couldn't create a request?")
    }
    resp, err := botta.Issue(req)
    if err != nil {
		if resp != nil {
			msg = resp.StringVal("error")
		    panic(fmt.Sprintf("Request failed: %s", resp.StringVal("error")))
		} else {
			panic(err.Error())
		}
	}

	// strVal = "this is a string value"
	strVal, err := resp.StringVal("stringVal")

	numVal, err := resp.NumVal("numVal")
	// i = 1234
	i, err := numVal.Int64()
	// f = 1234.0
	f := numVal.Float64()
	// s = "1234"
	s := numVal.String()
	
	// mapVal = map[string]interface{}{"key": "value"}
	mapVal, err := resp.MapVal("mapVal")

	// arrVal = []interface{}{"first","second",map[string]interface{}{"subkey":"subval"}}
	// arrVal, err := resp.ArrayVal("arraVal")

	// second = "second"
	second, err := resp.StringVal("arrayVal.[0]")

	// subval = "subval"
	subval, err := resp.StringVal("arrayVal.[2].subkey")
```
