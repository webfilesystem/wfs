package script

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	xj "github.com/basgys/goxml2json"
	"github.com/robertkrimen/otto"
)

func NewVM() *otto.Otto {
	vm := otto.New()

	setup := func(err error) {
		if err != nil {
			fmt.Println("Error set vm func", err)
		}
	}

	httpGet := func(call otto.FunctionCall) otto.Value {
		url := call.Argument(0).String()
		r, err := http.Get(url)
		if err != nil {
			panic(err)
		}
		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			panic(err)
		}
		v, _ := vm.ToValue(string(data))
		return v
	}
	setup(vm.Set("httpGet", httpGet))

	xmlToJson := func(call otto.FunctionCall) otto.Value {
		input := call.Argument(0).String()

		data := ""
		if strings.Contains(string(input), "<?xml") {
			xml := strings.NewReader(string(input))
			json, err := xj.Convert(xml)
			if err != nil {
				panic("That's embarrassing...")
			}
			data = json.String()
		}
		v, _ := vm.ToValue(string(data))
		return v
	}
	setup(vm.Set("xmlToJson", xmlToJson))

	require := func(call otto.FunctionCall) otto.Value {
		file := call.Argument(0).String()
		var path string = filepath.Join(os.Getenv("HOME"), ".wfs", "lib", file)
		data, err := ioutil.ReadFile(path)
		if err != nil {
			panic(err)
		}
		v, _ := vm.ToValue(string(data))
		return v
	}
	setup(vm.Set("require", require))

	return vm
}
