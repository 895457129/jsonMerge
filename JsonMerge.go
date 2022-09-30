package jsonMerge

import (
	"fmt"
	"math"
	"reflect"
)

type ChangeItem struct {
	Key  string `json:"key"`
	Desc string `json:"desc"`
}

type CompareResult int

type CompareReason int

func (c CompareResult) String() string {
	switch c {
	case COMPARE_RESULT_DELETE:
		return "删除"
	case COMPARE_RESULT_ADD:
		return "新增"
	default:
		return "N/A"
	}
}

func (c CompareReason) String() string {
	switch c {
	case COMPARE_REASON_MISS_SRC:
		return "Src不存在"
	case COMPARE_REASON_MISS_DST:
		return "Dst不存在"
	case COMPARE_REASON_INVAILD_SRC:
		return "Src不是可转换类型"
	case COMPARE_REASON_INVAILD_DST:
		return "Dst不是可转换类型"
	case COMPARE_REASON_TYPE_DIFF:
		return "类型不一致"
	default:
		return "N/A"
	}
}

const (
	COMPARE_REASON_MISS_SRC CompareReason = iota
	COMPARE_REASON_MISS_DST
	COMPARE_REASON_INVAILD_SRC
	COMPARE_REASON_INVAILD_DST
	COMPARE_REASON_TYPE_DIFF
)

const (
	COMPARE_RESULT_ADD CompareResult = iota
	COMPARE_RESULT_DELETE
)

type Compare func(dst interface{}, src interface{}, path string, reason CompareReason) (CompareResult, interface{})

func DefaultCompare(dst interface{}, src interface{}, path string, reason CompareReason) (CompareResult, interface{}) {
	typeSrc := reflect.TypeOf(src)
	srcIsEmptyMap := typeSrc.Kind() == reflect.Map && len(reflect.ValueOf(src).MapKeys()) == 0
	if srcIsEmptyMap {
		return COMPARE_RESULT_DELETE, nil
	}
	switch reason {
	case COMPARE_REASON_MISS_DST:
		return COMPARE_RESULT_ADD, src
	case COMPARE_REASON_MISS_SRC:
		return COMPARE_RESULT_DELETE, nil
	case COMPARE_REASON_TYPE_DIFF:
		return COMPARE_RESULT_ADD, src
	case COMPARE_REASON_INVAILD_SRC:
		return COMPARE_RESULT_ADD, src
	default:
		return COMPARE_RESULT_ADD, src
	}
}

type jsonMergeInfo struct {
	ChangeItems []ChangeItem
	Compare
}

func (mr *jsonMergeInfo) addChangeInfo(key string, desc string) {
	mr.ChangeItems = append(mr.ChangeItems, ChangeItem{
		Key:  key,
		Desc: desc,
	})
}

func (mr *jsonMergeInfo) merge(sourceDst interface{}, sourceSrc interface{}, path string) interface{} {
	srcType := reflect.TypeOf(sourceSrc)
	dstType := reflect.TypeOf(sourceDst)

	srcTypeKind := srcType.Kind()
	dstTypeKind := dstType.Kind()

	dstTypeKindIsMap := dstTypeKind == reflect.Map
	dstTypeKindIsSlice := dstTypeKind == reflect.Slice
	dstTypeKindIsArray := dstTypeKind == reflect.Array
	srcTypeKindIsMap := srcTypeKind == reflect.Map
	srcTypeKindIsSlice := srcTypeKind == reflect.Slice
	srcTypeKindIsArray := srcTypeKind == reflect.Array

	if !srcTypeKindIsMap && !srcTypeKindIsSlice && !srcTypeKindIsArray {
		mr.addChangeInfo(path, fmt.Sprintf("%s-->src: %+v", COMPARE_REASON_INVAILD_SRC, sourceSrc))
		action, val := mr.Compare(sourceDst, sourceSrc, path, COMPARE_REASON_INVAILD_SRC)
		if action == COMPARE_RESULT_DELETE {
			return nil
		}
		if action == COMPARE_RESULT_ADD {
			return val
		}
	}

	if srcTypeKindIsMap && !dstTypeKindIsMap {
		mr.addChangeInfo(path, fmt.Sprintf("%s-->src: %+v, dst: %+v", COMPARE_REASON_TYPE_DIFF, sourceSrc, sourceDst))
		action, val := mr.Compare(sourceDst, sourceSrc, path, COMPARE_REASON_TYPE_DIFF)
		if action == COMPARE_RESULT_DELETE {
			return nil
		}
		if action == COMPARE_RESULT_ADD {
			return val
		}
	}

	if (srcTypeKindIsSlice || srcTypeKindIsArray) && !(dstTypeKindIsSlice || dstTypeKindIsArray) {
		mr.addChangeInfo(path, fmt.Sprintf("%s-->src: %+v, dst: %+v", COMPARE_REASON_TYPE_DIFF, sourceSrc, sourceDst))
		action, val := mr.Compare(sourceDst, sourceSrc, path, COMPARE_REASON_TYPE_DIFF)
		if action == COMPARE_RESULT_DELETE {
			return nil
		}
		if action == COMPARE_RESULT_ADD {
			return val
		}
	}

	if srcTypeKindIsMap {
		return mr.mergeMap(sourceDst.(map[string]interface{}), sourceSrc.(map[string]interface{}), path)
	} else {
		newSourceDst := make([]interface{}, 0)
		newSourceSrc := make([]interface{}, 0)
		newSourceDstLen := reflect.ValueOf(sourceDst).Len()
		newSourceSrcLen := reflect.ValueOf(sourceSrc).Len()
		for i := 0; i < newSourceDstLen; i += 1 {
			newSourceDst = append(newSourceDst, reflect.ValueOf(sourceDst).Index(i).Interface())
		}
		for i := 0; i < newSourceSrcLen; i += 1 {
			newSourceSrc = append(newSourceSrc, reflect.ValueOf(sourceSrc).Index(i).Interface())
		}
		return mr.mergeSlice(newSourceDst, newSourceSrc, path)
	}
}

func (mr *jsonMergeInfo) mergeMap(dst map[string]interface{}, src map[string]interface{}, path string) interface{} {
	for key, srcVal := range src {
		nextKey := fmt.Sprintf("%s.%s", path, key)
		if dstVal, ok := dst[key]; !ok { // 原来的里面不存在， 直接添加
			compare, val := mr.Compare(nil, srcVal, nextKey, COMPARE_REASON_MISS_DST)
			mr.addChangeInfo(nextKey, fmt.Sprintf("%s-->src: %+v, dst: %+v", COMPARE_REASON_MISS_DST, srcVal, nil))
			if compare == COMPARE_RESULT_DELETE {
				continue
			}
			if compare == COMPARE_RESULT_ADD {
				dst[key] = val
			}
		} else {
			val := mr.merge(dstVal, srcVal, nextKey)
			if val == nil {
				delete(dst, key)
			} else {
				dst[key] = val
			}
		}
	}
	return dst
}

func (mr *jsonMergeInfo) mergeSlice(dst []interface{}, src []interface{}, path string) interface{} {
	dstLen := len(dst)
	srcLen := len(src)
	maxLen := int(math.Max(float64(dstLen), float64(srcLen)))
	result := make([]interface{}, 0)
	for i := 0; i < maxLen; i += 1 {
		nextKey := fmt.Sprintf("%s.%d", path, i)
		hasDst := i < dstLen
		hasSrc := i < srcLen
		if hasDst && !hasSrc {
			compare, val := mr.Compare(dst[i], nil, nextKey, COMPARE_REASON_MISS_SRC)
			if compare == COMPARE_RESULT_DELETE {
				mr.addChangeInfo(nextKey, fmt.Sprintf("%s:%+v", compare, dst[i]))
				continue
			}
			if compare == COMPARE_RESULT_ADD {
				mr.addChangeInfo(nextKey, fmt.Sprintf("%s:%+v", compare, val))
				result = append(result, val)
				continue
			}
		} else if !hasDst && hasSrc {
			compare, val := mr.Compare(nil, src[i], nextKey, COMPARE_REASON_MISS_DST)
			if compare == COMPARE_RESULT_DELETE {
				mr.addChangeInfo(nextKey, fmt.Sprintf("%s:%+v", compare, src[i]))
				continue
			}
			if compare == COMPARE_RESULT_ADD {
				mr.addChangeInfo(nextKey, fmt.Sprintf("%s:%+v", compare, val))
				result = append(result, val)
				continue
			}
		} else {
			val := mr.merge(dst[i], src[i], nextKey)
			if val == nil {
				continue
			} else {
				result = append(result, val)
			}
		}
	}
	return result
}

func JsonMerge(dst interface{}, src interface{}, compareFun ...Compare) (interface{}, []ChangeItem) {
	var compare Compare
	if compareFun == nil || len(compareFun) == 0 {
		compare = DefaultCompare
	} else {
		compare = compareFun[0]
	}
	merger := jsonMergeInfo{
		ChangeItems: nil,
		Compare:     compare,
	}
	return merger.merge(dst, src, "root"), merger.ChangeItems
}
