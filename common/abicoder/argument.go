// Copyright 2015 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package abicoder

import (
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"reflect"
)

// Argument holds the name of the argument and the corresponding type.
// Types are used when packing and testing arguments.
type Argument struct {
	Name    string
	Type    abi.Type
	Indexed bool // indexed is only used by events
}

type Arguments []Argument

type ArgumentMarshaling struct {
	Name         string
	Type         string
	InternalType string
	Components   []ArgumentMarshaling
	Indexed      bool
}

// NonIndexed returns the arguments with indexed arguments filtered out.
func (arguments Arguments) NonIndexed() Arguments {
	var ret []Argument
	for _, arg := range arguments {
		if !arg.Indexed {
			ret = append(ret, arg)
		}
	}
	return ret
}

// isTuple returns true for non-atomic constructs, like (uint,uint) or uint[].
func (arguments Arguments) isTuple() bool {
	return len(arguments) > 1
}

// Unpack performs the operation hexdata -> Go format.
func (arguments Arguments) Unpack(data []byte) ([]interface{}, error) {
	if len(data) == 0 {
		if len(arguments.NonIndexed()) != 0 {
			return nil, errors.New("abi: attempting to unmarshall an empty string while arguments are expected")
		}
		return make([]interface{}, 0), nil
	}
	return arguments.UnpackValues(data)
}

// UnpackIntoMap performs the operation hexdata -> mapping of argument name to argument value.
func (arguments Arguments) UnpackIntoMap(v map[string]interface{}, data []byte) error {
	// Make sure map is not nil
	if v == nil {
		return errors.New("abi: cannot unpack into a nil map")
	}
	if len(data) == 0 {
		if len(arguments.NonIndexed()) != 0 {
			return errors.New("abi: attempting to unmarshall an empty string while arguments are expected")
		}
		return nil // Nothing to unmarshal, return
	}
	marshalledValues, err := arguments.UnpackValues(data)
	if err != nil {
		return err
	}
	for i, arg := range arguments.NonIndexed() {
		v[arg.Name] = marshalledValues[i]
	}
	return nil
}

// Copy performs the operation go format -> provided struct.
func (arguments Arguments) Copy(v interface{}, values []interface{}) error {
	// make sure the passed value is arguments pointer
	if reflect.Ptr != reflect.ValueOf(v).Kind() {
		return fmt.Errorf("abi: Unpack(non-pointer %T)", v)
	}
	if len(values) == 0 {
		if len(arguments.NonIndexed()) != 0 {
			return errors.New("abi: attempting to copy no values while arguments are expected")
		}
		return nil // Nothing to copy, return
	}
	if arguments.isTuple() {
		return arguments.copyTuple(v, values)
	}
	return arguments.copyAtomic(v, values[0])
}

// unpackAtomic unpacks ( hexdata -> go ) a single value
func (arguments Arguments) copyAtomic(v interface{}, marshalledValues interface{}) error {
	dst := reflect.ValueOf(v).Elem()
	src := reflect.ValueOf(marshalledValues)

	if dst.Kind() == reflect.Struct {
		return set(dst.Field(0), src)
	}
	return set(dst, src)
}

// copyTuple copies a batch of values from marshalledValues to v.
//func (arguments Arguments) copyTuple(v interface{}, marshalledValues []interface{}) error {
//	value := reflect.ValueOf(v).Elem()
//	nonIndexedArgs := arguments.NonIndexed()
//
//	switch value.Kind() {
//	case reflect.Struct:
//		argNames := make([]string, len(nonIndexedArgs))
//		for i, arg := range nonIndexedArgs {
//			argNames[i] = arg.Name
//		}
//		var err error
//		abi2struct, err := mapArgNamesToStructFields(argNames, value)
//		if err != nil {
//			return err
//		}
//		for i, arg := range nonIndexedArgs {
//			var field reflect.Value
//			if arg.Name == "_lastCommittedBlockData" {
//				field = value.FieldByName("LastCommittedBlockData")
//			} else {
//				field = value.FieldByName("NewBlocksData")
//			}
//			if !field.IsValid() {
//				return fmt.Errorf("abi: field %s can't be found in the given value", arg.Name)
//			}
//			marshalledValue := reflect.ValueOf(marshalledValues[i])
//			if err := set(field, marshalledValue); err != nil {
//				return err
//			}
//		}
//	case reflect.Slice, reflect.Array:
//		if value.Len() < len(marshalledValues) {
//			return fmt.Errorf("abi: insufficient number of arguments for unpack, want %d, got %d", len(arguments), value.Len())
//		}
//		for i := range nonIndexedArgs {
//			if err := set(value.Index(i), reflect.ValueOf(marshalledValues[i])); err != nil {
//				return err
//			}
//		}
//	default:
//		return fmt.Errorf("abi:[2] cannot unmarshal tuple in to %v", value.Type())
//	}
//	return nil
//}

// copyTuple copies a batch of values from marshalledValues to v.
func (arguments Arguments) copyTuple(v interface{}, marshalledValues []interface{}) error {
	value := reflect.ValueOf(v).Elem()
	nonIndexedArgs := arguments.NonIndexed()

	switch value.Kind() {
	case reflect.Struct:
		argNames := make([]string, len(nonIndexedArgs))
		for i, arg := range nonIndexedArgs {
			argNames[i] = arg.Name
		}
		var err error
		abi2struct, err := mapArgNamesToStructFields(argNames, value)
		if err != nil {
			return err
		}
		for i, arg := range nonIndexedArgs {
			field := value.FieldByName(abi2struct[arg.Name])
			if !field.IsValid() {
				return fmt.Errorf("abi: field %s can't be found in the given value", arg.Name)
			}
			if err := set(field, reflect.ValueOf(marshalledValues[i])); err != nil {
				return err
			}
		}
	case reflect.Slice, reflect.Array:
		if value.Len() < len(marshalledValues) {
			return fmt.Errorf("abi: insufficient number of arguments for unpack, want %d, got %d", len(arguments), value.Len())
		}
		for i := range nonIndexedArgs {
			if err := set(value.Index(i), reflect.ValueOf(marshalledValues[i])); err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("abi:[2] cannot unmarshal tuple in to %v", value.Type())
	}
	return nil
}

// UnpackValues can be used to unpack ABI-encoded hexdata according to the ABI-specification,
// without supplying a struct to unpack into. Instead, this method returns a list containing the
// values. An atomic argument will be a list with one element.
func (arguments Arguments) UnpackValues(data []byte) ([]interface{}, error) {
	nonIndexedArgs := arguments.NonIndexed()
	retval := make([]interface{}, 0, len(nonIndexedArgs))
	virtualArgs := 0
	for index, arg := range nonIndexedArgs {
		marshalledValue, err := toGoType((index+virtualArgs)*32, arg.Type, data)
		if arg.Type.T == abi.ArrayTy && !isDynamicType(arg.Type) {
			// If we have a static array, like [3]uint256, these are coded as
			// just like uint256,uint256,uint256.
			// This means that we need to add two 'virtual' arguments when
			// we count the index from now on.
			//
			// Array values nested multiple levels deep are also encoded inline:
			// [2][3]uint256: uint256,uint256,uint256,uint256,uint256,uint256
			//
			// Calculate the full array size to get the correct offset for the next argument.
			// Decrement it by 1, as the normal index increment is still applied.
			virtualArgs += getTypeSize(arg.Type)/32 - 1
		} else if arg.Type.T == abi.TupleTy && !isDynamicType(arg.Type) {
			// If we have a static tuple, like (uint256, bool, uint256), these are
			// coded as just like uint256,bool,uint256
			virtualArgs += getTypeSize(arg.Type)/32 - 1
		}
		if err != nil {
			return nil, err
		}
		retval = append(retval, marshalledValue)
	}
	return retval, nil
}
