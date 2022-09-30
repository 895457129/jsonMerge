# merge json

# Useage 
```go
 type Compare func(dst interface{}, src interface{}, path string, reason CompareReason) (CompareResult, interface{})
 JsonMerge(dst interface{}, src interface{}, compareFun ...Compare) (interface{}, []ChangeItem) 
```
```go
dst := map[string]interface{}{
		"A": map[string]interface{}{
			"a0": "a0",
			"a1": 0,
			"a2": "a2",
		},
		"b": "b",
		"d": []string{
			"1",
			"2",
		},
		"e": []interface{}{
			"1",
			"2",
			map[string]interface{}{
				"id": 2,
			},
			[]string{
				"e1",
				"e2",
			},
		},
	}
	src := map[string]interface{}{
		"A": map[string]interface{}{
			"a1": "a1",
			"a2": "a2",
			"a3": "a3",
		},
		"b": "bb",
		"c": "c",
		"d": []interface{}{
			"1",
			2,
			3,
		},
		"e": []interface{}{
			"1",
			2,
			[]string{
				"e3",
				"e4",
			},
			"e5",
			"e6",
		},
		"f": map[string]interface{}{},
	}
result, changeInfo := JsonMerge(dst, src)
```
```json5
{
    "A": {
        "a0": "a0",
        "a1": "a1",
        "a2": "a2",
        "a3": "a3"
    },
    "b": "bb",
    "c": "c",
    "d": [
        "1",
        2,
        3
    ],
    "e": [
        "1",
        2,
        [
            "e3",
            "e4"
        ],
        "e5",
        "e6"
    ]
}
```

