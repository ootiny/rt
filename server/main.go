package main

import (
	"github.com/ootiny/rt/server/runtime"
	"github.com/ootiny/rt/server/runtime/api_system_city"
)

func init() {
	api_system_city.HookGetCityList(func(ctx *runtime.Context, country string) (api_system_city.CityList, runtime.Error) {
		return api_system_city.CityList{}, nil
	})
}

func main() {
	if err := runtime.NewHttpServer("0.0.0.0:8080", "", "", true).Run(); err != nil {
		panic(err)
	}
}
