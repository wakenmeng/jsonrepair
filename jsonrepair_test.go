package jsonrepair

import (
	"testing"
)

func TestParsing(t *testing.T) {
	ts := []struct {
		name  string
		cases []string
	}{
		{
			name: "parse full JSON object",
			cases: []string{
				`{"a":2.3e100,"b":"str","c":null,"d":false,"e":[1,2,3]}`,
			},
		},
		{
			name: "parse whitespace",
			cases: []string{
				"  { \n } \t ",
			},
		},
		{
			name: "parse object",
			cases: []string{
				`{}`,
				`{"a": {}}`,
				`{"a": "b"}`,
				`{"a": 2}`,
			},
		},
		{
			name: "parse array",
			cases: []string{
				`[]`,
				`[{}]`,
				`[1,2,3]`,
				`[ 1 , 2 , 3 ]`,
				`[1,2,[3,4,5]]`,
				`[{}]`,
				`{"a":[]}`,
				`[1, "hi", true, false, null, {}, []]`,
			},
		},
		{
			name: "parse number",
			cases: []string{
				`23`,
				`0`,
				`0e+2`,
				`0.0`,
				`-0`,
				`2.3`,
				`2300e3`,
				`2300e+3`,
				`2300e-3`,
				`-2`,
				`2e-3`,
				`2.3e-3`,
			},
		},
		{
			name: "parse string",
			cases: []string{
				`"str"`,
				`"\"\\\/\b\f\n\r\t"`,
				`"\u260E"`,
			},
		},
		{
			name: "parse keywords",
			cases: []string{
				`true`,
				`false`,
				`null`,
			},
		},
		{
			name: "correctly handle strings equaling a JSON delimiter",
			cases: []string{
				`""`,
				`"["`,
				`"]"`,
				`"{"`,
				`"}"`,
				`":"`,
				`","`,
			},
		},
		{
			name: "supports unicode characters in a string",
			cases: []string{
				`"★"`,
				`"\u2605"`,
				`"😀"`,
				`"\ud83d\ude00"`,
				`"йнформация"`,
			},
		},
		{
			name: "supports escaped unicode characters in a string",
			cases: []string{
				`"\u2605"`,
				`"\u2605A"`,
				`"\ud83d\ude00"`,
				`"\u0439\u043d\u0444\u043e\u0440\u043c\u0430\u0446\u0438\u044f"`,
			},
		},
		{
			name: "supports unicode characters in a key",
			cases: []string{
				`{"★":true}`,
				`{"\u2605":true}`,
				`{"😀":true}`,
				`{"\ud83d\ude00":true}`,
			},
		},
	}
	for _, tt := range ts {
		caseHasErr := false
		for _, text := range tt.cases {
			parsed, err := JSONRepair(text)
			if parsed != text {
				t.Errorf("failed on group: %s, case: %s, got: %s, err: %v", tt.name, text, parsed, err)
				caseHasErr = true
			}
		}
		if !caseHasErr {
			t.Logf("cases passed for group: %s", tt.name)
		}
	}
}

func TestRepairValidJSON(t *testing.T) {
	type Case struct {
		Input string
		Want  string
	}
	ts := []struct {
		name  string
		cases []Case
	}{
		{
			name: "should add missing quotes",
			cases: []Case{
				{`abc`, `"abc"`},
				{`hello   world`, `"hello   world"`},
				{`{a:2}`, `{"a":2}`},
				{`{a: 2}`, `{"a": 2}`},
				{`{2: 2}`, `{"2": 2}`},
				{`{true: 2}`, `{"true": 2}`},
				{"{\n  a: 2\n}", "{\n  \"a\": 2\n}"},
				{`[a,b]`, `["a","b"]`},
				{"[\na,\nb\n]", "[\n\"a\",\n\"b\"\n]"},
			},
		},
		{
			name: "should add missing end quote",
			cases: []Case{
				{`"abc`, `"abc"`},
				{`'abc`, `"abc"`},
				{"\u2018abc", `"abc"`},
			},
		},
		{
			name: "should repair truncated JSON",
			cases: []Case{
				{`"foo`, `"foo"`},
				{`[`, `[]`},
				{`["foo`, `["foo"]`},
				{`["foo"`, `["foo"]`},
				{`["foo",`, `["foo"]`},
				{`{"foo":"bar"`, `{"foo":"bar"}`},
				{`{"foo":"bar`, `{"foo":"bar"}`},
				{`{"foo":`, `{"foo":null}`},
				{`{"foo"`, `{"foo":null}`},
				{`{"foo`, `{"foo":null}`},
				{`{`, `{}`},
				{`2.`, `2.0`},
				{`2e`, `2e0`},
				{`2e+`, `2e+0`},
				{`2e-`, `2e-0`},
				{`{"foo":"bar\u20`, `{"foo":"bar"}`},
				{`"\u`, `""`},
				{`"\u2`, `""`},
				{`"\u260`, `""`},
				{`"\\u2605`, `"\\u2605"`},
				{`{"s \ud`, `{"s": null}`},
			},
		},
		{
			name: "should add missing start quote",
			cases: []Case{
				{`abc"`, `"abc"`},
				{`[a","b"]`, `["a","b"]`},
				{`[a",b"]`, `["a","b"]`},
				{`{"a":"foo","b":"bar"}`, `{"a":"foo","b":"bar"}`},
				{`{a":"foo","b":"bar"}`, `{"a":"foo","b":"bar"}`},
				{`{"a":"foo",b":"bar"}`, `{"a":"foo","b":"bar"}`},
				{`{"a":foo","b":"bar"}`, `{"a":"foo","b":"bar"}`},
			},
		},
		// {`[\n"abc,  \n"def"\n]'), '[\n"abc").toBe( \n"def"\n]')
		{
			name: "should stop at the first next return when missing an end quote",
			cases: []Case{
				{"[\n\"abc,\n\"def\"\n]", "[\n\"abc\",\n\"def\"\n]"},
				{"[\"abc]\n", "[\"abc\"]\n"},
				{"[\"abc  ]\n", "[\"abc\"  ]\n"},
			},
		},
		{
			name: "should replace single quotes with double quotes",
			cases: []Case{
				{"{'a':2}", `{"a":2}`},
				{"{'a':'foo'}", `{"a":"foo"}`},
				{`{"a":\'foo\'}`, `{"a":"foo"}`},
				{"{a:'foo',b:'bar'}", `{"a":"foo","b":"bar"}`},
			},
		},
		{
			name: `should replace special quotes with double quotes`,
			cases: []Case{
				{`{“a”:“b”}`, `{"a":"b"}`},
				{`{‘a’:‘b’}`, `{"a":"b"}`},
				{"{`a´:`b´}", `{"a":"b"}`},
			},
		},
		{
			name: "should not replace special quotes inside a normal string",
			cases: []Case{
				{`"Rounded “ quote"`, `"Rounded “ quote"`},
				{`'Rounded “ quote'`, `"Rounded “ quote"`},
				{`"Rounded ’ quote"`, `"Rounded ’ quote"`},
				{`'Rounded ’ quote'`, `"Rounded ’ quote"`},
				{`'Double " quote'`, `"Double \" quote"`},
			},
		},
		{
			name: "should not crash when repairing quotes",
			cases: []Case{
				{`{pattern: '’'}`, `{"pattern": "’"}`},
			},
		},
		{
			name: "should leave string content untouched",
			cases: []Case{
				{`"{a:b}"`, `"{a:b}"`},
			},
		},

		{
			name: "should add/remove escape characters",
			cases: []Case{
				{`"foo'bar"`, `"foo'bar"`},
				{`"foo\"bar"`, `"foo\"bar"`},
				{"'foo\"bar'", `"foo\"bar"`},
				{`'foo\'bar'`, `"foo'bar"`},
				{`"foo\'bar"`, `"foo'bar"`},
				{`"\a"`, `"a"`},
			},
		},
		{
			name: `should repair a missing object value`,
			cases: []Case{
				{`{"a":}`, `{"a":null}`},
				{`{"a":,"b":2}`, `{"a":null,"b":2}`},
				{`{"a":`, `{"a":null}`},
			},
		},
		{
			name: `should repair undefined values`,
			cases: []Case{
				{`{"a":undefined}`, `{"a":null}`},
				{`[undefined]`, `[null]`},
				{`undefined`, `null`},
			},
		},
		{
			name: `should escape unescaped control characters`,
			cases: []Case{
				{"\"hello\bworld\"", `"hello\bworld"`},
				{"\"hello\fworld\"", `"hello\fworld"`},
				{"\"hello\nworld\"", `"hello\nworld"`},
				{"\"hello\rworld\"", `"hello\rworld"`},
				{"\"hello\tworld\"", `"hello\tworld"`},
				{"{\"key\nafter\": \"foo\"}", "{\"key\\nafter\": \"foo\"}"},
				{"{\"value\n\": \"dc=hcm,dc=com\"}", `{"value\n": "dc=hcm,dc=com"}`},
				{"[\"hello\nworld\"]", `["hello\nworld"]`},
				{"[\"hello\nworld\"   ]", `["hello\nworld"   ]`},
				{"[\"hello\nworld\"\n]", "[\"hello\\nworld\"\n]"},
			},
		},
		{
			name: "should replace special white space characters",
			cases: []Case{
				{"{\"a\":\u00a0\"foo\u00a0bar\"}", "{\"a\": \"foo\u00a0bar\"}"},
				{"{\"a\":\u202F\"foo\"}", "{\"a\": \"foo\"}"},
				{"{\"a\":\u205F\"foo\"}", "{\"a\": \"foo\"}"},
				{"{\"a\":\u3000\"foo\"}", "{\"a\": \"foo\"}"},
			},
		},
		{
			name: "should replace non normalized left/right quotes",
			cases: []Case{
				{"\u2018foo\u2019", `"foo"`},
				{"\u201Cfoo\u201D", `"foo"`},
				{"\u0060foo\u00B4", `"foo"`},

				// mix single quotes
				{"\u0060foo'", `"foo"`},

				{"\u0060foo'", `"foo"`},
			},
		},
		{
			name: "should remove block comments",
			cases: []Case{
				{`/* foo */ {}`, ` {}`},
				{`{} /* foo */ `, `{}  `},
				{`{} /* foo `, `{} `},
				{"\n/* foo */\n{}", "\n\n{}"},
				{"{\"a\":\"foo\",/*hello*/\"b\":\"bar\"}", `{"a":"foo","b":"bar"}`},
			},
		},
		{
			name: "should remove line comments",
			cases: []Case{
				{`{} // comment`, `{} `},
				{"{\n\"a\":\"foo\",//hello\n\"b\":\"bar\"\n}", "{\n\"a\":\"foo\",\n\"b\":\"bar\"\n}"},
			},
		},
		{
			name: "should not remove comments inside a string",
			cases: []Case{
				{`"/* foo */"`, `"/* foo */"`},
			},
		},
		{
			name: "should strip JSONP notation",
			cases: []Case{
				// matching
				{`callback_123({});`, `{}`},
				{`callback_123([]);`, `[]`},
				{`callback_123(2);`, `2`},
				{`callback_123("foo");`, `"foo"`},
				{`callback_123(null);`, `null`},
				{`callback_123(true);`, `true`},
				{`callback_123(false);`, `false`},
				{`callback({}`, `{}`},
				{`/* foo bar */ callback_123 ({}`, ` {}`},
				{`/* foo bar */ callback_123 ({}`, ` {}`},
				{"/* foo bar */\ncallback_123({}", "\n{}"},
				{`/* foo bar */ callback_123 (  {}  )`, `   {}  `},
				{`  /* foo bar */   callback_123({});  `, `     {}  `},
				{"\n/* foo\nbar */\ncallback_123 ({});\n\n", "\n\n{}\n\n"},
			},
		},
		{
			name: "should repair escaped string contents",
			cases: []Case{
				{`\"hello world\"`, `"hello world"`},
				{`\"hello world\`, `"hello world"`},
				{`\"hello \\"world\\"\"`, `"hello \"world\""`},
				{`[\"hello \\"world\\"\"]`, `["hello \"world\""]`},
				{`{\"stringified\": \"hello \\"world\\"\"}`, `{"stringified": "hello \"world\""}`},
				// the following is sort of invalid: the end quote should be escaped too,
				// but the fixed result is most likely what you want in the end
				{`\"hello"`, `"hello"`},
			},
		},
		{
			name: "should strip trailing commas from an array",
			cases: []Case{
				{"[1,2,3,]", "[1,2,3]"},
				{"[1,2,3,\n]", "[1,2,3\n]"},
				{"[1,2,3,  \n  ]", "[1,2,3  \n  ]"},
				{"[1,2,3,/*foo*/]", `[1,2,3]`},
				{"{\"array\":[1,2,3,]}", `{"array":[1,2,3]}`},

				// not matching: inside a string
				{`"[1,2,3,]"`, `"[1,2,3,]"`},
			},
		},
		{
			name: "should strip trailing commas from an object",
			cases: []Case{
				{`{"a":2,}`, `{"a":2}`},
				{`{"a":2  ,  }`, `{"a":2    }`},
				{"{\"a\":2  , \n }", "{\"a\":2   \n }"},
				{`{"a":2/*foo*/,/*foo*/}`, `{"a":2}`},

				// not matching: inside a string
				{`"{a:2,}"`, `"{a:2,}"`},
			},
		},
		{
			name: "should strip trailing comma at the end",
			cases: []Case{
				{`4,`, `4`},
				{`4 ,`, `4 `},
				{`4 , `, `4  `},
				{`{"a":2},`, `{"a":2}`},
				{`[1,2,3],`, `[1,2,3]`},
			},
		},

		{
			name: "should add a missing closing bracket for an object",
			cases: []Case{
				{`{`, `{}`},
				{`{"a":2`, `{"a":2}`},
				{`{"a":2,`, `{"a":2}`},
				{`{"a":{"b":2}`, `{"a":{"b":2}}`},
				{"{\n  \"a\":{\"b\":2\n}", "{\n  \"a\":{\"b\":2\n}}"},
				{`[{"b":2]`, `[{"b":2}]`},
				{"[{\"b\":2\n]", "[{\"b\":2}\n]"},
				{`[{"i":1{"i":2}]`, `[{"i":1},{"i":2}]`},
				{`[{"i":1,{"i":2}]`, `[{"i":1},{"i":2}]`},
			},
		},

		{
			name: "should add a missing closing bracket for an array",
			cases: []Case{
				{`[`, `[]`},
				{`[1,2,3`, `[1,2,3]`},
				{`[1,2,3,`, `[1,2,3]`},
				{`[[1,2,3,`, `[[1,2,3]]`},
				{"{\n\"values\":[1,2,3\n}", "{\n\"values\":[1,2,3]\n}"},
				{"{\n\"values\":[1,2,3\n", "{\n\"values\":[1,2,3]}\n"},
			},
		},
		{
			name: "should strip MongoDB data types",
			cases: []Case{
				{`NumberLong("2")`, `"2"`},
				{`{"_id":ObjectId("123")}`, `{"_id":"123"}`},
				{
					"{\n" +
						"   \"_id\" : ObjectId(\"123\"),\n" +
						"   \"isoDate\" : ISODate(\"2012-12-19T06:01:17.171Z\"),\n" +
						"   \"regularNumber\" : 67,\n" +
						"   \"long\" : NumberLong(\"2\"),\n" +
						"   \"long2\" : NumberLong(2),\n" +
						"   \"int\" : NumberInt(\"3\"),\n" +
						"   \"int2\" : NumberInt(3),\n" +
						"   \"decimal\" : NumberDecimal(\"4\"),\n" +
						"   \"decimal2\" : NumberDecimal(4)\n" +
						"}",
					"{\n" +
						"   \"_id\" : \"123\",\n" +
						"   \"isoDate\" : \"2012-12-19T06:01:17.171Z\",\n" +
						"   \"regularNumber\" : 67,\n" +
						"   \"long\" : \"2\",\n" +
						"   \"long2\" : 2,\n" +
						"   \"int\" : \"3\",\n" +
						"   \"int2\" : 3,\n" +
						"   \"decimal\" : \"4\",\n" +
						"   \"decimal2\" : 4\n" +
						"}",
				},
			},
		},
		{
			name: "should replace Python constants None, True, False",
			cases: []Case{
				{"True", "true"},
				{"False", "false"},
				{"None", "null"},
			},
		},
		{
			name: "should turn unknown symbols into a string",
			cases: []Case{
				{"foo", "\"foo\""},
				{"[1,foo,4]", "[1,\"foo\",4]"},
				{"{foo: bar}", "{\"foo\": \"bar\"}"},
				{"foo 2 bar", "\"foo 2 bar\""},
				{"{greeting: hello world}", "{\"greeting\": \"hello world\"}"},
				{"{greeting: hello world\nnext: \"line\"}", "{\"greeting\": \"hello world\",\n\"next\": \"line\"}"},
				{"{greeting: hello world!}", "{\"greeting\": \"hello world!\"}"},
			},
		},
		{
			name: "should concatenate strings",
			cases: []Case{
				{"\"hello\" + \" world\"", "\"hello world\""},
				{"\"hello\" +\n \" world\"", "\"hello world\""},
				{"\"a\"+\"b\"+\"c\"", "\"abc\""},
				{"\"hello\" + /*comment*/ \" world\"", "\"hello world\""},
				{"{\n  \"greeting\": 'hello' +\n 'world'\n}",
					"{\n  \"greeting\": \"helloworld\"\n}"},
			},
		},
		{
			name: "should repair missing comma between array items",
			cases: []Case{
				{"{\"array\": [{}{}]}", "{\"array\": [{},{}]}"},
				{"{\"array\": [{} {}]}", "{\"array\": [{}, {}]}"},
				{"{\"array\": [{}\n{}]}", "{\"array\": [{},\n{}]}"},
				{"{\"array\": [\n{}\n{}\n]}", "{\"array\": [\n{},\n{}\n]}"},
				{"{\"array\": [\n1\n2\n]}", "{\"array\": [\n1,\n2\n]}"},
				{"{\"array\": [\n\"a\"\n\"b\"\n]}", "{\"array\": [\n\"a\",\n\"b\"\n]}"},
				// should leave normal array as is
				{"[\n{},\n{}\n]", "[\n{},\n{}\n]"},
			},
		},
		{
			name: "should repair missing comma between object properties",
			cases: []Case{
				// {"{\"a\":2\n\"b\":3\n}", "{\"a\":2,\n\"b\":3\n}",
				{"{\"a\":2\n\"b\":3\nc:4}", "{\"a\":2,\n\"b\":3,\n\"c\":4}"},
			},
		},
		{
			name: "should repair numbers at the end",
			cases: []Case{
				{"{\"a\":2.", "{\"a\":2.0}"},
				{"{\"a\":2e", "{\"a\":2e0}"},
				{"{\"a\":2e-", "{\"a\":2e-0}"},
				{"{\"a\":-", "{\"a\":-0}"},
			},
		},
		{
			name: "should repair missing colon between object key and value",
			cases: []Case{
				{"{\"a\" \"b\"}", "{\"a\": \"b\"}"},
				{"{\"a\" 2}", "{\"a\": 2}"},
				{"{\n\"a\" \"b\"\n}", "{\n\"a\": \"b\"\n}"},
				{`{"a" 'b'}`, `{"a": "b"}`},
				{"{'a' 'b'}", `{"a": "b"}`},
				{`{“a” “b”}`, `{"a": "b"}`},
				{`{a 'b'}`, `{"a": "b"}`},
				{`{a “b”}`, `{"a": "b"}`},
			},
		},
		{
			name: "should repair missing a combination of comma, quotes and brackets",
			cases: []Case{
				{"{\"array\": [\na\nb\n]}", "{\"array\": [\n\"a\",\n\"b\"\n]}"},
				{"1\n2", "[\n1,\n2\n]"},
				{"[a,b\nc]", "[\"a\",\"b\",\n\"c\"]"},
			},
		},
		{
			name: "should repair newline separated json (for example from MongoDB)",
			cases: []Case{
				{
					"" + "/* 1 */\n" + "{}\n" + "\n" + "/* 2 */\n" + "{}\n" + "\n" + "/* 3 */\n" + "{}\n",
					"[\n\n{},\n\n\n{},\n\n\n{}\n\n]",
				},
			},
		},
		{
			name: "should repair newline separated json having commas",
			cases: []Case{
				{
					"" + "/* 1 */\n" + "{},\n" + "\n" + "/* 2 */\n" + "{},\n" + "\n" + "/* 3 */\n" + "{}\n",
					"[\n\n{},\n\n\n{},\n\n\n{}\n\n]",
				},
			},
		},
		{
			name: "should repair newline separated json having commas and trailing comma",
			cases: []Case{
				{
					"" + "/* 1 */\n" + "{},\n" + "\n" + "/* 2 */\n" + "{},\n" + "\n" + "/* 3 */\n" + "{},\n",
					"[\n\n{},\n\n\n{},\n\n\n{}\n\n]",
				},
			},
		},
		{
			name: "should repair a comma separated list with value",
			cases: []Case{
				{`1,2,3`, "[\n1,2,3\n]"},
				{`1,2,3,`, "[\n1,2,3\n]"},
				{"1\n2\n3", "[\n1,\n2,\n3\n]"},
				{"a\nb", "[\n\"a\",\n\"b\"\n]"},
				{`a,b`, "[\n\"a\",\"b\"\n]"},
			},
		},
		{
			name: "should repair a number with leading zero",
			cases: []Case{
				{`0789`, `"0789"`},
				{`000789`, `"000789"`},
				{`001.2`, `"001.2"`},
				{`002e3`, `"002e3"`},
				{`[0789]`, `["0789"]`},
				{`{value:0789}`, `{"value":"0789"}`},
			},
		},
	}

	for _, tt := range ts {
		hasTestErr := false
		for _, c := range tt.cases {
			parsed, err := JSONRepair(c.Input)
			if parsed != c.Want {
				hasTestErr = true
				t.Errorf("failed on group: %s, case: %s, got: %s, expect: %s", tt.name, c.Input, parsed, c.Want)
			}
			if err != nil {
				hasTestErr = true
				t.Errorf("failed on group: %s, case: %s, err: %v", tt.name, c.Input, err)
			}
		}
		if !hasTestErr {
			t.Logf("cases passed for group: %s\n", tt.name)
		}
	}
}

func TestNonRepairable(t *testing.T) {
	ts := []struct {
		Input  string
		ErrStr string
	}{
		{
			Input:  "",
			ErrStr: "Unexpected end of json string at position 0",
		},
		{
			Input:  `{"a",`,
			ErrStr: `Colon expected at position 4`,
		},
		{
			Input:  `{:2}`,
			ErrStr: `Object key expected at position 1`,
		},
		{
			Input:  `{"a":2}{}`,
			ErrStr: `Unexpected character "{" at position 7`,
		},
		{
			Input:  `{"a" ]`,
			ErrStr: `Colon expected at position 5`,
		},
		{
			Input:  `{"a":2}foo`,
			ErrStr: `Unexpected character "f" at position 7`,
		},
		{
			Input:  `2.3.4`,
			ErrStr: `Unexpected character "." at position 3`,
		},
		{
			Input:  `234..5`,
			ErrStr: "Invalid number '234.', expecting a digit but got '.' at position 4",
		},
		{
			Input:  `2e3.4`,
			ErrStr: `Unexpected character "." at position 3`,
		},
		{
			Input:  `[2e,`,
			ErrStr: "Invalid number '2e', expecting a digit but got ',' at position 3", // TODO: position 2?: https://github.com/josdejong/jsonrepair/issues/98
		},
		{
			Input:  `[-,`,
			ErrStr: "Invalid number '-', expecting a digit but got ',' at position 2",
		},
		{
			Input:  `foo [`,
			ErrStr: `Unexpected character "[" at position 4`,
		},
		{
			Input:  `"\u26"`,
			ErrStr: `Invalid unicode character "\u26"" at position 1`, // TODO "\u26" instead of "\u26""
		},
		{
			Input:  `"\uZ000"`,
			ErrStr: `Invalid unicode character "\uZ000" at position 1`,
		},
	}

	hasTestErr := false
	for _, tt := range ts {
		repaired, err := JSONRepair(tt.Input)
		if err == nil {
			hasTestErr = true
			t.Errorf("an error is expected, but got nil for input %s, got repaird: %s", tt.Input, repaired)
		} else if err.Error() != tt.ErrStr {
			hasTestErr = true
			t.Errorf("case: %s, error is: [%s], expect: [%s]", tt.Input, err, tt.ErrStr)
		}
	}
	if !hasTestErr {
		t.Log("cases passed for group: should throw an exception in case of non-repairable issues")
	}
}
