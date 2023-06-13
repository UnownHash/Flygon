package routes

import (
	"github.com/gin-gonic/gin"
	"golang.org/x/exp/constraints"
	"net/http"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

func min[T constraints.Ordered](a, b T) T {
	if a < b {
		return a
	}
	return b
}

func getFieldPathByTagName(reflectedType reflect.Type, tag, tagValue string, fieldPath []reflect.StructField) ([]reflect.StructField, bool) {
	tagValues := strings.SplitN(tagValue, ".", 2)

	if fieldPath == nil {
		fieldPath = make([]reflect.StructField, 0)
	}

	for i := 0; i < reflectedType.NumField(); i++ {
		field := reflectedType.Field(i)
		// use split to ignore tag "options" like omitempty, etc.
		value := strings.Split(field.Tag.Get(tag), ",")[0]

		if len(tagValues[0]) == 0 || value == tagValues[0] {
			fieldPath = append(fieldPath, field)

			if len(tagValues) > 1 {
				return getFieldPathByTagName(field.Type, tag, tagValues[1], fieldPath)
			}

			return fieldPath, true
		}
	}

	return nil, false
}

func getValueByPath[S any](object S, path []reflect.StructField) reflect.Value {
	field := reflect.ValueOf(object)

	for _, structField := range path {
		field = field.FieldByIndex(structField.Index)
	}

	return field
}

func paginateAndSort[S any](context *gin.Context, collection []S) {
	collectionLength := len(collection)
	pageParam := context.DefaultQuery("page", "0")
	perPageParam := context.DefaultQuery("perPage", strconv.Itoa(collectionLength))
	sortByParam := context.DefaultQuery("sortBy", "")
	orderParam := context.DefaultQuery("order", "DESC")

	page, pageConversionError := strconv.Atoi(pageParam)
	perPage, perPageConversionError := strconv.Atoi(perPageParam)

	if pageConversionError != nil || perPageConversionError != nil || page < 0 || perPage < 0 {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Invalid pagination"})
		return
	}
	if orderParam != "DESC" && orderParam != "ASC" {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sort order parameter"})
		return
	}

	typeOfElement := reflect.TypeOf(collection).Elem()
	pathToField, hasField := getFieldPathByTagName(typeOfElement, "json", sortByParam, nil)
	if !hasField {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Cannot sort `" + typeOfElement.Name() + "` by `" + sortByParam + "`"})
		return
	}

	sort.Slice(collection, func(i, j int) bool {
		var valueA, valueB reflect.Value

		if orderParam == "ASC" {
			valueA = getValueByPath(collection[i], pathToField)
			valueB = getValueByPath(collection[j], pathToField)
		} else {
			valueA = getValueByPath(collection[j], pathToField)
			valueB = getValueByPath(collection[i], pathToField)
		}

		switch pathToField[len(pathToField)-1].Type.Name() {
		case "int":
			return int(valueA.Int()) < int(valueB.Int())
		case "string":
			return valueA.String() < valueB.String()
		case "bool":
			return valueA.Bool()
		default:
			return false
		}
	})

	firstIndex := page * perPage
	lastIndex := min((page+1)*perPage, collectionLength)

	context.JSON(http.StatusOK, gin.H{
		"data": collection[firstIndex:lastIndex],
		"pagination": gin.H{
			"total":       collectionLength,
			"hasNext":     lastIndex < collectionLength-1,
			"hasPrevious": page > 0,
		},
	})
}
