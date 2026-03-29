package main

import "github.com/marcbran/jsonnet-plugin-kubectl/kubectl"

func main() {
	kubectl.Plugin().Serve()
}
