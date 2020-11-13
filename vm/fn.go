package script

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	b64 "encoding/base64"

	xj "github.com/basgys/goxml2json"
	"github.com/dghubble/oauth1"
	"github.com/robertkrimen/otto"
	"golang.org/x/oauth2"
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

	httpGetOAuth1 := func(call otto.FunctionCall) otto.Value {
		url := call.Argument(0).String()
		key := call.Argument(1).String()
		keysecret := call.Argument(2).String()
		token := call.Argument(3).String()
		tokensecret := call.Argument(4).String()

		c := oauth1.NewConfig(key, keysecret)
		_token := oauth1.NewToken(token, tokensecret)
		client := c.Client(oauth1.NoContext, _token)

		r, err := client.Get(url)
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
	setup(vm.Set("httpGetOAuth1", httpGetOAuth1))

	httpGetOAuth2 := func(call otto.FunctionCall) otto.Value {
		url := call.Argument(0).String()
		token := call.Argument(1).String()

		ctx := context.Background()
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: token},
		)
		client := oauth2.NewClient(ctx, ts)

		r, err := client.Get(url)
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
	setup(vm.Set("httpGetOAuth2", httpGetOAuth2))

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

	encode := func(call otto.FunctionCall) otto.Value {
		str := call.Argument(0).String()
		enc := b64.StdEncoding.EncodeToString([]byte(str))
		v, _ := vm.ToValue(string(enc))
		return v
	}
	setup(vm.Set("btoa", encode))

	decode := func(call otto.FunctionCall) otto.Value {
		str := call.Argument(0).String()
		dec, _ := b64.URLEncoding.DecodeString(str)
		v, _ := vm.ToValue(string(dec))
		return v
	}
	setup(vm.Set("atob", decode))

	consoleLog := func(call otto.FunctionCall) otto.Value {
		str := call.Argument(0).String()
		fmt.Printf(fmt.Sprintf("%s\n", str))
		return otto.NullValue()
	}
	setup(vm.Set("log", consoleLog))

	return vm
}
