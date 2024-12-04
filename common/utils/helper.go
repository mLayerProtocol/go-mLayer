package utils

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"os"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/mlayerprotocol/go-mlayer/pkg/log"
)

var logger = &log.Logger
type errorStruct struct {
	EmptyFieldInStruct bool
}
func isZeroValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.String:
		return v.String() == ""
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Slice, reflect.Map, reflect.Array:
		return v.Len() == 0
	case reflect.Ptr, reflect.Interface:
		return v.IsNil()
	}
	return reflect.DeepEqual(v.Interface(), reflect.Zero(v.Type()).Interface())
}
func EnsureNotEmpty(s interface{}) any {
	empty := true
    v := reflect.ValueOf(s)
	typ := v.Type()
	if typ.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if !v.IsValid() {
		return errorStruct{EmptyFieldInStruct: true}
	}
	if v.NumField() == 0 {
		return errorStruct{EmptyFieldInStruct: true}
	}
	
    for i := 0; i < v.NumField(); i++ {
        field := v.Field(i)
		if !isZeroValue(field) {
			empty = false
		}
    }
	if empty {
		return errorStruct{EmptyFieldInStruct: true}
	}
    return s
}
func CheckEmpty(s interface{}) error {
	resp := EnsureNotEmpty(s)
	switch val := resp.(type) {
		case errorStruct:
			logger.Debugf("CheckEmpty: %v", val)
			return fmt.Errorf("is empty %s", "")
	}
	return nil
}
// get absolute difference
func Abs(a uint64, b uint64) uint64 {
	return IfThenElse(a >b, a-b,b-a)
}

func CopyStructValues(src, dst interface{}) error {
	srcVal := reflect.ValueOf(src)
	dstVal := reflect.ValueOf(dst).Elem()

	if srcVal.Kind() != reflect.Struct || dstVal.Kind() != reflect.Struct {
		return fmt.Errorf("input types must be structs")
	}

	for i := 0; i < srcVal.NumField(); i++ {
		fieldName := srcVal.Type().Field(i).Name
		srcField := srcVal.Field(i)
		dstField := dstVal.FieldByName(fieldName)

		if dstField.IsValid() && dstField.CanSet() && srcField.Type() == dstField.Type() {
			dstVal.FieldByName(fieldName).Set(srcField)
		}
	}

	return nil
}
func UpdateStruct(src, dst interface{}) {
	srcVal := reflect.ValueOf(src).Elem()
	dstVal := reflect.ValueOf(dst).Elem()

	for i := 0; i < srcVal.NumField(); i++ {
		srcField := srcVal.Field(i)
		dstField := dstVal.Field(i)

		// Skip if srcField is zero
		if IsZero(srcField) {
			continue
		}

		// Handle pointers: only set if non-nil
		if srcField.Kind() == reflect.Ptr {
			if !srcField.IsNil() {
				// If dstField is a pointer but nil, initialize it to avoid panic
				if dstField.IsNil() {
					dstField.Set(reflect.New(dstField.Type().Elem()))
				}
				dstField.Elem().Set(srcField.Elem())
			}
		} else {
			// For non-pointer fields, copy directly
			dstField.Set(srcField)
		}
	}
}

func IsZero(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Ptr:
		return v.IsNil()
	case reflect.Slice, reflect.Map, reflect.Array:
		return v.Len() == 0
	case reflect.Struct:
		// Check each field in the struct for "zero" status
		for i := 0; i < v.NumField(); i++ {
			if !IsZero(v.Field(i)) {
				return false
			}
		}
		return true
	default:
		// Check if the type is comparable before using Interface comparison
		if v.Type().Comparable() {
			return v.Interface() == reflect.Zero(v.Type()).Interface()
		}
	}
	return false
}
func StructToMap(input interface{}) map[string]interface{} {
	output := make(map[string]interface{})

	// val := reflect.ValueOf(input)
	// typ := reflect.TypeOf(input)

	// for i := 0; i < val.NumField(); i++ {
	// 	field := typ.Field(i)
	// 	value := val.Field(i).Interface()
	// 	output[field.Name] = value
	// }

	d, _ := json.Marshal(input)
	json.Unmarshal(d, &output)

	return output
}

func GetFieldValueByName(s interface{}, fieldName string) interface{} {
	// Get the reflect.Value of the struct
	v := reflect.ValueOf(s)

	// Ensure that the provided interface is a struct
	if v.Kind() == reflect.Struct {
		// Get the field by name
		field := v.FieldByName(fieldName)
		if field.IsValid() && field.CanInterface() {
			// Return the field value as an interface{}
			return field.Interface()
		}
	}

	// Return nil if the field doesn't exist or can't be accessed
	return nil
}
func SetDefaultValues(s interface{}) {
	v := reflect.ValueOf(s).Elem() // Get the actual value of the struct
	 t := v.Type()

	// Iterate over all fields in the struct
	logger.Debugf("SETTINGDEFAUTVALUE:  %v, %v", v.Addr(), t )
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		// fieldName := t.Field(i).Name
		
		// Check if the field is empty (has zero value)
		if reflect.DeepEqual(field.Interface(), reflect.Zero(field.Type()).Interface()) {
			// Assign default values based on field type
			
			switch field.Kind() {
			case reflect.String:
				field.SetString("Default___")
			case reflect.Int:
				field.SetInt(-99999)
			// Add other types as necessary (e.g., bool, float64, etc.)
			}
		}
	}
}
func To256Bits(b []byte) []byte {
	const size = 32
	padded := make([]byte, size)
	copy(padded[size-len(b):], b)
	return padded
}
func  ToUint256(num *big.Int) []byte {
	return num.FillBytes(make([]byte, 32))
}
func  Uint64ToUint256(num uint64) []byte {
	return new(big.Int).SetUint64(num).FillBytes(make([]byte, 32))
}

func Lcg(seed uint64)  (*big.Int) {
    a := uint64(1664525)
    c := uint64(1013904223)
    m := uint64(1 << 32)
	r := (a * seed + c) % m
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, r)
    return new(big.Int).SetBytes(b)
}

func BoolPtr(b bool) *bool {
    return &b
}
func TruePtr() *bool {
	var b = true
    return &b
}


func FalsePtr() *bool {
	var b = false
    return &b
}

func UuidToBytes(uuid string) ([]byte) {
	hexStr := strings.ReplaceAll(uuid, "-", "")
	if len(hexStr) == 0 {
		return []byte{}
	}
	b, _ :=  hex.DecodeString(hexStr)
	return b
}

func Contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}
func RandomString(length int) string {
	var str string;
	var err error;
	for {
		str, err = gonanoid.Generate("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890-_", length)
		if err == nil {
			break
		}
	}
	return str
}

func AddressToHex(addr string) string {
 return strings.ToLower(strings.Replace(addr, "0x", "", 1))
}
func RandomAplhaNumString(length int) string {
	var str string;
	var err error;
	for {
		str, err = gonanoid.Generate("abcdefghijklmnopqrstuvwxyz123456789", length)
		if err == nil {
			break
		}
	}
	return str
}
func RandomHexString(length int) string {
	var str string;
	var err error;
	for {
		str, err = gonanoid.Generate("abcdef1234567890", length)
		if err == nil {
			break
		}
	}
	return str
}


func TimestampMilli() uint64 {
	return uint64(time.Now().UnixNano() / 1e6)
}

func IsNumericInt(s string) bool {
    _, err := strconv.Atoi(s)
    return err == nil
}

func IsAlphaNumericDotNoNumberPrefix(str string) bool {
	pattern := `^[a-zA-Z][a-zA-Z0-9._]*$`
	matched, _ := regexp.MatchString(pattern, str)
	return matched
}

func IsAlphaNumericDot(str string) bool {
	pattern := `^[a-zA-Z0-9_.]+$`
	matched, _ := regexp.MatchString(pattern, str)
	return matched
}

func IsAlphaLowerNumericDot(str string) bool {
	pattern := `^[a-z0-9_.]+$`
	matched, _ := regexp.MatchString(pattern, str)
	return matched
}

func IsDomain(str string) bool {
	pattern := `[[A-Za-z0-9](?:[A-Za-z0-9\-]{0,61}[A-Za-z0-9])?`
	matched, _ := regexp.MatchString(pattern, str)
	return matched
}

func IfThenElse[T any](condition bool, a T, b T) T {
	if condition {
		return a
	}
	return b
}
func SafePointerValue[T any](b *T, defaultValue T) T {
	if b == nil {
		return defaultValue
	}
	return *b
}

func ToStringSlice(slice []interface{}) []string {
	stringSlice := make([]string, len(slice))

	// Convert each element from interface{} to string using type assertion
	for i, v := range slice {
		// Assert that the value is a string
		str := fmt.Sprintf("%v", v)
		stringSlice[i] = str
	}
	return stringSlice
}
func ParseQueryString(c *gin.Context) (*[]byte, error) {
	rawQuery := c.Request.URL.Query()
	var query map[string]any = map[string]any{}
	for key, v := range rawQuery {
		if len(v) > 0 {
			query[key] = v[0]
		}

	}
	b, requestErr := json.Marshal(query)
	if requestErr != nil {
		return nil, requestErr
	}
	return &b, nil
}

func ReadJSONFromFile(filePath string) (map[string]interface{}, error) {
	// file, err := os.Open(filePath)
	// if err != nil {
	// 	return nil, err
	// }
	// defer file.Close()

	// Read the entire file content
	bytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// Unmarshal the JSON into a map
	var data map[string]interface{}
	err = json.Unmarshal(bytes, &data)
	if err != nil {
		return nil, err
	}

	return data, nil
}
func WriteToCSV(filePath string, data [][]string) error {
	


	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	
    if err != nil {
       return fmt.Errorf("failed to create file: %v", err)
    }
    defer file.Close()

    // Initialize the CSV writer
    writer := csv.NewWriter(file)

    // Write all data to the CSV at once
    err = writer.WriteAll(data)
    if err != nil {
        return fmt.Errorf("failed to write data to CSV: %v", err)
    }

    // Flush any buffered data to the file
    writer.Flush()

    // Check for any errors during flushing
    if err := writer.Error(); err != nil {
        return fmt.Errorf("failed to flush data to CSV: %v", err)
    }
	return nil
}
// writeJSON writes the given map as JSON to a file.
func WriteJSONToFile(filePath string, data map[string]interface{}) error {
	// Marshal the map into JSON bytes
	
	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	// Write the JSON data back to the file
	return os.WriteFile(filePath, bytes, 0644)
}

func Find[T any](slice []T, predicate func(T) bool) (T, bool) {
	for _, v := range slice {
		if predicate(v) {
			return v, true
		}
	}
	var zero T
	return zero, false
}

func CompressToGzip(input []byte) ([]byte, error) {
	// Create a buffer to hold the compressed data
	var compressedData bytes.Buffer

	// Create a new Gzip writer using the buffer as the output
	gzipWriter := gzip.NewWriter(&compressedData)
	defer gzipWriter.Close() // Ensure the writer is closed to finish writing

	// Stream the input text into the Gzip writer
	_, err := io.Writer.Write(gzipWriter, input)
	if err != nil {
		return nil, fmt.Errorf("error writing to gzip writer: %w", err)
	}

	// Close the Gzip writer to ensure all data is flushed
	if err := gzipWriter.Close(); err != nil {
		return nil, fmt.Errorf("error closing gzip writer: %w", err)
	}

	// Return the compressed data as a byte slice
	return compressedData.Bytes(), nil
}

func DecompressGzip(compressedData []byte) ([]byte, error) {
	byteReader := bytes.NewReader(compressedData)
	gzipReader, err := gzip.NewReader(byteReader)
	if err != nil {
		return nil, err
	}
	defer gzipReader.Close()

	// Decompress the data into a buffer
	var result bytes.Buffer
	_, err = io.Copy(&result, gzipReader)
	if err != nil {
		return result.Bytes(), nil
	}
	return result.Bytes(), err
}

func MatchUrlPath(pattern, path string) (exist bool, params map[string]string) {
	params = make(map[string]string)
	patternSegments := strings.Split(pattern, "/")
	pathSegments := strings.Split(path, "/")

	// Handle different segment lengths
	if len(patternSegments) != len(pathSegments) {
		return false, nil
	}

	for i, segment := range patternSegments {
		// If the segment is a parameter (e.g., :id) or matches the path segment, continue
		if strings.HasPrefix(segment, ":") {
			params[segment[1:]] = pathSegments[i]
		}
		if strings.HasPrefix(segment, ":") || segment == pathSegments[i] {
			continue
		}

		// If it doesn't match and isn't a parameter, it's a mismatch
		return false, nil
	}
	return true, params
}

func IntMilliToTimestampString( milliseconds int64) string {
    // Convert microseconds to nanoseconds
    nanoseconds := milliseconds * 1_000_000
    timestamp := time.Unix(0, nanoseconds)
    return  timestamp.Format("20060102150405000") // No dot bef
}

func ListFilesInDir(directory string) ([]string, error) {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".") || !strings.HasSuffix(entry.Name(), ".dat") {
			continue
		}
		files = append(files, entry.Name())
	}

	// Sort alphabetically
	sort.Strings(files)
	return files, nil
}

func WriteBytesToFile(filePath string, data []byte) error {
	// Open the file in append mode, create if it doesn't exist
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.Write(data)
	return err
}