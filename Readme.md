# JSON Repair for Go

This is a Go implementation forked from TypeScript library [jsonrepair](https://github.com/josdejong/jsonrepair) by [@josdejong](https://github.com/josdejong).

## Usage

```
s := `{"a": "b" "c": "d"}`
repaired, err := jsonrepair.JSONRepair(s)
```
