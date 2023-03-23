package abicoder

import (
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi"
)

func (d *ABIDecoder) getArguments(name string, data []byte) (Arguments, error) {
	// since there can't be naming collisions with contracts and events,
	// we need to decide whether we're calling a method or an event
	args := make([]Argument, 0)
	if method, ok := d.myabi.Methods[name]; ok {
		if len(data)%32 != 0 {
			return nil, fmt.Errorf("abicoder: improperly formatted output: %s - Bytes: [%+v]", string(data), data)
		}
		for _, input := range method.Inputs {
			newInput := Argument{Name: input.Name, Type: input.Type, Indexed: input.Indexed}
			args = append(args, newInput)
		}
	}
	if args == nil {
		return nil, fmt.Errorf("abicoder: could not locate named method or event: %s", name)
	}
	return args, nil
}

//
// Unpack unpacks the output according to the abicoder specification.
func (d *ABIDecoder) Unpack(name string, data []byte) ([]interface{}, error) {
	args, err := d.getArguments(name, data)
	if err != nil {
		return nil, err
	}
	return args.Unpack(data)
}

func (d *ABIDecoder) UnpackIntoInterface(v interface{}, name string, data []byte) error {
	args, err := d.getArguments(name, data)
	if err != nil {
		return err
	}
	unpacked, err := args.Unpack(data)
	if err != nil {
		return err
	}
	return args.Copy(v, unpacked)
}

//
// UnpackIntoMap unpacks a log into the provided map[string]interface{}.
func (d *ABIDecoder) UnpackIntoMap(v map[string]interface{}, name string, data []byte) (err error) {
	args, err := d.getArguments(name, data)
	if err != nil {
		return err
	}
	return args.UnpackIntoMap(v, data)
}

// ABIDecoder ethereum transaction data decoder
type ABIDecoder struct {
	myabi abi.ABI
}

func NewABIDecoder(abi abi.ABI) *ABIDecoder {
	return &ABIDecoder{myabi: abi}
}
