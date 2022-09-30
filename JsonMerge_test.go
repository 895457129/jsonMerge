package jsonMerge

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
)

func TestJsonMerge(t *testing.T) {
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
	result, info := JsonMerge(dst, src)
	str, _ := json.Marshal(result)
	fmt.Println(result, reflect.TypeOf(result))
	fmt.Println(fmt.Sprintf("dst: %+v", dst))
	fmt.Println(fmt.Sprintf("src: %+v", src))
	fmt.Println(fmt.Sprintf("结果: %+v",string(str)))
	for _, item := range info{
		fmt.Println(item)
	}
}