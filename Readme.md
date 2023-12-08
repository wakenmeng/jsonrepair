# JSON Repair for Go

This is a Go implementation forked from TypeScript library [jsonrepair](https://github.com/josdejong/jsonrepair) by [@josdejong](https://github.com/josdejong).

## Version

Current version [v3.5.0](https://github.com/josdejong/jsonrepair/tree/v3.5.0), has NO streaming support yet.

## Usage

```
s := `{"a": "b" "c": "d"}`
repaired, err := jsonrepair.JSONRepair(s)
```
