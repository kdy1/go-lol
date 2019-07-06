# go-lol
A new, generated golang client for riot apis.

# Install
```sh
go get -u github.com/go-lol/go-lol
```

# Features
 - [x] Clean API. See [godoc][godoc]
   - [x] No global variable.
   - [x] Region. (lol.NA == lol.RegionByName("NA"))
 - [x] [net/context](https://godoc.org/golang.org/x/net/context) support.
 - [x] Google app engine support. (My usecase.)
 - [x] (Optional) Batching.
   - [ ] API to get single entity.


# FAQ
If your question is not listed here, please feel free to make an issue for it.

## Can I limit request rate?
No. As API key is per appplication instead of per server, you must use stuffs like task queue to limit it.

## Why do you generate instead of writing it by hand?
Rito api really sucks.
It's not documented, but they use multiple swagger api manifests internally, and it results in multiple classes with same name.
(I'm sure about this as swagger does not support multiple host with a manifest.)

 - No manifest provided.
 - No standard method names.
 - conflicting class names.
 - Response classes change frequently. (As game changes..)

One exception: 'SpellRange' is handwritten becuase
 - Class differs if 'version' parameter changes.
 ```
range: { // normal json object,
    self: false,
    ranges: [0, 0, 0, 0] // int
}
```
```
range: [0, 0, 0, 0] // int-array only
```
```
range: "self" // string. WTF?
```


See implementation: [go-lol-generator/lolregio/datas.go](https://github.com/go-lol/go-lol/blob/master/go-lol-generator/lolregi/datas.go)

# Build
```sh
# If you modify 'SpellRange', you **must** run this first.
go install github.com/go-lol/go-lol
go generate github.com/go-lol/go-lol
```

# License
Apache2



[godoc]:(https://godoc.org/github.com/go-lol/go-lol)
